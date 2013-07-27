package sitejsoncache

import (
	"akamai/atlas/linker"
	"akamai/atlas/sitelistcache"
	"akamai/atlas/stat"
	"akamai/atlas/svgtext"
	"bytes"
	"encoding/json"
	"github.com/golang/glog"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

func L(s string, v ...interface{}) {
	if glog.V(1) {
		glog.Infof("sjc "+s, v...)
	}
}

type Dep struct {
	name string
	fi   os.FileInfo
}

type Ent struct {
	text string
	deps []Dep
}

func (self Ent) MarshalJSON() ([]byte, error) {
	return json.Marshal(self.text)
}

type SiteJsonCache struct {
	*sitelistcache.SiteListCache
	ModTime time.Time
	Entries map[string]Ent
	Json    []byte
}

func New(siteListCache *sitelistcache.SiteListCache) *SiteJsonCache {
	return &SiteJsonCache{
		SiteListCache: siteListCache,
		ModTime:       time.Time{},
		Json:          nil,
		Entries:       map[string]Ent{},
	}
}

func (self *SiteJsonCache) Make() (built bool, err error) {
	L("make starting")
	built = false
	err = nil

	fresh, err := self.allFresh()
	L("make allFresh returned fresh %t, err %v", fresh, err)
	if err != nil {
		L("make exiting; allFresh returned err %v", err)
		return
	}

	if !fresh {
		built = true

		L("make rebuilding")
		err = self.rebuild()
		if err != nil {
			L("make exiting; rebuild returned err %v", err)
			return
		}
	}
	L("make done")
	return
}

func (self *SiteJsonCache) isFresh(a, b os.FileInfo) bool {
	return stat.IsFresh(a, b)
}

func (self *SiteJsonCache) allFresh() (fresh bool, err error) {
	L("allFresh")
	fresh = false
	err = nil

	if built, err := self.SiteListCache.Make(); built || err != nil {
		return false, err
	}

	for _, ent := range self.Entries {
		for _, dep := range ent.deps {
			var fi os.FileInfo
			fi, err = os.Stat(dep.name)
			if err != nil {
				if os.IsNotExist(err) {
					err = nil
				}
				return
			}
			fresh = self.isFresh(fi, dep.fi)
			if !fresh {
				return
			}
		}
	}

	return
}

func (self *SiteJsonCache) updateModTime(depModTime time.Time) bool {
	if depModTime.After(self.ModTime) {
		self.ModTime = depModTime
		return true
	}
	return false
}

func (self *SiteJsonCache) rebuild() (err error) {
	L("rebuild starting")
	self.Entries = map[string]Ent{}

	for name, slEnt := range self.SiteListCache.Entries {
		chart := slEnt.Chart
		if chart == nil {
			L("rebuild skipping name %s", name)
			continue
		}

		key := chart.Slug()
		L("rebuild found key %s", key)

		err = chart.Read()
		if err != nil {
			L("rebuild warning after read: %s", err)
			continue
		}

		chartBytes := chart.Bytes()

		dep := Dep{
			name: chart.Src(),
			fi:   chart.FileInfo(),
		}

		ent := Ent{}
		ent.text = string(chartBytes)
		ent.deps = []Dep{dep}

		if glog.V(2) {
			glog.Infof("HandleSiteJsonGet(): found body: %q", ent.text)
		}

		self.updateModTime(dep.fi.ModTime())

		linkRenderer := linker.NewLinkRenderer()
		extFlags := 0
		extFlags |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
		extFlags |= blackfriday.EXTENSION_TABLES
		extFlags |= blackfriday.EXTENSION_FENCED_CODE
		extFlags |= blackfriday.EXTENSION_AUTOLINK
		extFlags |= blackfriday.EXTENSION_STRIKETHROUGH
		extFlags |= blackfriday.EXTENSION_SPACE_HEADERS
		blackfriday.Markdown([]byte(chart.Body()), linkRenderer, extFlags)
		L("rebuild found links: %s", linkRenderer.Links)

		for _, link := range linkRenderer.Links {
			// BUG(mistone): directory traversal
			sfx := strings.HasSuffix(link.Href, "svg")
			if sfx {
				svgPath := path.Clean(path.Join(chart.Dir(), link.Href))
				L("rebuild found svg: %s", svgPath)

				svgFile, err := os.Open(svgPath)
				if err != nil {
					L("rebuild warning: unable to open svg: %q, err %v", svgPath, err)
					continue
				}
				defer svgFile.Close()

				svgFI, err := svgFile.Stat()
				if err != nil {
					L("rebuild warning: unable to stat svg: %q, err %v", svgPath, err)
					continue
				}

				svgBody, err := ioutil.ReadAll(svgFile)
				if err != nil {
					L("rebuild warning: unable to read svg: %s, error: %s", svgPath, err)
					continue
				}

				cdata, err := svgtext.GetCData(svgBody)
				if err != nil {
					L("rebuild warning: unable to parse svg: %s, error: %s", svgPath, err)
					continue
				}
				L("rebuild found svg cdata items: %d", len(cdata))
				//L("rebuild found svg cdata: %s", cdata)
				L("rebuild done with svg: %s", svgPath)

				var buf bytes.Buffer
				for _, datum := range cdata {
					buf.WriteString("svg: ")
					buf.WriteString(datum)
					buf.WriteRune('\n')
				}

				ent.text = ent.text + "\n" + buf.String()
				ent.deps = append(ent.deps, Dep{
					name: svgPath,
					fi:   svgFI,
				})
				self.updateModTime(svgFI.ModTime())
			}
		}

		self.Entries[key] = ent
	}

	L("rebuild produced cache: %q", self.Entries)

	self.Json, err = json.Marshal(self.Entries)
	L("rebuild produced json: %q", string(self.Json))
	if err != nil {
		return
	}

	return
}

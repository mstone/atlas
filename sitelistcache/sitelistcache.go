package sitelistcache

import (
	"akamai/atlas/chart"
	"akamai/atlas/stat"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
)

var logSiteListCache = flag.Bool("log.cache.sitelist", false, "log sitelist code")

func L(s string, v ...interface{}) {
	if *logSiteListCache {
		log.Printf("sitelist "+s, v...)
	}
}

type SiteEnt struct {
	Chart *chart.Chart
	fi    os.FileInfo
}

type SiteListCache struct {
	Entries map[string]SiteEnt
	Root    string
}

func New(root string) *SiteListCache {
	return &SiteListCache{
		Entries: map[string]SiteEnt{},
		Root:    root,
	}
}

func (self *SiteListCache) Make() (built bool, err error) {
	L("Make() starting")
	var fi os.FileInfo

	fresh, err := self.allFresh()
	if err != nil {
		return
	}

	if !fresh {
		built = true

		fi, err = os.Stat(self.Root)
		if err != nil {
			return
		}

		err = self.rechart(self.Root, fi, SiteEnt{}, false)
		if err != nil {
			return
		}
		err = self.rebuild(self.Root)
		if err != nil {
			return
		}
	}

	return
}

func (self *SiteListCache) allFresh() (fresh bool, err error) {
	fresh = false
	err = nil

	if len(self.Entries) == 0 {
		return
	}

	for key, ent := range self.Entries {
		fi, err := os.Stat(key)
		if err != nil {
			return fresh, err
		}

		fresh = self.isFresh(fi, ent.fi)
		if !fresh {
			return fresh, err
		}
	}

	return
}

func (self *SiteListCache) isFresh(a, b os.FileInfo) bool {
	return stat.IsFresh(a, b)
}

func (self *SiteListCache) rebuild(name string) (err error) {
	fis, err := ioutil.ReadDir(name)
	if err != nil {
		L("rebuild ReadDir -> %v", err)
		return
	}

	for idx, fi := range fis {
		L("rebuild name %q idx %d fi %v", name, idx, fi)

		childName := path.Join(name, fi.Name())

		ent, ok := self.Entries[childName]
		L("rebuild child %q ent %q ok %t", childName, ent, ok)
		if ok && self.isFresh(fi, ent.fi) {
			L("rebuild recurse %q", childName)
			err = self.rebuild(childName)
			if err != nil {
				return
			}
		} else {
			if fi.IsDir() {
				L("rebuild rechart %q", childName)
				err = self.rechart(childName, fi, ent, ok)
				if err != nil {
					return
				}
				L("rebuild recurse %q", childName)
				err = self.rebuild(childName)
				if err != nil {
					return
				}
			}
		}
	}
	return
}

func (self *SiteListCache) rechart(name string, fi os.FileInfo, ent SiteEnt, ok bool) (err error) {
	L("rechart name %q fi %v ent %v ok %t", name, fi, ent, ok)
	newChart := ent.Chart
	if !ok {
		newChart, err = chart.Resolve(name, self.Root)
	} else {
		if newChart != nil {
			_, err = ent.Chart.Make()
		} else {
			newChart, err = chart.Resolve(name, self.Root)
		}
	}

	if err != nil && os.IsNotExist(err) {
		err = nil
	}

	newEnt := SiteEnt{
		Chart: newChart,
		fi:    fi,
	}
	L("rechart newent %q", newEnt)

	self.Entries[name] = newEnt

	return
}

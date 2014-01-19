package web

import (
	"akamai/atlas/atom"
	"akamai/atlas/cfg"

	"github.com/golang/glog"

	"bytes"
	"encoding/xml"
	"net/http"
	"net/url"
	"sort"
	"time"
)

type EntriesByDate []*atom.Entry

func (self EntriesByDate) Len() int {
	return len(self)
}

func (self EntriesByDate) Less(i, j int) bool {
	return self[i].Updated > self[j].Updated
}

func (self EntriesByDate) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self *App) GetAbsoluteBaseUrl() (*url.URL, error) {
	scheme, err := cfg.String("approot.scheme")
	if err != nil {
		glog.Errorf("Unable to get approot.scheme, err %q", err)
		return nil, err
	}

	host, err := cfg.String("approot.host")
	if err != nil {
		glog.Errorf("Unable to get approot.host, err: %q", err)
		return nil, err
	}

	return &url.URL{
		Scheme: scheme,
		Host:   host,
	}, nil
}

func HandleSiteAtomGet(self *App, w http.ResponseWriter, r *http.Request) {
	glog.Infof("HandleSiteAtomGet()")

	_, err := self.SiteListCache.Make()
	checkHTTP(err)

	title, _ := cfg.String("atom.title")
	id := cfg.MustString("atom.id")

	baseUrl, err := self.GetAbsoluteBaseUrl()
	checkHTTP(err)

	lastUpdated := time.Time{}

	feed := atom.Feed{
		XMLName: xml.Name{"http://www.w3.org/2005/Atom", "feed"},
		Title:   title,
		ID:      id,
	}

	for name, ent := range self.SiteListCache.Entries {
		if ent.Chart != nil {
			err = ent.Chart.Read()
			if err != nil {
				glog.Errorf("unable to read chart %q, err %q", name, err)
				continue
			}

			link, err := self.GetChartUrl(ent.Chart)
			if err != nil {
				glog.Errorf("unable to get chart url %q, err %q", name, err)
				continue
			}

			absLink := baseUrl.ResolveReference(&link)

			modTime := ent.Chart.FileInfo().ModTime()

			feed.Entry = append(feed.Entry, &atom.Entry{
				Title: ent.Chart.Meta().Title,
				ID:    absLink.String(),
				Link: []atom.Link{atom.Link{
					Href: absLink.String(),
				}},
				Published: atom.Time(modTime),
				Updated:   atom.Time(modTime),
			})

			if lastUpdated.Before(modTime) {
				lastUpdated = modTime
			}
		}
	}

	sort.Sort(EntriesByDate(feed.Entry))

	feed.Updated = atom.Time(lastUpdated)

	bits, err := xml.Marshal(&feed)
	checkHTTP(err)

	http.ServeContent(w, r, "atom.xml", lastUpdated, bytes.NewReader(bits))
}

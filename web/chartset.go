package web

import (
	"akamai/atlas/chart"

	"github.com/golang/glog"

	"fmt"
	"net/http"
	"net/url"
	"sort"
	"time"
)

type vChartLink struct {
	ModTime time.Time
	chart.ChartMeta
	Link url.URL
}

type vChartLinkList []*vChartLink

func (self vChartLinkList) Len() int {
	return len(self)
}

func (self vChartLinkList) Less(i, j int) bool {
	return self[i].ModTime.After(self[j].ModTime)
}

func (self vChartLinkList) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

type vChartSet struct {
	*vRoot
	Charts vChartLinkList
}

func HandleChartSetGet(self *App, w http.ResponseWriter, r *http.Request) {
	glog.Infof("HandleChartSetGet(): start")

	var charts vChartLinkList = nil

	_, err := self.SiteListCache.Make()
	checkHTTP(err)

	for name, ent := range self.SiteListCache.Entries {
		if ent.Chart != nil {
			err = ent.Chart.Read()
			if err != nil {
				glog.Infof("HandleChartSetGet(): warning: unable to read chart %q err %v", name, err)
				continue
			}

			link, err := self.GetChartUrl(ent.Chart)
			if err != nil {
				glog.Infof("HandleChartSetGet(): warning: unable to get chart url %q err %v", name, err)
				continue
			}

			modTime := ent.Chart.FileInfo().ModTime()

			charts = append(charts, &vChartLink{
				ModTime:   modTime,
				ChartMeta: ent.Chart.Meta(),
				Link:      link,
			})
		}
	}

	now := time.Now()
	date := fmt.Sprintf("%s %0.2d, %d", now.Month().String(), now.Day(), now.Year())

	sort.Sort(charts)

	view := &vChartSet{
		vRoot:  newVRoot(self, "chart_set", "List of Charts", "Michael Stone", date),
		Charts: charts,
	}
	glog.Infof("HandleChartSetGet(): view: %s", view)

	self.renderTemplate(w, "chart_set", view)
}

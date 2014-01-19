package web

import (
	"github.com/golang/glog"

	"bytes"
	"net/http"
)

func HandleSiteJsonGet(self *App, w http.ResponseWriter, r *http.Request) {
	glog.Infof("HandleSiteJsonGet(): start")

	_, err := self.SiteJsonCache.Make()
	checkHTTP(err)

	http.ServeContent(w, r, "site.json", self.SiteJsonCache.ModTime, bytes.NewReader(self.SiteJsonCache.Json))
}

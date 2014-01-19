// Package web implements the Atlas's HTTP handlers and codecs.

package web

import (
	"akamai/atlas/cfg"
	"akamai/atlas/sitejsoncache"
	"akamai/atlas/sitelistcache"
	"akamai/atlas/templatecache"

	"github.com/golang/glog"

	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
)

const MAX_CHART_SIZE = 1000000

type App struct {
	StaticPath        string
	StaticRoot        string
	ChartsRoot        string
	HtmlPath          string
	ChartsPath        string
	EtherpadApiUrl    *url.URL
	EtherpadApiSecret string
	*templatecache.TemplateCache
	*sitelistcache.SiteListCache
	*sitejsoncache.SiteJsonCache
}

var errTooShort = errors.New("URL path too short.")
var errWrongPrefix = errors.New("URL has wrong prefix.")

func (self *App) RemoveUrlPrefix(urlPath string, prefix string) (string, error) {
	up := path.Clean(urlPath)
	sp := path.Clean("/" + prefix)

	glog.Infof("RemoveUrlPrefix(%q, %q)", urlPath, prefix)
	glog.Infof("RemoveUrlPrefix(): up: %q, sp: %q", up, sp)

	fp := ""
	err := errTooShort

	if up == sp {
		err = nil
	}

	if len(up) > len(sp) {
		if strings.HasPrefix(up, sp) {
			fp = up[len(sp):]
			err = nil
		} else {
			err = errWrongPrefix
		}
	}
	return fp, err
}

func (self *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer recoverHTTP(w, r)

	glog.Infof("HandleRootApp: path: %v", r.URL.Path)

	isStatic := strings.HasPrefix(r.URL.Path, path.Clean("/"+self.StaticRoot))
	if isStatic {
		self.HandleStatic(w, r)
		return
	}

	isChart := strings.HasPrefix(r.URL.Path, path.Clean("/"+self.ChartsRoot))
	if isChart {
		self.HandleChart(w, r)
		return
	}

	glog.Infof("warning: can't route path: %v", r.URL.Path)
}

// Serve initializes some variables on self and then delegates to net/http to
// to receive incoming HTTP requests. Requests are handled by self.ServeHTTP()
func (self *App) Serve() {
	httpAddr := cfg.MustString("http.addr")

	self.StaticRoot = path.Clean("/" + self.StaticRoot)

	self.SiteListCache = sitelistcache.New(self.ChartsPath)
	self.TemplateCache = templatecache.New(self.HtmlPath)
	self.SiteJsonCache = sitejsoncache.New(self.SiteListCache)

	fmt.Printf("App: %v\n", self)

	http.Handle("/", self)
	glog.Fatal(http.ListenAndServe(httpAddr, nil))
}

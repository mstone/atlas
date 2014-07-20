// Copyright (c) 2013, 2014 Akamai Technologies, Inc.

package web

import (
	"github.com/golang/glog"
	"net/http"
	"path"
)

// BUG(mistone): HandleStatic() directory traversal?

func (self *App) HandleStatic(w http.ResponseWriter, r *http.Request) {
	fp, err := self.RemoveUrlPrefix(r.URL.Path, self.StaticRoot)
	checkHTTP(err)
	glog.Infof("HandleStatic: file path: %v", fp)
	http.ServeFile(w, r, path.Join(self.StaticPath, fp))
}

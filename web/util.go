// Copyright (c) 2013, 2014 Akamai Technologies, Inc.

package web

import (
	"github.com/golang/glog"

	"net/http"
	"runtime/debug"
)

func recoverHTTP(w http.ResponseWriter, r *http.Request) {
	if rec := recover(); rec != nil {
		switch err := rec.(type) {
		case error:
			glog.Infof("error: %v, req: %v", err, r)
			debug.PrintStack()
			http.Error(w, err.Error(), http.StatusInternalServerError)
		default:
			glog.Infof("unknown error: %v, req: %v", err, r)
			debug.PrintStack()
			http.Error(w, "unknown error", http.StatusInternalServerError)
		}
	}
}

func checkHTTP(err error) {
	if err != nil {
		panic(err)
	}
}

func (self *App) renderTemplate(w http.ResponseWriter, templateName string, view interface{}) {
	_, err := self.TemplateCache.Make(templateName)
	checkHTTP(err)

	tmplEnt := self.TemplateCache.Entries[templateName]

	tmpl := tmplEnt.Template

	err = tmpl.ExecuteTemplate(w, templateName, view)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

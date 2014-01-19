package web

import (
	"net/http"
	"path"
)

type vRoot struct {
	PageName   string
	Title      string
	Authors    string
	Date       string
	StaticUrl  string
	ChartsRoot string
}

func newVRoot(self *App, pageName string, title string, authors string, date string) *vRoot {
	return &vRoot{
		PageName:   pageName,
		Title:      title,
		Authors:    authors,
		Date:       date,
		StaticUrl:  self.StaticRoot,
		ChartsRoot: path.Clean(self.ChartsRoot + "/"),
	}
}

func HandleRootGet(self *App, w http.ResponseWriter, r *http.Request) {
	view := newVRoot(self, "root", "", "", "")
	self.renderTemplate(w, "root", view)
}

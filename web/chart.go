// Copyright (c) 2013, 2014 Akamai Technologies, Inc.

package web

import (
	"akamai/atlas/chart"

	"github.com/golang/glog"
	"github.com/russross/blackfriday"

	"html/template"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

type vChart struct {
	*vRoot
	FullPath  string
	Url       string
	EditorUrl url.URL
	Html      template.HTML
}

func (self *App) GetChartUrl(chart *chart.Chart) (url.URL, error) {
	slug := chart.Slug()
	url := url.URL{}
	if slug != "" {
		url.Path = path.Clean(path.Join("/", self.ChartsRoot, slug)) + "/"
	} else {
		url.Path = path.Clean(path.Join("/", self.ChartsRoot))
	}
	return url, nil
}

func (self *App) HandleChartPost(w http.ResponseWriter, r *http.Request) {
	fp, err := self.RemoveUrlPrefix(r.URL.Path, self.ChartsRoot)
	checkHTTP(err)
	glog.Infof("HandleChartPost(): fp: %v\n", fp)

	if strings.HasPrefix(fp, "resumes/") {
		self.HandleResumePost(w, r)
	}
}

func (self *App) HandleChartGet(w http.ResponseWriter, r *http.Request) {
	chartUrl := path.Clean(r.URL.Path)
	glog.Infof("HandleChartGet(): chartUrl: %v\n", chartUrl)

	fullPath := path.Join(self.ChartsPath, chartUrl)

	if chartUrl == "/atom.xml" {
		switch r.Method {
		default:
			panic("method")
		case "GET":
			HandleSiteAtomGet(self, w, r)
		}
		return
	}

	if chartUrl == "/site.json" {
		switch r.Method {
		default:
			panic("method")
		case "GET":
			HandleSiteJsonGet(self, w, r)
		}
		return
	}

	if chartUrl == "/pages" {
		switch r.Method {
		default:
			panic("method")
		case "GET":
			HandleChartSetGet(self, w, r)
		}
		return
	}

	// anyway, assuming it's a chart, find the index.txt
	fi, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			checkHTTP(err)
		}
	}

	if !fi.IsDir() {
		fp3 := fullPath
		// BUG(mistone): don't set Content-Type blindly; also need to check Accept header
		// BUG(mistone): do we really want to sniff mime-types here?
		http.ServeFile(w, r, fp3)
		return
	} else {
		txtFile := "index.txt"
		name := path.Join(fullPath, txtFile)
		if _, err := os.Stat(name); os.IsNotExist(err) {
			txtFile = "index.text"
			name = path.Join(fullPath, txtFile)
			_, err = os.Stat(name)
			if err != nil && os.IsNotExist(err) {
				w.WriteHeader(http.StatusNotFound)
				return
			} else {
				checkHTTP(err)
			}
		}

		chart := chart.NewChart(name, self.ChartsPath)

		err = chart.Read()
		checkHTTP(err)

		editorUrl, err := self.GetChartUrl(chart)
		checkHTTP(err)
		editorUrl.Path = path.Join(editorUrl.Path, txtFile, "editor")

		// attempt to parse header lines
		meta := chart.Meta()

		htmlFlags := 0
		//htmlFlags |= blackfriday.HTML_USE_XHTML
		htmlFlags |= blackfriday.HTML_TOC
		htmlFlags |= blackfriday.HTML_SKIP_HTML // disable script tags!

		htmlRenderer := blackfriday.HtmlRenderer(htmlFlags, "", "")

		extFlags := 0
		extFlags |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
		extFlags |= blackfriday.EXTENSION_TABLES
		extFlags |= blackfriday.EXTENSION_FENCED_CODE
		extFlags |= blackfriday.EXTENSION_AUTOLINK
		extFlags |= blackfriday.EXTENSION_STRIKETHROUGH
		extFlags |= blackfriday.EXTENSION_SPACE_HEADERS

		html := blackfriday.Markdown([]byte(chart.Body()), htmlRenderer, extFlags)
		view := &vChart{
			vRoot: newVRoot(self, "chart", meta.Title, meta.Authors, meta.Date),
			//Url:          chartUrl.String(),
			FullPath:  fullPath,
			Url:       chartUrl,
			Html:      template.HTML(html),
			EditorUrl: editorUrl,
		}

		self.renderTemplate(w, "chart", view)
	}
}

func (self *App) HandleChart(w http.ResponseWriter, r *http.Request) {
	fp, err := self.RemoveUrlPrefix(r.URL.Path, self.ChartsRoot)
	checkHTTP(err)
	glog.Infof("HandleChart: file path: %v", fp)

	base := path.Base(fp)
	ext := path.Ext(path.Dir(fp))

	glog.Infof("HandleChart: base: %s, ext: %s", base, ext)

	isSvgEditor := (base == "editor") && (ext == ".svg")
	isTxtEditor := (base == "editor") && ((ext == ".txt") || (ext == ".text"))

	if isSvgEditor {
		switch r.Method {
		default:
			panic("method")
		case "GET":
			self.HandleSvgEditorGet(w, r)
		case "POST":
			self.HandleSvgEditorPost(w, r)
		}
		return
	}

	if isTxtEditor {
		switch r.Method {
		default:
			panic("method")
		case "GET":
			self.HandleTxtEditorGet(w, r)
		case "POST":
			self.HandleTxtEditorPost(w, r)
		}
		return
	}

	switch r.Method {
	default:
		panic("method")
	case "GET":
		self.HandleChartGet(w, r)
	case "POST":
		self.HandleChartPost(w, r)
	}
	return
}

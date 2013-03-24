// Package web implements the record wizard's controllers
// and views.
//
// Presently, there are controllers for these resources:
//
//     Root
//       QuestionSet
//         Question
//       FormSet
//         Form
//       RecordSet
//         Record
//
// Controllers are App struct methods named Handle*{Get|Post}.
//
// Controllers largely work by
//
//   1. looking up an entity to be modified or represented in the requested
//      response
//
//   2. constructing a private "view struct" with the data to be displayed
//
//   3. handing the view struct and the http.ResponseWriter to an appropriate
//      template for rendering.
//
// All *Set resources load and save their contained entities through
// corresponding *Repo interfaces on the App struct.
package web

import (
	"akamai/atlas/forms/chart"
	"akamai/atlas/forms/linker"
	"akamai/atlas/forms/shake"
	"akamai/atlas/forms/svgtext"
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/russross/blackfriday"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
)

const MAX_CHART_SIZE = 1000000

type App struct {
	StaticPath string
	StaticRoot string
	ChartsRoot string
	HtmlPath   string
	ChartsPath string
	HttpAddr   string
	templates  *template.Template
}

func recoverHTTP(w http.ResponseWriter, r *http.Request) {
	if rec := recover(); rec != nil {
		switch err := rec.(type) {
		case error:
			log.Printf("error: %v, req: %v", err, r)
			debug.PrintStack()
			http.Error(w, err.Error(), http.StatusInternalServerError)
		default:
			log.Printf("unknown error: %v, req: %v", err, r)
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

type vChart struct {
	*vRoot
	FullPath string
	Url      string
	Html     template.HTML
}

func HandleChartGet(self *App, w http.ResponseWriter, r *http.Request) {
	chartUrl := path.Clean(r.URL.Path)
	log.Printf("HandleChartGet(): chartUrl: %v\n", chartUrl)

	fullPath := path.Join(self.ChartsPath, chartUrl)

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
		name := path.Join(fullPath, "index.txt")
		if _, err := os.Stat(name); os.IsNotExist(err) {
			name = path.Join(fullPath, "index.text")
			_, err = os.Stat(name)
			checkHTTP(err)
		}

		chart := chart.NewChart(name, self.ChartsPath)

		err = chart.Read()
		checkHTTP(err)

		// attempt to parse header lines
		meta := chart.Meta()

		htmlFlags := 0
		//htmlFlags |= blackfriday.HTML_USE_XHTML
		htmlFlags |= blackfriday.HTML_TOC

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
			FullPath: fullPath,
			Url:      chartUrl,
			Html:     template.HTML(html),
		}

		self.renderTemplate(w, "chart", view)
	}
}

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

func (self *App) GetChartUrl(chart *chart.Chart) (url.URL, error) {
	slug := chart.Slug()
	return url.URL{
		Path: path.Clean(path.Join("/", self.ChartsRoot, slug)) + "/",
	}, nil
}

type vChartLink struct {
	chart.ChartMeta
	Link url.URL
}

type vChartLinkList []*vChartLink

type vChartSet struct {
	*vRoot
	Charts vChartLinkList
}

func HandleChartSetGet(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleChartSetGet(): start")

	var charts vChartLinkList = nil

	filepath.Walk(self.ChartsPath, func(name string, fi os.FileInfo, err error) error {
		log.Printf("HandleChartSetGet(): visiting path %s", name)
		if err != nil {
			return err
		}

		chart := chart.NewChart(name, self.ChartsPath)

		if !chart.IsChart() {
			return nil
		}

		err = chart.Read()
		if err != nil {
			return nil
		}

		link, err := self.GetChartUrl(chart)

		if err == nil {
			charts = append(charts, &vChartLink{
				ChartMeta: chart.Meta(),
				Link:      link,
			})
		}
		return nil
	})

	now := time.Now()
	date := fmt.Sprintf("%s %0.2d, %d", now.Month().String(), now.Day(), now.Year())

	view := &vChartSet{
		vRoot:  newVRoot(self, "chart_set", "List of Charts", "Michael Stone", date),
		Charts: charts,
	}
	log.Printf("HandleChartSetGet(): view: %s", view)

	self.renderTemplate(w, "chart_set", view)
}

func HandleSiteJsonGet(self *App, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleSiteJsonGet(): start")

	view := map[string]string{}

	filepath.Walk(self.ChartsPath, func(name string, fi os.FileInfo, err error) error {
		//log.Printf("HandleSiteJsonGet(): visiting path %s", name)
		if err != nil {
			return err
		}

		chart := chart.NewChart(name, self.ChartsPath)

		isChart := chart.IsChart()
		if !isChart {
			return nil
		}

		key := chart.Slug()
		log.Printf("HandleSiteJsonGet(): found key %s", key)

		err = chart.Read()
		if err != nil {
			log.Printf("HandleSiteJsonGet(): warning after read: %s", err)
			return nil
		}

		chartBytes := chart.Bytes()
		//log.Printf("HandleSiteJsonGet(): found body: %s", body)

		view[key] = string(chartBytes)

		linkRenderer := linker.NewLinkRenderer()
		extFlags := 0
		extFlags |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
		extFlags |= blackfriday.EXTENSION_TABLES
		extFlags |= blackfriday.EXTENSION_FENCED_CODE
		extFlags |= blackfriday.EXTENSION_AUTOLINK
		extFlags |= blackfriday.EXTENSION_STRIKETHROUGH
		extFlags |= blackfriday.EXTENSION_SPACE_HEADERS
		blackfriday.Markdown([]byte(chart.Body()), linkRenderer, extFlags)
		log.Printf("HandleSiteJsonGet(): found links: %s", linkRenderer.Links)

		for _, link := range linkRenderer.Links {
			// BUG(mistone): directory traversal
			sfx := strings.HasSuffix(link.Href, "svg")
			if sfx {
				svgPath := path.Clean(path.Join(chart.Dir(), link.Href))
				log.Printf("HandleSiteJsonGet(): found svg: %s", svgPath)

				svgBody, err := ioutil.ReadFile(svgPath)
				if err != nil {
					log.Printf("HandleSiteJsonGet(): unable to read svg: %s, error: %s", svgPath, err)
					continue
				}
				cdata, err := svgtext.GetCData(svgBody)
				if err != nil {
					log.Printf("HandleSiteJsonGet(): unable to parse svg: %s, error: %s", svgPath, err)
					continue
				}
				log.Printf("HandleSiteJsonGet(): found svg cdata items: %s", len(cdata))
				log.Printf("HandleSiteJsonGet(): found svg cdata: %s", cdata)
				log.Printf("HandleSiteJsonGet(): done with svg: %s", svgPath)
				var buf bytes.Buffer
				for _, datum := range cdata {
					buf.WriteString("svg: ")
					buf.WriteString(datum)
					buf.WriteRune('\n')
				}
				view[key] = view[key] + "\n" + buf.String()
			}
		}

		return nil
	})

	//log.Printf("HandleSiteJsonGet(): view: %s", view)

	writer := bufio.NewWriter(w)
	defer writer.Flush()

	encoder := json.NewEncoder(writer)
	err := encoder.Encode(&view)
	checkHTTP(err)

	//log.Printf("SiteJsonGet(): encoded view: %v", view)
}

func (self *App) renderTemplate(w http.ResponseWriter, tmpl string, p interface{}) {
	err := self.templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// BUG(mistone): XSS Safety for SVG editor?
func (self *App) HandleSvgEditorPost(w http.ResponseWriter, r *http.Request) {
	fp, err := self.RemoveUrlPrefix(r.URL.Path, self.ChartsRoot)
	checkHTTP(err)

	svgName := path.Clean(path.Dir(fp))
	log.Printf("HandleSvgEditorPost(): got svg: %s", svgName)

	if strings.HasPrefix(svgName, "..") {
		panic("HandleSvgEditorPost(): directory traversal")
	}

	svgDir := path.Dir(svgName)
	log.Printf("HandleSvgEditorPost(): got svg dir: %s", svgDir)

	realSvgDir := path.Join(self.ChartsPath, svgDir)
	err = os.MkdirAll(realSvgDir, 0755)
	checkHTTP(err)

	//now := time.Now()
	//date := fmt.Sprintf("%s %0.2d, %d", now.Month().String(), now.Day(), now.Year())

	svgBodyB64 := r.FormValue("filepath")
	log.Printf("HandleSvgEditorPost(): got svg body b64: %s", svgBodyB64)

	reader := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(svgBodyB64))

	realSvgName := path.Join(self.ChartsPath, svgName)
	svgFile, err := os.Create(realSvgName)
	checkHTTP(err)
	defer svgFile.Close()

	written, err := io.Copy(svgFile, reader)
	checkHTTP(err)

	log.Printf("HandleSvgEditorPost(): wrote %d bytes of svg body", written)
	w.WriteHeader(http.StatusNoContent)
}

type vSvgEditor struct {
	*vRoot
	SvgEditorUrl     url.URL
	StaticSvgEditUrl url.URL
}

func (self *App) GetSvgEditorUrl() (url.URL, error) {
	url := url.URL{
		Path: path.Clean(path.Join(self.StaticRoot, "svg-edit-2.6", "svg-editor.html")),
	}
	return url, nil
}

func (self *App) GetStaticSvgEditUrl() (url.URL, error) {
	url := url.URL{
		Path: path.Clean(path.Join(self.StaticRoot, "svg-edit-2.6")),
	}
	return url, nil
}

func (self *App) HandleSvgEditorGet(w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleSvgEditorGet(): starting")

	fp, err := self.RemoveUrlPrefix(r.URL.Path, self.ChartsRoot)
	checkHTTP(err)

	svg := path.Dir(fp)
	log.Printf("HandleSvgEditorGet(): handling svg: %s", svg)

	now := time.Now()
	date := fmt.Sprintf("%s %0.2d, %d", now.Month().String(), now.Day(), now.Year())

	editorUrl, err := self.GetSvgEditorUrl()
	checkHTTP(err)

	staticSvgEditUrl, err := self.GetStaticSvgEditUrl()
	checkHTTP(err)

	view := &vSvgEditor{
		vRoot:            newVRoot(self, "svg_editor", "SVG Editor", "Michael Stone", date),
		SvgEditorUrl:     editorUrl,
		StaticSvgEditUrl: staticSvgEditUrl,
	}
	log.Printf("HandleSvgEditorGet(): view: %s", view)

	self.renderTemplate(w, "svg_editor", view)
}

var errTooShort = errors.New("URL path too short.")
var errWrongPrefix = errors.New("URL has wrong prefix.")

func (self *App) RemoveUrlPrefix(urlPath string, prefix string) (string, error) {
	up := path.Clean(urlPath)
	sp := path.Clean("/" + prefix)

	log.Printf("RemoveUrlPrefix(%q, %q)", urlPath, prefix)
	log.Printf("RemoveUrlPrefix(): up: %q, sp: %q", up, sp)

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

func (self *App) HandleChart(w http.ResponseWriter, r *http.Request) {
	fp, err := self.RemoveUrlPrefix(r.URL.Path, self.ChartsRoot)
	checkHTTP(err)
	log.Printf("HandleChart: file path: %v", fp)

	base := path.Base(fp)
	ext := path.Ext(path.Dir(fp))

	log.Printf("HandleChart: base: %s, ext: %s", base, ext)

	isSvgEditor := (base == "editor") && (ext == ".svg")

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
	} else {
		HandleChartGet(self, w, r)
	}
}

type WebQuestion struct {
	http.ResponseWriter
	*http.Request
}

// BUG(mistone): WebQuestion's Key() method is really scary!
func (self WebQuestion) Key() (shake.Key, error) {
	return shake.Key(self.Method + " " + path.Clean(self.URL.Path)), nil
}

func (self *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer recoverHTTP(w, r)
	log.Printf("HandleRootApp: path: %v", r.URL.Path)

	question := WebQuestion{w, r}

	rules := shake.NewRuleSet()
	rules.Rules = []shake.Rule{
		&StaticContentRule{self},
		&ChartsContentRule{self},
	}

	_, err := rules.Make(question)
	switch err.(type) {
	default:
		panic(err)
	case nil:
		return
	case *shake.NoMatchingRuleError:
		break
	}

	log.Printf("warning: can't route path: %v", r.URL.Path)
}

// Serve initializes some variables on self and then delegates to net/http to
// to receive incoming HTTP requests. Requests are handled by self.ServeHTTP()
func (self *App) Serve() {
	self.templates = template.Must(
		template.ParseGlob(
			path.Join(self.HtmlPath, "*.html")))

	self.StaticRoot = path.Clean("/" + self.StaticRoot)

	fmt.Printf("App: %v\n", self)

	http.Handle("/", self)
	log.Fatal(http.ListenAndServe(self.HttpAddr, nil))
}

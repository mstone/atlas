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
	"akamai/atlas/chart"
	"akamai/atlas/resumes"
	"akamai/atlas/sitejsoncache"
	"akamai/atlas/sitelistcache"
	"akamai/atlas/templatecache"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
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
	"runtime/debug"
	"strings"
	"time"
)

const MAX_CHART_SIZE = 1000000

type App struct {
	StaticPath        string
	StaticRoot        string
	ChartsRoot        string
	HtmlPath          string
	ChartsPath        string
	HttpAddr          string
	EtherpadApiUrl    *url.URL
	EtherpadApiSecret string
	*templatecache.TemplateCache
	*sitelistcache.SiteListCache
	*sitejsoncache.SiteJsonCache
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
	FullPath  string
	Url       string
	EditorUrl url.URL
	Html      template.HTML
}

func (self *App) HandleResumePost(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	fp, err := self.RemoveUrlPrefix(r.URL.Path, self.ChartsRoot)
	checkHTTP(err)
	log.Printf("HandleResumePost(): fp: %v\n", fp)

	ext := path.Ext(fp)

	switch ext {
	default:
		log.Fatalf("HandleResumePost(): bad ext: %q", ext)
	case ".doc":
		break
	case ".docx":
		break
	case ".pdf":
		break
	}

	// BUG(mistone): other name validation?
	dstPath := path.Join(self.ChartsPath, fp, "upload"+ext)
	dstDir := path.Dir(dstPath)

	err = os.MkdirAll(dstDir, 0755)
	checkHTTP(err)

	dstFile, err := os.Create(dstPath)
	checkHTTP(err)
	defer dstFile.Close()

	_, err = io.Copy(dstFile, r.Body)
	checkHTTP(err)

	displayName := resumes.SimplifyName(path.Base(fp[:len(fp)-len(ext)]))

	log.Printf("HandleResumePost(): attempting to convert: %q -> %q", dstPath, dstDir)
	err = resumes.Convert(dstPath, dstDir, displayName)
	checkHTTP(err)

	chartName := path.Join(dstDir, "index.txt")
	chart := chart.NewChart(chartName, self.ChartsPath)

	if !chart.IsChart() {
		log.Fatalf("HandleResumePost(): missing chart: %q", chartName)
	}

	err = chart.Read()
	if err != nil {
		log.Fatalf("HandleResumePost(): bad chart: %q", chartName)
	}

	link, err := self.GetChartUrl(chart)

	if err == nil {
		http.Redirect(w, r, link.String(), http.StatusSeeOther)
	}
}

func (self *App) HandleChartPost(w http.ResponseWriter, r *http.Request) {
	fp, err := self.RemoveUrlPrefix(r.URL.Path, self.ChartsRoot)
	checkHTTP(err)
	log.Printf("HandleChartPost(): fp: %v\n", fp)

	if strings.HasPrefix(fp, "resumes/") {
		self.HandleResumePost(w, r)
	}
}

func (self *App) HandleChartGet(w http.ResponseWriter, r *http.Request) {
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
	url := url.URL{}
	if slug != "" {
		url.Path = path.Clean(path.Join("/", self.ChartsRoot, slug)) + "/"
	} else {
		url.Path = path.Clean(path.Join("/", self.ChartsRoot))
	}
	return url, nil
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

	_, err := self.SiteListCache.Make()
	checkHTTP(err)

	for name, ent := range self.SiteListCache.Entries {
		if ent.Chart != nil {
			err = ent.Chart.Read()
			if err != nil {
				log.Printf("HandleChartSetGet(): warning: unable to read chart %q err %v", name, err)
				continue
			}

			link, err := self.GetChartUrl(ent.Chart)
			if err != nil {
				log.Printf("HandleChartSetGet(): warning: unable to get chart url %q err %v", name, err)
				continue
			}

			charts = append(charts, &vChartLink{
				ChartMeta: ent.Chart.Meta(),
				Link:      link,
			})
		}
	}

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

	_, err := self.SiteJsonCache.Make()
	checkHTTP(err)

	http.ServeContent(w, r, "site.json", self.SiteJsonCache.ModTime, bytes.NewReader(self.SiteJsonCache.Json))
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

func (self *App) SvgEditFile(svgName string) (*os.File, error) {
	if strings.HasPrefix(svgName, "..") {
		panic(fmt.Sprintf("SvgEditFile(): directory traversal: %s", svgName))
	}

	svgDir := path.Dir(svgName)
	log.Printf("SvgEditFile(): got svg dir: %s", svgDir)

	realSvgDir := path.Join(self.ChartsPath, svgDir)
	err := os.MkdirAll(realSvgDir, 0755)
	checkHTTP(err)

	realSvgName := path.Join(self.ChartsPath, svgName)
	return os.Create(realSvgName)
}

// BUG(mistone): XSS Safety for SVG editor?
func (self *App) HandleSvgEditorPost(w http.ResponseWriter, r *http.Request) {
	fp, err := self.RemoveUrlPrefix(r.URL.Path, self.ChartsRoot)
	checkHTTP(err)

	svgName := path.Clean(path.Dir(fp))
	log.Printf("HandleSvgEditorPost(): got svg: %s", svgName)

	svgBodyB64 := r.FormValue("filepath")
	log.Printf("HandleSvgEditorPost(): got svg body b64: %s", svgBodyB64)

	reader := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(svgBodyB64))

	svgFile, err := self.SvgEditFile(svgName)
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

func (self *App) InitializeTxt(txtName string) error {
	txtFile, err := self.TxtEditFile(txtName)
	checkHTTP(err)
	defer txtFile.Close()

	_, err = txtFile.WriteString(`% Title
% Authors
% Date

[(edit this chart)](./index.txt/editor)

# Overview

# System Diagram     [ ](data:tkt,owner=&next_action=)

[(edit this diagram)](./system-diagram.svg/editor)

![System Diagram](./system-diagram.svg)

# Security Considerations

## Accidents         [ ](data:tkt,owner=&next_action=)

## Hazards           [ ](data:tkt,owner=&next_action=)

## Powers            [ ](data:tkt,owner=&next_action=)

## Controls          [ ](data:tkt,owner=&next_action=)

`)
	return err
}

func (self *App) GetTxtEditorUrl() (url.URL, error) {
	return url.URL{}, nil
}

func (self *App) GetStaticTxtEditUrl() (url.URL, error) {
	return url.URL{}, nil
}

func (self *App) TxtEditFile(txtName string) (*os.File, error) {
	if strings.HasPrefix(txtName, "..") {
		panic(fmt.Sprintf("TxtEditFile(): directory traversal: %s", txtName))
	}

	txtDir := path.Dir(txtName)
	log.Printf("TxtEditFile(): got txt dir: %s", txtDir)

	realTxtDir := path.Join(self.ChartsPath, txtDir)
	err := os.MkdirAll(realTxtDir, 0755)
	checkHTTP(err)

	realTxtName := path.Join(self.ChartsPath, txtName)
	return os.Create(realTxtName)
}

func (self *App) TxtOpenFile(txtName string) (*os.File, error) {
	if strings.HasPrefix(txtName, "..") {
		panic(fmt.Sprintf("TxtEditFile(): directory traversal: %s", txtName))
	}

	realTxtName := path.Join(self.ChartsPath, txtName)
	return os.Open(realTxtName)
}

type epResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

func (self *App) HandleTxtEditorPostSave(w http.ResponseWriter, r *http.Request, txtName string, padName string) {
	// get pad latest rev
	epUrl := *self.EtherpadApiUrl
	epUrl.Path = path.Join(epUrl.Path, "1.2.7", "getRevisionsCount")
	queryValues := url.Values{}
	queryValues.Set("apikey", self.EtherpadApiSecret)
	queryValues.Set("padID", padName)
	epUrl.RawQuery = queryValues.Encode()
	resp, err := http.Get(epUrl.String())
	checkHTTP(err)
	defer resp.Body.Close()
	log.Printf("HandleTxtEditorPost(): getRevisionsCount resp: %q", resp)
	decoder := json.NewDecoder(resp.Body)
	epResp := epResponse{}
	err = decoder.Decode(&epResp)
	checkHTTP(err)
	log.Printf("HandleTxtEditorPost(): getRevisionsCount epResp: %q", epResp)
	// XXX: check resp....
	// {"code":0,"message":"ok","data":null}
	// {"code":1,"message":"padID does already exist","data":null}

	// get pad latest rev
	epUrl = *self.EtherpadApiUrl
	epUrl.Path = path.Join(epUrl.Path, "1.2.7", "getRevisionsCount")
	queryValues = url.Values{}
	queryValues.Set("apikey", self.EtherpadApiSecret)
	queryValues.Set("padID", padName)
	epUrl.RawQuery = queryValues.Encode()
	resp, err = http.Get(epUrl.String())
	checkHTTP(err)
	defer resp.Body.Close()
	log.Printf("HandleTxtEditorPost(): getRevisionsCount resp: %q", resp)
	decoder = json.NewDecoder(resp.Body)
	epResp = epResponse{}
	err = decoder.Decode(&epResp)
	checkHTTP(err)
	log.Printf("HandleTxtEditorPost(): getRevisionsCount epResp: %q", epResp)
	if epResp.Code != 0 {
		panic("HandleTxtEditorPost(): epResp code != 0")
	}
	if epResp.Data == nil {
		panic("HandleTxtEditorPost(): no data")
	}
	revsIface := epResp.Data["revisions"]
	if revsIface == nil {
		panic("HandleTxtEditorPost(): no revisions field value")
	}
	revsFloat, ok := revsIface.(float64)
	if !ok {
		panic("HandleTxtEditorPost(): revisions field not a number")
	}
	rev := fmt.Sprintf("%d", int(revsFloat))

	// given rev, get pad text

	epUrl = *self.EtherpadApiUrl
	epUrl.Path = path.Join(epUrl.Path, "1.2.7", "getText")
	queryValues = url.Values{}
	queryValues.Set("apikey", self.EtherpadApiSecret)
	queryValues.Set("padID", padName)
	queryValues.Set("rev", rev)
	epUrl.RawQuery = queryValues.Encode()
	resp, err = http.Get(epUrl.String())
	checkHTTP(err)
	defer resp.Body.Close()
	log.Printf("HandleTxtEditorPost(): getText resp: %q", resp)
	decoder = json.NewDecoder(resp.Body)
	epResp = epResponse{}
	err = decoder.Decode(&epResp)
	checkHTTP(err)
	log.Printf("HandleTxtEditorPost(): getText epResp: %t", epResp)
	log.Printf("HandleTxtEditorPost(): getText epResp: %q", epResp)
	if epResp.Code != 0 {
		panic("HandleTxtEditorPost(): epResp code != 0")
	}
	if epResp.Data == nil {
		panic("HandleTxtEditorPost(): no data")
	}
	textIface := epResp.Data["text"]
	if textIface == nil {
		panic("HandleTxtEditorPost(): no text field value")
	}
	text, ok := textIface.(string)
	if !ok {
		panic("HandleTxtEditorPost(): text field not a string")
	}

	txtFile, err := self.TxtEditFile(txtName)
	checkHTTP(err)
	defer txtFile.Close()

	reader := bytes.NewBufferString(text)

	written, err := io.Copy(txtFile, reader)
	checkHTTP(err)

	log.Printf("HandleTxtEditorPost(): wrote %d bytes of txt body", written)
	w.WriteHeader(http.StatusNoContent)
}

func (self *App) HandleTxtEditorPostReload(w http.ResponseWriter, r *http.Request, txtName string, padName string) {
	self.ReloadPad(txtName, padName)
	http.Redirect(w, r, "", http.StatusSeeOther)
}

// BUG(mistone): XSS Safety for TXT editor?
func (self *App) HandleTxtEditorPost(w http.ResponseWriter, r *http.Request) {
	fp, err := self.RemoveUrlPrefix(r.URL.Path, self.ChartsRoot)
	checkHTTP(err)

	txtName := path.Clean(path.Dir(fp))
	log.Printf("HandleTxtEditorPost(): got txt: %s", txtName)

	// get pad id
	hash := sha1.New()
	hash.Write([]byte(txtName))
	padName := hex.EncodeToString(hash.Sum(nil))
	log.Printf("HandleTxtEditorPost(): calculated pad name: %s", padName)

	action := r.FormValue("action")
	log.Printf("HandleTxtEditorPost(): processing action: %s", action)

	switch action {
	default:
		panic(fmt.Sprintf("HandleTxtEditorPost(): unknown action: %s", action))
	case "save":
		self.HandleTxtEditorPostSave(w, r, txtName, padName)
	case "reload":
		self.HandleTxtEditorPostReload(w, r, txtName, padName)
	}
	return
}

type vTxtEditor struct {
	*vRoot
	TxtEditorUrl url.URL
	ChartUrl     url.URL
}

func (self *App) ReloadPad(txtName, padName string) error {
	txtFile, err := self.TxtOpenFile(txtName)
	checkHTTP(err)
	defer txtFile.Close()

	txtBodyRaw, err := ioutil.ReadAll(txtFile)
	checkHTTP(err)
	txtBody := string(txtBodyRaw)

	epUrl := *self.EtherpadApiUrl
	epUrl.Path = path.Join(epUrl.Path, "1.2.7", "setText")
	queryValues := url.Values{}
	queryValues.Set("apikey", self.EtherpadApiSecret)
	queryValues.Set("padID", padName)
	queryValues.Set("text", txtBody)
	epUrl.RawQuery = queryValues.Encode()
	resp, err := http.Get(epUrl.String())
	checkHTTP(err)
	defer resp.Body.Close()
	log.Printf("HandleTxtEditorGet(): setText resp: %q", resp)
	// XXX: check resp....
	return nil
}

func (self *App) HandleTxtEditorGet(w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleTxtEditorGet(): starting")

	fp, err := self.RemoveUrlPrefix(r.URL.Path, self.ChartsRoot)
	checkHTTP(err)

	txtName := path.Clean(path.Dir(fp))
	log.Printf("HandleTxtEditorGet(): handling txtName: %s", txtName)

	realTxtName := path.Join(self.ChartsPath, txtName)
	_, err = os.Stat(realTxtName)
	if err != nil && os.IsNotExist(err) {
		err = self.InitializeTxt(txtName)
	}

	hash := sha1.New()
	hash.Write([]byte(txtName))
	padName := hex.EncodeToString(hash.Sum(nil))
	log.Printf("HandleTxtEditorGet(): calculated pad name: %s", padName)

	// create the pad
	epUrl := *self.EtherpadApiUrl
	epUrl.Path = path.Join(epUrl.Path, "1.2.7", "createPad")
	queryValues := url.Values{}
	queryValues.Set("apikey", self.EtherpadApiSecret)
	queryValues.Set("padID", padName)
	epUrl.RawQuery = queryValues.Encode()
	resp, err := http.Get(epUrl.String())
	checkHTTP(err)
	defer resp.Body.Close()
	log.Printf("HandleTxtEditorPost(): createPad resp: %q", resp)
	decoder := json.NewDecoder(resp.Body)
	epResp := epResponse{}
	err = decoder.Decode(&epResp)
	checkHTTP(err)
	log.Printf("HandleTxtEditorPost(): createPad epResp: %t", epResp)
	if epResp.Code == 0 {
		err := self.ReloadPad(txtName, padName)
		checkHTTP(err)
	}
	// {"code":1,"message":"padID does already exist","data":null}
	if epResp.Code == 1 && epResp.Message != "padID does already exist" {
		panic("Unknown createPad response")
	}
	if epResp.Code != 0 && epResp.Code != 1 {
		panic("Unknown createPad response")
	}

	now := time.Now()
	date := fmt.Sprintf("%s %0.2d, %d", now.Month().String(), now.Day(), now.Year())

	editorUrl := *self.EtherpadApiUrl
	editorUrl.Path = path.Join(path.Dir(editorUrl.Path), "p", padName)
	editorValues := url.Values{}
	editorValues.Set("showControls", "true")
	editorValues.Set("showChat", "true")
	editorValues.Set("showLineNumbers", "true")
	editorValues.Set("useMonospaceFont", "true")
	editorUrl.RawQuery = editorValues.Encode()

	chart := chart.NewChart(path.Join(self.ChartsPath, txtName), self.ChartsPath)
	if !chart.IsChart() {
		panic("Not a chart!")
	}
	chartUrl, err := self.GetChartUrl(chart)
	checkHTTP(err)
	log.Printf("HandleTxtEditorGet(): found chart url: %q", chartUrl)

	slug := chart.Slug()
	title := "Chart Editor: "
	if slug == "" {
		title = title + "Root Chart"
	} else {
		title = title + slug[0:len(slug)-1]
	}

	view := &vTxtEditor{
		vRoot:        newVRoot(self, "txt_editor", title, "(none)", date),
		TxtEditorUrl: editorUrl,
		ChartUrl:     chartUrl,
	}
	log.Printf("HandleTxtEditorGet(): view: %s", view)

	self.renderTemplate(w, "txt_editor", view)
}

func (self *App) InitializeSvg(svgName string) error {
	svgFile, err := self.SvgEditFile(svgName)
	checkHTTP(err)
	defer svgFile.Close()

	_, err = svgFile.WriteString(`<?xml version="1.0"?>
<svg width="800" height="600" xmlns="http://www.w3.org/2000/svg">
 <metadata id="metadata7">image/svg+xml</metadata>
 <g>
  <title>Layer 1</title>
 </g>
</svg>
`)
	return err
}

func (self *App) HandleSvgEditorGet(w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleSvgEditorGet(): starting")

	fp, err := self.RemoveUrlPrefix(r.URL.Path, self.ChartsRoot)
	checkHTTP(err)

	svgName := path.Clean(path.Dir(fp))
	log.Printf("HandleSvgEditorGet(): handling svgName: %s", svgName)

	realSvgName := path.Join(self.ChartsPath, svgName)
	_, err = os.Stat(realSvgName)
	if err != nil && os.IsNotExist(err) {
		err = self.InitializeSvg(svgName)
	}

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

type WebQuestion struct {
	http.ResponseWriter
	*http.Request
}

func (self *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer recoverHTTP(w, r)
	log.Printf("HandleRootApp: path: %v", r.URL.Path)

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

	log.Printf("warning: can't route path: %v", r.URL.Path)
}

// Serve initializes some variables on self and then delegates to net/http to
// to receive incoming HTTP requests. Requests are handled by self.ServeHTTP()
func (self *App) Serve() {
	self.StaticRoot = path.Clean("/" + self.StaticRoot)

	self.SiteListCache = sitelistcache.New(self.ChartsPath)
	self.TemplateCache = templatecache.New(self.HtmlPath)
	self.SiteJsonCache = sitejsoncache.New(self.SiteListCache)

	fmt.Printf("App: %v\n", self)

	http.Handle("/", self)
	log.Fatal(http.ListenAndServe(self.HttpAddr, nil))
}

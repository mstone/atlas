// Copyright (c) 2013, 2014 Akamai Technologies, Inc.

package web

import (
	"akamai/atlas/chart"

	"github.com/golang/glog"

	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

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
	glog.Infof("TxtEditFile(): got txt dir: %s", txtDir)

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
	glog.Infof("HandleTxtEditorPost(): getRevisionsCount resp: %q", resp)
	decoder := json.NewDecoder(resp.Body)
	epResp := epResponse{}
	err = decoder.Decode(&epResp)
	checkHTTP(err)
	glog.Infof("HandleTxtEditorPost(): getRevisionsCount epResp: %q", epResp)
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
	glog.Infof("HandleTxtEditorPost(): getRevisionsCount resp: %q", resp)
	decoder = json.NewDecoder(resp.Body)
	epResp = epResponse{}
	err = decoder.Decode(&epResp)
	checkHTTP(err)
	glog.Infof("HandleTxtEditorPost(): getRevisionsCount epResp: %q", epResp)
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
	glog.Infof("HandleTxtEditorPost(): getText resp: %q", resp)
	decoder = json.NewDecoder(resp.Body)
	epResp = epResponse{}
	err = decoder.Decode(&epResp)
	checkHTTP(err)
	glog.Infof("HandleTxtEditorPost(): getText epResp: %t", epResp)
	glog.Infof("HandleTxtEditorPost(): getText epResp: %q", epResp)
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

	glog.Infof("HandleTxtEditorPost(): wrote %d bytes of txt body", written)
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
	glog.Infof("HandleTxtEditorPost(): got txt: %s", txtName)

	// get pad id
	hash := sha1.New()
	hash.Write([]byte(txtName))
	padName := hex.EncodeToString(hash.Sum(nil))
	glog.Infof("HandleTxtEditorPost(): calculated pad name: %s", padName)

	action := r.FormValue("action")
	glog.Infof("HandleTxtEditorPost(): processing action: %s", action)

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
	glog.Infof("HandleTxtEditorGet(): setText resp: %q", resp)
	// XXX: check resp....
	return nil
}

func (self *App) HandleTxtEditorGet(w http.ResponseWriter, r *http.Request) {
	glog.Infof("HandleTxtEditorGet(): starting")

	fp, err := self.RemoveUrlPrefix(r.URL.Path, self.ChartsRoot)
	checkHTTP(err)

	txtName := path.Clean(path.Dir(fp))
	glog.Infof("HandleTxtEditorGet(): handling txtName: %s", txtName)

	realTxtName := path.Join(self.ChartsPath, txtName)
	_, err = os.Stat(realTxtName)
	if err != nil && os.IsNotExist(err) {
		err = self.InitializeTxt(txtName)
	}

	hash := sha1.New()
	hash.Write([]byte(txtName))
	padName := hex.EncodeToString(hash.Sum(nil))
	glog.Infof("HandleTxtEditorGet(): calculated pad name: %s", padName)

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
	glog.Infof("HandleTxtEditorPost(): createPad resp: %q", resp)
	decoder := json.NewDecoder(resp.Body)
	epResp := epResponse{}
	err = decoder.Decode(&epResp)
	checkHTTP(err)
	glog.Infof("HandleTxtEditorPost(): createPad epResp: %t", epResp)
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
	glog.Infof("HandleTxtEditorGet(): found chart url: %q", chartUrl)

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
	glog.Infof("HandleTxtEditorGet(): view: %s", view)

	self.renderTemplate(w, "txt_editor", view)
}

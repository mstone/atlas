// Copyright (c) 2013, 2014 Akamai Technologies, Inc.

package web

import (
	"github.com/golang/glog"

	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

func (self *App) SvgEditFile(svgName string) (*os.File, error) {
	if strings.HasPrefix(svgName, "..") {
		panic(fmt.Sprintf("SvgEditFile(): directory traversal: %s", svgName))
	}

	svgDir := path.Dir(svgName)
	glog.Infof("SvgEditFile(): got svg dir: %s", svgDir)

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
	glog.Infof("HandleSvgEditorPost(): got svg: %s", svgName)

	svgBodyB64 := r.FormValue("filepath")
	glog.Infof("HandleSvgEditorPost(): got svg body b64: %s", svgBodyB64)

	reader := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(svgBodyB64))

	svgFile, err := self.SvgEditFile(svgName)
	checkHTTP(err)
	defer svgFile.Close()

	written, err := io.Copy(svgFile, reader)
	checkHTTP(err)

	glog.Infof("HandleSvgEditorPost(): wrote %d bytes of svg body", written)
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
	glog.Infof("HandleSvgEditorGet(): starting")

	fp, err := self.RemoveUrlPrefix(r.URL.Path, self.ChartsRoot)
	checkHTTP(err)

	svgName := path.Clean(path.Dir(fp))
	glog.Infof("HandleSvgEditorGet(): handling svgName: %s", svgName)

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
	glog.Infof("HandleSvgEditorGet(): view: %s", view)

	self.renderTemplate(w, "svg_editor", view)
}

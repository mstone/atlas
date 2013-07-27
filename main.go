// Package main calls the atlas web package.
//
// The web package defines the web UI and controllers.
//
// The main package ties the other packages together and
// configures them via command-line flags.
package main

import (
	"akamai/atlas/web"
	"flag"
	"io/ioutil"
	"net/url"
	"strings"
)

// htmlPath tells the web controller where to look for
// HTML templates to render.
var htmlPath = flag.String("html", "html/", "path to atlas-forms html templates")

// staticPath tells the web controller where to look for non-template static
// assets
var staticPath = flag.String("static", "static/", "path to atlas-forms static assets")

// staticRoot tells the web controller what URL prefix to discard
var staticRoot = flag.String("staticroot", "static", "static app url prefix")

// chartsRoot tells the web controller what URL prefix to discard
var chartsRoot = flag.String("chartsroot", "charts", "charts app url prefix")

// chartsPath tells us where to look for charts to render
var chartsPath = flag.String("charts", "charts/", "path to atlas charts")

// etherpadApiUrlStr tells us where to access etherpads for chart editing
var etherpadApiUrlStr = flag.String("etherpadApiUrl", "http://localhost:9001/api", "etherpad API url")

// etherpadApiSecretPath tells us where to look for the etherpad API key
var etherpadApiSecretPath = flag.String("etherpadApiSecretPath", "eplite/APIKEY.txt", "path to the etherpad API secret")

func main() {
	flag.Parse()

	etherpadApiSecretRaw, err := ioutil.ReadFile(*etherpadApiSecretPath)
	if err != nil {
		panic(err)
	}

	etherpadApiSecret := strings.Trim(string(etherpadApiSecretRaw), " \t\n")

	etherpadApiUrl, err := url.Parse(*etherpadApiUrlStr)
	if err != nil {
		panic(err)
	}

	web := &web.App{
		HtmlPath:          *htmlPath,
		StaticPath:        *staticPath,
		StaticRoot:        *staticRoot,
		ChartsRoot:        *chartsRoot,
		ChartsPath:        *chartsPath,
		EtherpadApiUrl:    etherpadApiUrl,
		EtherpadApiSecret: etherpadApiSecret,
	}

	web.Serve()
}

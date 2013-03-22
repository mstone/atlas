// Package main calls the atlas web package.
//
// The web package defines the web UI and controllers.
//
// The main package ties the other packages together and
// configures them via command-line flags.
package main

import (
	"akamai/atlas/forms/web"
	"flag"
)

// httpAddr tells the web controller on what address to
// listen for requests.
var httpAddr = flag.String("http", "127.0.0.1:3001", "addr:port")

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

func main() {
	flag.Parse()

	web := &web.App{
		HttpAddr:   *httpAddr,
		HtmlPath:   *htmlPath,
		StaticPath: *staticPath,
		StaticRoot: *staticRoot,
		ChartsRoot: *chartsRoot,
		ChartsPath: *chartsPath,
	}

	web.Serve()
}

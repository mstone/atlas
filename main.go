// Package main implements the atlas record wizard.
//
// The code is organized into four packages:
//
//    entity
//      ^ ^
//      |  \
//      |  persist
//      |  ^
//      | /
//     web
//      ^
//      |
//      |
//     main
//
//
// The entity package contains domain syntax: Questions,
// Records, (Question) Forms, Responses, etc. and their
// persistence interfaces.
//
// The persist package implements the *Repo interfaces
// defined in the entity model.
//
// The web package defines the web UI and calls the entity
// and persist code.
//
// The main package ties the other packages together and
// configures them via command-line flags.
//
// For more information on this architecture, see
//
//     http://manuel.kiessling.net/2012/09/28/applying-the-clean-architecture-to-go-applications/
package main

import (
	"akamai/atlas/forms/entity"
	"akamai/atlas/forms/persist"
	"akamai/atlas/forms/web"
	"flag"
)

// httpAddr tells the web controller on what address to
// listen for requests.
var httpAddr = flag.String("http", "127.0.0.1:3001", "addr:port")

// htmlPath tells the web controller where to look for
// HTML templates to render.
var htmlPath = flag.String("html", "html/", "path to atlas-forms html templates")

// formsPath tells us where to store our persisted
// entities.
var formsPath = flag.String("forms", "forms/", "path to forms database")

// formsRoot tells the web controller what URL prefix to discard
var formsRoot = flag.String("formsroot", "", "forms app url prefix")

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

	persist := persist.NewPersistJSON(*formsPath)

	web := &web.App{
		HttpAddr:     *httpAddr,
		QuestionRepo: entity.QuestionRepo(persist),
		FormRepo:     entity.FormRepo(persist),
		RecordRepo:   entity.RecordRepo(persist),
		FormsRoot:    *formsRoot,
		HtmlPath:     *htmlPath,
		StaticPath:   *staticPath,
		StaticRoot:   *staticRoot,
		ChartsRoot:   *chartsRoot,
		ChartsPath:   *chartsPath,
	}

	web.Serve()
}

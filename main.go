// Package main implements the atlas review wizard.
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
// Reviews, (Question) Profiles, Responses, etc. and their
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

// dataPath tells us where to store our persisted
// entities.
var dataPath = flag.String("data", "data/", "path to atlas-forms database")

// htmlPath tells the web controller where to look for
// HTML templates to render.
var htmlPath = flag.String("html", "html/", "path to atlas-forms html templates")

// httpAddr tells the web controller on what address to
// listen for requests.
var httpAddr = flag.String("http", "127.0.0.1:3001", "addr:port")
var appRoot  = flag.String("approot", "", "approot prefix")

func main() {
	flag.Parse()

	persist := persist.NewPersistJSON(*dataPath)

	web := &web.App{
		QuestionRepo: entity.QuestionRepo(persist),
		ProfileRepo:  entity.ProfileRepo(persist),
		ReviewRepo:   entity.ReviewRepo(persist),
		HtmlPath:     *htmlPath,
		HttpAddr:     *httpAddr,
		AppRoot:      *appRoot,
	}

	web.Serve()
}

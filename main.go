package main

import (
	"akamai/atlas/forms/entity"
	"akamai/atlas/forms/persist"
	"akamai/atlas/forms/web"
	"flag"
)

var dataPath = flag.String("data", "data/", "path to atlas-forms database")
var htmlPath = flag.String("html", "html/", "path to atlas-forms html templates")
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

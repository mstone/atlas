package main

import (
	"akamai/atlas/forms/web"
	"flag"
)

func main() {
	flag.Parse()
	web.Serve()
}

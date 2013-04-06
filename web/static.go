package web

import (
	"log"
	"net/http"
	"path"
)

// BUG(mistone): HandleStatic() directory traversal?

func (self *App) HandleStatic(w http.ResponseWriter, r *http.Request) {
	fp, err := self.RemoveUrlPrefix(r.URL.Path, self.StaticRoot)
	checkHTTP(err)
	log.Printf("HandleStatic: file path: %v", fp)
	http.ServeFile(w, r, path.Join(self.StaticPath, fp))
}

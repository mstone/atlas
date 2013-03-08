package chart

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
)

// BUG(mistone): Chart's methods are not goroutine-safe.
type Chart struct {
	dsnPath     string
	srcPath     string
	meta        ChartMeta
	body        string
	hasBeenRead bool
}

type ChartMeta struct {
	Title   string
	Authors string
	Date    string
}

func NewChart(srcPath, dsnPath string) *Chart {
	return &Chart{
		dsnPath: dsnPath,
		srcPath: srcPath,
	}
}

func (self *Chart) Read() error {
	log.Printf("Chart.Read(): reading path %s", self.srcPath)
	body, err := ioutil.ReadFile(self.srcPath)
	//log.Printf("Chart.Read(): read body %s", body)
	log.Printf("Chart.Read(): read err %s", err)
	if err == nil {
		self.hasBeenRead = true
		self.body = string(body)
	}
	return err
}

func (self *Chart) Body() string {
	return self.body
}

func (self *Chart) Meta() (ChartMeta, error) {
	if !self.hasBeenRead {
		err := self.Read()
		if err != nil {
			return ChartMeta{}, err
		}
	}
	return self.meta, nil
}

type NotAChart struct {
	path string
	err  error
}

func (self *NotAChart) Error() string {
	return fmt.Sprintf("Error: %s, %s is not a chart.", self.err, self.path)
}

func (self *Chart) Slug() (string, error) {
	dir := filepath.Dir(self.srcPath)
	base := filepath.Base(self.srcPath)
	pfx := path.Clean(self.dsnPath)

	if base != "index.txt" && base != "index.text" {
		return "", &NotAChart{
			path: self.srcPath,
			err:  nil,
		}
	}

	log.Printf("Chart.Slug(): srcPath: %s", self.srcPath)
	log.Printf("Chart.Slug(): dsnPath: %s", self.dsnPath)
	log.Printf("Chart.Slug(): dir : %s", dir)
	log.Printf("Chart.Slug(): base: %s", base)
	log.Printf("Chart.Slug(): pfx : %s", pfx)

	var sfx string
	if len(dir) > len(pfx) {
		sfx = dir[len(pfx)+1:] + "/"
	} else {
		sfx = ""
	}
	log.Printf("Chart.Slug(): sfx : %s", sfx)
	return sfx, nil
}

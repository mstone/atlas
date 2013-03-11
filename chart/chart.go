package chart

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"strings"
)

// BUG(mistone): Chart's methods are not goroutine-safe.
type Chart struct {
	dsnPath     string
	srcPath     string
	meta        ChartMeta
	bytes       []byte
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
	if err != nil {
		return err
	}

	self.hasBeenRead = true
	self.bytes = body
	self.body = string(body)

	lines := strings.Split(self.body, "\n")
	log.Printf("Chart.Read: found %d lines", len(lines))

	foundHeaderChars := false
	if len(lines) > 3 {
		foundHeaderChars = true
		for i := 0; i < 3; i++ {
			if len(lines[i]) < 1 || lines[i][0] != '%' {
				foundHeaderChars = false
			}
		}

		if foundHeaderChars {
			self.meta.Title = strings.TrimLeft(lines[0], "% ")
			self.meta.Authors = strings.TrimLeft(lines[1], "% ")
			self.meta.Date = strings.TrimLeft(lines[2], "% ")
			self.body = strings.SplitAfterN(self.body, "\n", 4)[3]
		}
	}

	log.Printf("Chart.Read: found title: %s", self.meta.Title)
	log.Printf("Chart.Read: found authors: %s", self.meta.Authors)
	log.Printf("Chart.Read: found date: %s", self.meta.Date)

	return nil
}

func (self *Chart) Body() string {
	return self.body
}

func (self *Chart) Bytes() []byte {
	return self.bytes
}

func (self *Chart) Meta() ChartMeta {
	return self.meta
}

type NotAChart struct {
	path string
	err  error
}

func (self *NotAChart) Error() string {
	return fmt.Sprintf("Error: %s, %s is not a chart.", self.err, self.path)
}

func (self *Chart) IsChart() bool {
	base := filepath.Base(self.srcPath)

	if base != "index.txt" && base != "index.text" {
		return false
	}
	return true
}

func (self *Chart) Slug() string {
	dir := filepath.Dir(self.srcPath)
	pfx := path.Clean(self.dsnPath)

	//log.Printf("Chart.Slug(): srcPath: %s", self.srcPath)
	//log.Printf("Chart.Slug(): dsnPath: %s", self.dsnPath)
	//log.Printf("Chart.Slug(): dir : %s", dir)
	//log.Printf("Chart.Slug(): base: %s", base)
	//log.Printf("Chart.Slug(): pfx : %s", pfx)

	var sfx string
	if len(dir) > len(pfx) {
		sfx = dir[len(pfx)+1:] + "/"
	} else {
		sfx = ""
	}
	log.Printf("Chart.Slug(): sfx : %s", sfx)
	return sfx
}

func (self *Chart) Dir() string {
	return path.Dir(self.srcPath)
}

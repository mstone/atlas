package chart

import (
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func L(s string, v ...interface{}) {
	if glog.V(1) {
		glog.Infof("chart "+s, v...)
	}
}

// BUG(mistone): Chart's methods are not goroutine-safe.
type Chart struct {
	dsnPath string
	srcPath string
	fi      os.FileInfo
	meta    ChartMeta
	bytes   []byte
	body    string
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
		fi:      nil,
	}
}

func (self *Chart) Read() (err error) {
	L("read path %s", self.srcPath)
	f, err := os.Open(self.srcPath)
	if err != nil {
		return
	}
	defer f.Close()

	self.fi, err = f.Stat()
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(f)
	if glog.V(2) {
		L("read body %s", body)
	}
	L("read err %s", err)
	if err != nil {
		return
	}

	self.bytes = body
	self.body = string(body)

	lines := strings.Split(self.body, "\n")
	L("read found %d lines", len(lines))

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

	L("read title %q", self.meta.Title)
	L("read authors %q", self.meta.Authors)
	L("read date %q", self.meta.Date)

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

func (self *Chart) Make() (built bool, err error) {
	return true, self.Read()
}

func (self *Chart) IsChart() bool {
	base := filepath.Base(self.srcPath)

	if base != "index.txt" && base != "index.text" {
		return false
	}
	return true
}

func Resolve(dirPath, dsnPath string) (*Chart, error) {
	var err error

	txtPath := path.Join(dirPath, "index.txt")
	_, err = os.Stat(txtPath)

	if err == nil {
		return NewChart(txtPath, dsnPath), nil
	} else {
		if os.IsNotExist(err) {
			textPath := path.Join(dirPath, "index.text")
			_, err := os.Stat(textPath)

			if err == nil {
				return NewChart(textPath, dsnPath), nil
			}
		}
	}
	return nil, err
}

func (self *Chart) Slug() string {
	dir := filepath.Dir(self.srcPath)
	pfx := path.Clean(self.dsnPath)

	L("slug srcPath: %q", self.srcPath)
	L("slug dsnPath: %q", self.dsnPath)
	L("slug dir : %q", dir)
	L("slug pfx : %q", pfx)

	var sfx string
	if len(dir) > len(pfx) {
		sfx = dir[len(pfx)+1:] + "/"
	} else {
		sfx = ""
	}
	L("slug sfx %q", sfx)
	return sfx
}

func (self *Chart) Dir() string {
	return path.Dir(self.srcPath)
}

func (self *Chart) Src() string {
	return self.srcPath
}

func (self *Chart) FileInfo() os.FileInfo {
	return self.fi
}

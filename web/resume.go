package web

import (
	"akamai/atlas/chart"
	"akamai/atlas/resumes"

	"github.com/golang/glog"

	"io"
	"net/http"
	"os"
	"path"
)

func (self *App) HandleResumePost(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	fp, err := self.RemoveUrlPrefix(r.URL.Path, self.ChartsRoot)
	checkHTTP(err)
	glog.Infof("HandleResumePost(): fp: %v\n", fp)

	ext := path.Ext(fp)

	switch ext {
	default:
		glog.Fatalf("HandleResumePost(): bad ext: %q", ext)
	case ".doc":
		break
	case ".docx":
		break
	case ".pdf":
		break
	}

	// BUG(mistone): other name validation?
	dstPath := path.Join(self.ChartsPath, fp, "upload"+ext)
	dstDir := path.Dir(dstPath)

	err = os.MkdirAll(dstDir, 0755)
	checkHTTP(err)

	dstFile, err := os.Create(dstPath)
	checkHTTP(err)
	defer dstFile.Close()

	_, err = io.Copy(dstFile, r.Body)
	checkHTTP(err)

	displayName := resumes.SimplifyName(path.Base(fp[:len(fp)-len(ext)]))

	glog.Infof("HandleResumePost(): attempting to convert: %q -> %q", dstPath, dstDir)
	err = resumes.Convert(dstPath, dstDir, displayName)
	checkHTTP(err)

	chartName := path.Join(dstDir, "index.txt")
	chart := chart.NewChart(chartName, self.ChartsPath)

	if !chart.IsChart() {
		glog.Fatalf("HandleResumePost(): missing chart: %q", chartName)
	}

	err = chart.Read()
	if err != nil {
		glog.Fatalf("HandleResumePost(): bad chart: %q", chartName)
	}

	link, err := self.GetChartUrl(chart)

	if err == nil {
		http.Redirect(w, r, link.String(), http.StatusSeeOther)
	}
}

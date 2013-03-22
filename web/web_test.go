package web

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
)

var normalApp *App

func init() {
	testPath := os.Getenv("ATLAS_TEST_PATH")

	if testPath == "" {
		testPath = "../"
	}

	httpAddr := "localhost:3001"
	htmlPath := path.Join(testPath, "html/")
	chartsPath := path.Join(testPath, "test/charts/")
	staticPath := path.Join(testPath, "static/")
	chartsRoot := ""
	staticRoot := "static/"

	normalApp = &App{
		HttpAddr:   httpAddr,
		HtmlPath:   htmlPath,
		StaticPath: staticPath,
		StaticRoot: staticRoot,
		ChartsPath: chartsPath,
		ChartsRoot: chartsRoot,
	}

	normalApp.templates = template.Must(
		template.ParseGlob(
			path.Join(htmlPath, "*.html")))
}

func TestChartsGet(t *testing.T) {
	t.Parallel()
	t.Log("TestChartsGet(): starting.")
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://localhost:3001/", nil)
	normalApp.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("TestChartsGet() failed: response code %d != 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Demo Atlas") {
		t.Fatalf("TestChartsGet() failed: body does not mention 'Demo Atlas':\n %s", w.Body)
	}
}

func TestChartsGetIndexText(t *testing.T) {
	t.Parallel()
	t.Log("TestChartsGet(): starting.")
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://localhost:3001/subchart/", nil)
	normalApp.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("TestChartsGetIndexText() failed: response code %d != 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "What could go wrong?") {
		t.Fatalf("TestChartsGetIndexText() failed: body does not mention 'What could go wrong?':\n %s", w.Body)
	}
}

func TestSiteJsonGet(t *testing.T) {
	t.Parallel()
	t.Log("TestSiteJsonGet(): starting.")
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://localhost:3001/site.json", nil)
	normalApp.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("TestSiteJsonGet() failed: response code %d != 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Demo Atlas") {
		t.Fatalf("TestSiteJsonGet() failed: body does not mention 'Demo Atlas':\n %s", w.Body)
	}
	if !strings.Contains(body, "ad tag") {
		t.Fatalf("TestSiteJsonGet() failed: body does not mention 'ad tag':\n %s", w.Body)
	}
}

func TestChartSetGet(t *testing.T) {
	t.Parallel()
	t.Log("TestChartSetGet(): starting.")
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://localhost:3001/pages/", nil)
	normalApp.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("TestChartSetGet() failed: response code %d != 200", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Demo Atlas") {
		t.Fatalf("TestChartSetGet() failed: body does not mention 'Demo Atlas':\n %s", w.Body)
	}
	if !strings.Contains(body, "Just for kicks...") {
		t.Fatalf("TestChartSetGet() failed: body does not mention 'Just for kicks...':\n %s", w.Body)
	}
	if !strings.Contains(body, "subchart") {
		t.Fatalf("TestChartSetGet() failed: body does not mention 'subchart...':\n %s", w.Body)
	}
	if !strings.Contains(body, "index.text") {
		t.Fatalf("TestChartSetGet() failed: body does not mention 'index.text...':\n %s", w.Body)
	}
}

func TestChartGet404(t *testing.T) {
	t.Parallel()
	t.Log("TestChartGet(): starting.")
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://localhost:3001/i.do.not.exist", nil)
	normalApp.ServeHTTP(w, r)
	if w.Code != 404 {
		t.Fatalf("TestChartSetGet() failed: response code %d != 404", w.Code)
	}
}

func TestSvgEditorGet(t *testing.T) {
	t.Parallel()
	t.Log("TestSvgEditorGet(): starting.")
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://localhost:3001/hades.svg/editor", nil)
	normalApp.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("TestSvgEditorGet() failed: response code %d != 404", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "svg-editor") {
		t.Fatalf("TestSvgEditorGet() failed: body does not mention 'svg-editor':\n %s", w.Body)
	}
}

func TestRemoveUrlPrefix(t *testing.T) {
	t.Parallel()
	t.Log("TestRemoveUrlPrefix(): starting.")

	fp, err := normalApp.RemoveUrlPrefix("/", "/")
	if fp != "" || err != nil {
		t.Fatalf("TestRemoveUrlPrefix() failed: (/ /) -> (%q %q)", fp, err)
	}

	fp, err = normalApp.RemoveUrlPrefix("/", "/abc")
	if fp != "" || err == nil {
		t.Fatalf("TestRemoveUrlPrefix() failed: (/ /) -> (%q %q)", fp, err)
	}

	fp, err = normalApp.RemoveUrlPrefix("/abc", "")
	if fp != "abc" || err != nil {
		t.Fatalf("TestRemoveUrlPrefix() failed: (/ /) -> (%q %q)", fp, err)
	}

	fp, err = normalApp.RemoveUrlPrefix("/abc", "/abc/")
	if fp != "" || err != nil {
		t.Fatalf("TestRemoveUrlPrefix() failed: (/ /) -> (%q %q)", fp, err)
	}
}

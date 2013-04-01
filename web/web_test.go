package web

import (
	"akamai/atlas/shake"
	"bytes"
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

	normalApp.Shake = shake.NewRuleSet()
	normalApp.Shake.Rules = []shake.Rule{
		&StaticContentRule{normalApp},
		&ChartsContentRule{normalApp},
		&TemplateRule{normalApp},
		&shake.ReadFileRule{""},
	}
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

func TestSvgEditorPost(t *testing.T) {
	t.Parallel()
	t.Log("TestSvgEditorPost(): starting.")

	// try to delete whatever we create
	defer os.RemoveAll(path.Join(normalApp.ChartsPath, "newchart"))

	svgUrl := "/newchart/new.svg"
	svgEditorUrl := path.Join(svgUrl, "editor")

	// check that new.svg doesn't exist yet...
	w0 := httptest.NewRecorder()
	r0, _ := http.NewRequest("GET", svgUrl, nil)
	normalApp.ServeHTTP(w0, r0)
	if w0.Code != 404 {
		t.Fatalf("TestSvgEditorPost() failed: response code %d != 404", w0.Code)
	}

	// create new.svg
	w1 := httptest.NewRecorder()
	svgBody := "filepath=PD94bWwgdmVyc2lvbj0iMS4wIj8%2BCjxzdmcgd2lkdGg9IjgwMCIgaGVpZ2h0PSI2MDAiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI%2BCgogPG1ldGFkYXRhIGlkPSJtZXRhZGF0YTciPmltYWdlL3N2Zyt4bWw8L21ldGFkYXRhPgogPGc%2BCiAgPHRpdGxlPkxheWVyIDE8L3RpdGxlPgogIDx0ZXh0IHN0cm9rZT0iIzAwMDAwMCIgdHJhbnNmb3JtPSJtYXRyaXgoMSAwIDAgMS4wNjg5NyAwIC0zLjU4NjIxKSIgeG1sOnNwYWNlPSJwcmVzZXJ2ZSIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZm9udC1mYW1pbHk9InNlcmlmIiBmb250LXNpemU9IjI0IiBpZD0ic3ZnXzEiIHk9IjMyNy4wMTYxMTkiIHg9IjI5NiIgc3Ryb2tlLXdpZHRoPSIwIiBmaWxsPSIjMDAwMDAwIj53aWRnZXQ8L3RleHQ%2BCiAgPHRleHQgeG1sOnNwYWNlPSJwcmVzZXJ2ZSIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZm9udC1mYW1pbHk9InNlcmlmIiBmb250LXNpemU9IjI0IiBpZD0ic3ZnXzIiIHk9IjIyMy41IiB4PSIzOTEiIHN0cm9rZS13aWR0aD0iMCIgc3Ryb2tlPSIjMDAwMDAwIiBmaWxsPSIjMDAwMDAwIj5naXptb3M8L3RleHQ%2BCiAgPHRleHQgeG1sOnNwYWNlPSJwcmVzZXJ2ZSIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZm9udC1mYW1pbHk9InNlcmlmIiBmb250LXNpemU9IjI0IiBpZD0ic3ZnXzMiIHk9IjM0Ni41IiB4PSI1MDkiIHN0cm9rZS13aWR0aD0iMCIgc3Ryb2tlPSIjMDAwMDAwIiBmaWxsPSIjMDAwMDAwIj5zcHJvY2tldDwvdGV4dD4KICA8bGluZSBzdHJva2U9IiM4MDAwMDAiIGlkPSJzdmdfNSIgeTI9IjI0MCIgeDI9IjM1OSIgeTE9IjMyMC45OTk5OTciIHgxPSIzMjIiIGZpbGw9Im5vbmUiLz4KICA8bGluZSBzdHJva2U9IiM4MDAwMDAiIGlkPSJzdmdfNiIgeTI9IjMxOSIgeDI9IjQ3MCIgeTE9IjIzNS4wMDAwMDQiIHgxPSI0MzIiIGZpbGw9Im5vbmUiLz4KIDwvZz4KPC9zdmc%2B&filename=drawing.svg&contenttype=application%2Fx-svgdraw"
	svgBodyReader := bytes.NewBufferString(svgBody)
	r1, _ := http.NewRequest("POST", svgEditorUrl, svgBodyReader)
	r1.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	normalApp.ServeHTTP(w1, r1)
	if w1.Code != 204 {
		t.Fatalf("TestSvgEditorPost() failed: response code %d != 204", w1.Code)
	}

	// check that new.svg now exists
	w2 := httptest.NewRecorder()
	r2, _ := http.NewRequest("GET", svgUrl, nil)
	normalApp.ServeHTTP(w2, r2)
	if w2.Code != 200 {
		t.Fatalf("TestSvgEditorPost() failed: response code %d != 200", w2.Code)
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

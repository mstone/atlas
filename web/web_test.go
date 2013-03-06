package web

import (
	"akamai/atlas/forms/entity"
	"akamai/atlas/forms/persist"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
)

var normalPersist *persist.PersistJSON
var normalApp *App

func init() {
	testPath := os.Getenv("ATLAS_TEST_PATH")

	if testPath == "" {
		testPath = "../"
	}

	httpAddr := "localhost:3001"
	dataPath := path.Join(testPath, "test/data/")
	htmlPath := path.Join(testPath, "html/")
	chartsPath := path.Join(testPath, "test/charts/")
	staticPath := path.Join(testPath, "static/")
	formsRoot := "forms/"
	chartsRoot := ""
	staticRoot := "static/"

	normalPersist = persist.NewPersistJSON(dataPath)

	normalApp = &App{
		HttpAddr:     httpAddr,
		QuestionRepo: entity.QuestionRepo(normalPersist),
		FormRepo:  entity.FormRepo(normalPersist),
		RecordRepo:   entity.RecordRepo(normalPersist),
		HtmlPath:     htmlPath,
		StaticPath:   staticPath,
		StaticRoot:   staticRoot,
		ChartsPath:   chartsPath,
		ChartsRoot:   chartsRoot,
		FormsRoot:    formsRoot,
	}

	normalApp.templates = template.Must(
		template.ParseGlob(
			path.Join(htmlPath, "*.html")))
}

func TestRecordSetGet(t *testing.T) {
	t.Parallel()
	t.Log("TestRecordSetGet(): starting.")
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://localhost:3001/forms/records/", nil)
	normalApp.ServeHTTP(w, r)
	t.Logf("TestRecordSetGet(): status code %d", w.Code)
	if w.Code != 200 {
		t.Fatalf("TestRecordSetGet() failed: %s", w)
	}
}

func TestQuestionSetGet(t *testing.T) {
	t.Parallel()
	t.Log("TestQuestionSetGet(): starting.")
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://localhost:3001/forms/questions/", nil)
	normalApp.ServeHTTP(w, r)
	t.Logf("TestQuestionGet(): status code %d", w.Code)
	if w.Code != 200 {
		t.Fatalf("TestRecordSetGet() failed: %s", w)
	}
}

func TestFormSetGet(t *testing.T) {
	t.Parallel()
	t.Log("TestFormSetGet(): starting.")
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://localhost:3001/forms/forms/", nil)
	normalApp.ServeHTTP(w, r)
	t.Logf("TestFormGet(): status code %d", w.Code)
	if w.Code != 200 {
		t.Fatalf("TestRecordSetGet() failed: %s", w)
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
}

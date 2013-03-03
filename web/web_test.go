package web

import (
	"akamai/atlas/forms/entity"
	"akamai/atlas/forms/persist"
	"html/template"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"
)

const (
	httpAddr   = "localhost:3001"
	dataPath   = "../test/data/"
	htmlPath   = "../html/"
	chartsPath = "../test/charts/"
	staticPath = "../static/"
	formsRoot  = "forms/"
	chartsRoot = ""
	staticRoot = "static/"
)

var normalPersist *persist.PersistJSON
var normalApp *App

func init() {
	normalPersist = persist.NewPersistJSON(dataPath)

	normalApp = &App{
		HttpAddr:     httpAddr,
		QuestionRepo: entity.QuestionRepo(normalPersist),
		ProfileRepo:  entity.ProfileRepo(normalPersist),
		ReviewRepo:   entity.ReviewRepo(normalPersist),
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

func TestReviewSetGet(t *testing.T) {
	t.Parallel()
	t.Log("TestReviewSetGet(): starting.")
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://localhost:3001/forms/reviews/", nil)
	normalApp.ServeHTTP(w, r)
	t.Logf("TestReviewSetGet(): status code %d", w.Code)
	if w.Code != 200 {
		t.Fatalf("TestReviewSetGet() failed: %s", w)
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
		t.Fatalf("TestReviewSetGet() failed: %s", w)
	}
}

func TestProfileSetGet(t *testing.T) {
	t.Parallel()
	t.Log("TestProfileSetGet(): starting.")
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "http://localhost:3001/forms/profiles/", nil)
	normalApp.ServeHTTP(w, r)
	t.Logf("TestProfileGet(): status code %d", w.Code)
	if w.Code != 200 {
		t.Fatalf("TestReviewSetGet() failed: %s", w)
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

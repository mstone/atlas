package chart

import (
	"os"
	"path"
	"strings"
	"testing"
)

var testPath string
var chartsPath string

func init() {
	testPath := os.Getenv("ATLAS_TEST_PATH")

	if testPath == "" {
		testPath = "../"
	}

	chartsPath = path.Join(testPath, "test/charts")
}

func TestChartRead(t *testing.T) {
	t.Parallel()
	chart := NewChart(path.Join(chartsPath, "index.txt"), chartsPath)

	err := chart.Read()
	if err != nil {
		t.Fatalf("TestChartsRead() failed: unable to read test chart: %s", err)
	}

	meta := chart.Meta()
	if meta.Title != "Demo Atlas" {
		t.Fatalf("TestChartsRead() failed: title is not 'Demo Atlas':\n %s", meta.Title)
	}
	if meta.Authors != "Michael Stone" {
		t.Fatalf("TestChartsRead() failed: authors is not 'Michael Stone':\n %s", meta.Authors)
	}
	if meta.Date != "March 3, 2013" {
		t.Fatalf("TestChartsRead() failed: date is not 'March 3, 2013':\n %s", meta.Date)
	}

	body := chart.Body()
	if !strings.Contains(body, "gotta start somewhere") {
		t.Fatalf("TestChartsRead() failed: body does not mention 'Demo Atlas':\n %s", body)
	}
}

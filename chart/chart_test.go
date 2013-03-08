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

	body := chart.Body()
	if !strings.Contains(body, "Demo Atlas") {
		t.Fatalf("TestChartsRead() failed: body does not mention 'Demo Atlas':\n %s", body)
	}
}

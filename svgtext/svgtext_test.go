package svgtext

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
)

var testPath string
var svgtextPath string

func init() {
	testPath := os.Getenv("ATLAS_TEST_PATH")

	if testPath == "" {
		testPath = "../"
	}

	svgtextPath = path.Join(testPath, "test/svgtext")
}

func TestGetCData(t *testing.T) {
	t.Parallel()
	text, err := ioutil.ReadFile(path.Join(svgtextPath, "hades.svg"))
	if err != nil {
		t.Fatalf("TestGetCData() failed: unable to read hades.svg: %s", err)
	}

	cdata, err := GetCData(text)
	if err != nil {
		t.Fatalf("TestGetCData() failed: unable to parse hades.svg: %s", err)
	}

	buf := bytes.Buffer{}
	for _, datum := range cdata {
		buf.WriteString(datum)
		buf.WriteRune('\n')
	}

	res := buf.String()
	//log.Printf("TestGetCData(): found cdata: %s", res)
	if !strings.Contains(res, "browser") {
		t.Fatalf("TestGetCData() failed: results string did not contain test vector 'browser'")
	}
}

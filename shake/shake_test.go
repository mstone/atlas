package shake

import (
	"os"
	"path"
	"testing"
	"time"
)

var testPath string
var shakePath string

func init() {
	testPath := os.Getenv("ATLAS_TEST_PATH")

	if testPath == "" {
		testPath = "../"
	}

	shakePath = path.Join(testPath, "test/shake")
}

type AlwaysRule struct{}

func (self *AlwaysRule) Matches(question Question, key Key) bool {
	return true
}

func (self *AlwaysRule) Make(question Question, key Key, rules *RuleSet) (Result, error) {
	value := "hi"
	result := Result{
		Key:     key,
		Changed: true,
		Value:   value,
		Rule:    self,
		Deps:    nil,
		Cookie:  nil,
	}
	return result, nil
}

func (self *AlwaysRule) Validate(key Key, cookie interface{}) error {
	return nil
}

func TestAlwaysKey(t *testing.T) {
	rs := NewRuleSet()
	rs.Rules = append(rs.Rules, &AlwaysRule{})
	res, err := rs.Make(StringQuestion("i.do.exit!"))
	if err != nil {
		t.Fatalf("Shake failed w/ a default rule!")
	}
	t.Logf("res: %s", res)
}

func TestMissingKey(t *testing.T) {
	rs := NewRuleSet()
	_, err := rs.Make(StringQuestion("i.do.not.exist"))
	if err == nil {
		t.Fatalf("Shake failed to fail!")
	}
}

func TestFileKey(t *testing.T) {
	name := path.Join(shakePath, "demo.txt")
	rs := NewRuleSet()
	rs.Rules = append(rs.Rules, &ReadFileRule{Pattern: name})

	// First we read a test file...
	question := ReadFileQuestion(name)
	res, err := rs.Make(question)
	if err != nil {
		t.Fatalf("Shake failed w/ FileRule!")
	}
	val, ok := res.Value.(string)
	if !ok {
		t.Fatalf("shake.TestFileKey() failed: result /value/ type != string, res: %s", res)
	}
	if val != "Hi\n" {
		t.Fatalf("shake.TestFileKey() failed: result value != Hi, res: %q", res)
	}

	// ...and then we read it again to make sure our cache validation logic
	// works...
	res2, err := rs.Make(question)
	if err != nil {
		t.Fatalf("Shake failed w/ FileRule!")
	}
	if res2.Changed {
		t.Fatalf("shake.TestFileKey() failed: bad file change reported: %s", res2)
	}
	val2, ok := res2.Value.(string)
	if !ok {
		t.Fatalf("shake.TestFileKey() failed: result /value/ type != string, res: %s", res2)
	}
	if val2 != "Hi\n" {
		t.Fatalf("shake.TestFileKey() failed: result value != Hi, res: %q", res2)
	}

	// ...and then we touch it...
	now := time.Now()
	err = os.Chtimes(name, now, now)
	if err != nil {
		t.Fatalf("shake.TestFileKey(): unable to touch test file: %s", err)
	}

	// ...and then we read it a third time to check that our cache
	// *in*validation works.
	res3, err := rs.Make(question)
	if err != nil {
		t.Fatalf("Shake failed w/ FileRule!")
	}
	if !res3.Changed {
		t.Fatalf("shake.TestFileKey() failed: no file change reported: %s", res3)
	}
	val3, ok := res3.Value.(string)
	if !ok {
		t.Fatalf("shake.TestFileKey() failed: result /value/ type != string, res: %s", res3)
	}
	if val3 != "Hi\n" {
		t.Fatalf("shake.TestFileKey() failed: result value != Hi, res: %q", res3)
	}

}

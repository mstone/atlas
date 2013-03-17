package shake

import (
	"io/ioutil"
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

func (self *AlwaysRule) Matches(key Key) bool {
	return true
}

func (self *AlwaysRule) Make(key Key, rules *RuleSet) (Result, error) {
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
	res, err := rs.Make("i.do.exit!")
	if err != nil {
		t.Fatalf("Shake failed w/ a default rule!")
	}
	t.Logf("res: %s", res)
}

func TestMissingKey(t *testing.T) {
	rs := NewRuleSet()
	_, err := rs.Make("i.do.not.exist")
	if err == nil {
		t.Fatalf("Shake failed to fail!")
	}
}

type FileRule struct {
	Pattern Key
}

func (self *FileRule) Matches(key Key) bool {
	return key == self.Pattern
}

func (self *FileRule) Make(key Key, rules *RuleSet) (Result, error) {
	file, err := os.Open(string(key))
	if err != nil {
		return Result{}, err
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return Result{}, err
	}

	value, err := ioutil.ReadAll(file)
	if err != nil {
		return Result{}, err
	}

	text := string(value)

	result := Result{
		Key:     key,
		Changed: true,
		Value:   text,
		Rule:    self,
		Deps:    nil,
		Cookie:  fi,
	}
	return result, nil
}

func (self *FileRule) Validate(key Key, cookie interface{}) error {
	fi, err := os.Stat(string(key))
	if err != nil {
		return err
	}

	oldFi, ok := cookie.(os.FileInfo)
	if !ok {
		return &BadCookieError{
			Key: key,
		}
	}

	err = nil

	sizeOk := fi.Size() == oldFi.Size()
	modeOk := fi.Mode() == oldFi.Mode()
	modTimeOk := fi.ModTime() == oldFi.ModTime()

	if !(sizeOk && modeOk && modTimeOk) {
		err = &OutOfDateError{
			Key: key,
		}
	}
	return err
}

func TestFileKey(t *testing.T) {
	name := path.Join(shakePath, "demo.txt")
	rs := NewRuleSet()
	rs.Rules = append(rs.Rules, &FileRule{Pattern: Key(name)})

	// First we read a test file...
	res, err := rs.Make(Key(name))
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
	res2, err := rs.Make(Key(name))
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
	res3, err := rs.Make(Key(name))
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

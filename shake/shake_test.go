package shake

import (
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
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
		Epoch:   0,
		Type:    reflect.TypeOf(value),
		Value:   value,
		Deps:    nil,
		Cookie:  nil,
	}
	return result, nil
}

func (self *AlwaysRule) Validate(key Key, cookie interface{}) error {
	return nil
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
		Epoch:   0,
		Type:    reflect.TypeOf(text),
		Value:   text,
		Deps:    nil,
		Cookie:  fi,
	}
	return result, nil
}

func (self *FileRule) Validate(key Key, cookie interface{}) error {
	return &OutOfDateError{}
}

func TestFileKey(t *testing.T) {
	name := path.Join(shakePath, "demo.txt")
	rs := NewRuleSet()
	rs.Rules = append(rs.Rules, &FileRule{Pattern: Key(name)})
	res, err := rs.Make(Key(name))
	if err != nil {
		t.Fatalf("Shake failed w/ FileRule!")
	}
	if res.Type != reflect.TypeOf("") {
		t.Fatalf("shake.TestFileKey() failed: result type != string, res: %s", res)
	}
	val, ok := res.Value.(string)
	if !ok {
		t.Fatalf("shake.TestFileKey() failed: result /value/ type != string, res: %s", res)
	}
	if val != "Hi\n" {
		t.Fatalf("shake.TestFileKey() failed: result value != Hi, res: %q", res)
	}
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

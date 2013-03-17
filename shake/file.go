package shake

import (
	"io/ioutil"
	"os"
)

// ReadFileRule is a rule for reading the contents of a given file.
// ReadFileRule's cookies are os.FileInfo. ReadFileRule cookies are valid if the
// Size(), Mode(), and ModTime() match. Presently, ReadFileRule patterns are
// matched against StringQuestions for simple string equality.
type ReadFileRule struct {
	Pattern string
}

func (self *ReadFileRule) Matches(question Question, key Key) bool {
	ret := false
	if val, ok := question.(StringQuestion); ok {
		ret = string(val) == self.Pattern
	}
	return ret
}

func (self *ReadFileRule) Make(question Question, key Key, rules *RuleSet) (Result, error) {
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

func (self *ReadFileRule) Validate(key Key, cookie interface{}) error {
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

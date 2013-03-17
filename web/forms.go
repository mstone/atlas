package web

import (
	"akamai/atlas/forms/shake"
	"log"
	"path"
	"strings"
)

type FormsContentRule struct {
	*App
}

func (self *FormsContentRule) Matches(question shake.Question, key shake.Key) bool {
	ret := false
	if q, ok := question.(WebQuestion); ok {
		ret = strings.HasPrefix(q.URL.Path, path.Clean("/"+self.FormsRoot))
	}
	if ret {
		log.Printf("FormsContentRule.Matches(): true.")
	}
	return ret
}

func (self *FormsContentRule) Make(question shake.Question, key shake.Key, rules *shake.RuleSet) (shake.Result, error) {
	if q, ok := question.(WebQuestion); ok {
		self.HandleForms(q.ResponseWriter, q.Request)
		result := shake.Result{
			Key:     key,
			Changed: true,
			Value:   nil,
			Rule:    self,
			Deps:    nil,
			Cookie:  nil,
		}
		return result, nil
	}
	return shake.Result{}, &shake.BadQuestionError{
		Key: key,
	}
}

func (self *FormsContentRule) Validate(key shake.Key, cookie interface{}) error {
	return &shake.OutOfDateError{key}
}

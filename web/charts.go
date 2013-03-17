package web

import (
	"akamai/atlas/forms/shake"
	"log"
	"path"
	"strings"
)

type ChartsContentRule struct {
	*App
}

func (self *ChartsContentRule) Matches(question shake.Question, key shake.Key) bool {
	ret := false
	if q, ok := question.(WebQuestion); ok {
		ret = strings.HasPrefix(q.URL.Path, path.Clean("/"+self.ChartsRoot))
	}
	if ret {
		log.Printf("ChartsContentRule.Matches(): true.")
	}
	return ret
}

func (self *ChartsContentRule) Make(question shake.Question, key shake.Key, rules *shake.RuleSet) (shake.Result, error) {
	if q, ok := question.(WebQuestion); ok {
		self.HandleChart(q.ResponseWriter, q.Request)
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

func (self *ChartsContentRule) Validate(key shake.Key, cookie interface{}) error {
	return &shake.OutOfDateError{key}
}

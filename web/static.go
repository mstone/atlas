package web

import (
	"akamai/atlas/shake"
	"log"
	"net/http"
	"path"
	"strings"
)

// BUG(mistone): HandleStatic() directory traversal?

func (self *App) HandleStatic(w http.ResponseWriter, r *http.Request) {
	fp, err := self.RemoveUrlPrefix(r.URL.Path, self.StaticRoot)
	checkHTTP(err)
	log.Printf("HandleStatic: file path: %v", fp)
	http.ServeFile(w, r, path.Join(self.StaticPath, fp))
}

type StaticContentRule struct {
	*App
}

func (self *StaticContentRule) Matches(question shake.Question, key shake.Key) bool {
	ret := false
	if q, ok := question.(WebQuestion); ok {
		ret = strings.HasPrefix(q.URL.Path, path.Clean("/"+self.StaticRoot))
	}
	if ret {
		log.Printf("StaticContentRule.Matches(): true.")
	}
	return ret
}

func (self *StaticContentRule) Make(question shake.Question, key shake.Key, rules *shake.RuleSet) (shake.Result, error) {
	if q, ok := question.(WebQuestion); ok {
		self.HandleStatic(q.ResponseWriter, q.Request)
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

func (self *StaticContentRule) Validate(key shake.Key, cookie interface{}) error {
	return &shake.OutOfDateError{key}
}

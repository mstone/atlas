package web

import (
	"akamai/atlas/shake"
	"html/template"
	"path"
)

type TemplateQuestion struct {
	templateName string
}

func (self TemplateQuestion) Key() (shake.Key, error) {
	return shake.Key("aka-tmpl://" + self.templateName), nil
}

type TemplateRule struct {
	*App
}

func (self *TemplateRule) Matches(question shake.Question, key shake.Key) bool {
	_, ok := question.(TemplateQuestion)
	return ok
}

func (self *TemplateRule) Make(question shake.Question, key shake.Key, rules *shake.RuleSet) (shake.Result, error) {
	if _, ok := question.(TemplateQuestion); ok {
		tmpls := template.Must(
			template.ParseGlob(
				path.Join(self.HtmlPath, "*.html")))

		result := shake.Result{
			Key:     key,
			Changed: true,
			Value:   tmpls,
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

func (self *TemplateRule) Validate(key shake.Key, cookie interface{}) error {
	return &shake.OutOfDateError{key}
}

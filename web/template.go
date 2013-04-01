package web

import (
	"akamai/atlas/shake"
	"fmt"
	"html/template"
	"log"
	"path"
	"text/template/parse"
)

type TemplateQuestion struct {
	templateName string
}

type TemplateAnswer struct {
	Template *template.Template
	Tree     *parse.Tree
}

func (self TemplateQuestion) Key() (shake.Key, error) {
	return shake.Key("atlas-template://" + self.templateName), nil
}

type TemplateRule struct {
	*App
}

func (self *TemplateRule) Matches(question shake.Question, key shake.Key) bool {
	_, ok := question.(TemplateQuestion)
	return ok
}

func findDepsBranch(node parse.BranchNode, deps *[]string) error {
	if node.List != nil {
		for _, kid := range node.List.Nodes {
			err := findDeps(kid, deps)
			if err != nil {
				return err
			}
		}
	}
	if node.ElseList != nil {
		for _, kid := range node.ElseList.Nodes {
			err := findDeps(kid, deps)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func findDeps(node parse.Node, deps *[]string) error {
	switch node.Type() {
	default:
		break
	case parse.NodeList:
		v, ok := node.(*parse.ListNode)
		if !ok {
			return fmt.Errorf("Node %t not a ListNode!", node)
		}
		for _, kid := range v.Nodes {
			err := findDeps(kid, deps)
			if err != nil {
				return err
			}
		}
	case parse.NodeTemplate:
		v, ok := node.(*parse.TemplateNode)
		if !ok {
			return fmt.Errorf("Node %t not a TemplateNode!", node)
		}
		*deps = append(*deps, v.Name)
		return nil
	case parse.NodeIf:
		v, ok := node.(*parse.IfNode)
		if !ok {
			return fmt.Errorf("Node %t not a IfNode!", node)
		}
		findDepsBranch(v.BranchNode, deps)
	case parse.NodeRange:
		v, ok := node.(*parse.RangeNode)
		if !ok {
			return fmt.Errorf("Node %t not a RangeNode!", node)
		}
		findDepsBranch(v.BranchNode, deps)
	case parse.NodeWith:
		v, ok := node.(*parse.WithNode)
		if !ok {
			return fmt.Errorf("Node %t not a WithNode!", node)
		}
		findDepsBranch(v.BranchNode, deps)
	}
	return nil
}

func (self *TemplateRule) Make(question shake.Question, key shake.Key, rules *shake.RuleSet) (shake.Result, error) {
	if q, ok := question.(TemplateQuestion); ok {
		// tmpls := template.Must(
		// 	template.ParseGlob(
		// 		path.Join(self.HtmlPath, "*.html")))

		// ick
		rootTmplPath := path.Join(self.HtmlPath, q.templateName+".html")
		textQuestion := shake.ReadFileQuestion(rootTmplPath)
		textAnswer, err := self.Shake.Make(textQuestion)
		if err != nil {
			return shake.Result{}, err
		}

		text, ok := textAnswer.Value.(string)
		if !ok {
			log.Printf("TemplateRule.Make(): got textAnswer: %t", textAnswer)
			return shake.Result{}, fmt.Errorf("TemplateRule.Make(); text not a string.")
		}

		treeSet, err := parse.Parse(q.templateName, text, "", "")
		log.Printf("TemplateRule.Make(): got treeSet: %t, err: %t", treeSet, err)

		tree := treeSet[q.templateName]
		if tree == nil {
			log.Printf("TemplateRule.Make(): couldn't find tree at key: %q", q.templateName)
			return shake.Result{}, fmt.Errorf("TemplateRule.Make(): couldn't find tree at key: %q", q.templateName)
		}

		deps := []string{}
		err = findDeps(parse.Node(tree.Root), &deps)
		if err != nil {
			log.Printf("TemplateRule.Make(): failed to find deps: %s", err)
		}
		log.Printf("TemplateRule.Make(): FOUND DEPS: %q", deps)

		//htmlTmpl := template.New(q.templateName)
		htmlTmpl, err := template.New("").Parse("")
		if err != nil {
			log.Printf("TemplateRule.Make(): failed to initialize empty root template: %s", err)
		}
		checkHTTP(err)
		if htmlTmpl == nil {
			panic("TemplateRule.Make(): Unable to allocate new \"html/template\".Template")
		}

		log.Printf("BOOM: %t", htmlTmpl)
		log.Printf("BOOM: %t", q.templateName)
		log.Printf("BOOM: %t", tree)

		_, err = htmlTmpl.AddParseTree(q.templateName, tree)
		checkHTTP(err)

		// BUG(mistone): no protection from infinite build loops!
		for _, dep := range deps {
			depTmplQuestion := TemplateQuestion{dep}
			depTmplResult, err := self.Shake.Make(depTmplQuestion)
			if err != nil {
				log.Printf("TemplateRule.Make(): failed to build DEP: %q, %q", dep, err)
				return shake.Result{}, err
			}
			depTmplAnswer, ok := depTmplResult.Value.(TemplateAnswer)
			if !ok {
				log.Printf("TemplateRule.Make(): building DEP didn't return a TemplateAnswer: %q, %q", dep, depTmplAnswer)
			}
			depTree := depTmplAnswer.Tree
			_, err = htmlTmpl.AddParseTree(depTree.Name, depTree)
			if err != nil {
				log.Printf("TemplateRule.Make(): failed to merge DEP: %q, %q, %q", dep, depTmplAnswer, err)
				return shake.Result{}, err
			}
		}

		answer := TemplateAnswer{
			Template: htmlTmpl,
			Tree:     tree,
		}

		result := shake.Result{
			Key:     key,
			Changed: true,
			Value:   answer,
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

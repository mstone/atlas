package templatecache

import (
	"akamai/atlas/stat"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path"
	"text/template/parse"
)

var logTemplateCache = flag.Bool("log.cache.template", false, "log template cache ops")

func L(s string, v ...interface{}) {
	if *logTemplateCache {
		log.Printf("tc "+s, v...)
	}
}

type TemplateEnt struct {
	Template *template.Template
	Tree     *parse.Tree
	fi       os.FileInfo
	deps     []string
}

type TemplateCache struct {
	HtmlPath string
	Entries  map[string]TemplateEnt
}

func New(htmlPath string) *TemplateCache {
	return &TemplateCache{
		HtmlPath: htmlPath,
		Entries:  map[string]TemplateEnt{},
	}
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

func (self *TemplateCache) Make(templateName string) (built bool, err error) {
	built = false
	err = nil
	rootTmplPath := path.Join(self.HtmlPath, templateName+".html")

	fresh, err := self.allFresh(templateName, rootTmplPath)
	if err != nil {
		return
	}

	if !fresh {
		built = true

		var fi os.FileInfo
		fi, err = os.Stat(rootTmplPath)
		if err != nil {
			return
		}

		err = self.reread(templateName, rootTmplPath, fi)
		if err != nil {
			return
		}
	}
	return
}

func (self *TemplateCache) isFresh(a, b os.FileInfo) bool {
	return stat.IsFresh(a, b)
}

func (self *TemplateCache) allFresh(templateName string, fileName string) (fresh bool, err error) {
	L("allFresh templateName %q fileName %q", templateName, fileName)
	fresh = false
	err = nil

	ent, ok := self.Entries[templateName]
	if !ok {
		return
	}

	var fi os.FileInfo
	fi, err = os.Stat(fileName)
	if err != nil {
		return
	}

	fresh = self.isFresh(fi, ent.fi)
	if !fresh {
		return
	}

	for _, dep := range ent.deps {
		// BUG(mistone): may duplicate work if deps are shared...
		fresh, err = self.allFresh(dep, path.Join(self.HtmlPath, dep+".html"))
		if err != nil {
			return
		}
		if !fresh {
			return
		}
	}

	return
}

func (self *TemplateCache) reread(templateName string, fileName string, fi os.FileInfo) (err error) {
	L("reread templateName %q fileName %q, fi %v", templateName, fileName, fi)

	body, err := ioutil.ReadFile(fileName)
	if err != nil {
		return
	}

	text := string(body)

	treeSet, err := parse.Parse(templateName, text, "", "")
	L("reread got treeSet: %t, err: %t", treeSet, err)
	if err != nil {
		return
	}

	tree := treeSet[templateName]
	if tree == nil {
		L("reread couldn't find tree at key: %q", templateName)
		err = fmt.Errorf("TemplateRule.Make(): couldn't find tree at key: %q", templateName)
		return
	}

	deps := []string{}
	err = findDeps(parse.Node(tree.Root), &deps)
	if err != nil {
		L("reread failed to find deps: %s", err)
		return
	}
	L("reread FOUND DEPS: %q", deps)

	htmlTmpl, err := template.New("").Parse("")
	if err != nil {
		L("reread failed to initialize empty root template: %s", err)
		return
	}
	if htmlTmpl == nil {
		err = fmt.Errorf("TemplateCache.reread(): Unable to allocate new \"html/template\".Template")
	}

	_, err = htmlTmpl.AddParseTree(templateName, tree)
	if err != nil {
		return
	}

	// BUG(mistone): no protection from infinite build loops!
	for _, dep := range deps {
		L("reread recursing on dep %q", dep)
		_, err = self.Make(dep)
		if err != nil {
			L("reread recursion on dep %q failed; err %v", dep, err)
			return
		}

		depEnt, ok := self.Entries[dep]
		if !ok {
			L("reread recursion: dep entry not found: %q", dep)
			err = fmt.Errorf("TemplateCache.reread recursion: dep entry not found: %q", dep)
			return
		}

		depTree := depEnt.Tree
		_, err = htmlTmpl.AddParseTree(depTree.Name, depTree)
		if err != nil {
			L("reread failed to merge DEP: %q, %q, %q", dep, depEnt, err)
			return
		}
	}

	self.Entries[templateName] = TemplateEnt{
		Template: htmlTmpl,
		Tree:     tree,
		fi:       fi,
		deps:     deps,
	}

	return nil
}

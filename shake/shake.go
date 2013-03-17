package shake

import (
	"errors"
	"log"
	"reflect"
)

// argh; can't map []byte because == isn't defined...
type Key string

var ErrNoMatchingRule = errors.New("shake: no matching rule.")

type Rule interface {
	Matches(key Key) bool
	Make(key Key, rules *RuleSet) (Result, error)
}

type State struct {
	Result
	Deps []State
}

type Result struct {
	Key     Key
	Epoch   int64
	Changed bool
	Type    reflect.Type
	Value   interface{}
}

// RuleSet is basically a parser.
type RuleSet struct {
	Rules []Rule
	State map[Key]State
}

func (self *RuleSet) Make(key Key) (Result, error) {
	log.Printf("Rules.Answer(): key: %s", key)
	for _, rule := range self.Rules {
		if rule.Matches(key) {
			return rule.Make(key, self)
		}
	}
	return Result{}, ErrNoMatchingRule
}

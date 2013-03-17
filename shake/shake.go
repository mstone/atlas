package shake

import (
	"errors"
	"fmt"
	"log"
	"reflect"
)

// argh; can't map []byte because == isn't defined...
type Key string

var ErrNoMatchingRule = errors.New("shake: no matching rule.")

type OutOfDateError struct {
	Key Key
}

func (self *OutOfDateError) Error() string {
	return fmt.Sprintf("shake: old dep: key %s", self.Key)
}

type ForgottenDepError struct {
	Key Key
}

func (self *ForgottenDepError) Error() string {
	return fmt.Sprintf("shake: forgot dep: key %s", self.Key)
}

type BadCookieError struct {
	Key Key
}

func (self *BadCookieError) Error() string {
	return fmt.Sprintf("shake: bad cookie: key %s", self.Key)
}

func IsOutOfDate(err error) bool {
	_, ok := err.(*OutOfDateError)
	return ok
}

type Rule interface {
	Matches(key Key) bool
	Make(key Key, rules *RuleSet) (Result, error)
	Validate(key Key, cookie interface{}) error
}

type Dep struct {
	Key    Key
	Cookie interface{}
}

type Result struct {
	Key     Key
	Epoch   int64
	Changed bool
	Type    reflect.Type
	Value   interface{}
	Rule    Rule
	Deps    []Result
	Cookie  interface{}
}

func (self *Result) Validate(rules *RuleSet) error {
	for _, dep := range self.Deps {
		if err := dep.Validate(rules); err != nil {
			return err
		}
	}
	return self.Rule.Validate(self.Key, self.Cookie)
}

// RuleSet is basically a parser.
type RuleSet struct {
	Rules []Rule
	State map[Key]Result
}

func NewRuleSet() *RuleSet {
	return &RuleSet{
		Rules: nil,
		State: map[Key]Result{},
	}
}

func (self *RuleSet) Make(key Key) (Result, error) {
	log.Printf("Rules.Answer(): key: %s", key)

	if result, ok := self.State[key]; ok {
		if err := result.Validate(self); err == nil {
			res := result
			res.Changed = false
			return res, nil
		} else {
			log.Printf("shake: warning: %s", err)
		}
	}

	for _, rule := range self.Rules {
		if rule.Matches(key) {
			result, err := rule.Make(key, self)
			if err != nil {
				return Result{}, err
			}
			self.State[key] = result
			return result, nil
		}
	}
	return Result{}, ErrNoMatchingRule
}

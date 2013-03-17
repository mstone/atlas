package shake

import (
	"log"
)

// A Key is a question; i.e., a potentially buildable target.
type Key string

// A Rule is a way to get answers to a language of questions, even when those
// answers depend on potentially-stale cached answers to other questions.
type Rule interface {
	// Matches determines whether this rule can answer the given question.
	Matches(key Key) bool

	// Make causes this rule and any other rules provided in the given
	// RuleSet to attempt to answer the given question, perhaps by way of
	// other secondary questions (dependencies).
	Make(key Key, rules *RuleSet) (Result, error)

	// Validate uses the information stored in cookie to determine whether a
	// the cached response that provided the cookie is still fresh.
	Validate(key Key, cookie interface{}) error
}

// A Result is a cache entry.
type Result struct {
	Key     Key         // cache key -- the question
	Changed bool        // true if the cached value was rebuilt
	Value   interface{} // cached value -- the answer
	Cookie  interface{} // validator for cached value
	Rule    Rule        // rule to use to validate the cookie
	Deps    []Result    // deps to validate before validating the cookie
}

// Validate recursively validates all the dependent results contributing to this
// result and then uses this result's recorded Rule to validate the stored
// Cookie. Validate returns nil to indicate that the current result is fresh,
// returns an error satisfying IsOutOfDate() to indicate that the current result
// is stale, or returns other errors to indicate that validation failed.
func (self *Result) Validate(rules *RuleSet) error {
	for _, dep := range self.Deps {
		if err := dep.Validate(rules); err != nil {
			return err
		}
	}
	return self.Rule.Validate(self.Key, self.Cookie)
}

// A RuleSet is a combination of a parser and a cache.
type RuleSet struct {
	Rules []Rule
	State map[Key]Result
}

// NewRuleSet returns a pointer to a new RuleSet with no rules and an empty (but
// initialized) cache.
func NewRuleSet() *RuleSet {
	return &RuleSet{
		Rules: nil,
		State: map[Key]Result{},
	}
}

// BUG(mistone): At this time, RuleSet.Make is not goroutine-safe.

// Make attempts to use the given RuleSet to efficiently answer the given
// question (key). NOTE: at this time, Make is not goroutine-safe.
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
	return Result{}, &NoMatchingRuleError{
		Key: key,
	}
}

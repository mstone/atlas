package shake

import (
	"fmt"
)

// A NoMatchingRuleError occurs when no ask a RuleSet to answer a question (key)
// and no rule in the RuleSet successfully matches the key.
type NoMatchingRuleError struct {
	Key Key
}

func (self *NoMatchingRuleError) Error() string {
	return fmt.Sprintf("shake: no matching rule: key %s", self.Key)
}

// An OutOfDateError occurs when a Validate method determines that a Result
// Cookie is stale;, e.g., when the information contained in the cookie no
// longer matches the information provided by the external resource described by
// the cache key.
type OutOfDateError struct {
	Key Key
}

func (self *OutOfDateError) Error() string {
	return fmt.Sprintf("shake: old dep: key %s", self.Key)
}

// IsOutOfDate determines whether or not err is an OutOfDateError.
func IsOutOfDate(err error) bool {
	_, ok := err.(*OutOfDateError)
	return ok
}

// A BadCookieError signals that one of your rules was unable to typecast
// (typically its own) cookie to the expected type.
type BadCookieError struct {
	Key Key
}

func (self *BadCookieError) Error() string {
	return fmt.Sprintf("shake: bad cookie: key %s", self.Key)
}

// A BadQuestionError signals that one of your rules was unable to typecast
// (typically its own) question to the expected type.
type BadQuestionError struct {
	Key Key
}

func (self *BadQuestionError) Error() string {
	return fmt.Sprintf("shake: bad question: key %s", self.Key)
}

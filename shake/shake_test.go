package shake

import (
	"reflect"
	"testing"
)

type AlwaysRule struct{}

func (self *AlwaysRule) Matches(key Key) bool {
	return true
}

func (self *AlwaysRule) Make(key Key, rules *RuleSet) (Result, error) {
	val := "hi"
	res := Result{
		Key:     key,
		Changed: true,
		Epoch:   0,
		Type:    reflect.TypeOf(val),
		Value:   val,
	}
	return res, nil
}

func TestAlwaysKey(t *testing.T) {
	rs := RuleSet{}
	rs.Rules = append(rs.Rules, &AlwaysRule{})
	res, err := rs.Make("i.do.exit!")
	if err != nil {
		t.Fatalf("Shake failed w/ a default rule!")
	}
	t.Logf("res: %s", res)
}

func TestMissingKey(t *testing.T) {
	rs := RuleSet{}
	_, err := rs.Make("i.do.not.exist")
	if err == nil {
		t.Fatalf("Shake failed to fail!")
	}
}

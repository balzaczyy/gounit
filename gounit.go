package gounit

import (
	"fmt"
	"testing"
)

type T struct {
	delegate *testing.T
}

func wrapT(t *testing.T) *T {
	return &T{t}
}

func (c *T) Error(args ...interface{}) {
	if c.delegate != nil {
		c.delegate.Error(args)
	} else {
		fmt.Println(args)
	}
}

func (c *T) Assert(ok bool) {
	c.Assert2(ok, "assert fail")
}

func (c *T) Assert2(ok bool, msg string, args ...interface{}) {
	if !ok {
		c.Error(fmt.Sprintf(msg, args...))
	}
}

// Mimic JUnit's TestRule
type TestRule interface {
	Apply(t *testing.T, f func(t *T)) func(t *T)
}

// Mimic Junit's RuleChain
type _RuleChain struct {
	rules []TestRule
}

var testRuleChain *_RuleChain

func TestRuleChain(rule TestRule) *_RuleChain {
	testRuleChain = &_RuleChain{[]TestRule{rule}}
	return testRuleChain
}

func (rc *_RuleChain) Around(rule TestRule) *_RuleChain {
	rc.rules = append(rc.rules, rule)
	return rc
}

func (rc *_RuleChain) Apply(t *testing.T, f func(t *T)) func(t *T) {
	for _, rule := range rc.rules {
		f = rule.Apply(t, f)
	}
	return f
}

func Test(t *testing.T, f func(t *T)) {
	if testRuleChain == nil {
		f(wrapT(t))
	} else {
		testRuleChain.Apply(t, f)(wrapT(t))
	}
}

// Workaround Junit's TestRule for class
type ClassTestRule interface {
	Before() error
	After() error
}

type _ClassRuleChain struct {
	rules []ClassTestRule
}

var classRuleChain *_ClassRuleChain

func ClassRuleChain(rule ClassTestRule) *_ClassRuleChain {
	if rule == nil {
		panic("initial rule cannot be nil")
	}
	classRuleChain = &_ClassRuleChain{[]ClassTestRule{rule}}
	return classRuleChain
}

var classRuleChainErrors []error

func BeforeSuite(t *testing.T) {
	for _, rule := range classRuleChain.rules {
		if err := rule.Before(); err != nil {
			classRuleChainErrors = append(classRuleChainErrors, err)
		}
	}
}

var suiteClosers []func() error

// Registers a Closeable resource that shold be closed after the suite completes.
func CloseAfterSuite(closer func() error) {
	suiteClosers = append(suiteClosers, closer)
}

func AfterSuite(t *testing.T) {
	for i := len(classRuleChain.rules) - 1; i >= 0; i-- {
		if err := classRuleChain.rules[i].After(); err != nil {
			classRuleChainErrors = append(classRuleChainErrors, err)
		}
	}
	if len(classRuleChainErrors) > 0 {
		panic(fmt.Sprintf("Errors during afterSuite(): %v", classRuleChainErrors))
	}
	for _, closer := range suiteClosers {
		closer() // ignore error
	}
}

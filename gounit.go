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
	c.delegate.Error(args)
}

func (c *T) Assert(ok bool) {
	if !ok {
		c.Error("assert fail")
	}
}

func (c *T) Assert2(ok bool, msg string) {
	if !ok {
		c.Error(msg)
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
	testRuleChain.Apply(t, f)(wrapT(t))
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
	classRuleChain = &_ClassRuleChain{[]ClassTestRule{rule}}
	return classRuleChain
}

var classRuleChainErrors []error

func BeforeSuite(t *testing.T) {
	for _, rule := range classRuleChain.rules {
		classRuleChainErrors = append(classRuleChainErrors, rule.Before())
	}
}

var suiteClosers []func() error

// Registers a Closeable resource that shold be closed after the suite completes.
func CloseAfterSuite(closer func() error) {
	suiteClosers = append(suiteClosers, closer)
}

func AfterSuite(t *testing.T) {
	for i := len(classRuleChain.rules) - 1; i >= 0; i-- {
		classRuleChainErrors = append(classRuleChainErrors, classRuleChain.rules[i].After())
	}
	if len(classRuleChainErrors) > 0 {
		panic(fmt.Sprintf("Errors during afterSuite(): %v", classRuleChainErrors))
	}
	for _, closer := range suiteClosers {
		closer() // ignore error
	}
}

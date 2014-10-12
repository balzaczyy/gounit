package gounit

import (
	"testing"
)

func It(t *testing.T) *T {
	return &T{t}
}

type Assert struct {
	delegate *testing.T
	msg      string
	args     []interface{}
}

func (t *T) Should(msg string, args ...interface{}) *Assert {
	return &Assert{
		delegate: t.delegate,
		msg:      msg,
		args:     args,
	}
}

func (t *Assert) Verify(ok bool) {
	if !ok {
		t.delegate.Errorf(t.msg, t.args...)
	}
}

func (t *Assert) Assert(ok bool) {
	if !ok {
		t.delegate.Fatalf(t.msg, t.args...)
	}
}

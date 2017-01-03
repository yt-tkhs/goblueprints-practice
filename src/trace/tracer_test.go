package trace

import (
	"testing"
	"bytes"
)

func TestNew(t *testing.T) {
	var buf bytes.Buffer

	tracer := New(&buf)

	if tracer == nil {
		t.Error("return value from new() is nil.")
	} else {
		tracer.Trace("Hello, trace package")
		if buf.String() != "Hello, trace package\n" {
			t.Errorf("wrong value '%s' is outed", buf.String())
		}
	}
}

func TestOff(t *testing.T) {
	var silentTracer Tracer = Off()
	silentTracer.Trace("data")
}
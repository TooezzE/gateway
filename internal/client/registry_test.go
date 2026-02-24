package client

import (
	"context"
	"testing"
)

type stubInvoker struct {
	resp []byte
	err  error
}

func (s *stubInvoker) Invoke(_ context.Context, _ string, _ []byte) ([]byte, error) {
	return s.resp, s.err
}

func TestRegistry_GetKnownService(t *testing.T) {
	stub := &stubInvoker{resp: []byte("ok")}
	r := New(map[string]Invoker{"svc": stub})

	inv, err := r.Get("svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv != stub {
		t.Fatal("returned wrong invoker")
	}
}

func TestRegistry_GetUnknownService(t *testing.T) {
	r := New(map[string]Invoker{})

	_, err := r.Get("unknown")
	if err == nil {
		t.Fatal("expected error for unknown service")
	}
}

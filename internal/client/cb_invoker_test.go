package client

import (
	"context"
	"errors"
	"testing"

	"github.com/sony/gobreaker"
)

var errDownstream = errors.New("downstream error")

type failingInvoker struct{ err error }

func (f *failingInvoker) Invoke(_ context.Context, _ string, _ []byte) ([]byte, error) {
	return nil, f.err
}

func TestCBInvoker_Success(t *testing.T) {
	cb := NewCBInvoker("test", &stubInvoker{resp: []byte("pong")})

	resp, err := cb.Invoke(context.Background(), "method", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp) != "pong" {
		t.Fatalf("expected pong, got %s", resp)
	}
}

func TestCBInvoker_TripsAfter3ConsecutiveFailures(t *testing.T) {
	cb := NewCBInvoker("test", &failingInvoker{err: errDownstream})

	for i := 0; i < 3; i++ {
		_, err := cb.Invoke(context.Background(), "method", nil)
		if !errors.Is(err, errDownstream) {
			t.Fatalf("call %d: expected downstream error, got %v", i+1, err)
		}
	}

	_, err := cb.Invoke(context.Background(), "method", nil)
	if !errors.Is(err, gobreaker.ErrOpenState) {
		t.Fatalf("expected ErrOpenState after breaker trips, got %v", err)
	}
}

func TestCBInvoker_OpenStateHidesUnderlyingError(t *testing.T) {
	cb := NewCBInvoker("test", &failingInvoker{err: errDownstream})

	for i := 0; i < 3; i++ {
		cb.Invoke(context.Background(), "method", nil) //nolint:errcheck
	}

	_, err := cb.Invoke(context.Background(), "method", nil)
	if errors.Is(err, errDownstream) {
		t.Fatal("open breaker should not expose the underlying error")
	}
}

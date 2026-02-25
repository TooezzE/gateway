package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	gatewayv1 "github.com/TooezzE/contracts/gen/go/gateway/v1"
	"github.com/TooezzE/gateway/internal/client"
)

// mockPolicyClient satisfies policy.Client.
type mockPolicyClient struct {
	timeout time.Duration
	err     error
}

func (m *mockPolicyClient) GetTimeout(_ context.Context, _ string) (time.Duration, error) {
	return m.timeout, m.err
}

// slowInvoker blocks until context is done.
type slowInvoker struct{ delay time.Duration }

func (s *slowInvoker) Invoke(ctx context.Context, _ string, _ []byte) ([]byte, error) {
	select {
	case <-time.After(s.delay):
		return []byte("ok"), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func newHandler(pc *mockPolicyClient, invokers map[string]client.Invoker) *Handler {
	return New(pc, client.New(invokers))
}

func TestHandle_HappyPath(t *testing.T) {
	h := newHandler(
		&mockPolicyClient{timeout: time.Second},
		map[string]client.Invoker{
			"svc": stubInvoker("pong"),
		},
	)

	resp, err := h.Handle(context.Background(), &gatewayv1.Request{
		ServiceName: "svc",
		MethodName:  "GET",
		Payload:     []byte("ping"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(resp.Payload) != "pong" {
		t.Fatalf("expected pong, got %s", resp.Payload)
	}
}

func TestHandle_PolicyError(t *testing.T) {
	h := newHandler(
		&mockPolicyClient{err: errors.New("policy down")},
		map[string]client.Invoker{},
	)

	_, err := h.Handle(context.Background(), &gatewayv1.Request{ServiceName: "svc"})
	if err == nil {
		t.Fatal("expected error from policy client")
	}
}

func TestHandle_UnknownService(t *testing.T) {
	h := newHandler(
		&mockPolicyClient{timeout: time.Second},
		map[string]client.Invoker{},
	)

	_, err := h.Handle(context.Background(), &gatewayv1.Request{ServiceName: "unknown"})
	if err == nil {
		t.Fatal("expected error for unknown service")
	}
}

func TestHandle_InvokerError(t *testing.T) {
	invokeErr := errors.New("invoker failed")
	h := newHandler(
		&mockPolicyClient{timeout: time.Second},
		map[string]client.Invoker{
			"svc": &errInvoker{err: invokeErr},
		},
	)

	_, err := h.Handle(context.Background(), &gatewayv1.Request{ServiceName: "svc"})
	if !errors.Is(err, invokeErr) {
		t.Fatalf("expected invoker error, got %v", err)
	}
}

func TestHandle_TimeoutFromPolicyIsApplied(t *testing.T) {
	h := newHandler(
		&mockPolicyClient{timeout: 10 * time.Millisecond},
		map[string]client.Invoker{
			"svc": &slowInvoker{delay: 500 * time.Millisecond},
		},
	)

	_, err := h.Handle(context.Background(), &gatewayv1.Request{ServiceName: "svc"})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

// helpers

type fixedInvoker struct{ payload []byte }

func (f *fixedInvoker) Invoke(_ context.Context, _ string, _ []byte) ([]byte, error) {
	return f.payload, nil
}

func stubInvoker(payload string) client.Invoker {
	return &fixedInvoker{payload: []byte(payload)}
}

type errInvoker struct{ err error }

func (e *errInvoker) Invoke(_ context.Context, _ string, _ []byte) ([]byte, error) {
	return nil, e.err
}

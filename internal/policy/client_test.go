package policy

import (
	"context"
	"errors"
	"testing"
	"time"

	policyv1 "github.com/TooezzE/contracts/gen/go/policy/v1"
	"google.golang.org/grpc"
)

type mockPolicyServiceClient struct {
	resp *policyv1.GetPolicyResponse
	err  error
}

func (m *mockPolicyServiceClient) GetPolicy(_ context.Context, _ *policyv1.GetPolicyRequest, _ ...grpc.CallOption) (*policyv1.GetPolicyResponse, error) {
	return m.resp, m.err
}

func TestGetTimeout_ConvertsMsToDuration(t *testing.T) {
	c := New(&mockPolicyServiceClient{
		resp: &policyv1.GetPolicyResponse{TimeoutMs: 500},
	})

	d, err := c.GetTimeout(context.Background(), "svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 500*time.Millisecond {
		t.Fatalf("expected 500ms, got %v", d)
	}
}

func TestGetTimeout_ZeroMs(t *testing.T) {
	c := New(&mockPolicyServiceClient{
		resp: &policyv1.GetPolicyResponse{TimeoutMs: 0},
	})

	d, err := c.GetTimeout(context.Background(), "svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 0 {
		t.Fatalf("expected 0, got %v", d)
	}
}

func TestGetTimeout_PropagatesGRPCError(t *testing.T) {
	c := New(&mockPolicyServiceClient{err: errors.New("grpc unavailable")})

	_, err := c.GetTimeout(context.Background(), "svc")
	if err == nil {
		t.Fatal("expected error from gRPC client")
	}
}

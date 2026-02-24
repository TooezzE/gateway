package policy

import (
	"context"
	"time"

	policyv1 "github.com/TooezzE/contracts/gen/go/policy/v1"
)

type Client interface {
	GetTimeout(ctx context.Context, serviceName string) (time.Duration, error)
}

type client struct {
	grpc policyv1.PolicyServiceClient
}

func New(grpc policyv1.PolicyServiceClient) Client {
	return &client{grpc: grpc}
}

func (c *client) GetTimeout(ctx context.Context, serviceName string) (time.Duration, error) {
	resp, err := c.grpc.GetPolicy(ctx, &policyv1.GetPolicyRequest{
		ServiceName: serviceName,
	})
	if err != nil {
		return 0, err
	}

	return time.Duration(resp.TimeoutMs) * time.Millisecond, nil
}

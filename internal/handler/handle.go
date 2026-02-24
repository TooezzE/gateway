package handler

import (
	"context"

	gatewayv1 "github.com/TooezzE/contracts/gen/go/gateway/v1"
	"github.com/TooezzE/gateway/internal/client"
	"github.com/TooezzE/gateway/internal/policy"
)

type Handler struct {
	gatewayv1.UnimplementedGatewayServiceServer
	policyClient policy.Client
	registry     *client.ClientsRegistry
}

func New(policyClient policy.Client, registry *client.ClientsRegistry) *Handler {
	return &Handler{
		policyClient: policyClient,
		registry:     registry,
	}
}

func (h *Handler) Handle(ctx context.Context, req *gatewayv1.Request) (*gatewayv1.Response, error) {
	timeout, err := h.policyClient.GetTimeout(ctx, req.ServiceName)
	if err != nil {
		return nil, err
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	invoker, err := h.registry.Get(req.ServiceName)
	if err != nil {
		return nil, err
	}

	respBytes, err := invoker.Invoke(ctxWithTimeout, req.MethodName, req.Payload)
	if err != nil {
		return nil, err
	}

	return &gatewayv1.Response{
		Payload: respBytes,
	}, nil
}

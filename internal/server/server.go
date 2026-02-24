package server

import (
	"fmt"
	"log/slog"
	"net"

	gatewayv1 "github.com/TooezzE/contracts/gen/go/gateway/v1"
	"google.golang.org/grpc"
)

type Gateway struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(log *slog.Logger, port int) *Gateway {
	gRPCServer := grpc.NewServer()

	return &Gateway{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (g *Gateway) MustRun(handler gatewayv1.GatewayServiceServer) {
	if err := g.Run(handler); err != nil {
		panic(err)
	}
}

func (g *Gateway) Run(handler gatewayv1.GatewayServiceServer) error {
	const op = "grpcapp.Run"

	log := g.log.With(slog.String("op", op))

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", g.port))
	if err != nil {
		return fmt.Errorf("%s: %w:", op, err)
	}

	gatewayv1.RegisterGatewayServiceServer(g.gRPCServer, handler)

	log.Info("gRPC server is running", slog.String("addr", l.Addr().String()))

	if err := g.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w:", op, err)
	}

	return nil
}

func (g *Gateway) Stop() {
	const op = "grpcapp.Stop"

	g.log.With(slog.String("op", op)).Info("stopping gRPC server", slog.Int("port", g.port))

	g.gRPCServer.GracefulStop()
}

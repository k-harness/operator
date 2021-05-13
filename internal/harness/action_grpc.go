package harness

import (
	"context"

	"github.com/k-harness/operator/api/v1alpha1/models/action"
	"github.com/k-harness/operator/internal/executor/grpcexec"
)

type grpcAction struct {
	*action.GRPC
}

func NewGRPC(in *action.GRPC) ActionInterface {
	return &grpcAction{GRPC: in}
}

func (g *grpcAction) Call(ctx context.Context, request []byte) (*ActionResult, error) {
	gc := grpcexec.New()

	code, body, err := gc.Call(ctx, g.GRPC.Addr, grpcexec.Path{
		Package: g.GRPC.Package,
		Service: g.GRPC.Service,
		RPC:     g.GRPC.RPC,
	}, request)

	if err != nil {
		return nil, err
	}

	return &ActionResult{Code: code.String(), Body: body}, nil
}

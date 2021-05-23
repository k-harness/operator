package harness

import (
	"context"
	"fmt"

	"github.com/k-harness/operator/api/v1alpha1/models/action"
	executor2 "github.com/k-harness/operator/pkg/executor"
	grpcexec2 "github.com/k-harness/operator/pkg/executor/grpcexec"
)

type grpcRequest struct {
	*action.GRPC
}

func NewGRPCRequest(in *action.GRPC) RequestInterface {
	return &grpcRequest{GRPC: in}
}

func (g *grpcRequest) Call(ctx context.Context, request *executor2.Request) (*ActionResult, error) {
	gc := grpcexec2.New()

	// prepare headers
	headers := make([]string, 0, len(request.Header))
	for key, val := range request.Header {
		headers = append(headers, fmt.Sprintf("%s: %s", key, val))
	}

	code, body, err := gc.Call(ctx, g.GRPC.Addr, grpcexec2.Path{
		Package: g.GRPC.Package,
		Service: g.GRPC.Service,
		RPC:     g.GRPC.RPC,
	}, request.Body, headers)

	if err != nil {
		return nil, err
	}

	return &ActionResult{Code: code.String(), Body: body}, nil
}

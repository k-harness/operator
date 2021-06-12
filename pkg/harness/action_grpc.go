package harness

import (
	"context"
	"fmt"

	"github.com/k-harness/operator/api/v1alpha1"
	"github.com/k-harness/operator/api/v1alpha1/models/action"
	grpcexec2 "github.com/k-harness/operator/pkg/executor/grpcexec"
	"github.com/k-harness/operator/pkg/harness/stuff"
)

type grpcRequest struct {
	*action.GRPC
}

func NewGRPCRequest(in *action.GRPC) RequestInterface {
	return &grpcRequest{GRPC: in}
}

func (g *grpcRequest) Call(ctx context.Context, request *v1alpha1.Request) (*stuff.Response, error) {
	gc := grpcexec2.New()

	// prepare headers
	headers := make([]string, 0, len(request.Header))
	for key, val := range request.Header {
		headers = append(headers, fmt.Sprintf("%s: %s", key, val))
	}

	req, err := stuff.ScenarioBody(&request.Body).Get()
	if err != nil {
		return nil, fmt.Errorf("action can't exstract body: %w", err)
	}

	code, body, err := gc.Call(ctx, g.GRPC.Addr, grpcexec2.Path{
		Package: g.GRPC.Package,
		Service: g.GRPC.Service,
		RPC:     g.GRPC.RPC,
	}, req, headers)

	if err != nil {
		return nil, err
	}

	return &stuff.Response{Code: code.String(), Body: body}, nil
}

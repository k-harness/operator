package harness

import (
	"context"
	"errors"
	"fmt"

	"github.com/k-harness/operator/api/v1alpha1"
	"github.com/k-harness/operator/internal/executor"
)

var (
	ErrNoConnectionData = errors.New("no connection data")
)

func Call(ctx context.Context, c v1alpha1.Connect, r executor.Request) (*ActionResult, error) {
	if c.GRPC != nil {
		res, err := NewGRPCRequest(c.GRPC).Call(ctx, r)
		if err != nil {
			return nil, fmt.Errorf("grpc call: %w", err)
		}

		return res, nil
	}

	if c.HTTP != nil {
		res, err := NewHttpRequest(c.HTTP).Call(ctx, r)
		if err != nil {
			return nil, fmt.Errorf("http call: %w", err)
		}

		return res, nil
	}

	return nil, ErrNoConnectionData
}

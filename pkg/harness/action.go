package harness

import (
	"context"
	"errors"
	"fmt"

	"github.com/k-harness/operator/api/v1alpha1"
	"github.com/k-harness/operator/pkg/harness/stuff"

	"github.com/k-harness/operator/pkg/harness/variables"
)

var (
	ErrNoConnectionData = errors.New("no connection data")
)

type RequestInterface interface {
	Call(ctx context.Context, request *v1alpha1.Request) (*stuff.Response, error)
}

type stepRequest struct {
	Name string

	*v1alpha1.Request
	vars *variables.Store
}

func NewRequest(name string, a *v1alpha1.Request, variables *variables.Store) *stepRequest {
	return &stepRequest{Name: name, Request: a, vars: variables}
}

func (a *stepRequest) Call(ctx context.Context) (*stuff.Response, error) {
	// translate all request
	a.vars.RequestTranslate(a.Request)

	return a.Do(ctx, a.Connect)
}

func (a *stepRequest) Do(ctx context.Context, c v1alpha1.Connect) (*stuff.Response, error) {
	if c.GRPC != nil {
		res, err := NewGRPCRequest(c.GRPC).Call(ctx, a.Request)
		if err != nil {
			return nil, fmt.Errorf("grpc call: %w", err)
		}

		return res, nil
	}

	if c.HTTP != nil {
		// template path
		if c.HTTP.Path != nil {
			p := a.vars.Template(*c.HTTP.Path)
			c.HTTP.Path = &p
		}

		for k, v := range c.HTTP.Query {
			c.HTTP.Query[k] = a.vars.Template(v)
		}

		res, err := NewHttpRequest(c.HTTP).Call(ctx, a.Request)
		if err != nil {
			return nil, fmt.Errorf("http call: %w", err)
		}

		return res, nil
	}

	return nil, ErrNoConnectionData
}

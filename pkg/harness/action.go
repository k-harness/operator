package harness

import (
	"context"
	"errors"
	"fmt"

	"github.com/k-harness/operator/api/v1alpha1"
	executor2 "github.com/k-harness/operator/pkg/executor"
	"github.com/k-harness/operator/pkg/harness/stuff"

	"github.com/k-harness/operator/pkg/harness/variables"
)

var (
	ErrNoConnectionData = errors.New("no connection data")
)

type RequestInterface interface {
	Call(ctx context.Context, request *executor2.Request) (*stuff.Response, error)
}

type stepRequest struct {
	Name string

	*v1alpha1.Request
	vars *variables.Store
}

func NewRequest(name string, a *v1alpha1.Request, variables *variables.Store) *stepRequest {
	return &stepRequest{Name: name, Request: a, vars: variables}
}

// GetRequest take stora sync.Map and fill
func (a *stepRequest) GetRequest() (*executor2.Request, error) {
	body, err := stuff.ScenarioBody(&a.Request.Body).Get()
	if err != nil {
		return nil, fmt.Errorf("action can't exstract body: %w", err)
	}

	req := a.vars.TemplateBytesOrReturnWithout(body)
	headers := a.vars.TemplateMapOrReturnWhatPossible(a.Request.Header)

	return &executor2.Request{
		Body:   req,
		Type:   a.Body.Type,
		Header: headers,
	}, nil
}

func (a *stepRequest) Call(ctx context.Context) (*stuff.Response, error) {
	req, err := a.GetRequest()
	if err != nil {
		return nil, err
	}

	return a.Do(ctx, a.Connect, req)
}

func (a *stepRequest) Do(ctx context.Context, c v1alpha1.Connect, r *executor2.Request) (*stuff.Response, error) {
	if c.GRPC != nil {
		res, err := NewGRPCRequest(c.GRPC).Call(ctx, r)
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

		res, err := NewHttpRequest(c.HTTP).Call(ctx, r)
		if err != nil {
			return nil, fmt.Errorf("http call: %w", err)
		}

		return res, nil
	}

	return nil, ErrNoConnectionData
}

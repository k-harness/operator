package harness

import (
	"context"
	"errors"
	"fmt"

	"github.com/k-harness/operator/api/v1alpha1"
	executor2 "github.com/k-harness/operator/pkg/executor"
	checker2 "github.com/k-harness/operator/pkg/harness/checker"
	"github.com/k-harness/operator/pkg/harness/variables"
)

type RequestInterface interface {
	Call(ctx context.Context, request *executor2.Request) (*ActionResult, error)
}

var (
	ErrNoKey       = errors.New("provided key not exists in result")
	ErrBadJsonPath = errors.New("bad json path formula")
)

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
	body, err := checker2.Body(&a.Request.Body).Get()
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

func (a *stepRequest) Call(ctx context.Context) (*ActionResult, error) {
	req, err := a.GetRequest()
	if err != nil {
		return nil, err
	}

	return a.Do(ctx, a.Connect, req)
}

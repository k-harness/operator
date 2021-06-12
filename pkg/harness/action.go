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

type Action struct {
	Name string

	v1alpha1.Action
	vars *variables.Store
}

func NewStep(name string, a v1alpha1.Action, variables *variables.Store) *Action {
	return &Action{Name: name, Action: a, vars: variables}
}

// GetRequest take stora sync.Map and fill
func (a *Action) GetRequest() (*executor2.Request, error) {
	body, err := checker2.Body(&a.Request.Body).Get()
	if err != nil {
		return nil, fmt.Errorf("action can't exstract body: %w", err)
	}

	req := a.vars.TemplateBytesOrReturnWithout(body)
	headers := a.vars.TemplateMapOrReturnWhatPossible(a.Request.Header)

	return &executor2.Request{
		Body:   req,
		Type:   a.Request.Body.Type,
		Header: headers,
	}, nil
}

func (a *Action) Call(ctx context.Context) (*ActionResult, error) {
	req, err := a.GetRequest()
	if err != nil {
		return nil, err
	}

	return a.Do(ctx, a.Connect, req)
}

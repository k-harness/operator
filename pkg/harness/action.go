package harness

import (
	"context"
	"errors"
	"fmt"

	"github.com/k-harness/operator/api/v1alpha1"
	executor2 "github.com/k-harness/operator/pkg/executor"
	checker2 "github.com/k-harness/operator/pkg/harness/checker"
)

type RequestInterface interface {
	Call(ctx context.Context, request executor2.Request) (*ActionResult, error)
}

var (
	ErrNoKey       = errors.New("provided key not exists in result")
	ErrBadJsonPath = errors.New("bad json path formula")
)

type Action struct {
	Name string

	v1alpha1.Action
	vars map[string]string
}

func NewAction(name string, a v1alpha1.Action, variables map[string]string) *Action {
	return &Action{Name: name, Action: a, vars: variables}
}

// GetBody take stora sync.Map and fill
func (a *Action) GetBody() ([]byte, error) {
	return checker2.Body(&a.Request.Body).GetBody(a.vars)
}

func (a *Action) Call(ctx context.Context) (*ActionResult, error) {
	body, err := a.GetBody()
	if err != nil {
		return nil, fmt.Errorf("action can't exstract body: %w", err)
	}

	return Call(ctx, a.Connect, executor2.Request{
		Body:   body,
		Type:   a.Request.Body.Type,
		Header: a.Request.Header,
	})
}

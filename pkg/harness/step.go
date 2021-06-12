package harness

import (
	"context"
	"errors"
	"fmt"

	"github.com/k-harness/operator/api/v1alpha1"
	"github.com/k-harness/operator/pkg/harness/checker"
	"github.com/k-harness/operator/pkg/harness/stuff"
	"github.com/k-harness/operator/pkg/harness/variables"
)

type step struct {
	v1alpha1.Step
	v    *variables.Store
	bind map[string]string
}

func NewStep(s v1alpha1.Step, v *variables.Store) *step {
	return &step{Step: s, v: v, bind: make(map[string]string)}
}

func (s *step) Go(ctx context.Context) error {
	r, err := s.request(ctx)
	if err != nil {
		return err
	}

	if err = s.checkComplete(r); err != nil {
		return fmt.Errorf("condition error: %w", err)
	}

	return nil
}

func (s *step) request(ctx context.Context) (r *stuff.Response, err error) {
	if s.Action == nil {
		return &stuff.Response{}, nil
	}

	action := NewRequest(s.Name, s.Action.Request, s.v)

	res, err := action.Call(ctx)
	switch {
	case err == nil:
		// if condition completion is empty, we can ignore that?
	case errors.Is(err, ErrNoConnectionData) && s.Complete == nil:
	default:
		return r, fmt.Errorf("step call error: %w", err)
	}

	// only if completion is OK we're able to bind action
	for variable, jpath := range s.Action.BindResult {
		val, err := res.GetKeyValue(jpath)
		if err != nil {
			return nil, fmt.Errorf("binding result key %s err %w", variable, err)
		}

		s.v.Update(variable, val) // give availability to use value in completion condition
		s.bind[variable] = val
	}

	return res, nil
}

func (s *step) checkComplete(result *stuff.Response) error {
	if s.Complete == nil {
		return nil
	}

	for _, condition := range s.Complete.Condition {
		if condition.Response != nil {
			if err := checker.Res(condition.Response, s.v, result).
				Is(); err != nil {
				return fmt.Errorf("res check: %w", err)
			}
		}

		if condition.Variables != nil {
			if err := checker.Vars(condition.Variables, s.v).Is(); err != nil {
				return fmt.Errorf("var check: %w", err)
			}
		}
	}

	return nil
}

func (s *step) UpdateStore(store map[string]string) {
	for k, v := range s.bind {
		store[k] = v
	}
}

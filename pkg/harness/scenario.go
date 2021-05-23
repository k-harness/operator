package harness

import (
	"context"
	"errors"
	"fmt"

	"github.com/k-harness/operator/api/v1alpha1"
	checker2 "github.com/k-harness/operator/pkg/harness/checker"
	"github.com/k-harness/operator/pkg/harness/variables"
)

type scenarioProcessor struct {
	*v1alpha1.Scenario
}

func NewScenarioProcessor(item *v1alpha1.Scenario) *scenarioProcessor {
	item.Status.Of = len(item.Spec.Events)
	item.Status.Progress = sFmt(item.Status.Step, len(item.Spec.Events))
	item.Status.State = v1alpha1.Ready

	if item.Status.Variables == nil {
		item.Status.Variables = make(map[string]string)
	}

	for k, v := range item.Spec.Variables {
		item.Status.Variables[k] = v
	}

	return &scenarioProcessor{Scenario: item}
}

func (s *scenarioProcessor) Step(ctx context.Context) error {
	if len(s.Spec.Events) == 0 {
		return nil
	}

	s.Status.State = v1alpha1.InProgress
	ev := s.Spec.Events

	defer func() {
		s.Status.Progress = sFmt(s.Status.Step, len(ev))
	}()

	e := s.Spec.Events[s.Status.Step]
	if err := s.process(ctx, e); err != nil {
		s.Status.State = v1alpha1.Failed
		s.Status.Message = err.Error()

		return fmt.Errorf("%s: %w", e.Name, err)
	}

	if s.Status.Of <= s.Status.Step {
		s.Status.State = v1alpha1.Complete
		return nil
	}

	return nil
}

func (s *scenarioProcessor) process(ctx context.Context, event v1alpha1.Event) error {
	v := variables.New(s.Status.Variables, nil)
	action := NewAction(event.Name, event.Action, v)

	res, err := action.Call(ctx)
	switch {
	case err == nil:
		// if condition completion is empty, we can ignore that?
	case errors.Is(err, ErrNoConnectionData) && len(event.Complete.Condition) == 0:
	default:
		return fmt.Errorf("action call error: %w", err)
	}

	if err = s.checkComplete(event.Complete, res, v); err != nil {
		return fmt.Errorf("check completion: %w", err)
	}

	// only if completion is OK we're able to bind action
	for variable, jpath := range event.Action.BindResult {
		val, err := res.GetKeyValue(jpath)
		if err != nil {
			return fmt.Errorf("binding result key %s err %w", variable, err)
		}

		s.Status.Variables[variable] = val
	}

	return nil
}

func (s *scenarioProcessor) checkComplete(c v1alpha1.Completion, result *ActionResult, v *variables.Store) error {
	for _, condition := range c.Condition {
		if condition.Response != nil {
			if err := checker2.ResCheck(v, condition.Response).
				Is(result.Code, result.Body); err != nil {
				return err
			}
		}
	}

	s.Status.Repeat++
	if c.Repeat > 1 && c.Repeat > s.Status.Repeat {
		return nil
	}

	s.Status.Step++
	s.Status.Repeat = 0

	return nil
}

func sFmt(start, end int) string {
	return fmt.Sprintf("%d of %d", start, end)
}

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
	protected map[string]string
}

func NewScenarioProcessor(item *v1alpha1.Scenario, protected map[string]string) *scenarioProcessor {
	item.Status.Of = len(item.Spec.Events)
	item.Status.EventName = item.EventName()
	item.Status.StepName = item.StepName()

	item.Status.State = v1alpha1.Ready

	if item.Status.Variables == nil {
		item.Status.Variables = make(map[string]string)
	}

	for k, v := range item.Spec.Variables {
		item.Status.Variables[k] = v
	}

	return &scenarioProcessor{Scenario: item, protected: protected}
}

func (s *scenarioProcessor) Step(ctx context.Context) error {
	if len(s.Spec.Events) == 0 {
		s.Status.State = v1alpha1.Complete
		return nil
	}

	s.Status.State = v1alpha1.InProgress

	if err := s.process(ctx); err != nil {
		s.Status.State = v1alpha1.Failed
		return err
	}

	if s.Next() {
		s.Status.State = v1alpha1.Complete
	}

	return nil
}

func (s *scenarioProcessor) process(ctx context.Context) error {
	if len(s.Spec.Events[s.Status.Idx].Step) == 0 {
		return nil
	}

	step := s.Spec.Events[s.Status.Idx].Step[s.Status.Step]

	if err := s.step(ctx, step); err != nil {
		return err
	}

	return nil
}

func (s *scenarioProcessor) step(ctx context.Context, step v1alpha1.Step) error {
	v := variables.New(s.Status.Variables, s.protected)

	action := NewStep(step.Name, step.Action, v)

	res, err := action.Call(ctx)
	switch {
	case err == nil:
		// if condition completion is empty, we can ignore that?
	case errors.Is(err, ErrNoConnectionData) && len(step.Complete.Condition) == 0:
	default:
		return fmt.Errorf("action call error: %w", err)
	}

	if err = s.checkComplete(step.Complete, res, v); err != nil {
		return fmt.Errorf("check completion: %w", err)
	}

	// only if completion is OK we're able to bind action
	for variable, jpath := range step.Action.BindResult {
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

	return nil
}

func sFmt(start, end int) string {
	return fmt.Sprintf("%d of %d", start, end)
}

package harness

import (
	"context"

	"github.com/k-harness/operator/api/v1alpha1"
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

	ss := s.Spec.Events[s.Status.Idx].Step[s.Status.Step]
	v := variables.New(s.Status.Variables, s.protected)
	stepper := NewStep(ss, v)

	if err := stepper.Go(ctx); err != nil {
		return err
	}

	stepper.UpdateStore(s.Status.Variables)

	return nil
}

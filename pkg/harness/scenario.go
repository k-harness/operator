package harness

import (
	"context"
	"fmt"
	"sync"

	"github.com/k-harness/operator/api/v1alpha1"
	"github.com/k-harness/operator/pkg/harness/variables"
	"golang.org/x/sync/errgroup"
)

type scenarioProcessor struct {
	*v1alpha1.Scenario
	protected map[string]string
	statusMx  sync.Mutex
}

func NewScenarioProcessor(item *v1alpha1.Scenario, protected map[string]string) *scenarioProcessor {
	item.Status.Of = len(item.Spec.Events)
	item.Status.Progress = fmt.Sprintf("%s/%s", item.EventName(), item.StepName(0))
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
	e := s.Spec.Events[s.Status.Idx]

	wg := errgroup.Group{}
	for i := 0; i < e.Concurrency(); i++ {
		wg.Go(func() error {
			return s.process(ctx)
		})
	}

	if err := wg.Wait(); err != nil {
		s.Status.State = v1alpha1.Failed
		return err
	}

	if s.Next(e.Concurrency()) {
		s.Status.State = v1alpha1.Complete
	}

	return nil
}

func (s *scenarioProcessor) process(ctx context.Context) error {
	if len(s.Spec.Events[s.Status.Idx].Step) == 0 {
		return nil
	}

	v := variables.New(s.Status.Variables, s.protected)

	e := s.Spec.Events[s.Status.Idx]
	for idx, step := range e.Step {
		stepper := NewStep(step, v)

		if err := stepper.Go(ctx); err != nil {
			return fmt.Errorf("event %q step %q err: %w", s.EventName(), s.StepName(idx), err)
		}

		s.statusMx.Lock()
		stepper.UpdateStore(s.Status.Variables)
		s.statusMx.Unlock()
	}

	return nil
}

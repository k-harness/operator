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

	// should we duplicate variables to status?
	if item.Status.Variables == nil {
		item.Status.Variables = make(v1alpha1.ThreadVariables, 0)
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
	if len(e.StepVariables) > 0 && len(e.StepVariables) != e.Concurrency() {
		return fmt.Errorf("step variable not equal concurent")
	}

	wg := errgroup.Group{}
	for i := 0; i < e.Concurrency(); i++ {
		func(i int) {
			wg.Go(func() error {
				return s.process(ctx, i)
			})
		}(i)
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

func (s *scenarioProcessor) process(ctx context.Context, threadID int) error {
	if len(s.Spec.Events[s.Status.Idx].Step) == 0 {
		return nil
	}

	e := s.Spec.Events[s.Status.Idx]
	var threadVars v1alpha1.Variables
	if len(e.StepVariables) > threadID {
		threadVars = e.StepVariables[threadID]
	}

	v := variables.New(
		s.Spec.Variables, s.Status.Variables.GetOrCreate(threadID), s.protected, e.Variables, threadVars)

	for idx, step := range e.Step {
		stepper := NewStep(step.DeepCopy(), v)

		if err := stepper.Go(ctx); err != nil {
			return fmt.Errorf("event %q step %q err: %w", s.EventName(), s.StepName(idx), err)
		}

		stepper.UpdateStore(s.Status.Variables.GetOrCreate(threadID))
	}

	return nil
}

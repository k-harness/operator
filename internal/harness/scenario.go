package harness

import (
	"context"
	"fmt"

	"github.com/k-harness/operator/api/v1alpha1"
	"github.com/k-harness/operator/internal/harness/checker"
	"github.com/k-harness/operator/internal/harness/models"
	"k8s.io/klog/v2"
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

	if err := s.process(ctx, s.Spec.Events[s.Status.Step]); err != nil {
		s.Status.State = v1alpha1.Failed
		s.Status.Message = err.Error()

		return fmt.Errorf("process error: %w", err)
	}

	if s.Status.Of <= s.Status.Step {
		s.Status.State = v1alpha1.Complete
		return nil
	}

	return nil
}

func (s *scenarioProcessor) process(ctx context.Context, event v1alpha1.Event) error {
	res, err := s.action(ctx, models.NewAction(event.Name, event.Action))
	if err != nil {
		return fmt.Errorf("action %w", err)
	}

	return s.checkComplete(event.Complete.Condition, res)
}

func (s *scenarioProcessor) action(ctx context.Context, a *models.Action) (res *ActionResult, err error) {
	res = OK()

	resp, err := a.GetBody(s.Spec.Variables)
	if err != nil {
		return nil, err
	}

	if a.Connect.GRPC != nil {
		res, err = NewGRPC(a.Connect.GRPC).Call(ctx, resp)
		if err != nil {
			klog.Errorf("scenario progress with action %q grpc call error %v", a.Name, err)
			// ok=true:  we want to try again
			return nil, err
		}
	}

	for variable, jpath := range a.BindResult {
		val, err := res.GetKeyValue(jpath)
		if err != nil {
			return nil, fmt.Errorf("binding result key %s err %w", variable, err)
		}

		s.Status.Variables[variable] = val
	}

	return res, nil
}

func (s *scenarioProcessor) checkComplete(c []v1alpha1.Condition, result *ActionResult) error {
	for _, condition := range c {
		if condition.Response != nil {
			if err := checker.ResCheck(s.Status.Variables, condition.Response).
				Is(result.Code, result.Body); err != nil {
				return err
			}
		}
	}

	s.Status.Step++

	return nil
}

func sFmt(start, end int) string {
	return fmt.Sprintf("%d of %d", start, end)
}

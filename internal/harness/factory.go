package harness

import (
	"context"
	"fmt"
	"sync"

	"github.com/k-harness/operator/api/v1alpha1"
	"github.com/k-harness/operator/internal"
)

type Harness struct {
	// key: object
	// value: context.CancelFunc
	cancels sync.Map

	store sync.Map
}

func (h *Harness) Factory(root context.Context, c internal.Kube, key string, obj interface{}) error {
	// duplicate check
	if cur, ok := h.GetProcessor(key); ok {
		switch item := cur.(type) {
		case Processor:
			x := item.(*scenarioProcessor)
			if x.entity.ObjectMeta.ResourceVersion == obj.(*v1alpha1.Scenario).ObjectMeta.ResourceVersion {
				return nil
			}
		default:
			panic(">>fail")
		}
	}

	ctx, cancel := context.WithCancel(root)
	h.cancels.Store(key, cancel)

	switch item := obj.(type) {
	case *v1alpha1.Scenario:
		p := newScenarioProcessor(c, item)
		h.store.Store(key, p)

		go p.Start(ctx)

		return c.Update(item)
	default:
		panic(fmt.Errorf("upredictable object: %v[%[1]T]", item))
	}
}

func (h *Harness) GetProcessor(key string) (interface{}, bool) {
	return h.store.Load(key)
}

func (h *Harness) Stop(key string) {
	obj, ok := h.cancels.Load(key)
	if !ok {
		return
	}

	obj.(context.CancelFunc)()
}

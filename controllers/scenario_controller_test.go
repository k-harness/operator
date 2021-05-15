package controllers

import (
	"context"
	"time"

	scenariosv1alpha1 "github.com/k-harness/operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// https://book.kubebuilder.io/cronjob-tutorial/writing-tests.html
const (
	name = "test-scenario"
	ns   = "default"

	timeout  = time.Second * 10
	interval = time.Millisecond * 250
)

var _ = Context("empty scenario", func() {
	ctx := context.Background()

	key := types.NamespacedName{Name: name, Namespace: ns}

	created := &scenariosv1alpha1.Scenario{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: scenariosv1alpha1.ScenarioSpec{
			Name:        "123",
			Description: "dx",
			Events:      []scenariosv1alpha1.Event{},
			Variables:   map[string]string{},
		},
	}

	It("Should create successfully", func() {
		// Create
		Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

		By("Expecting submitted")

		sc1 := &scenariosv1alpha1.Scenario{}
		Eventually(func() bool {
			err := k8sClient.Get(ctx, key, sc1)
			if err != nil {
				return false
			}

			return sc1.Status.State != "" || sc1.Status.Progress != "" || sc1.Status.Message != ""
		}, timeout, interval).Should(BeTrue())
	})
})

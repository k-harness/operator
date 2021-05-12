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
var _ = Describe("Scenario controller", func() {
	const (
		name = "test-cronjob"
		ns   = "default"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	ctx := context.Background()

	Context("When scenario created", func() {
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
				Variables:   map[string]scenariosv1alpha1.Any{},
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
				return true
			}, timeout, interval).Should(BeTrue())
		})
	})
})

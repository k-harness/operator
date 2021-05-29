package controllers

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	scenariosv1alpha1 "github.com/k-harness/operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
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

var _ = Context("configMap scenario from yaml file", func() {
	ctx := context.Background()
	key := types.NamespacedName{Name: "example-config", Namespace: ns}
	cfgPath := filepath.Join("..", "config", "test", "configmap.yaml")

	It("Should create successfully", func() {
		By("loading from file")
		Expect(loadFixtures(cfgPath)).Should(Succeed())

		By("should be completed")
		sc1 := &scenariosv1alpha1.Scenario{}
		Eventually(func() bool {
			err := k8sClient.Get(ctx, key, sc1)
			return err != nil || sc1.Status.State == scenariosv1alpha1.Complete
		}, timeout, interval).Should(BeTrue())

		Expect(sc1.Status.Variables).ShouldNot(BeEmpty())
		By("not save secret config map into variable")
		Expect(sc1.Status.Variables).ShouldNot(ConsistOf("key1", "key2"))
	})
})

var _ = Context("secret + scenario from yaml file", func() {
	ctx := context.Background()
	key := types.NamespacedName{Name: "example-secret", Namespace: ns}
	cfgPath := filepath.Join("..", "config", "test", "secret.yaml")

	It("Should create successfully", func() {
		By("loading from file")
		Expect(loadFixtures(cfgPath)).Should(Succeed())

		By("should be completed")
		sc1 := &scenariosv1alpha1.Scenario{}
		Eventually(func() bool {
			err := k8sClient.Get(ctx, key, sc1)
			return err != nil || sc1.Status.State == scenariosv1alpha1.Complete
		}, timeout, interval).Should(BeTrue())

		Expect(sc1.Status.Variables).ShouldNot(BeEmpty())
		By("not save secret into variable")
		Expect(sc1.Status.Variables).ShouldNot(ConsistOf("key1", "key2"))

		By("ToDo: how check that protected variable available in variable system?")
	})
})

// loadFixtures yaml file and create into test cluster
func loadFixtures(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("file: %q open read: %w", path, err)
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("file: %q  read: %w", path, err)
	}

	reader := yaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(data)))
	for {
		// Read document
		doc, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			return fmt.Errorf("file: %q  unmarshal reader: %w", path, err)
		}

		item := &unstructured.Unstructured{}
		if err = yaml.Unmarshal(doc, item); err != nil {
			return fmt.Errorf("file: %q  unmarshal doc: %w", path, err)
		}

		if err = k8sClient.Create(context.Background(), item); err != nil {
			return fmt.Errorf("file: %q  item: %q k8s create: %w", path, item.GetName(), err)
		}
	}

	return nil
}

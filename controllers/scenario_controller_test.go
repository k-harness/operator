package controllers

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	scenariosv1alpha1 "github.com/k-harness/operator/api/v1alpha1"
	httpexec2 "github.com/k-harness/operator/pkg/executor/httpexec"
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

			return sc1.Status.State != "" || sc1.Status.Idx > 0
		}, timeout, interval).Should(BeTrue())
	})
})

var _ = Context("configMap scenario from yaml file", func() {
	ctx := context.Background()
	key := types.NamespacedName{Name: "example-config", Namespace: ns}
	key1 := types.NamespacedName{Name: "example-config-fail", Namespace: ns}
	key2 := types.NamespacedName{Name: "example-config-fail2", Namespace: ns}

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

		//Expect(sc1.Status.Variables.GetOrCreate(0)).ShouldNot(BeEmpty())
		By("not save secret config map into variable")
		Expect(sc1.Status.Variables).ShouldNot(ConsistOf("key1", "key2"))

		By("fail as completion scenario not contain required value")
		Eventually(func() bool {
			err := k8sClient.Get(ctx, key1, sc1)
			fmt.Println(sc1.Status)
			return err != nil || sc1.Status.State == scenariosv1alpha1.Failed
		}, timeout, interval).Should(BeTrue())

		By("fail as completion scenario not equal required value")
		Eventually(func() bool {
			err := k8sClient.Get(ctx, key2, sc1)
			return err != nil || sc1.Status.State == scenariosv1alpha1.Failed
		}, timeout, interval).Should(BeTrue())
	})
})

var _ = Describe("", func() {
	var (
		l   net.Listener
		srv *http.Server
	)

	BeforeEach(func() {
		By("prepare http mock echo server")

		var err error
		l, srv, err = httpexec2.CreateMockServer(&httpexec2.Fixture{
			Addr: ":8888"})

		Expect(err).ShouldNot(HaveOccurred())

		Eventually(func() bool {
			_, err = http.Get("http://127.0.0.1:8888/echo")
			return err == nil

		}, timeout, interval).Should(BeTrue())
	})

	AfterEach(func() {
		By("close mock server")

		_ = l.Close()
		_ = srv.Close()
	})

	var _ = Context("concurrency", func() {
		ctx := context.Background()
		key := types.NamespacedName{Name: "concurrency", Namespace: ns}

		It("with step_variables > 0", func() {
			// from manifest
			const concurrency = 5
			const repeat = 30

			cfgPath := filepath.Join("..", "config", "test", "concurrency.yaml")
			By("loading from file")
			Expect(loadFixtures(cfgPath)).Should(Succeed())

			sc1 := &scenariosv1alpha1.Scenario{}
			By("1. step variable for every concurrency thread use own variable and send it to echo server with key ping")
			By("2. bind response to thread variable PING")
			By("3. next step use variable PING to send it in key pong to echo server")
			By("4. bind response to thread variable PONG")

			When("processor handle it", func() {
				By("check progress")
				Eventually(func() bool {
					err := k8sClient.Get(ctx, key, sc1)

					By("threaded variable status should be equal to concurrency number")
					return err == nil && len(sc1.Status.Variables) == concurrency &&
						sc1.Status.State == scenariosv1alpha1.Complete
				}, timeout, interval).Should(BeTrue())
			})

			By("every thread variable contains PING/PONG vars with values equal to their thread number")
			for i, variable := range sc1.Status.Variables {
				val := fmt.Sprintf("%d", i)
				Expect(variable).Should(HaveKeyWithValue("PONG", val),
					HaveKeyWithValue("PING", val))
			}
		})
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

		//Expect(sc1.Status.Variables.GetOrCreate(0)).ShouldNot(BeEmpty())
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

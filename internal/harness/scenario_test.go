package harness_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/k-harness/operator/api/v1alpha1"
	"github.com/k-harness/operator/api/v1alpha1/models/action"
	"github.com/k-harness/operator/internal/executor/grpcexec"
	"github.com/k-harness/operator/internal/executor/httpexec"
	"github.com/k-harness/operator/internal/harness"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

var _ = Describe("scenario coverage", func() {
	_ = Context("event with 2 events. 1 repeat == 2, second = 1", func() {
		It("in should increase status.step only after exceed repeat counter", func() {
			By("prepare simple asset with repeat completion")

			item := &v1alpha1.Scenario{Spec: v1alpha1.ScenarioSpec{
				Events: []v1alpha1.Event{
					{Complete: v1alpha1.Completion{Repeat: 2}},
					{Complete: v1alpha1.Completion{Repeat: 1}},
				},
			}}

			By("creating processor")
			processor := harness.NewScenarioProcessor(item)
			Expect(processor).ShouldNot(BeNil())

			By("run 1 step call")
			Expect(processor.Step(context.Background())).ShouldNot(HaveOccurred())

			By("still should be in 0 stage", func() {
				Expect(item.Status.Step).Should(Equal(0))
			})

			By("run 2 step call")
			Expect(processor.Step(context.Background())).ShouldNot(HaveOccurred())

			By("should shift to next stage", func() {
				Expect(item.Status.Step).Should(Equal(1))

				By("reseting repeat status counter")
				Expect(item.Status.Repeat).Should(Equal(0))
			})

			By("run 3 step call")
			Expect(processor.Step(context.Background())).ShouldNot(HaveOccurred())

			By("it should finish all steps")
			Expect(item.Status.State).Should(Equal(v1alpha1.Complete))
		})
	})

	_ = Context("http call with evn and binding", func() {
		It("should send http and save bind request to status variables store", func() {
			token := uuid.New().String()

			fx := &httpexec.Fixture{
				Res:    map[string]string{"token": token},
				Status: http.StatusOK,
			}

			By("prepare http mock server which send token in message")
			l, srv, err := httpexec.CreateMockServer(fx)

			Expect(err).ShouldNot(HaveOccurred())

			defer func() {
				_ = l.Close()
				_ = srv.Close()
			}()

			By("prepare scenario asset")
			actionZero := v1alpha1.Action{
				Request: v1alpha1.Request{
					Header: map[string]string{
						"Secret": "my-secret",
					},
					Body: v1alpha1.Body{
						Type: "json", // under hood clien send context-type application/json
						Row:  `{"id": "{{ uuid }}", "client": "{{.CLIENT_ID}}", "rnd_str": "{{ rnd_str 32 }}"}`,
					},
				},
				Connect: v1alpha1.Connect{
					HTTP: &action.HTTP{
						Addr:   fmt.Sprintf("http://%s", l.Addr().String()),
						Method: http.MethodPost,
						Path:   strRef("/auth"),
					},
				},
				BindResult: map[string]string{
					"TOKEN": "{.token}",
				},
			}
			item := &v1alpha1.Scenario{Spec: v1alpha1.ScenarioSpec{
				Events: []v1alpha1.Event{
					{
						Action: actionZero,
						Complete: v1alpha1.Completion{
							Condition: []v1alpha1.Condition{{
								Response: &v1alpha1.ConditionResponse{
									Status: "200",
								}},
							},
						},
					},
				},
				Variables: map[string]string{
					"CLIENT_ID": "123",
				},
			}}

			By("creating processor")
			processor := harness.NewScenarioProcessor(item)
			Expect(processor).ShouldNot(BeNil())

			By("run step")
			err = processor.Step(context.Background())
			Expect(err).ShouldNot(HaveOccurred())

			By("expect step will increased")
			Expect(item.Status.Step).Should(Equal(1))

			By("expect mock server took correct header what was described in request headers")
			Expect(fx.RequestAccepted.Headers).Should(
				HaveKeyWithValue("Secret", "my-secret"))

			By("expect mock server took correct Content-Type header because of json request.body.type")
			Expect(fx.RequestAccepted.Headers).Should(
				HaveKeyWithValue("Content-Type", "application/json"),
			)

			By("expect server got body with CLIENT_ID value")
			Expect(fx.RequestAccepted.BodyMap).Should(HaveKeyWithValue("client", "123"))

			By("expect server got body with id value which is uuid function")
			Expect(fx.RequestAccepted.BodyMap).Should(HaveKey("id"))
			Expect(fx.RequestAccepted.BodyMap["id"]).To(HaveLen(36))

			By("check function rnd_str with 32 len")
			Expect(fx.RequestAccepted.BodyMap).Should(HaveKey("rnd_str"))
			Expect(fx.RequestAccepted.BodyMap["rnd_str"]).To(HaveLen(32))

			By("expecting token appears in variables store")
			Expect(item.Status.Variables).Should(ConsistOf(token, "123"))
		})
	})

	_ = Context("grpc call with evn and binding", func() {
		It("should send grpc and save bind request to status variables store", func() {
			By("prepare grpc mock server which send token in message")
			token := uuid.New().String()
			fx := grpcexec.Fixture{
				Res: &pb.HelloReply{Message: token},
				CB:  func(request *pb.HelloRequest) {},
			}

			l, srv := grpcexec.CreateMockServer(fx)
			defer func() {
				srv.Stop()
				_ = l.Close()
			}()

			By("prepare scenario asset")
			item := &v1alpha1.Scenario{Spec: v1alpha1.ScenarioSpec{
				Events: []v1alpha1.Event{
					{
						Action: v1alpha1.Action{
							Request: v1alpha1.Request{},
							Connect: v1alpha1.Connect{
								GRPC: &action.GRPC{
									Addr:    l.Addr().String(),
									Package: "helloworld",
									Service: "Greeter",
									RPC:     "SayHello",
								},
							},
							BindResult: map[string]string{
								"TOKEN": "{.message}",
							},
						},
						Complete: v1alpha1.Completion{
							Condition: []v1alpha1.Condition{{
								Response: &v1alpha1.ConditionResponse{
									Status: "OK",
								}},
							},
						},
					},
				},
				Variables: map[string]string{
					"CLIENT_ID": "123",
				},
			}}

			By("creating processor")
			processor := harness.NewScenarioProcessor(item)
			Expect(processor).ShouldNot(BeNil())

			By("run step")
			err := processor.Step(context.Background())
			Expect(err).ShouldNot(HaveOccurred())

			By("expecting token appears in variables store")
			Expect(item.Status.Variables).Should(ConsistOf(token, "123"))
		})
	})
})

func strRef(in string) *string {
	return &in
}

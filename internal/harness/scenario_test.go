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
						Row:  `{"id": "{{ uuid }}", "client": "{{.CLIENT_ID}}"}`,
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

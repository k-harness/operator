/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/k-harness/operator/internal/harness"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	scenariosv1alpha1 "github.com/k-harness/operator/api/v1alpha1"
)

// ScenarioReconciler reconciles a Scenario object
type ScenarioReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	harness harness.Harness
}

//+kubebuilder:rbac:groups=scenarios.karness.io,resources=scenarios,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=scenarios.karness.io,resources=scenarios/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=scenarios.karness.io,resources=scenarios/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Scenario object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *ScenarioReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.WithValues("scenario", req.NamespacedName, "name", req.Name, "ns", req.Namespace).Info("")

	// your logic here
	s := &scenariosv1alpha1.Scenario{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, s); err != nil {
		return ctrl.Result{RequeueAfter: 5 * time.Second}, err
	}

	//fmt.Printf(">>>>>> %+v", s)
	if err := r.harness.Factory(ctx, r, req.Name, s); err != nil {
		r.Log.Error(err, ">>>>>>>")
	}

	return ctrl.Result{}, nil
}

func (r *ScenarioReconciler) Update(item *scenariosv1alpha1.Scenario) error {
	r.Log.WithValues("name", item.Name, "ns", item.Namespace, "XXX", ">>>>>>> UPDATE")

	return r.Client.Update(context.TODO(), item)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScenarioReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&scenariosv1alpha1.Scenario{}).
		Complete(r)
}

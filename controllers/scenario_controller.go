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
	"fmt"
	"time"

	"github.com/go-logr/logr"
	scenariosv1alpha1 "github.com/k-harness/operator/api/v1alpha1"
	harness2 "github.com/k-harness/operator/pkg/harness"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ScenarioReconciler reconciles a Scenario object
type ScenarioReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Log      logr.Logger
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.karness.io,resources=scenarios,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.karness.io,resources=scenarios/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.karness.io,resources=scenarios/finalizers,verbs=update

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
	item := &scenariosv1alpha1.Scenario{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, item); err != nil {
		return ctrl.Result{RequeueAfter: 5 * time.Second}, err
	}

	if item.IsBeingDeleted() ||
		item.Status.State == scenariosv1alpha1.Complete {
		return ctrl.Result{}, nil
	}

	step := item.EventName()

	//
	protected, err := r.loadConfig(ctx, item.Spec.FromConfigMap, req.Namespace)
	if err != nil {
		r.Log.Error(err, "load config map", "step", step)
		r.Recorder.Event(item, corev1.EventTypeWarning, "load config map", err.Error())
	}

	protectedSecret, err := r.loadSecret(ctx, item.Spec.FromSecret, req.Namespace)
	if err != nil {
		r.Log.Error(err, "load secret", "step", step)
		r.Recorder.Event(item, corev1.EventTypeWarning, "load secret", err.Error())
	}

	for k, v := range protectedSecret {
		protected[k] = v
	}

	if err := harness2.NewScenarioProcessor(item, protected).Step(ctx); err != nil {
		r.Log.Error(err, "scenario process",
			"step", step, "status", item.Status, "meta", item.TypeMeta, "obg-meta", item.ObjectMeta)

		r.Recorder.Event(item, corev1.EventTypeWarning, "processor start",
			fmt.Sprintf("event: %q error: %s", step, err.Error()))

		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	r.Recorder.Event(item, corev1.EventTypeNormal, "step complete", step)
	r.Log.Info("Complete step", "step", item.Status.Idx, "of", item.Status.Of,
		"variables", item.Status.Variables,
		"state", item.Status.State, "repeat", item.Status.Repeat,
		"event", step,
	)

	// ToDo: crd:v1beta1 and v1 has different flow for saving
	// for v1beta1 we should call r.Update method
	// there as for v1 we should call special method r.Status().Update
	if err := r.Status().Update(ctx, item.DeepCopy()); err != nil {
		r.Log.Error(err, "status update error", "event", step,
			"status", item.Status, "meta", item.TypeMeta, "obg-meta", item.ObjectMeta)

		r.Recorder.Event(item, corev1.EventTypeWarning, "processor update",
			fmt.Sprintf("event: %q error: %s", item.EventName(), err.Error()))
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScenarioReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&scenariosv1alpha1.Scenario{}).
		Complete(r)
}

func (r *ScenarioReconciler) loadConfig(
	ctx context.Context,
	list []scenariosv1alpha1.NamespacedName,
	defaultNS string,
) (map[string]string, error) {
	protected := make(map[string]string)

	for _, v := range list {
		cm := corev1.ConfigMap{}
		ns := v.Namespace
		if ns == "" {
			ns = defaultNS
		}

		if err := r.Get(ctx, types.NamespacedName{Namespace: ns, Name: v.Name}, &cm); err != nil {
			return nil, fmt.Errorf("load %q error: %w", v.Name, err)
		}

		for key, bytes := range cm.Data {
			protected[key] = bytes
			r.Log.Info("load config", "key", key, "val", bytes)
		}
	}

	return protected, nil
}

func (r *ScenarioReconciler) loadSecret(
	ctx context.Context,
	list []scenariosv1alpha1.NamespacedName,
	defaultNS string,
) (map[string]string, error) {
	protected := make(map[string]string)

	for _, v := range list {
		cm := corev1.Secret{}
		ns := v.Namespace
		if ns == "" {
			ns = defaultNS
		}

		if err := r.Get(ctx, types.NamespacedName{Namespace: ns, Name: v.Name}, &cm); err != nil {
			return nil, fmt.Errorf("load %q error: %w", v.Name, err)
		}

		for key, bytes := range cm.Data {
			protected[key] = string(bytes)
			r.Log.Info("load secret", "key", key, "val", string(bytes))
		}
	}

	return protected, nil
}

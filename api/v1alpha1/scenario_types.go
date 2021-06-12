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

package v1alpha1

import (
	"github.com/k-harness/operator/api/v1alpha1/models/action"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type State string

const (
	Ready      State = "READY"
	InProgress State = "IN_PROGRESS"
	Complete   State = "COMPLETE"
	Failed     State = "FAILED"
)

// ScenarioSpec defines the desired state of Scenario
type ScenarioSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Scenario. Edit scenario_types.go to remove/update
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	Events []Event `json:"events"`

	Variables     map[string]string `json:"variables,omitempty"`
	FromSecret    []NamespacedName  `json:"from_secret,omitempty"`
	FromConfigMap []NamespacedName  `json:"from_config_map,omitempty"`
}

// ScenarioStatus defines the observed state of Scenario
//
// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
// Important: Run "make" to regenerate code after modifying this file
type ScenarioStatus struct {
	// Idx current scenario in progress
	Idx int `json:"idx"`

	Step int `json:"step"`

	// Of total events in scenario list
	Of int `json:"of"`

	EventName string `json:"event_name"`

	StepName string `json:"step_name"`

	// Count of repeat current state
	Repeat int `json:"repeat"`

	State State `json:"state"`

	// storage
	Variables map[string]string `json:"variables"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//-kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector
//+kubebuilder:printcolumn:name="Event",type="string",JSONPath=".status.event_name",description="Event name"
//+kubebuilder:printcolumn:name="Step",type="string",JSONPath=".status.step",description="Step name"
//+kubebuilder:printcolumn:name="Idx",type="integer",JSONPath=".status.idx",description="Current execution progress"
//+kubebuilder:printcolumn:name="Of",type="integer",JSONPath=".status.of",description="Total events in queue"
//+kubebuilder:printcolumn:name="Repeat",type="integer",JSONPath=".status.repeat",description="Repeat number"
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.state",description="Status where is current progress"

// Scenario is the Schema for the scenarios API
type Scenario struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScenarioSpec   `json:"spec,omitempty"`
	Status ScenarioStatus `json:"status,omitempty"`
}

// Next shift step and event counter, returns true if complete all
func (in *Scenario) Next() bool {
	in.Status.Step++

	if in.Status.Step >= len(in.Spec.Events[in.Status.Idx].Step) {
		in.Status.Step = 0
		in.Status.Repeat++

		if in.Status.Repeat >= in.Spec.Events[in.Status.Idx].Repeat {
			in.Status.Idx++
			in.Status.Repeat = 0
		}
	}

	return in.Status.Idx >= len(in.Spec.Events)
}

func (in *Scenario) EventName() string {
	if in.Status.Idx < len(in.Spec.Events) {
		return in.Spec.Events[in.Status.Idx].Name
	}

	return ""
}

func (in *Scenario) StepName() string {
	if in.Status.Idx < len(in.Spec.Events) {
		if e := in.Spec.Events[in.Status.Idx]; in.Status.Step < len(e.Step) {
			return e.Step[in.Status.Step].Name
		}
	}

	return ""
}

func (in *Scenario) IsBeingDeleted() bool {
	return !in.ObjectMeta.DeletionTimestamp.IsZero()
}

//+kubebuilder:object:root=true

// ScenarioList contains a list of Scenario
type ScenarioList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Scenario `json:"items"`
}

//
type Event struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	// Repeat current item times
	// +kubebuilder:validation:Minimum=1
	Repeat int `json:"repeat,omitempty"`

	// ToDo: concurrency
	// Run step paralel times
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10
	Concurrent int `json:"concurrent,omitempty"`

	Step []Step `json:"step"`
}

type Step struct {
	Name string `json:"name,omitempty"`

	Action   *Action     `json:"action,omitempty"`
	Complete *Completion `json:"complete,omitempty"`
}

type Request struct {
	Header map[string]string `json:"header,omitempty"`
	Body   Body              `json:"body"`
	// Connect transport used by actor
	Connect Connect `json:"connect"`
}

type Body struct {
	// +kubebuilder:validation:Enum=json;xml
	Type string `json:"type"`

	// ToDo: validate oneOF
	KV   map[string]string `json:"kv,omitempty"`
	Byte []byte            `json:"byte,omitempty"`
	Row  string            `json:"row,omitempty"`
}

type Connect struct {
	GRPC *action.GRPC `json:"grpc,omitempty"`
	HTTP *action.HTTP `json:"http,omitempty"`
}

type Action struct {
	// makes requests via some transport protocols
	Request *Request `json:"request,omitempty"`

	// BindResult save result KV representation in global variable storage
	// This works only when result returns as JSON or maybe anything marshalable
	// Right now only JSON supposed to be
	// Key: result_key
	// Val: variable name for binding
	BindResult map[string]string `json:"bind_result,omitempty"`
}

type Any string

type Completion struct {
	Description string `json:"description,omitempty"`

	Condition []Condition `json:"condition"`
}

// Condition of complete show reason
type Condition struct {
	// Response of condition check
	Response *ConditionResponse `json:"response"`
}

// ConditionResponse contains competition condition for source
type ConditionResponse struct {
	Status string `json:"status"`
	Body   Body   `json:"body"`
}

type KV struct {
	Field []KVFieldMatch `json:"field_match"`
}

// KVFieldMatch mean that key sh
type KVFieldMatch struct {
	Key   string `json:"key"`
	Value Any    `json:"value"`
}

func init() {
	SchemeBuilder.Register(&Scenario{}, &ScenarioList{})
}

type NamespacedName struct {
	Name string `json:"name"`
	// by default use ns where scenario located
	Namespace string `json:"namespace,omitempty"`
}

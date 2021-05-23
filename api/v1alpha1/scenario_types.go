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
	Description string `json:"description"`

	Events    []Event           `json:"events"`
	Variables map[string]string `json:"variables"`
}

// ScenarioStatus defines the observed state of Scenario
type ScenarioStatus struct {
	// Step current scenario in progress
	Step int `json:"step"`

	// Of total events in scenario list
	Of int `json:"of"`

	// Count of repeat current state
	Repeat int `json:"repeat"`

	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Progress string `json:"progress"`
	State    State  `json:"state"`
	Message  string `json:"message"`

	// storage
	Variables map[string]string `json:"variables"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//-kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector
//+kubebuilder:printcolumn:name="Step",type="integer",JSONPath=".status.step",description="Current execution progress"
//+kubebuilder:printcolumn:name="Of",type="integer",JSONPath=".status.of",description="Total events in queue"
//+kubebuilder:printcolumn:name="Repeat",type="integer",JSONPath=".status.repeat",description="Repeat number"
//+kubebuilder:printcolumn:name="Progress",type="string",JSONPath=".status.progress",description="Progress of scenario"
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.state",description="Status where is current progress"
//+kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.message",description="Information related to some issues"

// Scenario is the Schema for the scenarios API
type Scenario struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScenarioSpec   `json:"spec,omitempty"`
	Status ScenarioStatus `json:"status,omitempty"`
}

func (in *Scenario) CurrentStepName() string {
	if in.Status.Step < len(in.Spec.Events) {
		return in.Spec.Events[in.Status.Step].Name
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

	Action   Action     `json:"action"`
	Complete Completion `json:"complete,omitempty"`
}

type Request struct {
	Header map[string]string `json:"header,omitempty"`
	Body   Body              `json:"body"`
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
	// request containment
	Request Request `json:"request"`

	// Connect transport used by actor
	Connect Connect `json:"connect"`

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

	// Repeat current item times
	// +kubebuilder:validation:Minimum=1
	Repeat int `json:"repeat,omitempty"`

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

// KVFields mean that key sh
type KVFieldMatch struct {
	Key   string `json:"key"`
	Value Any    `json:"value"`
}

func init() {
	SchemeBuilder.Register(&Scenario{}, &ScenarioList{})
}

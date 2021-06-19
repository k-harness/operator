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
	Description string `json:"description,omitempty"`

	Events []Event `json:"events"`

	Variables     Variables        `json:"variables,omitempty"`
	FromSecret    []NamespacedName `json:"from_secret,omitempty"`
	FromConfigMap []NamespacedName `json:"from_config_map,omitempty"`
}

// Variables is simple key/value storage
type Variables map[string]string

// ThreadVariables represent values per thread used only
type ThreadVariables []Variables

// ScenarioStatus defines the observed state of Scenario
//
// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
// Important: Run "make" to regenerate code after modifying this file
type ScenarioStatus struct {
	// Idx current scenario in progress
	Idx int `json:"idx"`

	// Of total events in scenario list
	Of int `json:"of"`

	Progress string `json:"progress"`

	// Count of repeat current state
	Repeat int `json:"repeat"`

	State State `json:"state"`

	// storage based on concurrency
	Variables ThreadVariables `json:"variables"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//-kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector
//+kubebuilder:printcolumn:name="Current Step",type="string",JSONPath=".status.progress",description="Event/Step name"
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
func (in *Scenario) Next(i int) bool {
	in.Status.Repeat += i

	if in.Status.Repeat >= in.Spec.Events[in.Status.Idx].Repeat {
		in.Status.Idx++
		in.Status.Repeat = 0
	}

	return in.Status.Idx >= len(in.Spec.Events)
}

func (in *Scenario) EventName() string {
	if in.Status.Idx < len(in.Spec.Events) {
		return in.Spec.Events[in.Status.Idx].Name
	}

	return ""
}

func (in *Scenario) StepName(idx int) string {
	if in.Status.Idx < len(in.Spec.Events) {
		if e := in.Spec.Events[in.Status.Idx]; idx < len(e.Step) {
			return e.Step[idx].Name
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

// Event ...
type Event struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	// variables used in current event within all steps
	// variables common for all steps
	Variables Variables `json:"variables,omitempty"`

	// variables unique for every step
	StepVariables []Variables `json:"step_variables,omitempty"`

	// Repeat current item times
	// +kubebuilder:validation:Minimum=1
	Repeat int `json:"repeat,omitempty"`

	// Run step parallel times
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=50
	Concurrent int `json:"concurrent,omitempty"`

	// sequence of operation performed by one worked
	Step []Step `json:"step"`
}

func (in Event) Concurrency() int {
	if in.Concurrent > 1 {
		return in.Concurrent
	}

	return 1
}

type Step struct {
	Name string `json:"name,omitempty"`

	Action   *Action     `json:"action,omitempty"`
	Complete *Completion `json:"complete,omitempty"`
}

type Request struct {
	Header map[string]string `json:"header,omitempty"`
	Body   Body              `json:"body,omitempty"`

	// Connect transport used by actor
	Connect Connect `json:"connect"`
}

type Body struct {
	// +kubebuilder:validation:Enum=json;xml
	Type string `json:"type"`

	// ToDo: validate oneOF
	// http form required KV only
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
	Response *ConditionResponse `json:"response,omitempty"`

	Variables *ConditionVariables `json:"variables,omitempty"`
}

// ConditionResponse contains competition condition for source
type ConditionResponse struct {
	Status   *string                   `json:"status,omitempty"`
	Body     *Body                     `json:"body,omitempty"`
	JSONPath map[string]VarOptionCheck `json:"JSONPath,omitempty"`
}

type Operator string

const (
	Required Operator = "required"
	Equal    Operator = "equal"
)

type VarOptionCheck struct {
	// +kubebuilder:validation:Enum=required;equal
	// how should we check value or key provided
	Operator Operator `json:"operator"`

	Value string `json:"value,omitempty"`
}

type ConditionVariables struct {
	// key of map represent storage key
	KV map[string]VarOptionCheck `json:"kv,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Scenario{}, &ScenarioList{})
}

type NamespacedName struct {
	Name string `json:"name"`

	// by default use ns where scenario located
	Namespace string `json:"namespace,omitempty"`
}

func (in *ThreadVariables) GetOrCreate(threadID int) Variables {
	if *in == nil {
		*in = make(ThreadVariables, threadID)
	}

	if len(*in) <= threadID {
		*in = append(*in, make(Variables))
		return in.GetOrDefault(threadID)
	}

	if (*in)[threadID] == nil {
		(*in)[threadID] = make(Variables)
	}

	return (*in)[threadID]
}

func (in ThreadVariables) GetOrDefault(threadID int) Variables {
	if len(in) > threadID {
		return in[threadID]
	}

	return in.GetOrCreate(0)
}

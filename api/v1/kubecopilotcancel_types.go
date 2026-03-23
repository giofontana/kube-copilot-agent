/*
Copyright 2026.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KubeCopilotCancelSpec defines the desired state of KubeCopilotCancel
type KubeCopilotCancelSpec struct {
	// AgentRef is the name of the KubeCopilotAgent in the same namespace.
	// +required
	AgentRef string `json:"agentRef"`

	// SendRef is the name of the KubeCopilotSend to cancel.
	// +required
	SendRef string `json:"sendRef"`
}

// KubeCopilotCancelStatus defines the observed state of KubeCopilotCancel.
type KubeCopilotCancelStatus struct {
	// Phase: Pending, Cancelled, Error.
	// +optional
	Phase string `json:"phase,omitempty"`

	// ErrorMessage contains error details when Phase is Error.
	// +optional
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="SendRef",type="string",JSONPath=".spec.sendRef"
// +kubebuilder:printcolumn:name="AgentRef",type="string",JSONPath=".spec.agentRef"

// KubeCopilotCancel cancels an in-flight KubeCopilotSend request.
type KubeCopilotCancel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubeCopilotCancelSpec   `json:"spec"`
	Status KubeCopilotCancelStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KubeCopilotCancelList contains a list of KubeCopilotCancel
type KubeCopilotCancelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubeCopilotCancel `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubeCopilotCancel{}, &KubeCopilotCancelList{})
}

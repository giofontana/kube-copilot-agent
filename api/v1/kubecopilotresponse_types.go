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

// KubeCopilotResponseSpec defines the content of a KubeCopilotResponse.
// This object is immutable once created by the operator webhook.
type KubeCopilotResponseSpec struct {
	// AgentRef is the name of the KubeCopilotAgent that produced this response.
	// +required
	AgentRef string `json:"agentRef"`

	// SessionID is the conversation session this response belongs to.
	// +optional
	SessionID string `json:"sessionID,omitempty"`

	// Prompt is the user message that triggered this response.
	// +required
	Prompt string `json:"prompt"`

	// Response is the agent's reply.
	// +required
	Response string `json:"response"`

	// SendRef is the name of the KubeCopilotSend that initiated this request.
	// +optional
	SendRef string `json:"sendRef,omitempty"`
}

// KubeCopilotResponseStatus defines the observed state of KubeCopilotResponse.
type KubeCopilotResponseStatus struct {
	// CreatedAt is the timestamp when this response was recorded.
	// +optional
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="AgentRef",type="string",JSONPath=".spec.agentRef"
// +kubebuilder:printcolumn:name="SessionID",type="string",JSONPath=".spec.sessionID"
// +kubebuilder:printcolumn:name="SendRef",type="string",JSONPath=".spec.sendRef"

// KubeCopilotResponse is the Schema for the kubecopilotresponses API
type KubeCopilotResponse struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubeCopilotResponseSpec   `json:"spec"`
	Status KubeCopilotResponseStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KubeCopilotResponseList contains a list of KubeCopilotResponse
type KubeCopilotResponseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubeCopilotResponse `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubeCopilotResponse{}, &KubeCopilotResponseList{})
}

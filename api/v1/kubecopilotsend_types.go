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

// KubeCopilotSendSpec defines the desired state of KubeCopilotSend
type KubeCopilotSendSpec struct {
	// AgentRef is the name of the KubeCopilotAgent in the same namespace.
	// +required
	AgentRef string `json:"agentRef"`

	// Message is the prompt to send to the copilot agent asynchronously.
	// +required
	Message string `json:"message"`

	// SessionID optionally continues an existing conversation session.
	// +optional
	SessionID string `json:"sessionID,omitempty"`
}

// KubeCopilotSendStatus defines the observed state of KubeCopilotSend.
type KubeCopilotSendStatus struct {
	// Phase: Pending, Queued, Done, Error.
	// +optional
	Phase string `json:"phase,omitempty"`

	// QueueID is the identifier returned by the agent for this queued request.
	// +optional
	QueueID string `json:"queueID,omitempty"`

	// ErrorMessage contains error details when Phase is Error.
	// +optional
	ErrorMessage string `json:"errorMessage,omitempty"`

	// Conditions represent the current state of the send request.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="AgentRef",type="string",JSONPath=".spec.agentRef"
// +kubebuilder:printcolumn:name="QueueID",type="string",JSONPath=".status.queueID"

// KubeCopilotSend is the Schema for the kubecopilotsends API
type KubeCopilotSend struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubeCopilotSendSpec   `json:"spec"`
	Status KubeCopilotSendStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KubeCopilotSendList contains a list of KubeCopilotSend
type KubeCopilotSendList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubeCopilotSend `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubeCopilotSend{}, &KubeCopilotSendList{})
}

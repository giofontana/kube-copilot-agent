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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KubeCopilotMessageSpec defines the desired state of KubeCopilotMessage
type KubeCopilotMessageSpec struct {
	// AgentRef is the name of the KubeCopilotAgent in the same namespace.
	// +required
	AgentRef string `json:"agentRef"`

	// Message is the prompt/message to send to the copilot-cli agent.
	// +required
	Message string `json:"message"`

	// SessionID optionally continues an existing conversation session.
	// Leave empty to start a new session.
	// +optional
	SessionID string `json:"sessionID,omitempty"`
}

// KubeCopilotMessageStatus defines the observed state of KubeCopilotMessage.
type KubeCopilotMessageStatus struct {
	// Phase: Pending, Processing, Done, Error.
	// +optional
	Phase string `json:"phase,omitempty"`

	// Response contains the agent's reply.
	// +optional
	Response string `json:"response,omitempty"`

	// SessionID is the session identifier to use for follow-up commands.
	// +optional
	SessionID string `json:"sessionID,omitempty"`

	// ErrorMessage contains error details when Phase is Error.
	// +optional
	ErrorMessage string `json:"errorMessage,omitempty"`

	// Conditions represent the current state of the command.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="AgentRef",type="string",JSONPath=".spec.agentRef"
// +kubebuilder:printcolumn:name="SessionID",type="string",JSONPath=".status.sessionID"

// KubeCopilotMessage is the Schema for the kubecopilotmessages API
type KubeCopilotMessage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubeCopilotMessageSpec   `json:"spec"`
	Status KubeCopilotMessageStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KubeCopilotMessageList contains a list of KubeCopilotMessage
type KubeCopilotMessageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubeCopilotMessage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubeCopilotMessage{}, &KubeCopilotMessageList{})
}

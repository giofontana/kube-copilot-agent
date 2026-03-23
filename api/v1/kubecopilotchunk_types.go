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

// KubeCopilotChunkSpec defines a single streaming output chunk from the agent.
type KubeCopilotChunkSpec struct {
	// AgentRef is the name of the KubeCopilotAgent that produced this chunk.
	// +required
	AgentRef string `json:"agentRef"`

	// SessionID is the conversation session this chunk belongs to.
	// +optional
	SessionID string `json:"sessionID,omitempty"`

	// SendRef is the name of the KubeCopilotSend that initiated this stream.
	// +optional
	SendRef string `json:"sendRef,omitempty"`

	// Sequence is the ordering index of this chunk within a send request.
	// +required
	Sequence int `json:"sequence"`

	// ChunkType categorises the chunk: thinking, tool_call, tool_result, response, info, error.
	// +required
	ChunkType string `json:"chunkType"`

	// Content is the human-readable text for this chunk.
	// +required
	Content string `json:"content"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="SendRef",type="string",JSONPath=".spec.sendRef"
// +kubebuilder:printcolumn:name="Seq",type="integer",JSONPath=".spec.sequence"
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.chunkType"

// KubeCopilotChunk is the Schema for the kubecopilotchunks API
type KubeCopilotChunk struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec KubeCopilotChunkSpec `json:"spec"`
}

// +kubebuilder:object:root=true

// KubeCopilotChunkList contains a list of KubeCopilotChunk
type KubeCopilotChunkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubeCopilotChunk `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubeCopilotChunk{}, &KubeCopilotChunkList{})
}

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

// PolicyRuleType defines the type of policy rule.
// +kubebuilder:validation:Enum=deny;require-approval
type PolicyRuleType string

const (
	// PolicyRuleTypeDeny blocks the send request entirely.
	PolicyRuleTypeDeny PolicyRuleType = "deny"

	// PolicyRuleTypeRequireApproval pauses the send request until explicitly approved.
	PolicyRuleTypeRequireApproval PolicyRuleType = "require-approval"
)

// PolicyRule defines a single guardrail rule evaluated against agent send requests.
type PolicyRule struct {
	// Name is a human-readable identifier for this rule.
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Name string `json:"name"`

	// Type is the action taken when this rule matches: "deny" or "require-approval".
	// +required
	Type PolicyRuleType `json:"type"`

	// CommandPatterns is a list of glob patterns matched against the send message.
	// A rule triggers if any pattern matches. Supports "*" and "?" wildcards.
	// Examples: "kubectl delete namespace prod-*", "* --force *"
	// +optional
	// +listType=set
	CommandPatterns []string `json:"commandPatterns,omitempty"`

	// Message is the explanation shown to the user when this rule triggers.
	// +optional
	Message string `json:"message,omitempty"`
}

// KubeCopilotPolicySpec defines the desired state of KubeCopilotPolicy.
type KubeCopilotPolicySpec struct {
	// AgentRef optionally binds this policy to a specific KubeCopilotAgent.
	// When empty, the policy applies to all agents in the namespace.
	// +optional
	AgentRef string `json:"agentRef,omitempty"`

	// Rules is the list of policy rules. Deny rules are always evaluated before
	// require-approval rules (deny takes precedence regardless of order).
	// +required
	// +kubebuilder:validation:MinItems=1
	Rules []PolicyRule `json:"rules"`
}

// KubeCopilotPolicyStatus defines the observed state of KubeCopilotPolicy.
type KubeCopilotPolicyStatus struct {
	// Active indicates whether this policy is currently being enforced.
	// +optional
	Active bool `json:"active,omitempty"`

	// RuleCount is the number of rules in this policy.
	// +optional
	RuleCount int `json:"ruleCount,omitempty"`

	// Conditions represent the current state of the policy.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Active",type="boolean",JSONPath=".status.active"
// +kubebuilder:printcolumn:name="AgentRef",type="string",JSONPath=".spec.agentRef"
// +kubebuilder:printcolumn:name="Rules",type="integer",JSONPath=".status.ruleCount"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// KubeCopilotPolicy is the Schema for the kubecopilotpolicies API.
// It defines guardrails and safety constraints for AI agent operations.
type KubeCopilotPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubeCopilotPolicySpec   `json:"spec"`
	Status KubeCopilotPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KubeCopilotPolicyList contains a list of KubeCopilotPolicy
type KubeCopilotPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubeCopilotPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KubeCopilotPolicy{}, &KubeCopilotPolicyList{})
}

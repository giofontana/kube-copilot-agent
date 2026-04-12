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

package controller

import (
	"context"
	"path/filepath"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"

	agentv1 "github.com/gfontana/kube-copilot-agent/api/v1"
)

// PolicyDecision represents the result of evaluating policies against a send request.
type PolicyDecision string

const (
	// PolicyDecisionAllow means no rules matched; proceed with dispatch.
	PolicyDecisionAllow PolicyDecision = "allow"

	// PolicyDecisionDeny means a deny rule matched; block the request.
	PolicyDecisionDeny PolicyDecision = "deny"

	// PolicyDecisionRequireApproval means an approval rule matched; pause until approved.
	PolicyDecisionRequireApproval PolicyDecision = "require-approval"
)

// PolicyEvalResult contains the outcome of policy evaluation.
type PolicyEvalResult struct {
	Decision   PolicyDecision
	RuleName   string
	PolicyName string
	Message    string
}

// EvaluatePolicies checks all applicable policies against a send request.
// Deny rules always take precedence over require-approval rules.
func EvaluatePolicies(ctx context.Context, c client.Client, namespace, agentRef, message string) (*PolicyEvalResult, error) {
	policies := &agentv1.KubeCopilotPolicyList{}
	if err := c.List(ctx, policies, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	var approvalResult *PolicyEvalResult

	for i := range policies.Items {
		policy := &policies.Items[i]

		// Skip inactive policies
		if !policy.Status.Active {
			continue
		}

		// Filter by agentRef (empty agentRef matches all agents)
		if policy.Spec.AgentRef != "" && policy.Spec.AgentRef != agentRef {
			continue
		}

		for j := range policy.Spec.Rules {
			rule := &policy.Spec.Rules[j]
			if !matchesPatterns(message, rule.CommandPatterns) {
				continue
			}

			// Deny always wins immediately
			if rule.Type == agentv1.PolicyRuleTypeDeny {
				return &PolicyEvalResult{
					Decision:   PolicyDecisionDeny,
					RuleName:   rule.Name,
					PolicyName: policy.Name,
					Message:    rule.Message,
				}, nil
			}

			// Record first matching approval rule (deny can still override later)
			if rule.Type == agentv1.PolicyRuleTypeRequireApproval && approvalResult == nil {
				approvalResult = &PolicyEvalResult{
					Decision:   PolicyDecisionRequireApproval,
					RuleName:   rule.Name,
					PolicyName: policy.Name,
					Message:    rule.Message,
				}
			}
		}
	}

	if approvalResult != nil {
		return approvalResult, nil
	}

	return &PolicyEvalResult{Decision: PolicyDecisionAllow}, nil
}

// matchesPatterns checks if the message matches any of the glob patterns.
// Uses case-insensitive matching for user-friendliness.
func matchesPatterns(message string, patterns []string) bool {
	lower := strings.ToLower(message)
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(strings.ToLower(pattern), lower); matched {
			return true
		}
		// Also check if pattern appears as a substring (glob patterns are short,
		// messages can be long). This handles multi-word messages where the
		// pattern describes a command fragment.
		if containsPattern(lower, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// containsPattern checks if any contiguous word window in the message matches the glob.
func containsPattern(message, pattern string) bool {
	words := strings.Fields(message)
	patWords := strings.Fields(pattern)
	if len(patWords) == 0 || len(words) < len(patWords) {
		return false
	}
	for i := 0; i <= len(words)-len(patWords); i++ {
		window := strings.Join(words[i:i+len(patWords)], " ")
		if matched, _ := filepath.Match(pattern, window); matched {
			return true
		}
	}
	return false
}

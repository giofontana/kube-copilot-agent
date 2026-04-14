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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	agentv1 "github.com/gfontana/kube-copilot-agent/api/v1"
)

var _ = Describe("KubeCopilotPolicy Controller", func() {
	var reconciler *KubeCopilotPolicyReconciler

	BeforeEach(func() {
		reconciler = &KubeCopilotPolicyReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}
	})

	Context("When creating a valid KubeCopilotPolicy", func() {
		It("Should set Active status to true", func() {
			policy := &agentv1.KubeCopilotPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deny-policy",
					Namespace: "default",
				},
				Spec: agentv1.KubeCopilotPolicySpec{
					Rules: []agentv1.PolicyRule{
						{
							Name:            "block-delete-ns",
							Type:            agentv1.PolicyRuleTypeDeny,
							CommandPatterns: []string{"kubectl delete namespace *"},
							Message:         "Namespace deletion is not allowed",
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, policy)).To(Succeed())
			defer func() { _ = k8sClient.Delete(ctx, policy) }()

			_, err := reconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{Name: policy.Name, Namespace: policy.Namespace},
			})
			Expect(err).NotTo(HaveOccurred())

			// Re-fetch to see status
			updated := &agentv1.KubeCopilotPolicy{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: policy.Name, Namespace: policy.Namespace}, updated)).To(Succeed())
			Expect(updated.Status.Active).To(BeTrue())
			Expect(updated.Status.RuleCount).To(Equal(1))
		})
	})

	Context("When creating a policy with empty command patterns", func() {
		It("Should set Active to false with validation error", func() {
			policy := &agentv1.KubeCopilotPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-invalid-policy",
					Namespace: "default",
				},
				Spec: agentv1.KubeCopilotPolicySpec{
					Rules: []agentv1.PolicyRule{
						{
							Name: "no-patterns",
							Type: agentv1.PolicyRuleTypeDeny,
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, policy)).To(Succeed())
			defer func() { _ = k8sClient.Delete(ctx, policy) }()

			_, err := reconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{Name: policy.Name, Namespace: policy.Namespace},
			})
			Expect(err).NotTo(HaveOccurred())

			updated := &agentv1.KubeCopilotPolicy{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: policy.Name, Namespace: policy.Namespace}, updated)).To(Succeed())
			Expect(updated.Status.Active).To(BeFalse())
		})
	})

	Context("When creating a policy with agentRef", func() {
		It("Should be active and scoped to the referenced agent", func() {
			policy := &agentv1.KubeCopilotPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-scoped-policy",
					Namespace: "default",
				},
				Spec: agentv1.KubeCopilotPolicySpec{
					AgentRef: "my-agent",
					Rules: []agentv1.PolicyRule{
						{
							Name:            "block-force",
							Type:            agentv1.PolicyRuleTypeDeny,
							CommandPatterns: []string{"* --force *"},
							Message:         "Force operations are not allowed",
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, policy)).To(Succeed())
			defer func() { _ = k8sClient.Delete(ctx, policy) }()

			_, err := reconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{Name: policy.Name, Namespace: policy.Namespace},
			})
			Expect(err).NotTo(HaveOccurred())

			updated := &agentv1.KubeCopilotPolicy{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: policy.Name, Namespace: policy.Namespace}, updated)).To(Succeed())
			Expect(updated.Status.Active).To(BeTrue())
			Expect(updated.Spec.AgentRef).To(Equal("my-agent"))
		})
	})
})

var _ = Describe("Policy Evaluator", func() {
	var (
		reconciler *KubeCopilotPolicyReconciler
	)

	BeforeEach(func() {
		reconciler = &KubeCopilotPolicyReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}
	})

	Context("Deny rules", func() {
		It("Should deny a matching send message", func() {
			policy := &agentv1.KubeCopilotPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "eval-deny-policy",
					Namespace: "default",
				},
				Spec: agentv1.KubeCopilotPolicySpec{
					Rules: []agentv1.PolicyRule{
						{
							Name:            "block-delete-ns",
							Type:            agentv1.PolicyRuleTypeDeny,
							CommandPatterns: []string{"kubectl delete namespace *"},
							Message:         "Cannot delete namespaces",
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, policy)).To(Succeed())
			defer func() { _ = k8sClient.Delete(ctx, policy) }()

			// Activate the policy
			_, err := reconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{Name: policy.Name, Namespace: policy.Namespace},
			})
			Expect(err).NotTo(HaveOccurred())

			result, err := EvaluatePolicies(ctx, k8sClient, "default", "any-agent", "kubectl delete namespace production")
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Decision).To(Equal(PolicyDecisionDeny))
			Expect(result.RuleName).To(Equal("block-delete-ns"))
			Expect(result.Message).To(Equal("Cannot delete namespaces"))
		})

		It("Should allow a non-matching message", func() {
			result, err := EvaluatePolicies(ctx, k8sClient, "default", "any-agent", "kubectl get pods")
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Decision).To(Equal(PolicyDecisionAllow))
		})
	})

	Context("Require-approval rules", func() {
		It("Should require approval for matching message", func() {
			policy := &agentv1.KubeCopilotPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "eval-approval-policy",
					Namespace: "default",
				},
				Spec: agentv1.KubeCopilotPolicySpec{
					Rules: []agentv1.PolicyRule{
						{
							Name:            "approve-scale",
							Type:            agentv1.PolicyRuleTypeRequireApproval,
							CommandPatterns: []string{"kubectl scale *"},
							Message:         "Scaling requires approval",
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, policy)).To(Succeed())
			defer func() { _ = k8sClient.Delete(ctx, policy) }()

			_, err := reconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{Name: policy.Name, Namespace: policy.Namespace},
			})
			Expect(err).NotTo(HaveOccurred())

			result, err := EvaluatePolicies(ctx, k8sClient, "default", "any-agent", "kubectl scale deployment nginx --replicas=10")
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Decision).To(Equal(PolicyDecisionRequireApproval))
			Expect(result.RuleName).To(Equal("approve-scale"))
		})
	})

	Context("Deny takes precedence over require-approval", func() {
		It("Should deny even when approval rule also matches", func() {
			denyPolicy := &agentv1.KubeCopilotPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "eval-precedence-deny",
					Namespace: "default",
				},
				Spec: agentv1.KubeCopilotPolicySpec{
					Rules: []agentv1.PolicyRule{
						{
							Name:            "block-delete",
							Type:            agentv1.PolicyRuleTypeDeny,
							CommandPatterns: []string{"kubectl delete *"},
						},
					},
				},
			}

			approvalPolicy := &agentv1.KubeCopilotPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "eval-precedence-approval",
					Namespace: "default",
				},
				Spec: agentv1.KubeCopilotPolicySpec{
					Rules: []agentv1.PolicyRule{
						{
							Name:            "approve-delete",
							Type:            agentv1.PolicyRuleTypeRequireApproval,
							CommandPatterns: []string{"kubectl delete *"},
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, denyPolicy)).To(Succeed())
			Expect(k8sClient.Create(ctx, approvalPolicy)).To(Succeed())
			defer func() {
				_ = k8sClient.Delete(ctx, denyPolicy)
				_ = k8sClient.Delete(ctx, approvalPolicy)
			}()

			for _, name := range []string{"eval-precedence-deny", "eval-precedence-approval"} {
				_, err := reconciler.Reconcile(ctx, ctrl.Request{
					NamespacedName: types.NamespacedName{Name: name, Namespace: "default"},
				})
				Expect(err).NotTo(HaveOccurred())
			}

			result, err := EvaluatePolicies(ctx, k8sClient, "default", "any-agent", "kubectl delete pod nginx")
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Decision).To(Equal(PolicyDecisionDeny))
		})
	})

	Context("AgentRef scoping", func() {
		It("Should skip policy when agentRef does not match", func() {
			policy := &agentv1.KubeCopilotPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "eval-scoped-agent",
					Namespace: "default",
				},
				Spec: agentv1.KubeCopilotPolicySpec{
					AgentRef: "specific-agent",
					Rules: []agentv1.PolicyRule{
						{
							Name:            "block-all",
							Type:            agentv1.PolicyRuleTypeDeny,
							CommandPatterns: []string{"*"},
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, policy)).To(Succeed())
			defer func() { _ = k8sClient.Delete(ctx, policy) }()

			_, err := reconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: types.NamespacedName{Name: policy.Name, Namespace: policy.Namespace},
			})
			Expect(err).NotTo(HaveOccurred())

			// Different agent — should not match
			result, err := EvaluatePolicies(ctx, k8sClient, "default", "other-agent", "any message")
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Decision).To(Equal(PolicyDecisionAllow))

			// Matching agent — should deny
			result, err = EvaluatePolicies(ctx, k8sClient, "default", "specific-agent", "any message")
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Decision).To(Equal(PolicyDecisionDeny))
		})
	})
})

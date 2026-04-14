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

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	agentv1 "github.com/gfontana/kube-copilot-agent/api/v1"
)

// KubeCopilotPolicyReconciler reconciles a KubeCopilotPolicy object
type KubeCopilotPolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=kubecopilot.io,resources=kubecopilotpolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubecopilot.io,resources=kubecopilotpolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kubecopilot.io,resources=kubecopilotpolicies/finalizers,verbs=update

func (r *KubeCopilotPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	policy := &agentv1.KubeCopilotPolicy{}
	if err := r.Get(ctx, req.NamespacedName, policy); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Validate rules
	var validationErr string
	for i, rule := range policy.Spec.Rules {
		if rule.Name == "" {
			validationErr = "rule at index " + string(rune('0'+i)) + " has empty name"
			break
		}
		if len(rule.CommandPatterns) == 0 {
			validationErr = "rule \"" + rule.Name + "\" has no command patterns"
			break
		}
	}

	if validationErr != "" {
		log.Info("Policy validation failed", "reason", validationErr)
		policy.Status.Active = false
		policy.Status.RuleCount = len(policy.Spec.Rules)
		meta.SetStatusCondition(&policy.Status.Conditions, metav1.Condition{
			Type:               "Ready",
			Status:             metav1.ConditionFalse,
			Reason:             "ValidationFailed",
			Message:            validationErr,
			ObservedGeneration: policy.Generation,
		})
		return ctrl.Result{}, r.Status().Update(ctx, policy)
	}

	// Policy is valid — activate it
	policy.Status.Active = true
	policy.Status.RuleCount = len(policy.Spec.Rules)
	meta.SetStatusCondition(&policy.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             "Active",
		Message:            "Policy is active and enforcing rules",
		ObservedGeneration: policy.Generation,
	})

	log.Info("Policy reconciled", "rules", policy.Status.RuleCount, "agentRef", policy.Spec.AgentRef)
	return ctrl.Result{}, r.Status().Update(ctx, policy)
}

// SetupWithManager sets up the controller with the Manager.
func (r *KubeCopilotPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&agentv1.KubeCopilotPolicy{}).
		Named("kubecopilotpolicy").
		Complete(r)
}

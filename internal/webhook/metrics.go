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

package webhook

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// tokensTotal tracks cumulative token usage broken down by agent, model, and direction
	// (direction is either "input" or "output").
	tokensTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kubecopilot_tokens_total",
			Help: "Total number of LLM tokens consumed, partitioned by agent, model, and direction (input/output).",
		},
		[]string{"agent", "model", "direction"},
	)

	// estimatedCostTotal tracks cumulative estimated USD cost broken down by agent and model.
	estimatedCostTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kubecopilot_estimated_cost_total",
			Help: "Cumulative estimated USD cost of LLM calls, partitioned by agent and model.",
		},
		[]string{"agent", "model"},
	)

	// sessionsTotal counts completed agent responses (one increment per KubeCopilotResponse).
	// A session may span multiple responses; use this counter as a proxy for request volume
	// grouped by agent.
	sessionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kubecopilot_sessions_total",
			Help: "Total number of completed agent responses received, partitioned by agent. A conversation session may produce multiple responses.",
		},
		[]string{"agent"},
	)
)

func init() {
	metrics.Registry.MustRegister(tokensTotal, estimatedCostTotal, sessionsTotal)
}

# KubeCopilotPolicy — Agent Guardrails

KubeCopilotPolicy defines safety constraints and guardrails for AI agent operations.
Policies are evaluated as **pre-dispatch screening** — they inspect the send message
before it reaches the agent. This is not a security boundary; agents decide what
commands to run and could interpret messages differently.

## Rule Types

| Type | Behavior |
|------|----------|
| `deny` | Blocks the send immediately. Status set to `Denied`. |
| `require-approval` | Pauses the send. Status set to `PendingApproval`. Set `spec.approved: true` on the Send to proceed. |

**Precedence:** Deny rules always win. If both deny and require-approval match, the request is denied.

## Example: Safe Mode Policy

```yaml
apiVersion: kubecopilot.io/v1
kind: KubeCopilotPolicy
metadata:
  name: safe-mode
spec:
  rules:
    - name: block-namespace-deletion
      type: deny
      commandPatterns:
        - "kubectl delete namespace *"
      message: "Namespace deletion is blocked"

    - name: approve-scaling
      type: require-approval
      commandPatterns:
        - "kubectl scale *"
      message: "Scaling requires approval"
```

## Scoping Policies

- **Namespace-wide:** Omit `spec.agentRef` — the policy applies to all agents.
- **Agent-specific:** Set `spec.agentRef` to restrict to one agent.

Multiple policies stack: all applicable policies are evaluated.

## Approving a Paused Send

When a send matches a `require-approval` rule, its status becomes `PendingApproval`:

```bash
# Check the status
kubectl get kubecopilotsend my-send -o jsonpath='{.status.phase}'
# PendingApproval

# Approve it
kubectl patch kubecopilotsend my-send --type merge -p '{"spec":{"approved":true}}'
```

After approval, the Send controller re-evaluates and proceeds with dispatch.

## Pattern Matching

`commandPatterns` uses glob-style matching:
- `*` matches any sequence of characters
- `?` matches a single character
- Matching is **case-insensitive**
- Patterns are matched against the full message and as sliding windows over message words

Examples:
- `"kubectl delete namespace *"` — matches any namespace deletion
- `"* --force *"` — matches any command with `--force` anywhere
- `"kubectl scale *"` — matches any scale command

## Limitations

- Policies screen the **message text**, not the actual commands the agent executes
- An agent may interpret "please clean up everything" differently than its literal text
- For hard RBAC boundaries, use Kubernetes RBAC via `spec.rbac` on KubeCopilotAgent
- Policies do not prevent the agent from running commands if it receives an approved message

# kube-copilot-agent

A Kubernetes operator that deploys a Copilot CLI as an AI agent inside your cluster, controlled entirely through Kubernetes CRDs. Users interact with the agent by creating Kubernetes resources — no direct pod access required.

## Overview

`kube-copilot-agent` wraps a Copilot CLI, such as [GitHub Copilot CLI](https://docs.github.com/en/copilot/using-github-copilot/using-github-copilot-in-the-command-line), in a container and exposes it as a Kubernetes-native AI agent. It supports:

- **Multi-turn conversations** with session continuity
- **Real-time streaming** of agent activity via `KubeCopilotChunk` CRDs
- **Custom skills** loaded from a ConfigMap (bash scripts the agent can invoke)
- **Custom instructions** via an `AGENT.md` ConfigMap
- **Cancellation** of in-flight requests
- **A web UI** for chatting with agents and browsing session history


`kube-copilot-agent` is designed to be extensible. Check for more information at [Agent Server Container](#agent-server-container).

### Architecture

```
User / Web UI
     │
     ▼
KubeCopilotSend (CR)          ← user creates this to send a message
     │
     ▼
Operator (controller-manager)
     │  POST /asyncchat
     ▼
Agent Pod (copilot CLI wrapper)
     │  tails events.jsonl, POSTs chunks + response to webhook
     ▼
Webhook server (inside operator)
     │  creates CRDs
     ├──► KubeCopilotChunk    (real-time streaming events)
     └──► KubeCopilotResponse (final answer)
```

### CRDs

| CRD | Purpose |
|---|---|
| `KubeCopilotAgent` | Declares an agent instance (image, credentials, skills, instructions) |
| `KubeCopilotSend` | Send a message to an agent; triggers copilot CLI execution |
| `KubeCopilotResponse` | Final response from the agent (written by operator webhook) |
| `KubeCopilotChunk` | Real-time streaming events (thinking, tool calls, results) |
| `KubeCopilotCancel` | Cancel an in-flight request |
| `KubeCopilotMessage` | Legacy single-turn message CRD |

---

## Getting Started

### Prerequisites

- Go v1.24+
- kubectl v1.20+
- Access to a Kubernetes/OpenShift cluster
- A GitHub account with Copilot access
- A GitHub Personal Access Token (PAT) with `copilot` scope
- Podman or Docker for building images

### 1. Create the namespace

```sh
kubectl apply -f config/samples/namespace.yaml
```

### 2. Install CRDs

```sh
make install
```

### 3. Build and push images

```sh
# Operator
make container-build container-push

# Agent container (GitHub Copilot CLI wrapper)
make container-build-agent container-push-agent

# Web UI
make container-build-ui container-push-ui
```

Image names are configured in the `Makefile`:

```makefile
IMG       ?= quay.io/gfontana/kube-copilot-agent:v1.0
AGENT_IMG ?= quay.io/gfontana/kube-github-copilot-agent-server:v1.0
UI_IMG    ?= quay.io/gfontana/kube-copilot-agent-ui:v1.0
```

You can run commands setting variables above to your registry, e.g.:

```sh
# Operator
make container-build container-push IMG=<your-registry>/<img-repo>/kube-copilot-agent:<tag>

# Agent container (GitHub Copilot CLI wrapper)
make container-build-agent container-push-agent AGENT_IMG=<your-registry>/<img-repo>/kube-copilot-agent:<tag>

# Web UI
make container-build-ui container-push-ui UI_IMG=<your-registry>/<img-repo>/kube-copilot-agent:<tag>
```

### 4. Deploy the operator

```sh
make deploy
```

### 5. Create credentials

Create a secret with your GitHub PAT:

```sh
kubectl create secret generic github-token \
  --from-literal=GITHUB_TOKEN=<your-pat> \
  -n kube-copilot-agent
```

Optionally, provide a kubeconfig so the agent can inspect your cluster:

```sh
kubectl create secret generic cluster-kubeconfig \
  --from-file=config=<path-to-kubeconfig> \
  -n kube-copilot-agent
```

### 6. Create skills and agent instructions ConfigMaps

```sh
kubectl apply -f config/samples/skills-configmap.yaml
kubectl apply -f config/samples/agent-md-configmap.yaml
```

### 7. Deploy an agent

```sh
kubectl apply -f config/samples/agent_v1_kubecopilotagent.yaml
```

Example `KubeCopilotAgent`:

```yaml
apiVersion: kubecopilot.io/v1
kind: KubeCopilotAgent
metadata:
  name: github-copilot-agent
  namespace: kube-copilot-agent
spec:
  image: quay.io/gfontana/kube-github-copilot-agent-server:v1.0
  githubTokenSecretRef:
    name: github-token
  kubeconfigSecretRef:       # optional
    name: cluster-kubeconfig
  skillsConfigMap: copilot-skills
  agentConfigMap: copilot-agent-md
  storageSize: "1Gi"
```

The operator creates a Pod and Service for the agent automatically.

### 8. Deploy the Web UI

```sh
make deploy-ui
```

Access via the OpenShift Route or expose the service manually on plain Kubernetes.

---

## Interacting with an Agent

### Via the Web UI

Open the route URL in a browser, select your agent, and start chatting. The UI supports:

- Multi-turn conversations with session history in the sidebar
- **Running Sessions** panel showing in-progress requests
- **Agent Activity** tab showing real-time chunk streaming
- **Stop** button to cancel an in-flight request

### Via kubectl (CRDs directly)

**Send a message:**

```yaml
apiVersion: kubecopilot.io/v1
kind: KubeCopilotSend
metadata:
  name: my-question
  namespace: kube-copilot-agent
spec:
  agentRef: github-copilot-agent
  message: "What is the overall health of the cluster?"
  sessionID: ""   # leave empty to start a new session
```

```sh
kubectl apply -f my-question.yaml
```

**Watch real-time activity:**

```sh
kubectl get kubecopilotchunks -n kube-copilot-agent -w
```

**Read the response:**

```sh
kubectl get kubecopilotresponses -n kube-copilot-agent -o yaml
```

**Resume a session:** set `spec.sessionID` to the session ID returned in a previous `KubeCopilotResponse`.

**Cancel a request:**

```yaml
apiVersion: kubecopilot.io/v1
kind: KubeCopilotCancel
metadata:
  name: cancel-my-question
  namespace: kube-copilot-agent
spec:
  sendRef: my-question
  agentRef: github-copilot-agent
```

---

## Custom Skills

Skills are bash snippets the agent can invoke as tools. Define them in a ConfigMap with a `skills.md` key:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: copilot-skills
  namespace: kube-copilot-agent
data:
  skills.md: |
    ## Skill: List unhealthy pods
    Lists all pods that are not Running or Completed.
    ```bash
    kubectl get pods -A | grep -vE "Running|Completed"
    ```
```

See `config/samples/skills-configmap.yaml` for a full example with Kubernetes operations skills.

---

## Custom Agent Instructions (AGENT.md)

Shape agent behaviour with persistent instructions:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: copilot-agent-md
  namespace: kube-copilot-agent
data:
  AGENT.md: |
    # Agent Instructions
    - Always confirm the current cluster context before acting.
    - Never modify resources in production namespaces (prefixed with `prod-`).
    - Prefer read-only operations unless explicitly asked to make changes.
```

---

## Chunk Types (Real-time Streaming)

`KubeCopilotChunk` resources are created as the agent works:

| `chunkType` | Description |
|---|---|
| `thinking` | Agent's internal reasoning |
| `tool_call` | Agent invoking a skill or tool |
| `tool_result` | Result returned from the tool |
| `response` | Final answer text |
| `info` | Processing status (e.g. "Processing: ...") |
| `error` | Error during processing or cancellation |

---

## Agent Server Container

The `agent-server-container/` directory contains the HTTP shim that bridges the Kubernetes operator with an AI CLI tool running inside the agent pod. Each subdirectory implements the shim for a specific AI agent binary.

```
agent-server-container/
  github-copilot/       ← GitHub Copilot CLI implementation
    server.py           ← FastAPI HTTP shim
    entrypoint.sh       ← Container entrypoint (auth setup, skill staging)
    Containerfile       ← Container image definition
```

### How It Works

The operator communicates with the agent pod over HTTP. The shim (`server.py`) exposes a small API that the operator calls and that POSTs results back to the operator's internal webhook:

```
Operator                    Agent Pod (shim)
   │                             │
   │  POST /asyncchat            │
   │ ──────────────────────────► │  enqueue message
   │  ← { queue_id }             │
   │                             │  spawn AI CLI subprocess
   │                             │  tail events.jsonl / stdout
   │  ◄── POST /chunk (webhook)  │  stream chunk events
   │  ◄── POST /chunk (webhook)  │
   │  ◄── POST /response         │  final answer
   │                             │
   │  DELETE /cancel/{queue_id}  │
   │ ──────────────────────────► │  SIGTERM to process group
```

**API surface the shim must expose:**

| Endpoint | Method | Description |
|---|---|---|
| `/health` | GET | Liveness probe — return `{"status":"ok"}` |
| `/asyncchat` | POST | Enqueue a message; returns `{"queue_id": "..."}` immediately |
| `/cancel/{queue_id}` | DELETE | Kill the in-flight subprocess for the given queue |

**Webhook payloads the shim must POST to `$WEBHOOK_URL`:**

Chunk (streamed during execution):
```json
{
  "queue_id": "<uuid>",
  "seq": 1,
  "type": "thinking|tool_call|tool_result|response|info|error",
  "content": "...",
  "session_id": "<copilot-session-id>",
  "send_ref": "...",
  "namespace": "...",
  "agent_ref": "..."
}
```

Final response (POST to `$WEBHOOK_URL`):
```json
{
  "queue_id": "<uuid>",
  "response": "full answer text",
  "session_id": "<session-id>",
  "send_ref": "...",
  "namespace": "...",
  "agent_ref": "..."
}
```

**Environment variables injected by the operator:**

| Variable | Description |
|---|---|
| `GITHUB_TOKEN` | Auth token for the AI CLI |
| `WEBHOOK_URL` | URL of the operator's internal webhook (`http://<svc>/response`) |
| `COPILOT_HOME` | Persistent storage root (backed by a PV) |
| `KUBECONFIG` | Path to kubeconfig if a `kubeconfigSecretRef` is set |

**Skills and AGENT.md** are mounted into the container as ConfigMaps:
- Skills ConfigMap → `/copilot-skills-staging/` → `entrypoint.sh` stages them into `$COPILOT_HOME/skills/<name>/SKILL.md`
- AGENT.md ConfigMap → `$COPILOT_HOME/AGENT.md`

---

### Creating a New Agent Image (e.g., Claude Code)

To support a different AI CLI (such as [Claude Code](https://docs.anthropic.com/en/docs/claude-code)), create a new subdirectory under `agent-server-container/`:

```
agent-server-container/
  github-copilot/    ← existing
  claude-code/       ← new
    server.py
    entrypoint.sh
    Containerfile
```

#### 1. Write `entrypoint.sh`

Set up auth and launch `server.py`, example:

```bash
#!/bin/bash
set -e

export ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY}"
export AGENT_HOME="${AGENT_HOME:-/agent}"

mkdir -p "${AGENT_HOME}/sessions" "${AGENT_HOME}/.cache"

# Stage skills (same pattern as github-copilot)
if [ -d /copilot-skills-staging ]; then
  for f in /copilot-skills-staging/*.md; do
    [ -f "$f" ] || continue
    skill_name="$(basename "$f" .md)"
    mkdir -p "${AGENT_HOME}/skills/${skill_name}"
    cp "$f" "${AGENT_HOME}/skills/${skill_name}/SKILL.md"
  done
fi

exec /opt/venv/bin/python /server.py
```

#### 2. Write `server.py`

Implement the three required endpoints. Example:

```python
import asyncio, httpx, json, os, subprocess, uuid
from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI()
WEBHOOK_URL = os.environ.get("WEBHOOK_URL", "")
_active_procs = {}

class AsyncChatRequest(BaseModel):
    message: str
    session_id: str | None = None
    send_ref: str | None = None
    namespace: str | None = None
    agent_ref: str | None = None

@app.get("/health")
async def health():
    return {"status": "ok"}

@app.post("/asyncchat")
async def asyncchat(req: AsyncChatRequest):
    queue_id = str(uuid.uuid4())
    asyncio.create_task(process(queue_id, req))
    return {"queue_id": queue_id, "status": "queued"}

@app.delete("/cancel/{queue_id}")
async def cancel(queue_id: str):
    proc = _active_procs.get(queue_id)
    if proc:
        proc.terminate()
        _active_procs.pop(queue_id, None)
        return {"status": "cancelled", "queue_id": queue_id}
    return {"status": "not_found", "queue_id": queue_id}

async def process(queue_id: str, req: AsyncChatRequest):
    chunk_url = WEBHOOK_URL.replace("/response", "/chunk")
    # Launch Claude Code CLI — adapt flags to the actual binary
    cmd = ["claude", "--no-interactive", "--output-format", "stream-json",
           req.message]
(.. ommitted ..)
```

#### 3. Write `Containerfile`

```dockerfile
FROM python:3.12-slim

RUN pip install --no-cache-dir fastapi uvicorn httpx && \
    # Install the Claude Code CLI (adjust to actual install method)
    pip install claude-code

RUN useradd -m -s /bin/bash agent
WORKDIR /home/agent

COPY entrypoint.sh /entrypoint.sh
COPY server.py /server.py
RUN chmod +x /entrypoint.sh

USER agent
EXPOSE 8080
ENTRYPOINT ["/entrypoint.sh"]
```

#### 4. Add a Makefile target

```makefile
CLAUDE_IMG ?= quay.io/yourorg/kube-claude-code-agent-server:v1.0

.PHONY: container-build-claude container-push-claude
container-build-claude:
	$(CONTAINER_TOOL) build -t $(CLAUDE_IMG) ./agent-server-container/claude-code/

container-push-claude:
	$(CONTAINER_TOOL) push $(CLAUDE_IMG)
```

#### 5. Create a `KubeCopilotAgent` CR pointing to the new image

```yaml
apiVersion: kubecopilot.io/v1
kind: KubeCopilotAgent
metadata:
  name: claude-code-agent
  namespace: kube-copilot-agent
spec:
  image: quay.io/yourorg/kube-claude-code-agent-server:v1.0
  githubTokenSecretRef:   # reuse field for ANTHROPIC_API_KEY via a secret
    name: anthropic-token
  skillsConfigMap: claude-skills
  agentConfigMap: claude-agent-md
  storageSize: "1Gi"
```

The operator treats every `KubeCopilotAgent` the same way regardless of which CLI runs inside — as long as the shim implements the three-endpoint contract above, the full UI, streaming, session history, and cancellation features work automatically.

---

## Development

### Run locally

```sh
make install   # install CRDs into current cluster
make run       # run operator locally against current kubeconfig context
```

### Regenerate CRDs and RBAC after changing API types

```sh
make manifests
make generate
```

### Build and test

```sh
make build
make test
```

---

## Project Structure

```
api/v1/                          CRD type definitions
  kubecopilotagent_types.go      KubeCopilotAgent
  kubecopilotcancel_types.go     KubeCopilotCancel
  kubecopilotchunk_types.go      KubeCopilotChunk
  kubecopilotmessage_types.go    KubeCopilotMessage
  kubecopilotresponse_types.go   KubeCopilotResponse
  kubecopilotsend_types.go       KubeCopilotSend

internal/controller/
  kubecopilotagent_controller.go    Creates/manages agent pods and services
  kubecopilotsend_controller.go     Forwards messages to the agent
  kubecopilotcancel_controller.go   Cancels in-flight requests

internal/webhook/
  server.go                    HTTP server receiving chunks + responses from agent pod

agent-server-container/github-copilot/
  server.py                    FastAPI server wrapping the copilot CLI
  Containerfile                Agent container image definition

web-ui/
  app/main.py                  FastAPI web UI backend
  app/k8s_client.py            Kubernetes API client
  app/static/index.html        Single-page chat UI
  deploy/base/                 Kustomize manifests for the web UI

config/
  crd/bases/                   Generated CRD manifests (do not edit)
  rbac/                        RBAC roles and bindings
  manager/                     Operator deployment manifests
  samples/                     Example CRs and supporting resources
```

---

## Uninstall

```sh
kubectl delete -k config/samples/
make undeploy
make uninstall
kubectl delete namespace kube-copilot-agent
```

---

## License

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

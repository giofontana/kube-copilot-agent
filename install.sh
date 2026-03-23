#!/bin/bash

echo "Starting Kube Copilot Agent installation..."

# Build and push images
echo "[1/10] Building and pushing operator, agent, and UI container images..."
make container-build container-push container-build-agent container-push-agent container-build-ui container-push-ui
echo "Container images built and pushed."

# Create namespace
echo "[2/10] Creating namespace..."
kubectl apply -f config/samples/namespace.yaml
echo "Namespace applied."

# Install CRDs
echo "[3/10] Installing CRDs..."
make install
echo "CRDs installed."

# Deploy the operator
echo "[4/10] Deploying operator..."
make deploy
echo "Operator deployed."

# Wait for user to create the GitHub token secret
echo "[5/10] Manual step required: GitHub token secret"
echo "Please create the GitHub token secret by running:"
echo "kubectl apply -f config/samples/github-token-secret.yaml"
echo "Press Enter once you have created the secret..."
read
echo "GitHub token secret step completed."

# Wait for user to create the cluster kubeconfig secret
echo "[6/10] Manual step required: cluster kubeconfig secret"
echo "Please create the cluster kubeconfig secret by running:"
echo "kubectl apply -f config/samples/cluster-kubeconfig-secret.yaml"
echo "Press Enter once you have created the secret..."
read
echo "Cluster kubeconfig secret step completed."

### Create skills and agent instructions ConfigMaps
echo "[7/10] Applying skills and agent instruction ConfigMaps..."
kubectl apply -f config/samples/skills-configmap.yaml
kubectl apply -f config/samples/agent-md-configmap.yaml
echo "ConfigMaps applied."

### Deploy an agent
echo "[8/10] Deploying sample agent resource..."
kubectl apply -f config/samples/agent_v1_kubecopilotagent.yaml
echo "Sample agent deployed."

### Deploy the Web UI
echo "[9/10] Deploying Web UI..."
make deploy-ui
echo "Web UI deployed."

### Test the setup by creating a KubecopilotSend resource
echo "[10/10] Creating a test KubeCopilotSend resource..."
cat <<EOF | kubectl apply -f -
apiVersion: kubecopilot.io/v1
kind: KubeCopilotSend
metadata:
  name: my-question
  namespace: kube-copilot-agent
spec:
  agentRef: github-copilot-agent
  message: "What is 2 + 2?"
  sessionID: ""   # leave empty to start a new session
EOF
echo "Test resource created."

### Watch kubecopilotchunks and present a success message once the response is received
echo "Waiting for chunks..."
while true; do
  if kubectl get kubecopilotchunks -n kube-copilot-agent -o name 2>/dev/null | grep -q .; then
    echo "Received chunks!"
    break
  fi
  sleep 2
done

### Watch kubecopilotresponse and present a success message once the response is received
echo "Waiting for response..."
while true; do
  if kubectl get kubecopilotresponse -n kube-copilot-agent -o name 2>/dev/null | grep -q .; then
    echo "Received response!"
    break
  fi
  sleep 2
done

echo "Installation flow completed successfully."
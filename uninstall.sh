#!/bin/bash

echo "Uninstalling Kube Copilot Agent..."

helm uninstall kube-copilot-console-plugin --namespace kube-copilot-agent  # if installed
helm uninstall kube-copilot-ui      --namespace kube-copilot-agent
helm uninstall my-agent             --namespace kube-copilot-agent
helm uninstall kube-copilot-agent   --namespace kube-copilot-agent
kubectl delete namespace kube-copilot-agent
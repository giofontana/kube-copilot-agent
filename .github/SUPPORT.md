# Support

## Getting Help

If you need help with KubeCopilot, here are your options:

### 📖 Documentation
Start with the [project documentation](https://github.com/giofontana/kube-copilot-agent/tree/main/docs):
- [Architecture](docs/architecture.md)
- [Installation Guide](docs/installation.md)
- [Configuration](docs/configuration.md)
- [Agent Server](docs/agent-server.md)
- [Usage Guide](docs/usage.md)
- [Development Guide](docs/development.md)

### 💬 Discussions
For questions, ideas, and community conversations, use [GitHub Discussions](https://github.com/giofontana/kube-copilot-agent/discussions).

### 🐛 Bug Reports
Found a bug? [Open an issue](https://github.com/giofontana/kube-copilot-agent/issues/new?template=bug_report.yml) using our bug report template.

### 💡 Feature Requests
Have an idea? [Open a feature request](https://github.com/giofontana/kube-copilot-agent/issues/new?template=feature_request.yml).

### 🔒 Security Issues
For security vulnerabilities, please see our [Security Policy](SECURITY.md). Do NOT open a public issue.

## What to Include in a Support Request

When asking for help, include:
- KubeCopilot version (operator image tag)
- Kubernetes/OpenShift version
- Relevant CRD YAML
- Operator logs: `kubectl logs -n kube-copilot-agent -l control-plane=controller-manager`
- Agent logs: `kubectl logs -n kube-copilot-agent -l app=github-copilot-agent`

## Commercial Support

KubeCopilot is an open-source project maintained by volunteers. There is no commercial support at this time.

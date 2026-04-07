# Release Process

This document describes the release process for KubeCopilot.

## Versioning

KubeCopilot follows [Semantic Versioning](https://semver.org/):
- **MAJOR** — Breaking changes to CRD schemas, API contract, or Helm values
- **MINOR** — New features, new CRDs, backward-compatible additions
- **PATCH** — Bug fixes, security patches, documentation updates

## Container Images

| Component | Registry | Image |
|---|---|---|
| Operator | `quay.io/gfontana/kube-copilot-agent` | Kubernetes operator |
| Agent Server | `quay.io/gfontana/kube-github-copilot-agent-server` | GitHub Copilot SDK engine |
| Web UI | `quay.io/gfontana/kube-copilot-agent-ui` | Chat interface |

## Cutting a Release

1. **Update version** in Helm chart `Chart.yaml` files and image tags
2. **Run tests**: `make test`
3. **Build images**: `make container-build container-build-agent container-build-ui`
4. **Push images**: `make container-push container-push-agent container-push-ui`
5. **Tag the release**: `git tag -a v<version> -m "Release v<version>"`
6. **Push the tag**: `git push origin v<version>`
7. **Create GitHub Release** with changelog

## Helm Charts

Charts are versioned independently from the operator. Update `Chart.yaml` version and `appVersion` for each chart in `helm/`.

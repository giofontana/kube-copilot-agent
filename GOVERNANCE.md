# Governance

This document describes the governance model for the KubeCopilot project.

## Overview

KubeCopilot follows a **maintainer council** governance model, where project decisions are made by consensus among active maintainers. This model is inspired by [CNCF governance best practices](https://contribute.cncf.io/projects/best-practices/governance/).

## Principles

- **Openness** — All discussions, decisions, and roadmap items are public
- **Meritocracy** — Contributions and sustained engagement lead to increased responsibility
- **Consensus** — Decisions are made by consensus; if consensus cannot be reached, a majority vote among maintainers decides
- **Transparency** — Meeting notes, design decisions, and release plans are documented publicly

## Roles

### Contributors
Anyone who contributes code, documentation, bug reports, or feature requests. No formal process is required.

### Reviewers
Contributors who have demonstrated sustained, quality contributions may be invited as reviewers. Reviewers can approve PRs but cannot merge.

### Maintainers
Reviewers who have demonstrated deep understanding of the project architecture, consistent review quality, and commitment to the project may be nominated as maintainers. Maintainers can merge PRs, cut releases, and vote on governance decisions.

See the [Contributor Ladder](CONTRIBUTOR_LADDER.md) for the full path.

## Decision Making

1. **Lazy consensus** — Most decisions (bug fixes, minor features, docs) are merged after approval from one maintainer
2. **Active consensus** — Significant changes (new CRDs, architecture changes, breaking changes) require approval from at least two maintainers
3. **Voting** — If consensus cannot be reached, a simple majority vote among active maintainers decides. Each maintainer has one vote.

## Adding or Removing Maintainers

- **Adding**: Any maintainer may nominate a contributor. Approval requires consensus or majority vote among existing maintainers.
- **Removing**: A maintainer may step down voluntarily. A maintainer may be removed for sustained inactivity (6+ months) or Code of Conduct violations, by majority vote.
- **Emeritus**: Maintainers who step down are moved to the Emeritus list in [MAINTAINERS.md](MAINTAINERS.md).

## Conflict Resolution

If a disagreement cannot be resolved through discussion:
1. The maintainers hold a vote
2. If still unresolved, an external mediator from the CNCF community may be consulted

## Changes to Governance

Changes to this document require approval from all active maintainers.

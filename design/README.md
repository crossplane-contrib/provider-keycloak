# Design Proposals

This directory holds design proposals for provider-keycloak: drafts that are
discussed and refined *before* code is written.

Proposals are **not** user documentation. User-facing documentation lives in
`docs/content/` and is published to
<https://crossplane-contrib.github.io/provider-keycloak/>. Keeping proposals out
of the Hugo tree avoids publishing ideas that were never implemented, and avoids
a second place that has to be kept in sync with the shipped behaviour.

## Conventions

- One file per proposal, named `NNNN-short-title.md`.
- Each proposal states: problem, goals, non-goals, the proposed design,
  rejected alternatives, and a phased rollout.
- A proposal keeps a `Status` field: `Draft`, `Accepted`, `Implemented` or
  `Rejected`. Once a proposal is implemented, the durable parts of it move into
  `docs/content/` (and, where relevant, the agent instructions) and the proposal
  is only kept as a record of *why*.

## Index

| Proposal | Status | Summary |
|----------|--------|---------|
| [0001 – Schema-driven resource onboarding](0001-schema-driven-resource-onboarding.md) | Draft | Use `config/schema.json` to *detect* configuration drift and multitype candidates; scaffold the rest with one `make` command, keeping every decision explicit in Go. |
| [0002 – AI skills with a single source of truth](0002-ai-skills-single-source-of-truth.md) | Draft | Author agent guidance once under `ai/skills/`, generate `AGENTS.md`, `CLAUDE.md`, `SKILL.md` and the docs page from it. |

# 0002 – AI Skills With a Single Source of Truth

- **Status:** Draft
- **Issue:** [#712](https://github.com/crossplane-contrib/provider-keycloak/issues/712) (comment)
- **Scope:** `ai/`, `AGENTS.md`, `SKILL.md`, `docs/content/docs/ai-usage/`, `Makefile`, CI

## Problem

Agent-facing guidance is currently written three times:

| File | Lines | Audience |
|------|-------|----------|
| `AGENTS.md` | 350 | Copilot / Cursor / Codex, auto-loaded from the repo root |
| `SKILL.md` | 149 | Claude-style skill discovery |
| `docs/content/docs/ai-usage/agents.md` | 290 | published docs site, feeds `llms.txt` |

They are near-copies of one another, and they have already drifted:

- `SKILL.md` says the Keycloak Terraform provider "must not be updated via
  Renovate". `.github/renovate.json` groups it into a weekly PR and
  **auto-merges** minor/patch/digest updates; only major bumps are held back.
  `AGENTS.md` describes the correct behaviour. One of the three files is simply
  wrong, and an agent reading it will draw a wrong conclusion.
- The docs page documents `make test` and `make docs-freshness-check` in its
  task table; `AGENTS.md` does not.
- Paragraphs differ only in line wrapping, which makes real differences
  invisible in review.

Nobody did anything wrong here — three hand-maintained copies of the same
content will always diverge. The fix has to be structural.

At the same time the guidance is organised as one big page per agent, while the
actual work is a small set of highly repetitive procedures ("add a resource",
"add an e2e demo"). Those are exactly the tasks a cheap model can do reliably
*if* it is handed a precise, verifiable procedure instead of a 350-line
document.

## Goals

- One authored source for agent guidance; every agent-specific file is
  generated from it and marked as generated.
- Staleness is a CI failure, not a discovery months later.
- Guidance is split into **task-shaped skills** with explicit verification
  commands, so a small model can follow them and prove it succeeded.
- A skill exists for *research* as well as for editing, so the default answer to
  an upstream limitation is "report it upstream", not "work around it".
- No new tooling to maintain beyond one small generator, in the style of the
  generators the repo already has (`docs/scripts/gen-llms.sh`,
  `cmd/generatedlist`, `scripts/e2e_dag.py`).

## Non-Goals

- Inventing a skill format. Follow the conventions the consuming tools already
  expect (front matter with `name` and `description`, as `SKILL.md` already
  uses).
- Teaching agents general Crossplane/Upjet concepts. Skills link to upstream
  docs; they do not restate them.
- Replacing the published docs site. `docs/content/` stays the source for the
  *user* documentation; `ai/` is the source for *contributor/agent* procedure.

## Proposed Design

### Layout

```
ai/
  context/                 authored, shared prose included by skills
    repository.md          layout, what is generated, hard rules
    generation.md          make generate pipeline, gates
    pitfalls.md            known constraints and troubleshooting table
  skills/
    add-resource/SKILL.md          expose a Terraform resource as an MR
    add-e2e-test/SKILL.md          add a demo + case-list entry, verify selection
    research-upstream/SKILL.md     investigate Crossplane / Upjet / TF provider
    write-docs/SKILL.md            author a resource page, regenerate llms.txt
  templates/               snippets skills point at (config stub, demo skeleton)
```

Every skill is a single markdown file with:

1. front matter (`name`, `description`) so it is discoverable,
2. **When to use / when not to use**,
3. **Prerequisites** (e.g. `make submodules` on a fresh clone),
4. **Steps**, each one a concrete command or a concrete file to edit,
5. **Verification** — the exact commands that must pass
   (`make generate && git diff --exit-code`, `make test`,
   `make e2e-cases-check`, `python3 scripts/e2e_dag.py select …`),
6. **Failure modes** and what they mean,
7. **Escalation** — when to stop and open an upstream issue instead of patching
   around the problem.

The `add-resource` skill is the direct consumer of `make new-resource` from
[0001](0001-schema-driven-resource-onboarding.md): scaffold, fill the `TODO`s,
run the verification commands. That combination is what makes the task tractable
for a cheap model — the scaffolding removes the "where do I put things?" problem
and the gates remove the "did I get it right?" problem.

### Generation

```
make ai-gen     # regenerate all agent-specific files from ai/
make ai-check   # fail if any generated file is stale   (CI, and generate.done)
```

Generated, each with a `<!-- generated from ai/ — do not edit -->` header:

| Output | Shape |
|--------|-------|
| `AGENTS.md` | hard rules inline + index of skills with one-line descriptions |
| `CLAUDE.md` | same, Claude-flavoured front matter |
| `.github/copilot-instructions.md` | same |
| `SKILL.md` | index pointing into `ai/skills/` |
| `docs/content/docs/ai-usage/agents.md` | rendered with Hugo front matter, so it keeps feeding `llms.txt` |

Two open shapes for the root files, to be decided when refining:

- **Thin pointer**: the file only lists the skills and the handful of
  non-negotiable rules, everything else is read on demand from `ai/`. Cheapest,
  smallest context footprint, but relies on the agent actually opening the
  linked file.
- **Full render**: the file inlines the shared context. Guarantees the agent
  sees it, but burns context on every task and re-creates today's 350-line page.

The recommendation is *thin pointer plus hard rules inline*: the rules whose
violation is expensive (never edit `apis/`, `package/crds/`,
`examples-generated/`, `config/generated.lst`; always `make generate`; no
workarounds) are short enough to inline, and everything task-specific is pulled
in only when that task is being done.

Hooking `ai-check` into CI mirrors `generated-lst-check` and
`docs-freshness-check`, so the anti-staleness mechanism is one the repo already
operates and trusts.

## Rejected Alternatives

- **Symlink `CLAUDE.md` → `AGENTS.md`.** Cheapest possible option, but symlinks
  render poorly on github.com, break on Windows checkouts without developer
  mode, and are not followed by every agent runtime. It also cannot solve the
  docs-site copy, which needs Hugo front matter.
- **Keep three hand-maintained files and add a review checklist.** This is the
  status quo; the Renovate contradiction above shows how it ends.
- **Put the source of truth in `docs/content/` and generate the root files from
  it.** Tempting since the docs pipeline exists, but it forces contributor
  procedure onto the public user-facing site and couples agent guidance to Hugo
  front matter and shortcodes. `ai/` is plain markdown that both humans and
  agents can read without a build step.
- **One giant skill file.** Defeats the purpose: the value for cheap models is a
  short, task-scoped procedure with verification commands.

## Rollout

1. Create `ai/` and move the existing `AGENTS.md` content into
   `ai/context/*.md` unchanged, resolving the three-way drift explicitly (the
   Renovate statement is the first thing to fix).
2. Add `make ai-gen` / `make ai-check`, generate `AGENTS.md`, `SKILL.md` and the
   docs page; wire `ai-check` into CI and `generate.done`.
3. Extract `add-resource` and `add-e2e-test` skills from the existing prose;
   validate them by having an agent add one real resource end to end with no
   other context.
4. Add `research-upstream` and `write-docs`.
5. Add `CLAUDE.md` / `.github/copilot-instructions.md` outputs once the format
   has settled.

## Open Questions

- Which agent-specific outputs are actually worth generating today? Every extra
  output is another file in the root directory; adding one is cheap once the
  generator exists, so it may be better to start with the three that exist.
- Should skills be validated mechanically (front matter present, every `make`
  target referenced actually exists in the `Makefile`, every referenced path
  exists)? That is a ~50-line check and would prevent the most common form of
  rot: instructions pointing at files that were renamed.
- Do we want `ai/skills/` to be shipped in the Crossplane package, or is it
  repository-only tooling?

---
title: Config Audit
description: Find missing references, missing multitypes and configuration drift with make config-audit
weight: 6
---

# Config Audit

`make config-audit` cross-checks the Terraform schema against the provider
configuration the generator builds, and reports reference wiring that is
inconsistent, incomplete or missing.

It needs neither a Keycloak instance nor network access: it builds
`config.GetProvider(true)` from `config/schema.json` and inspects the result.

```bash
make config-audit                                              # human-readable report
make config-audit CONFIG_AUDIT_ARGS='--show-all'               # include suppressed findings
make config-audit CONFIG_AUDIT_ARGS='--format=json'            # machine-readable report
make config-audit CONFIG_AUDIT_ARGS='--fail-on=drift'          # exit non-zero on findings
```

The audit **reports, it does not decide**. A finding is a question — "why is
this attribute treated differently here?" — and the answer is a Keycloak
semantics decision that belongs in `config/` as reviewed Go code. See
[design/0001](https://github.com/crossplane-contrib/provider-keycloak/blob/main/design/0001-schema-driven-resource-onboarding.md)
for the rationale.

## Detectors

### `drift` — the same attribute treated in more than one way

Attributes are grouped by name **and schema shape** (type plus optionality), so
only attributes a user would expect to behave identically are compared. A group
is reported when some resources wire the attribute and others do not, or when it
resolves to more than one target type.

| Class | Meaning |
|-------|---------|
| `gap` | wired to one target on some resources, unwired on others with an identical schema |
| `multitype` | one attribute name resolving to several target types — the signal for `config/multitypes` |

A resource that is itself the reference target is never counted as unwired:
`keycloak_realm.realm` and `keycloak_openid_client.client_id` name the object
they configure rather than pointing at another one.

### `missing-multitype` — a reference to one member of a type family

Type families are derived from the configuration itself, so the detector only
generalises from an existing precedent. A family is formed when

- one attribute name resolves to different targets on different resources, or
- a `config/multitypes` field wires several targets on one resource through its
  synthetic instances (`client_id` + `saml_client_id`).

Every reference pointing at one family member while the resource wires none of
the others is reported. Findings on resources whose name is prefixed
`keycloak_openid_` / `keycloak_saml_` are marked `protocolSpecific`: an OpenID
protocol mapper cannot attach to a SAML client, so those are shown only with
`--show-all` and never fail the gate.

### `unclassified` — a reference-shaped attribute that is not wired

Non-computed, non-sensitive `*_id`, `*_ids` and `*_alias` attributes without a
reference. Roughly a third of these are correct omissions (`provider_id`,
`tenant_id`, `entity_id`, `key_alias`), which is why this detector reports but
is not meant to gate CI until there is a place to record the reason.

## Working with the findings

1. Run `make config-audit` and pick a finding.
2. Establish the semantics from the pinned upstream sources rather than
   guessing — the Terraform provider docs are checked out by `make pull-docs`
   into `.work/keycloak/keycloak/`, and its Go source is at
   `go list -m -f '{{.Dir}}' github.com/keycloak/terraform-provider-keycloak`.
3. Record the decision in `config/`: wire the reference, apply
   `config/multitypes`, or document why the attribute is not a reference.
4. Run `make generate` and review the `package/crds/` diff.

`--format=json` prints every finding with a stable `Key()`-style identity
(`detector/resource/attribute`), so repeated runs can be deduplicated when
filing follow-up issues.

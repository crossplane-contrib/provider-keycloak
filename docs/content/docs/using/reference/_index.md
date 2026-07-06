---
title: Reference
weight: 4
---

Reference pages collect reusable provider concepts that apply across many
resources, such as credentials, ProviderConfig, troubleshooting, and common
Crossplane manifest patterns.

For resource-specific field schemas, use the generated CRDs in `package/crds/`.

{{< cards >}}
  {{< card link="provider-config/" title="ProviderConfig" icon="adjustments" subtitle="Connection details and authentication for a Keycloak instance" >}}
  {{< card link="credentials/" title="Credentials" icon="lock-closed" subtitle="All supported credential fields and authentication methods" >}}
  {{< card link="common-patterns/" title="Common Patterns" icon="puzzle" subtitle="Cross-cutting Crossplane patterns: references, selectors, policies" >}}
  {{< card link="troubleshooting/" title="Troubleshooting" icon="support" subtitle="Diagnose and resolve common issues" >}}
{{< /cards >}}

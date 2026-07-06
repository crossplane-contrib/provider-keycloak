---
title: Provider Keycloak
description: Crossplane provider for declarative Keycloak management
layout: hextra-home
---

<div class="hx:mt-6 hx:mb-6">
{{< hextra/hero-headline >}}Declarative Keycloak on Kubernetes{{< /hextra/hero-headline >}}
{{< hextra/hero-subtitle >}}
Manage Keycloak realms, clients, users, and roles as Kubernetes resources.
Built on [Crossplane](https://crossplane.io/) and [Upjet](https://github.com/crossplane/upjet).
{{< /hextra/hero-subtitle >}}
</div>

<div class="hx:mb-12 hx:flex hx:flex-wrap hx:gap-3">
{{< hextra/hero-button text="Get Started" link="docs/using/getting-started/installation/" >}}
{{< hextra/hero-badge link="https://github.com/crossplane-contrib/provider-keycloak" >}}⭐ GitHub{{< /hextra/hero-badge >}}
{{< hextra/hero-badge link="https://github.com/crossplane-contrib/provider-keycloak/releases" >}}📦 Releases{{< /hextra/hero-badge >}}
</div>

{{< cards cols="3" >}}
  {{< card link="docs/using/getting-started/installation/" title="Installation" icon="server" subtitle="Add the provider to your Crossplane cluster" >}}
  {{< card link="docs/using/getting-started/configuration/" title="Configuration" icon="adjustments" subtitle="Connect to your Keycloak instance" >}}
  {{< card link="docs/using/getting-started/first-realm/" title="First Realm" icon="academic-cap" subtitle="Create a realm, client, and user" >}}
{{< /cards >}}

{{< cards cols="2" >}}
  {{< card link="docs/using/resources/" title="Managed Resources" icon="book-open" subtitle="Reference for all CRD types: realms, clients, users, roles, groups, identity providers, and more" >}}
  {{< card link="docs/using/reference/" title="Reference" icon="puzzle" subtitle="ProviderConfig, credentials, common patterns, and troubleshooting" >}}
  {{< card link="docs/developing/" title="Developing" icon="terminal" subtitle="Set up a local dev environment, contribute code, or work on the docs" >}}
  {{< card link="docs/ai-usage/" title="AI Usage" icon="sparkles" subtitle="llms.txt, agents.md, and AI-oriented entry points for this project" >}}
{{< /cards >}}

# LLM Files

[llms.txt](https://crossplane-contrib.github.io/provider-keycloak/pr-preview/pr-760/llms.txt)


The docs site ships AI-oriented reference files:

- [`llms.txt`](https://crossplane-contrib.github.io/provider-keycloak/llms.txt):
  an index of all documentation pages and section landing pages, linking directly
  to their Markdown representations.
- [`llms-full.txt`](https://crossplane-contrib.github.io/provider-keycloak/llms-full.txt):
  the full documentation content in one file.

HTML pages advertise the index and their Markdown alternative in link elements.
Markdown pages also link back to `llms.txt`. Guides are available at `.md` URLs
(for example, [Realms](https://crossplane-contrib.github.io/provider-keycloak/docs/using/resources/realms.md));
section landing pages use `index.md` (for example,
[Resources](https://crossplane-contrib.github.io/provider-keycloak/docs/using/resources/index.md)).

The site is hosted on GitHub Pages, which does not support `Accept: text/markdown`
content negotiation. Request the explicit Markdown URL rather than sending that
header to an HTML URL.



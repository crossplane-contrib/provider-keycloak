# Documentation

This directory contains the provider-keycloak documentation site built with [Hugo](https://gohugo.io/) and the [Hextra](https://github.com/imfing/hextra) theme.

## Local Development

```bash
cd docs
hugo server --buildDrafts
```

This starts a local development server at `http://localhost:1313/`.

## Build

```bash
hugo --minify
```

## Content Model

- Write guides and examples by hand for real workflows.
- Keep resource pages curated and focused on common usage.
- Treat `../package/crds/` as the generated source of truth for complete field
  schemas.
- See `content/docs/developing/documentation-model.md` before adding large
  reference sections.

## Deployment

The documentation is automatically deployed to GitHub Pages when changes are pushed to the `main` branch via the `.github/workflows/deploy-docs.yml` workflow.

Live site: https://crossplane-contrib.github.io/provider-keycloak/

## AI discovery and Markdown

Every documentation page has a Markdown representation at the equivalent `.md`
URL (for example,
`https://crossplane-contrib.github.io/provider-keycloak/docs/using/resources/realms.md`).
Section landing pages and the home page instead use `index.md` beneath their HTML
URLs. Empty taxonomy pages are disabled to keep the sitemap focused on documentation.
HTML advertises both the Markdown alternative and the site's `llms.txt` index,
with a visible index link in the banner at the top of every page;
Markdown outputs also link to that index.

Run `make docs-gen` from the repository root after changing content to regenerate
`docs/static/llms.txt` and `docs/static/llms-full.txt`. The index includes landing
pages as well as guides and links directly to their Markdown representations.
`make docs-freshness-check` verifies that the committed indexes are current.

The PR preview workflow regenerates both indexes with `DOCS_BASE_URL` set to
the same base URL passed to Hugo. Committed indexes always use the production
URL. Discovery links in Hugo templates use the active site base URL.
Markdown output expands Hugo shortcodes so card titles, descriptions, and hero
content are available rather than leaving unprocessed template syntax.

GitHub Pages serves static files and does not support `Accept: text/markdown`
content negotiation. Clients must request the advertised Markdown URL.
Passing a content-negotiation check on the HTML URLs requires a different host
or a reverse proxy in front of the site: it must serve the corresponding Markdown
file when Markdown is preferred in `Accept`, return `Content-Type: text/markdown`,
and set `Vary: Accept` so caches keep HTML and Markdown responses separate.
Hugo output templates alone cannot change GitHub Pages' response handling.

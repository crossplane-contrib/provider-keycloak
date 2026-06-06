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

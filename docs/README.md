# Documentation

This directory contains the provider-keycloak documentation site built with [Docusaurus](https://docusaurus.io/).

## Local Development

```bash
cd docs
npm install
npm start
```

This starts a local development server at `http://localhost:3000/provider-keycloak/`.

## Build

```bash
npm run build
```

## Deployment

The documentation is automatically deployed to GitHub Pages when changes are pushed to the `main` branch via the `.github/workflows/deploy-docs.yml` workflow.

Live site: https://crossplane-contrib.github.io/provider-keycloak/

# Local Docs Development

[llms.txt](https://crossplane-contrib.github.io/provider-keycloak/pr-preview/pr-760/llms.txt)


Run the documentation site locally:

```bash
cd docs
hugo server --buildDrafts
```

Build the static site:

```bash
cd docs
hugo --minify
```

The generated site is written to `docs/public/`.



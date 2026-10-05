# Local Docs Development

[llms.txt](https://crossplane-contrib.github.io/provider-keycloak/llms.txt)


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



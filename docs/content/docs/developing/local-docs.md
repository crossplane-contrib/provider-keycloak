---
title: Local Docs Development
---

# Local Docs Development

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

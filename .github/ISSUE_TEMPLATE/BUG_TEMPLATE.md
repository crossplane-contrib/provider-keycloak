---
name: Bug Report
about: Use this template to track bugs
title: "[BUG]"
labels: bug, needs triage
assignees: 
type: bug
---

## Metadata

> [!TIP]
> Please always try to use the latest version of the provider before reporting a bug.

- keycloak version: `<your keycloak version>`
- crossplane version: `<your crossplane version>`
- keycloak provider version:  `<your keycloak provider version>`

## Reproducible example

> [!TIP]
> We always need a reproducible example to investigate your issue. 
> The easiest for us is if you setup the local dev environment and provide the YAML file you used to create the resource that is causing the error, see [local-environment](https://github.com/crossplane-contrib/provider-keycloak?tab=readme-ov-file#local-environment). This is not required tho.

- Here is my full YAML file that I used to create the resource:
```yaml
<-- Paste your YAML here -->
``` 


## What do you expect? 
- Describe what you expected to happen.


## What is the error? 

> [!TIP]
> Include any relevant error messages, logs. 
> Run `kubectl describe <resource>` and include any relevant events.


## Does this error also happen if you are using the keycloak terraform provider? 

- If yes, you might have found an issue with the terraform provider itself and should report it there.
- If no, please provide the terraform code and steps to reproduce the error: 

```hcl
<your terraform code here>
```


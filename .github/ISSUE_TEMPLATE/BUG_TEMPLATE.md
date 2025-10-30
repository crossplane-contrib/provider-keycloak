---
name: Bug Report
about: Use this template to track bugs
title: "[BUG]"
labels: bug, needs triage
assignees: 
---

## Metadata
- keycloak version: `<your keycloak version>`
- crossplane version: `<your crossplane version>`
- keycloak provider version:  `<your keycloak provider version>`

## Reproducible example
- Here is my full YAML file that I used to create the resource:

```yaml
<-- Paste your YAML here -->
``` 


## What do you expect? 
- Describe what you expected to happen.


## What is the error? 
- Include any relevant error messages, logs. 
- run `kubectl describe <resource>` and include any relevant events.


## Does this error also happen if you are using the keycloak terraform provider? 

- If yes, please provide the terraform code and steps to reproduce the error: 

```hcl
<your terraform code here>
``` 


---
sidebar_position: 3
title: Troubleshooting
description: Common issues and how to resolve them
---

# Troubleshooting

## Common Issues

### Resource Stuck in "Creating" State

**Symptoms**: Resource shows `SYNCED: False` and never becomes `READY`.

**Diagnosis**:
```bash
kubectl describe <resource-type> <resource-name>
```

Look at the `Events` section and `status.conditions` for error messages.

**Common causes**:
- Invalid `providerConfigRef` — ensure the referenced `ProviderConfig` exists
- Incorrect credentials — verify username/password or client secret
- Network connectivity — the provider pod cannot reach the Keycloak URL
- Invalid field values — check the resource spec against Keycloak's requirements

### Authentication Failures

**Symptoms**: Events show `401 Unauthorized` or `403 Forbidden`.

**Steps**:
1. Verify the credentials secret exists and has correct data:
   ```bash
   kubectl get secret keycloak-credentials -n crossplane-system -o jsonpath='{.data.credentials}' | base64 -d
   ```
2. Test connectivity from the provider pod:
   ```bash
   kubectl exec -it -n crossplane-system <provider-pod> -- curl -k https://keycloak.example.com
   ```
3. Confirm the `client_id` has admin privileges in Keycloak

### Drift Detection Loops

**Symptoms**: Resource keeps updating even when no changes are made.

**Common causes**:
- Fields with server-side defaults that differ from your spec
- Timestamp or ordering differences

**Solution**: Ensure your spec exactly matches the desired state. Use `kubectl describe` to compare `spec.forProvider` with `status.atProvider`.

### Unexpected `make generate` Diffs

**Symptoms**: `make generate` produces unexpectedly large or stale diffs.

**Common cause**: Stale local generator cache/artifacts.

**Solution**:
1. Remove `.work/` and `config/schema.json`.
2. Run `make generate` again.

### Provider Pod CrashLoopBackOff

**Steps**:
1. Check pod logs:
   ```bash
   kubectl logs -n crossplane-system -l pkg.crossplane.io/revision --tail=100
   ```
2. Common causes:
   - Insufficient memory (increase via `DeploymentRuntimeConfig`)
   - Missing CRDs (reinstall the provider)

### TLS Certificate Errors

**Symptoms**: `x509: certificate signed by unknown authority`

**Solutions**:
- Add the CA certificate to the credentials using `root_ca_certificate`
- Or mount the CA certificate into the provider pod via `DeploymentRuntimeConfig`

## Useful Commands

```bash
# List all Keycloak CRDs
kubectl get crd | grep keycloak.crossplane.io

# Check provider health
kubectl get providers.pkg.crossplane.io

# View provider logs
kubectl logs -n crossplane-system -l pkg.crossplane.io/revision --tail=50

# Describe a specific resource for debugging
kubectl describe realm my-realm

# List all managed resources
kubectl get managed -l crossplane.io/provider=provider-keycloak
```

## Getting Help

- [Open an issue](https://github.com/crossplane-contrib/provider-keycloak/issues) on GitHub
- Join the [Crossplane Slack](https://slack.crossplane.io/) community
- Check the [Resource Reference](/docs/using/resources/) for CRD documentation and examples

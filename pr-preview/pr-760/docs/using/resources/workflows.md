# Workflows

[llms.txt](https://crossplane-contrib.github.io/provider-keycloak/pr-preview/pr-760/llms.txt)


Use workflows when you want Keycloak 26.5&#43; to react automatically to realm events. They are a good fit for onboarding notifications, password-policy enforcement, or custom event-driven logic when users are created, updated, or perform specific actions. Key fields are `name`, `enabled`, `on` for the trigger event, `step` for the ordered actions, and `realmRef` for the target realm.

## API Reference

| Kind | API Group | Terraform | CRD Explorer |
|------|-----------|-----------|---|
| `Workflow` | `workflow.keycloak.crossplane.io/v1alpha1` | [`keycloak_workflow`](https://registry.terraform.io/providers/keycloak/keycloak/latest/docs/resources/workflow) | [View CRD Schema](https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak/latest/resources/workflow.keycloak.crossplane.io/Workflow/v1alpha1) |

## Working YAML examples

### Workflow

```yaml
apiVersion: workflow.keycloak.crossplane.io/v1alpha1
kind: Workflow
metadata:
  name: onboarding
spec:
  deletionPolicy: Delete
  forProvider:
    enabled: true
    name: onboarding-new-users
    &#34;on&#34;: user_created
    realmRef:
      name: &#34;dev&#34;
      policy:
        resolve: Always
    step:
      - config:
          message: &#34;Welcome to ${realm.displayName}!&#34;
        uses: notify-user
  providerConfigRef:
    name: &#34;keycloak-provider-config&#34;
```

## Related Resources

- [Realms](./realms.md)
- [Authentication Flows](./authentication-flows.md)
- [Users](./users.md)



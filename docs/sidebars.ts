import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docsSidebar: [
    {
      type: 'category',
      label: 'Getting Started',
      items: [
        'getting-started/installation',
        'getting-started/configuration',
        'getting-started/first-realm',
      ],
      collapsed: false,
    },
    {
      type: 'category',
      label: 'Resources',
      items: [
        'resources/realms',
        'resources/clients',
        'resources/users',
        'resources/roles',
        'resources/groups',
        'resources/protocol-mappers',
        'resources/identity-providers',
        'resources/user-federation',
      ],
    },
    {
      type: 'category',
      label: 'Guides',
      items: [
        'guides/sso-with-argocd',
        'guides/kubernetes-oidc',
        'guides/ldap-integration',
        'guides/external-secrets-operator',
        'guides/end-to-end-oidc-kind',
      ],
    },
    {
      type: 'category',
      label: 'Reference',
      items: [
        'reference/provider-config',
        'reference/credentials',
        'reference/troubleshooting',
      ],
    },
  ],
};

export default sidebars;

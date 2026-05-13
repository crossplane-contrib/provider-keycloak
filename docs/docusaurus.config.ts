import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'Provider Keycloak',
  tagline: 'Crossplane provider for declarative Keycloak management',
  favicon: 'img/favicon.ico',

  future: {
    v4: true,
  },

  url: 'https://crossplane-contrib.github.io',
  baseUrl: process.env.DOCUSAURUS_BASE_URL || '/provider-keycloak/',

  organizationName: 'crossplane-contrib',
  projectName: 'provider-keycloak',

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          editUrl:
            'https://github.com/crossplane-contrib/provider-keycloak/tree/main/docs/',
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themes: [
    [
      require.resolve("@easyops-cn/docusaurus-search-local"),
      {
        hashed: true,
        language: ["en"],
        indexBlog: false,
        docsRouteBasePath: "/docs",
      },
    ],
  ],

  themeConfig: {
    colorMode: {
      defaultMode: 'light',
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'Provider Keycloak',
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docsSidebar',
          position: 'left',
          label: 'Documentation',
        },
        {
          href: 'https://github.com/crossplane-contrib/provider-keycloak',
          label: 'GitHub',
          position: 'right',
        },
        {
          href: 'https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak',
          label: 'Marketplace',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Documentation',
          items: [
            {label: 'Getting Started', to: '/docs/getting-started/installation'},
            {label: 'Resources', to: '/docs/resources/realms'},
            {label: 'Guides', to: '/docs/guides/sso-with-argocd'},
          ],
        },
        {
          title: 'Community',
          items: [
            {label: 'GitHub Issues', href: 'https://github.com/crossplane-contrib/provider-keycloak/issues'},
            {label: 'Crossplane Slack', href: 'https://slack.crossplane.io/'},
          ],
        },
        {
          title: 'More',
          items: [
            {label: 'Crossplane Docs', href: 'https://docs.crossplane.io/'},
            {label: 'Keycloak Docs', href: 'https://www.keycloak.org/documentation'},
            {label: 'Upbound Marketplace', href: 'https://marketplace.upbound.io/providers/crossplane-contrib/provider-keycloak'},
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} Crossplane Contributors. Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['yaml', 'bash', 'json'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;

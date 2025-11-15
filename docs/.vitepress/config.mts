import { useSidebar } from 'vitepress-openapi'
import spec from '../src/openapi.json' with { type: 'json' }
import { defineConfig } from 'vitepress'
import { withMermaid } from "vitepress-plugin-mermaid";

const sidebar = useSidebar({
  spec: spec,
  linkPrefix: '/operations/'
})

export default withMermaid(
  defineConfig({
    title: "Nixopus Docs",
    description: "documentation",
    head: [['link', { rel: 'icon', href: '/favicon.png' }]],
    themeConfig: {
      search: {
        provider: 'local',
        options: {
          locales: {
            zh: {
              translations: {
                button: {
                  buttonText: '搜索',
                  buttonAriaLabel: '搜索'
                },
                modal: {
                  displayDetails: '显示详细列表',
                  resetButtonTitle: '重置搜索',
                  backButtonTitle: '关闭搜索',
                  noResultsText: '没有结果',
                  footer: {
                    selectText: '选择',
                    selectKeyAriaLabel: '输入',
                    navigateText: '导航',
                    navigateUpKeyAriaLabel: '上箭头',
                    navigateDownKeyAriaLabel: '下箭头',
                    closeText: '关闭',
                    closeKeyAriaLabel: 'esc'
                  }
                }
              }
            }
          }
        }
      },
      editLink: {
        pattern: 'https://github.com/raghavyuva/nixopus/edit/master/docs/:path',
        text: "Edit this page on Github"
      },
      nav: [
        { text: 'Get Started', link: '/install/index.md' },
        { text: "CLI", link: '/cli/index.md' },
        { text: 'Blog', link: '/blog/' },
      ],
      footer: {
        message: `<img src="https://madewithlove.now.sh/in?heart=true&colorA=%23ff671f&colorB=%23046a38&text=Open%20Source" alt="Made with love" style="display:block;margin:0 auto;" /><br>Released under the Functional Source License (FSL)`,
        copyright: 'Copyright © 2025–present Nixopus'
      },
      sidebar: [
        {
          text: "Get Started",
          items: [
            { text: "Introduction", link: "/introduction/index.md" },
            { text: "Installation", link: "/install/index.md" },
            { text: "Roadmap", link: "https://roadmap.nixopus.com" },
            { text: "Preferences", link: "/preferences/index.md" }
          ]
        },
        {
          text: 'Features',
          items: [
            { text: "Extensions", link: "/extensions/index.md" },
            { text: "Hosting Projects", link: "/self-host/index.md" },
            { text: 'Terminal', link: '/terminal/index.md' },
            { text: "File Manager", link: "/file-manager/index.md" },
            { text: "Notifications", link: "/notifications/index.md" }
          ]
        },
        {
          text: 'CLI',
          items: [
            { text: 'Overview', link: '/cli/index.md' },
            { text: 'Installation', link: '/cli/installation.md' },
            { text: 'Configuration', link: '/cli/config.md' },
            { text: 'Reference', link: '/cli/cli-reference.md' }
          ]
        },
        {
          text: 'Blog',
          items: [
            { text: 'Latest Posts', link: '/blog/' }
          ]
        },
        {
          text: 'Development',
          items: [
            { text: 'Development Guide', link: '/contributing/index.md' },
            { text: "Workflows", link: "/workflows/index.md" },
            { text: "Code of Conduct", link: "/code-of-conduct/index.md" },
          ]
        },
        {
          text: "Policy",
          items: [
            { text: "Changelogs", link: "https://github.com/raghavyuva/nixopus/releases" },
            { text: "License", link: "/license/index.md" },
            { text: "Privacy Policy", link: "/privacy-policy/index.md" },
            { text: "Credits", link: "/credits/index.md" }
          ]
        },
        {
          text: "Support",
          items: [
            { text: "Sponsor", link: '/sponsor/index.md' },
            { text: "Contact", link: '/contact/index.md' }
          ]
        },
        {
          text: "API Reference",
          items: [
            ...sidebar.generateSidebarGroups().map((group) => ({
              ...group,
              collapsed: true,
              items: group.items.map((item) => {
                const operationId = item.link.split('/').pop() || '';
                const endpointMatch = operationId.match(/^[A-Z]+_(.+)$/);
                const endpoint = endpointMatch ? endpointMatch[1] : operationId;
                const httpMethod = extractHttpMethods(item.link);
                const methodSpan = `
                <span class="OASidebarItem group/oaSidebarItem">
                  <span class="OASidebarItem-badge OAMethodBadge--${httpMethod}">${httpMethod.toUpperCase()}</span>
                  <span class="OASidebarItem-text text">${endpoint}</span>
                </span>`;

                return {
                  ...item,
                  link: `${item.link}`,
                  text: methodSpan
                };
              }),
            })),
          ]
        }
      ],
      socialLinks: [
        { icon: 'github', link: 'https://github.com/raghavyuva/nixopus' },
        { icon: "discord", link: "https://discord.gg/skdcq39Wpv" },
      ]
    }
  })
)

function extractHttpMethods(link: string): string {
  const operationPath = link.replace('/operations/', '');
  const methodMatch = operationPath.match(/^(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)_/);

  return methodMatch?.[1].toLowerCase() || 'METHOD_MISSING';
}

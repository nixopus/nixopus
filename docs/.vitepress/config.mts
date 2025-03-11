import { useSidebar } from 'vitepress-openapi'
import spec from '../public/openapi.json' assert { type: 'json' }
import { defineConfigWithTheme } from 'vitepress'

const sidebar = useSidebar({
  spec,
})

export default defineConfigWithTheme({
  title: "Nixopus Docs",
  description: "documentation",
  lastUpdated: true,

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
      pattern: 'https://github.com/nixopus/nixopus/edit/main/docs/:path',
      text: "Edit this page on Github"
    },
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Get Started', link: '/install/index.md' }
    ],
    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2024-present Nixopus'
    },
    sidebar: [
      {
        text: "Get Started",
        items: [
          { text: "Introduction", link: "/introduction/index.md" },
          { text: "Installation", link: "/install/index.md" },
          { text: "Themes and Fonts", link: "/themes-and-fonts/index.md" }
        ]
      },
      {
        text: 'Features',
        items: [
          { text: 'Terminal', link: '/terminal/index.md' },
          { text: "File Manager", link: "/file-manager/index.md" },
          { text: "Hosting Projects", link: "/self-host/index.md" },
          { text: "Marketplace", link: "/marketplace/index.md" },
          { text: "Cron Jobs", link: "/cron-jobs/index.md" },
          { text: "Firewall", link: "/firewall/index.md" },
          { text: "Mail", link: "/mail/index.md" },
        ]
      },
      {
        text: 'Development',
        items: [
          { text: 'Contribution', link: '/contributing/index.md' },
          { text: "Code of Conduct", link: "/code-of-conduct/index.md" },
          { text: "Change Log", link: "/changelog/index.md" },
          { text: "License", link: "/license/index.md" },
          { text: "Privacy Policy", link: "/privacy-policy/index.md" },
          { text: "Credits", link: "/credits/index.md" },
          { text: "Self Hosting", link: "/hosting/index.md" }
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
              const endpoint = item.link.split('/').pop();
              const methodSpan = `
                <span class="OASidebarItem group/oaSidebarItem">
                  <span class="OASidebarItem-badge OAMethodBadge--${extractHttpMethods(item.text)}">${extractHttpMethods(item.text)}</span>
                  <span class="OASidebarItem-text text">${endpoint}</span>
                </span>`;
              
              return {
                ...item,
                link: `${item.link}`,
                text:methodSpan
              };
            }),
          })),
        ]
      }
    ],
    socialLinks: [
      { icon: 'github', link: 'https://github.com/nixopus/nixopus' },
      { icon: "discord", link: "https://github.com/nixopus/nixopus" },
      { icon: "x", link: "https://github.com/nixopus/nixopus" },
      { icon: "youtube", link: "https://github.com/nixopus/nixopus" },
    ]
  }
})

function extractHttpMethods(text) {
  const methodRegex = /OAMethodBadge--(\w+)/g;
  const methods : string[] = [];
  let match;

  while ((match = methodRegex.exec(text)) !== null) {
      methods.push(match[1].toUpperCase());
  }

  return methods[0]
}

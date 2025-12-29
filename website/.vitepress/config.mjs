import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'Promenade Platform',
  description: 'Modern backend platform for customer management, orders, and business workflows with clean DDD architecture',
  base: '/promenade/',
  
  // Ignore dead links (localhost URLs and pages under construction)
  ignoreDeadLinks: 'localhostLinks',
  
  head: [
    ['link', { rel: 'icon', href: '/promenade/favicon.ico' }],
    ['meta', { name: 'theme-color', content: '#3b82f6' }],
  ],

  themeConfig: {
    logo: '/logo.svg',
    
    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      { 
        text: 'Concepts',
        items: [
          { text: 'Clean Architecture', link: '/concepts/clean-architecture' },
          { text: 'Event-Driven', link: '/concepts/event-driven' },
          { text: 'Bounded Contexts', link: '/concepts/bounded-contexts' },
        ]
      },
      { text: 'Contexts', link: '/contexts/identity' },
      { text: 'Packages', link: '/packages/bus' },
      {
        text: 'v2.0',
        items: [
          { text: 'Changelog', link: 'https://github.com/basilex/promenade/releases' },
          { text: 'Contributing', link: '/guide/contributing' }
        ]
      }
    ],

    sidebar: {
      '/guide/': [
        {
          text: 'Introduction',
          items: [
            { text: 'Getting Started', link: '/guide/getting-started' },
            { text: 'Quick Start', link: '/guide/quick-start' },
            { text: 'Architecture', link: '/guide/architecture' },
          ]
        },
        {
          text: 'Guides',
          items: [
            { text: 'Testing Patterns', link: '/guide/testing' },
            { text: 'RBAC', link: '/guide/rbac' },
            { text: 'Health Checks', link: '/docs/guides/health-checks' },
            { text: 'Rate Limiting', link: '/docs/guides/rate-limiting' },
            { text: 'API Reference', link: '/guide/api-reference' },
            { text: 'Contributing', link: '/guide/contributing' },
          ]
        }
      ],
      
      '/concepts/': [
        {
          text: 'Core Concepts',
          items: [
            { text: 'Clean Architecture', link: '/concepts/clean-architecture' },
            { text: 'Event-Driven Architecture', link: '/concepts/event-driven' },
            { text: 'Bounded Contexts', link: '/concepts/bounded-contexts' },
          ]
        }
      ],
      
      '/reference/': [
        {
          text: 'Technical Reference',
          items: [
            { text: 'Test Coverage Report', link: '/reference/test-coverage-report' },
            { text: 'Bus Test Coverage', link: '/reference/bus-test-coverage' },
            { text: 'Refactoring Roadmap', link: '/reference/refactoring-roadmap' },
          ]
        }
      ],
      
      '/contexts/': [
        {
          text: 'Bounded Contexts',
          items: [
            { text: 'Overview', link: '/contexts/' },
            { text: 'Identity', link: '/contexts/identity' },
            { text: 'Shared', link: '/contexts/shared' },
            { text: 'Customer Management', link: '/contexts/customer' },
            { text: 'Order Management', link: '/contexts/order' },
            { text: 'Billing', link: '/contexts/billing' },
          ]
        }
      ],
      
      '/packages/': [
        {
          text: 'Package Library',
          items: [
            { text: 'Overview', link: '/packages/' },
            { text: 'Event Bus', link: '/packages/bus' },
            { text: 'JWT', link: '/packages/jwt' },
            { text: 'Logger', link: '/packages/logger' },
            { text: 'UUID v7', link: '/packages/uuidv7' },
            { text: 'Value Objects', link: '/packages/valueobject' },
            { text: 'Response', link: '/packages/response' },
          ]
        }
      ]
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/basilex/promenade' }
    ],

    footer: {
      message: 'Built with Domain-Driven Design and Go',
      copyright: 'Copyright © 2024-2025 Promenade Platform'
    },

    search: {
      provider: 'local'
    },

    editLink: {
      pattern: 'https://github.com/basilex/promenade/edit/dev/website/:path',
      text: 'Edit this page on GitHub'
    },

    lastUpdated: {
      text: 'Updated at',
      formatOptions: {
        dateStyle: 'full',
        timeStyle: 'medium'
      }
    }
  }
})

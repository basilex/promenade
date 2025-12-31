import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'Promenade Platform',
  description: 'Modern backend platform for customer management, orders, and business workflows with clean DDD architecture',
  base: '/promenade/',
  
  // Ignore dead links (internal README files from docs/ don't exist in website/)
  ignoreDeadLinks: true,
  
  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/promenade/favicon.svg' }],
    ['meta', { name: 'theme-color', content: '#818cf8' }],
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
          { text: 'Customer Management', link: '/concepts/customer-management' },
          { text: 'Deal Management', link: '/concepts/deal-management' },
          { text: 'Order Management', link: '/concepts/order-management' },
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
          text: 'Implementation Guides',
          items: [
            { text: 'RBAC & Authorization', link: '/guide/rbac' },
            { text: 'Rate Limiting', link: '/guide/rate-limiting' },
            { text: 'Health Checks', link: '/guide/health-checks' },
            { text: 'Caching Layer', link: '/guide/caching' },
            { text: 'Testing Patterns', link: '/guide/testing-patterns' },
            { text: 'Testing Quick Reference', link: '/guide/testing-quick-reference' },
          ]
        },
        {
          text: 'Reference',
          items: [
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
            { text: 'Customer Management', link: '/concepts/customer-management' },
            { text: 'Deal Management', link: '/concepts/deal-management' },
            { text: 'Order Management', link: '/concepts/order-management' },
          ]
        }
      ],
      
      '/reference/': [
        {
          text: 'Technical Reference',
          items: [
            { text: 'Test Coverage Report', link: '/reference/test-coverage-report' },
            { text: 'Bus Test Coverage', link: '/reference/bus-test-coverage' },
          ]
        }
      ],
      
      '/contexts/': [
        {
          text: 'Bounded Contexts',
          items: [
            { text: 'Overview', link: '/contexts/' },
            { text: 'Identity Context', link: '/contexts/identity' },
            { text: 'Shared Context', link: '/contexts/shared' },
            { text: 'Customer Management', link: '/contexts/customer' },
            { text: 'Order Management', link: '/contexts/order' },
          ]
        },
        {
          text: 'Planned Contexts',
          items: [
            { text: 'Billing', link: '/contexts/billing' },
            { text: 'Warehouse', link: '/contexts/warehouse' },
          ]
        }
      ],
      
      '/packages/': [
        {
          text: 'Infrastructure',
          items: [
            { text: 'Overview', link: '/packages/' },
            { text: 'Event Bus', link: '/packages/bus' },
            { text: 'JWT Authentication', link: '/packages/jwt' },
            { text: 'Cache', link: '/packages/cache' },
            { text: 'Logger', link: '/packages/logger' },
          ]
        },
        {
          text: 'Domain Primitives',
          items: [
            { text: 'UUID v7', link: '/packages/uuidv7' },
            { text: 'Value Objects', link: '/packages/valueobject' },
            { text: 'Aggregates', link: '/packages/aggregate' },
            { text: 'Saga Pattern', link: '/packages/saga' },
          ]
        },
        {
          text: 'Utilities',
          items: [
            { text: 'Response Helpers', link: '/packages/response' },
            { text: 'JSONB Utilities', link: '/packages/jsonb' },
            { text: 'Migrations', link: '/packages/migration' },
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

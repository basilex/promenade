# Promenade Platform Documentation

🚀 **Built with VitePress** - Modern, fast, and beautiful documentation for Promenade Platform.

## Quick Start

### Development

```bash
cd website
npm install
npm run docs:dev
```

Visit **http://localhost:5173/**

### Build for Production

```bash
npm run docs:build
npm run docs:preview
```

### Deploy to GitHub Pages

**Manual Deploy:**

```bash
./deploy.sh
```

This will:
1. Build the VitePress site
2. Create/update the gh-pages branch
3. Push to GitHub
4. Site goes live at: https://basilex.github.io/promenade/

**Note**: Links work with `/promenade/` base path configured in `.vitepress/config.mjs`

**Automated via GitHub Actions** (coming soon)

## Project Structure

```
website/
├── .vitepress/
│   ├── config.mjs           # VitePress configuration
│   └── theme/              # Custom theme (optional)
├── index.md                # Landing page (Hero + Features)
├── guide/                  # Documentation guides
│   ├── getting-started.md
│   ├── architecture.md
│   ├── testing.md
│   └── api-reference.md
├── contexts/               # Bounded Contexts docs
│   ├── identity.md
│   ├── shared.md
│   └── customer.md
├── packages/               # Package library docs
│   ├── bus.md
│   ├── jwt.md
│   └── logger.md
└── public/                 # Static assets (images, icons)
```

## Adding New Pages

1. Create markdown file in appropriate section:
   ```bash
   touch website/guide/new-topic.md
   ```

2. Add to sidebar in `.vitepress/config.mjs`:
   ```js
   sidebar: {
     '/guide/': [
       {
         text: 'Guide',
         items: [
           { text: 'New Topic', link: '/guide/new-topic' }
         ]
       }
     ]
   }
   ```

## Markdown Features

### Code Blocks with Syntax Highlighting

```go
func Example() {
    fmt.Println("Hello, Promenade!")
}
```

### Custom Containers

::: info
This is an info box.
:::

::: tip
This is a tip.
:::

::: warning
This is a warning.
:::

::: danger
This is a dangerous warning.
:::

### Code Groups

::: code-group
```go [usecase.go]
func (uc *UseCase) Execute() error {
    return nil
}
```

```go [handler.go]
func (h *Handler) Handle(c *gin.Context) {
    c.JSON(200, gin.H{"status": "ok"})
}
```
:::

## Configuration

Edit `.vitepress/config.mjs`:

```js
export default defineConfig({
  title: 'Promenade Platform',
  description: 'Your description',
  themeConfig: {
    nav: [...],
    sidebar: {...}
  }
})
```

## Resources

- [VitePress Documentation](https://vitepress.dev/)
- [Markdown Extensions](https://vitepress.dev/guide/markdown)
- [Theme Configuration](https://vitepress.dev/reference/default-theme-config)

---

**Built with ❤️ and VitePress**

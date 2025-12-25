# Promenade Website

Static website built with [Hugo](https://gohugo.io/) - Fast, modern static site generator written in Go.

## 🚀 Features

- **Tailwind-inspired design** - Clean, minimal, and responsive
- **Multilingual** - English, Ukrainian (Українська), German (Deutsch)
- **Fast** - Static HTML generation, ~2ms per page
- **Documentation integration** - Auto-syncs from `../docs/`
- **GitHub Discussions** - Community testimonials and feedback
- **Zero config deployment** - GitHub Actions + GitHub Pages

## 📁 Structure

```
website/
├── content/           # Page content (Markdown)
│   ├── _index.md      # Homepage
│   ├── features/      # Features pages
│   └── docs/          # Documentation (linked from ../docs/)
├── layouts/           # HTML templates
│   ├── _default/      # Default layouts
│   │   ├── baseof.html    # Base template
│   │   ├── list.html      # List pages
│   │   └── single.html    # Single pages
│   ├── partials/      # Reusable components
│   │   ├── header.html    # Site header
│   │   └── footer.html    # Site footer
│   └── index.html     # Homepage layout
├── static/            # Static assets
│   ├── css/           # Stylesheets
│   │   └── main.css   # Main stylesheet
│   ├── js/            # JavaScript (if needed)
│   └── favicon.svg    # Site icon
├── hugo.toml          # Hugo configuration
└── README.md          # This file
```

## 🛠️ Development

### Prerequisites

- **Hugo Extended** v0.153.2+ ([install](https://gohugo.io/installation/))
- **Go** 1.21+ (optional, for Hugo modules)

### Quick Start

```bash
# Navigate to website directory
cd website

# Start development server
hugo server --buildDrafts

# Or use the shorthand
hugo server -D

# Server will be available at http://localhost:1313
```

### Build for Production

```bash
# Build static site
hugo --gc --minify

# Output will be in public/ directory
```

### Adding Content

```bash
# Create new feature page
hugo new content features/my-feature.md

# Create new documentation page
hugo new content docs/my-guide.md
```

## 🌍 Multilingual Support

The site supports three languages:

- **English (en)** - Default language
- **Ukrainian (uk)** - Content in `content/uk/`
- **German (de)** - Content in `content/de/`

### Adding Translated Content

1. Create language-specific content directory:

   ```bash
   mkdir -p content/uk/features
   mkdir -p content/de/features
   ```

2. Add translated content:

   ```bash
   cp content/features/my-feature.md content/uk/features/my-feature.md
   # Edit content/uk/features/my-feature.md with Ukrainian translation
   ```

3. Hugo will automatically detect and link translations

## 📚 Documentation Integration

Documentation is sourced from the main `docs/` directory:

```bash
# Option 1: Symbolic link (development)
ln -s ../../docs content/docs

# Option 2: Copy files (production build)
cp -r ../docs content/docs

# Option 3: Hugo mount (in hugo.toml)
[module]
  [[module.mounts]]
    source = "../docs"
    target = "content/docs"
```

Currently using **Option 3** for seamless integration.

## 🎨 Styling

The site uses a custom Tailwind-inspired CSS framework:

- **No build step** - Pure CSS, no Tailwind CLI needed
- **CSS variables** - Easy theming via `static/css/main.css`
- **Responsive** - Mobile-first design
- **Minimal** - ~8KB CSS (gzipped: ~2KB)

### Customizing Colors

Edit `static/css/main.css`:

```css
:root {
  --color-primary: #0ea5e9; /* Sky blue */
  --color-secondary: #8b5cf6; /* Purple */
  /* ... more variables ... */
}
```

## 🚀 Deployment

### GitHub Pages (Automatic)

1. **Enable GitHub Pages**:

   - Go to repository Settings → Pages
   - Source: GitHub Actions

2. **Push to `main` or `dev`**:

   ```bash
   git add website/
   git commit -m "feat: add website"
   git push origin dev
   ```

3. **GitHub Actions** will build and deploy automatically
4. Site will be live at `https://<username>.github.io/<repo>/`

### Custom Domain

1. Add `CNAME` file:

   ```bash
   echo "promenade.com.ua" > static/CNAME
   ```

2. Configure DNS:

   - Add CNAME record: `promenade.com.ua` → `<username>.github.io`
   - Or A records for apex domain

3. Update `hugo.toml`:
   ```toml
   baseURL = 'https://promenade.com.ua/'
   ```

### Other Hosting

**Netlify:**

```bash
# netlify.toml
[build]
  command = "cd website && hugo --gc --minify"
  publish = "website/public"
```

**Vercel:**

```json
{
  "buildCommand": "cd website && hugo --gc --minify",
  "outputDirectory": "website/public"
}
```

**Cloudflare Pages:**

- Build command: `cd website && hugo --gc --minify`
- Build output: `website/public`

## 🔧 Configuration

Main configuration in `hugo.toml`:

```toml
baseURL = 'https://promenade.com.ua/'
title = 'Promenade - Production-Ready REST API Framework'

[params]
  description = 'Production-ready REST API...'
  github = 'https://github.com/basilex/promenade'

[languages]
  [languages.en]
    weight = 1
  [languages.uk]
    weight = 2
  [languages.de]
    weight = 3
```

## 📊 Analytics (Optional)

Add analytics in `layouts/partials/analytics.html`:

```html
<!-- Google Analytics -->
<script async src="https://www.googletagmanager.com/gtag/js?id=G-XXXXXXXXXX"></script>

<!-- Plausible (privacy-friendly) -->
<script defer data-domain="promenade.com.ua" src="https://plausible.io/js/script.js"></script>
```

Then include in `layouts/_default/baseof.html`:

```html
{{ partial "analytics.html" . }}
```

## 🤝 Contributing

1. Make changes in `website/` directory
2. Test locally: `hugo server -D`
3. Commit and push
4. GitHub Actions will deploy automatically

## 📝 License

Same as Promenade - MIT License

## 🔗 Links

- **Live Site**: https://promenade.com.ua (after deployment)
- **Hugo Docs**: https://gohugo.io/documentation/
- **GitHub**: https://github.com/basilex/promenade
- **Discussions**: https://github.com/basilex/promenade/discussions

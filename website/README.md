# Promenade Website

Simple, single-page landing site for Promenade REST API Framework.

## Approach

**Single Source of Truth** = GitHub Repository

- All detailed documentation lives on GitHub
- Website is a beautiful business card with links to GitHub
- No content duplication = no synchronization issues
- Easy to maintain during active development

## Structure

```
website/
 hugo.toml              # Hugo config with multi-language support
 layouts/
    index.html         # Single-page landing (EN/UK/DE)
 i18n/                  # Translations
    en.yaml            # English
    uk.yaml            # Ukrainian
    de.yaml            # German
 static/                # Assets (images, favicon, etc.)
```

## Features

### One-Page Design

- **Hero Section** - Eye-catching title + CTA buttons
- **Features** - Key highlights (6 cards)
- **Stats** - 500+ tests, 8 modules, 100% passing
- **Modules** - Available modules (Free/Commercial)
- **Quick Start** - Code snippet + link to GitHub
- **Contact Form** - Simple contact form at bottom

### Multi-Language

- English (EN) - Primary
- Ukrainian () - Full translation
- German () - Full translation
- Language switcher in navigation

### Links to GitHub

- All "Learn More" → GitHub README
- Documentation → GitHub docs/
- Module details → GitHub internal/modules/

## Development

```bash
# Install Hugo
brew install hugo

# Start dev server
cd website
hugo server -D

# Build static site
hugo
```

## Deployment

Site is built automatically on push to `dev` branch via GitHub Actions.

**GitHub Pages URL:** https://basilex.github.io/promenade/

## Contact Form Setup

Form uses [Formspree](https://formspree.io/) for handling submissions.

1. Create free Formspree account
2. Get form endpoint ID
3. Update `layouts/index.html`:
   ```javascript
   fetch('https://formspree.io/f/YOUR_FORM_ID', {
   ```

## Design Principles

1. **Simple & Fast** - Single HTML page, no complex navigation
2. **Visual Appeal** - Modern gradient hero, hover effects
3. **Mobile-First** - Responsive design for all devices
4. **Performance** - Minimal JS, inline CSS, fast loading
5. **SEO Ready** - Proper meta tags, semantic HTML

## Why This Approach?

### Problem (Old Site)

- Internal links constantly breaking
- Content duplication between site/GitHub
- Hard to maintain during active development
- Inconsistent information

### Solution (New Site)

- Single landing page (візитка)
- Links to GitHub for details
- Easy to maintain
- Always consistent (GitHub is source of truth)
- Beautiful and functional

## What Changed?

**Removed:**

- `/features/` page - now cards on homepage
- `/modules/` page - now cards on homepage
- `/docs/` section - links to GitHub
- `/blog/` section - not needed yet
- Complex navigation - single page scroll

**Kept:**

- Beautiful hero section
- Feature highlights
- Module showcase
- Contact form
- Multi-language support

## Future Enhancements

- Add customer testimonials section
- Add use cases / case studies
- Blog when ready (separate Hugo content)
- Integration examples (after modules stabilize)

---

**Built with:** Hugo + Custom HTML/CSS
**Hosted on:** GitHub Pages
**Updated:** December 26, 2025

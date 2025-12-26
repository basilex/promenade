# Promenade Website - Changelog

## December 26, 2025 - Complete Redesign

### Philosophy Change

**Before:** Complex multi-page site with duplicated content
**After:** Simple single-page landing with links to GitHub

### Why This Change?

1. **Single Source of Truth** - All documentation on GitHub
2. **No Duplication** - No sync issues between site and GitHub
3. **Easier Maintenance** - During active development
4. **Always Consistent** - GitHub README is authoritative

### What Was Removed

- `/features/` - Multi-page feature documentation
- `/modules/` - Individual module pages
- `/docs/` - Documentation section (now links to GitHub)
- `/blog/` - Blog (not needed yet)
- Complex navigation menus
- Multiple HTML templates
- Theme system

### What Was Kept & Improved

- ✅ Beautiful hero section with gradient
- ✅ Feature highlights (6 cards)
- ✅ Module showcase (7 modules)
- ✅ Quick start code snippet
- ✅ Contact form
- ✅ Multi-language support (EN/UK/DE)
- ✅ Responsive mobile design
- ✅ Modern styling with hover effects

### New Structure

```
website/
├── hugo.toml           # Simple multi-language config
├── layouts/
│   └── index.html      # Single-page landing
├── i18n/               # Translations
│   ├── en.yaml
│   ├── uk.yaml
│   └── de.yaml
├── static/             # Minimal assets
└── README.md           # This file
```

### Technical Details

- **Framework:** Hugo (static site generator)
- **Pages:** 1 (single page with sections)
- **Languages:** 3 (English, Ukrainian, German)
- **External Links:** All detailed docs → GitHub
- **Contact Form:** Formspree integration
- **Hosting:** GitHub Pages

### Links Strategy

| Old Approach             | New Approach                                                         |
| ------------------------ | -------------------------------------------------------------------- |
| `/features/architecture` | https://github.com/basilex/promenade#architecture-overview           |
| `/modules/posts`         | https://github.com/basilex/promenade/tree/dev/internal/modules/posts |
| `/docs/quickstart`       | https://github.com/basilex/promenade#quick-start                     |
| `/docs/testing`          | https://github.com/basilex/promenade/blob/dev/docs/TESTING_GUIDE.md  |

### Benefits

1. **Fast Loading** - Single HTML page, inline CSS
2. **Easy Updates** - Change one file vs multiple pages
3. **No Broken Links** - Links go to stable GitHub URLs
4. **Multilingual** - Same structure in 3 languages
5. **Mobile First** - Responsive by design
6. **SEO Friendly** - Proper meta tags, semantic HTML

### Deployment

- **URL:** https://basilex.github.io/promenade/
- **Build:** GitHub Actions on push to `dev`
- **CDN:** GitHub Pages global CDN
- **SSL:** Automatic HTTPS

### Future Enhancements (Maybe)

- Customer testimonials section
- Use case examples
- Blog (when content is ready)
- Video demo embedding
- Live API playground

### Maintenance Notes

- Update stats when new modules added
- Update translations in sync
- Contact form uses Formspree (free tier)
- All feature descriptions link to GitHub
- No need to maintain internal documentation

---

**Old Site:** 20+ pages, complex navigation, duplicated content
**New Site:** 1 page, simple navigation, links to GitHub
**Result:** Faster, easier, more maintainable ✅

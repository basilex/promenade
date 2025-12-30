# Favicon Files

This directory contains favicon files for the Promenade Platform website.

## Current Files

- ✅ **favicon.svg** - Vector favicon (48x48, preferred for modern browsers)
  - Same visual style as main logo
  - Simplified "P" letter with gradient
  - Perfect for high-DPI displays

## Future Additions (Optional)

For maximum compatibility across all devices and browsers:

- **favicon.ico** - Classic 32x32 ICO format (for old browsers)
- **apple-touch-icon.png** - 180x180 PNG for iOS home screen

**Note**: Modern browsers (Chrome, Firefox, Safari, Edge) support SVG favicons natively.
The SVG favicon provides the best quality across all screen resolutions.

## Generation Commands

If PNG versions are needed in future:

```bash
# Convert SVG to PNG using ImageMagick
convert -background none -density 300 favicon.svg -resize 32x32 favicon-32x32.png
convert -background none -density 300 favicon.svg -resize 180x180 apple-touch-icon.png

# Convert PNG to ICO
convert favicon-32x32.png favicon.ico
```

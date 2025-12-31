#!/bin/bash
set -e

echo "🚀 Deploying VitePress site to GitHub Pages..."

# Build the site
echo "📦 Building site..."
npm run docs:build

# Navigate to build output
cd .vitepress/dist

# Create .nojekyll file (tells GitHub Pages not to use Jekyll)
touch .nojekyll

# Initialize git if not already initialized
if [ ! -d .git ]; then
  git init
  git checkout -b gh-pages
fi

# Add all files
git add -A

# Commit changes
git commit -m "Deploy VitePress site to GitHub Pages - $(date '+%Y-%m-%d %H:%M:%S')"

# Push to gh-pages branch
echo "🚢 Pushing to gh-pages..."
git push -f git@github.com:basilex/promenade.git gh-pages

echo "✅ Deployment complete!"
echo "🌐 Site will be available at: https://basilex.github.io/promenade/"

# GitHub Pages Setup Instructions

## Enable GitHub Pages with GitHub Actions

1. Go to your repository: https://github.com/basilex/promenade
2. Click **Settings** → **Pages** (left sidebar)
3. Under **Build and deployment**:
   - **Source**: Select **GitHub Actions** (not "Deploy from a branch")
4. Click **Save**

## Trigger First Deployment

The workflow will automatically deploy on:
- Push to `dev` or `main` branch
- Manual trigger via workflow_dispatch

### Option 1: Push Empty Commit (Recommended)

```bash
git commit --allow-empty -m "chore: Trigger GitHub Pages deployment"
git push origin dev
```

### Option 2: Manual Trigger

1. Go to **Actions** tab
2. Select **Deploy VitePress site to GitHub Pages** workflow
3. Click **Run workflow** → **Run workflow**

## Verify Deployment

1. Go to **Actions** tab
2. Check the latest workflow run
3. Wait for "Deploy VitePress site to GitHub Pages" to complete (~2 minutes)
4. Site will be available at: **https://basilex.github.io/promenade/**

## Troubleshooting

### Issue: "pages build and deployment" action not found
**Solution**: Make sure **Source** is set to **GitHub Actions** (not "Deploy from a branch")

### Issue: Permission denied
**Solution**: 
1. Go to **Settings** → **Actions** → **General**
2. Scroll to **Workflow permissions**
3. Select **Read and write permissions**
4. Check **Allow GitHub Actions to create and approve pull requests**
5. Click **Save**

### Issue: Build fails
**Solution**: Check the workflow logs in Actions tab

## Expected Workflow Output

```
✓ Build (build)
  - Checkout code
  - Setup Node.js
  - Install dependencies
  - Build VitePress
  - Upload artifact

✓ Deploy (deploy)
  - Deploy to GitHub Pages
```

## Site URL

After successful deployment, your site will be available at:
**https://basilex.github.io/promenade/**

The workflow is configured to:
- Build from `website/` directory
- Deploy `.vitepress/dist` folder
- Auto-deploy on push to `dev` or `main` branches

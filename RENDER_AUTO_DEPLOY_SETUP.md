# Render Auto-Deploy Setup Guide

## Step-by-Step: Enable Auto-Deployment on Render

### Step 1: Login to Render Dashboard
1. Go to https://dashboard.render.com/
2. Sign in with your account

### Step 2: Create New Web Service (If Not Already Created)

#### Option A: If Service Doesn't Exist Yet
1. Click **"New +"** button (top right)
2. Select **"Web Service"**
3. Click **"Connect GitHub"** (if not already connected)
4. Authorize Render to access your GitHub repositories
5. Find and select: **`UserMenagmentSystem_Backend`**
6. Click **"Connect"**

#### Option B: If Service Already Exists
1. Go to your existing service in the dashboard
2. Click **"Settings"** tab
3. Scroll to **"Build & Deploy"** section

### Step 3: Configure Auto-Deploy Settings

In the service settings, configure:

#### Basic Settings:
- **Name:** `zenithive-backend` (or your preferred name)
- **Region:** Oregon (or closest to your location)
- **Branch:** `main` ⚠️ **IMPORTANT: Must be "main"**
- **Root Directory:** Leave blank (uses repository root)

#### Build Settings:
- **Build Command:**
  ```bash
  go build -o ums-backend main.go
  ```
  
- **Start Command:**
  ```bash
  ./ums-backend
  ```

#### Auto-Deploy Settings:
- **Auto-Deploy:** ✅ **Enable this!**
  - Toggle should be **ON** (blue/green)
  - This makes Render automatically deploy on every push to `main` branch

### Step 4: Add Environment Variables

Click **"Environment"** tab and add these variables:

```
DB_URL=postgresql://user:password@host:port/database?sslmode=require
JWT_SECRET=your_secret_key_here_min_32_chars
PORT=8082
EMAIL_FROM=your-email@gmail.com
EMAIL_PASSWORD=your_app_password
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
```

**Important Notes:**
- Click **"Add Environment Variable"** for each one
- Don't use quotes around values
- For Gmail, use an [App Password](https://support.google.com/accounts/answer/185833)

### Step 5: Verify render.yaml is Detected

Render should automatically detect your `render.yaml` file. Verify:

1. In service settings, look for: **"Using render.yaml"**
2. If detected, it will show: ✅ **"Blueprint detected"**
3. Your `render.yaml` already has `autoDeploy: true` configured

### Step 6: Initial Deployment

1. Click **"Manual Deploy"** → **"Deploy latest commit"**
2. Wait for deployment to complete (5-10 minutes first time)
3. Check logs for:
   - ✅ "Database connection established successfully"
   - ✅ "Migrations ran successfully!"

### Step 7: Test Auto-Deploy

1. Make a small change to your code (e.g., add a comment)
2. Commit and push to `main` branch:
   ```bash
   git add .
   git commit -m "test: verify auto-deploy"
   git push origin main
   ```
3. Go to Render Dashboard → Your Service → **"Events"** tab
4. You should see a new deployment triggered automatically!

## Verification Checklist

After setup, verify these settings in Render Dashboard:

### Settings Tab:
- ✅ Branch: `main`
- ✅ Auto-Deploy: **Enabled** (toggle is ON)
- ✅ Build Command: `go build -o ums-backend main.go`
- ✅ Start Command: `./ums-backend`

### Environment Tab:
- ✅ All 7 environment variables are set
- ✅ No quotes around values
- ✅ DB_URL includes `?sslmode=require`

### Events Tab:
- ✅ Shows deployment history
- ✅ Latest deployment status: "Live"

## How Auto-Deploy Works

```
┌─────────────────────────────────────────────────────────────┐
│                    Auto-Deploy Flow                          │
└─────────────────────────────────────────────────────────────┘

1. Developer pushes code to GitHub main branch
   ↓
2. GitHub Actions runs (verifies build)
   ↓
3. Render detects push via GitHub webhook
   ↓
4. Render automatically starts deployment:
   - Clones latest code
   - Runs: go build -o ums-backend main.go
   - Runs migrations
   - Starts: ./ums-backend
   ↓
5. New version is live! 🚀
```

## Troubleshooting

### Auto-Deploy Not Triggering

**Check 1: Auto-Deploy Toggle**
- Go to Settings → Build & Deploy
- Ensure "Auto-Deploy" is **ON** (enabled)

**Check 2: Branch Name**
- Verify you're pushing to `main` branch
- Render only watches the configured branch
- Check: `git branch` (should show `* main`)

**Check 3: GitHub Connection**
- Settings → Connected Accounts
- Ensure GitHub is connected
- Try disconnecting and reconnecting

**Check 4: Webhook**
- Settings → Build & Deploy → Webhooks
- Should show GitHub webhook is active
- If missing, reconnect GitHub repository

### Deployment Fails

**Check Build Logs:**
1. Go to service → Logs tab
2. Look for error messages
3. Common issues:
   - Missing environment variables
   - Database connection failed
   - Migration errors
   - Build errors

**Check Environment Variables:**
- All 7 variables are set
- DB_URL is correct format
- No typos in variable names

### Service Not Starting

**Check Start Command:**
- Should be: `./ums-backend`
- Not: `go run main.go` (won't work in production)

**Check Logs:**
- Look for: "Database connection established"
- Look for: "Migrations ran successfully"
- Check for panic or error messages

## Manual Deploy (If Auto-Deploy Fails)

If auto-deploy isn't working, you can manually deploy:

1. Go to your service in Render Dashboard
2. Click **"Manual Deploy"** button (top right)
3. Select **"Deploy latest commit"**
4. Or select specific commit from dropdown

## Disable Auto-Deploy (If Needed)

To disable auto-deploy:
1. Go to Settings → Build & Deploy
2. Toggle **"Auto-Deploy"** to OFF
3. You'll need to manually deploy each time

## Best Practices

✅ **DO:**
- Keep auto-deploy enabled for `main` branch
- Use feature branches for development
- Test locally before pushing to `main`
- Monitor deployment logs after each push
- Set up health checks

❌ **DON'T:**
- Push broken code to `main`
- Commit `.env` file (it's in .gitignore)
- Change environment variables without testing
- Deploy during high-traffic periods

## Next Steps

After auto-deploy is working:

1. **Set up health checks** (already configured in render.yaml)
2. **Monitor logs** regularly
3. **Set up alerts** in Render Dashboard
4. **Configure custom domain** (if needed)
5. **Enable HTTPS** (automatic with Render)

## Support

- Render Docs: https://render.com/docs
- Render Status: https://status.render.com/
- Community: https://community.render.com/

---

**Summary:** Once auto-deploy is enabled in Render Dashboard, every push to `main` branch will automatically trigger a deployment. No GitHub Actions secrets needed!

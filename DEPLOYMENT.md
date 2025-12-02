# Deployment Guide for Render

## Prerequisites
- GitHub repository connected to Render
- Render account with service created
- Environment variables configured in Render dashboard

## Setup GitHub Secrets

Go to your GitHub repository → Settings → Secrets and variables → Actions → New repository secret

Add these secrets:

1. **RENDER_API_KEY**
   - Get from: Render Dashboard → Account Settings → API Keys
   - Create a new API key if you don't have one

2. **RENDER_SERVICE_ID**
   - Get from: Your service URL or Render dashboard
   - Format: `srv-xxxxxxxxxxxxx`

## Render Configuration

### Option 1: Using render.yaml (Recommended)

The `render.yaml` file in the root directory will automatically configure your service.

Make sure these environment variables are set in Render Dashboard:
- `DB_URL` - Your PostgreSQL connection string
- `JWT_SECRET` - Secret key for JWT tokens
- `EMAIL_FROM` - Email address for sending notifications
- `EMAIL_PASSWORD` - Email password/app password
- `SMTP_HOST` - SMTP server host
- `SMTP_PORT` - SMTP server port

### Option 2: Manual Configuration

If not using render.yaml:

1. **Build Command:**
   ```bash
   go build -o ums-backend main.go
   ```

2. **Start Command:**
   ```bash
   ./ums-backend
   ```

3. **Environment:**
   - Go

4. **Region:**
   - Oregon (or your preferred region)

## Deployment Process

### Automatic Deployment (CI/CD)

When you push to the `main` branch:
1. GitHub Actions runs tests
2. Builds the application
3. Triggers Render deployment via API

### Manual Deployment

1. Push your code to GitHub:
   ```bash
   git add .
   git commit -m "Your commit message"
   git push origin main
   ```

2. Render will automatically detect the push and deploy

3. Or manually trigger from Render Dashboard:
   - Go to your service
   - Click "Manual Deploy" → "Deploy latest commit"

## Troubleshooting

### Deployment Fails

1. **Check Build Logs in Render Dashboard**
   - Look for Go version issues
   - Check for missing dependencies
   - Verify migration files are included

2. **Database Connection Issues**
   - Verify `DB_URL` environment variable
   - Check database is accessible from Render
   - Ensure SSL mode is correct in connection string

3. **Migration Issues**
   - Ensure `pkg/migration` folder is included in deployment
   - Check migration files are valid SQL
   - Verify goose is running correctly

### CI/CD Not Triggering

1. **Check GitHub Actions**
   - Go to repository → Actions tab
   - Look for failed workflows
   - Check secrets are configured correctly

2. **Verify Render API Key**
   - Ensure `RENDER_API_KEY` secret is valid
   - Check `RENDER_SERVICE_ID` matches your service

3. **Check Workflow File**
   - Ensure `.github/workflows/go-ci.yml` is correct
   - Verify branch name matches (main vs master)

## Health Check

After deployment, verify:
1. Service is running: Check Render dashboard
2. Database migrations ran: Check logs for "Migrations ran successfully!"
3. API is accessible: Test endpoints

## Rollback

If deployment fails:
1. Go to Render Dashboard
2. Click on your service
3. Go to "Events" tab
4. Find a previous successful deployment
5. Click "Rollback to this version"

## Important Notes

- **Never commit `.env` file** - It's in .gitignore
- **Use Render environment variables** for sensitive data
- **Monitor logs** in Render dashboard for errors
- **Database migrations** run automatically on startup
- **Free tier** services may sleep after inactivity

## Support

If issues persist:
- Check Render status: https://status.render.com/
- Review Render docs: https://render.com/docs
- Check application logs in Render dashboard

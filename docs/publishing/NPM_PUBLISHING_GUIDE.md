# NPM Publishing Guide for mailos

## First-Time Publishing Setup

### Prerequisites
1. Create an npm account at https://www.npmjs.com/signup
2. Verify your email address
3. Install npm CLI (comes with Node.js)

### Step 1: Login to npm
```bash
npm login
```
Enter your username, password, and email when prompted.

### Step 2: Check Package Name Availability
```bash
npm view mailos
```
If you get a 404 error, the name is available. If not, you'll need to:
- Choose a different name in package.json
- Or use a scoped package name like `@anduimagui/mailos`

### Step 3: Initial Manual Publish
Navigate to the npm directory and publish:
```bash
cd npm

# First, ensure the package.json version matches your release tag
# Current version should be 0.1.14

# Temporarily create a dummy bin file for first publish
mkdir -p bin
echo '#!/usr/bin/env node\nconsole.log("Installing mailos...")' > bin/mailos
chmod +x bin/mailos

# Publish publicly (first time only)
npm publish --access public

# Clean up
rm -rf bin
```

### Step 4: Generate npm Token for GitHub Actions
1. Go to https://www.npmjs.com/settings/[your-username]/tokens
2. Click "Generate New Token"
3. Choose "Automation" token type
4. Copy the token

### Step 5: Add Token to GitHub Secrets
1. Go to https://github.com/anduimagui/emailos/settings/secrets/actions
2. Click "New repository secret"
3. Name: `NPM_TOKEN`
4. Value: Paste your npm token
5. Click "Add secret"

## Troubleshooting Common Issues

### Issue: Package name not available
**Solution**: Use a scoped package name
```json
{
  "name": "@anduimagui/mailos",
  "version": "0.1.14",
  ...
}
```

### Issue: Authentication failed
**Solution**: 
1. Verify npm login: `npm whoami`
2. Re-login if needed: `npm logout` then `npm login`

### Issue: Package already exists with different user
**Solution**:
1. Contact npm support to claim the package if it's abandoned
2. Or choose a different name

### Issue: GitHub Actions still failing after first publish
**Solution**: 
1. Ensure NPM_TOKEN is correctly set in GitHub secrets
2. Check the token hasn't expired
3. Verify the token has publish permissions

## Automated Publishing (After First Publish)

Once the package is published for the first time, GitHub Actions will handle subsequent releases automatically when you push a new tag:

```bash
# Update version in npm/package.json
# Commit changes
git add npm/package.json
git commit -m "Release v0.1.15"

# Create and push tag
git tag v0.1.15
git push origin main
git push origin v0.1.15
```

## Alternative: Using Scoped Package

If "mailos" is taken, use a scoped package:

1. Update npm/package.json:
```json
{
  "name": "@anduimagui/mailos",
  ...
}
```

2. Update .github/workflows/release.yml if needed to reference the new package name

3. Publish with:
```bash
npm publish --access public
```

## Verifying Publication

After publishing, verify your package:
```bash
# View package info
npm view mailos

# Test installation
npm install -g mailos

# Check installed version
mailos --version
```

## Notes
- The `prepublishOnly` script in package.json removes the bin directory before publishing
- The actual binary is downloaded during `postinstall` from GitHub releases
- Make sure GitHub releases are created before npm publish runs in CI/CD
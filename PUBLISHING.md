# Publishing Guide for MailOS

This guide explains how to publish MailOS to Homebrew and npm package registries.

## Prerequisites

Before publishing, ensure you have:

1. **NPM Account**: Create an account at https://www.npmjs.com/
2. **NPM Token**: Generate an access token from npm settings
3. **GitHub Repository**: Set up the repository at https://github.com/anduimagui/emailos
4. **Homebrew Tap**: Create a tap repository (e.g., `homebrew-mailos`)

## Setup Secrets

Add these secrets to your GitHub repository settings:

1. Go to Settings → Secrets and variables → Actions
2. Add the following secrets:
   - `NPM_TOKEN`: Your npm access token

## Publishing Process

### 1. Prepare for Release

1. Update version in relevant files:
   ```bash
   # Update version in go.mod if needed
   # Update version in npm/package.json
   ```

2. Commit all changes:
   ```bash
   git add .
   git commit -m "Prepare for v1.0.0 release"
   git push
   ```

### 2. Create a Git Tag

```bash
# Create and push a version tag
git tag v1.0.0
git push origin v1.0.0
```

This will trigger the GitHub Actions workflow which will:
- Build binaries for all platforms
- Create a GitHub release with the binaries
- Publish to npm automatically
- Update the Homebrew formula

### 3. Manual npm Publishing (if needed)

If automatic publishing fails, you can publish manually:

```bash
cd npm
npm login
npm publish --access public
```

### 4. Homebrew Formula

The Homebrew formula is automatically updated by the GitHub Actions workflow. For manual updates:

1. Calculate SHA256 of the darwin-amd64 release:
   ```bash
   shasum -a 256 mailos-darwin-amd64.tar.gz
   ```

2. Update `Formula/mailos.rb` with:
   - New version number
   - New SHA256 hash

3. Test the formula locally:
   ```bash
   brew install --build-from-source Formula/mailos.rb
   ```

4. Submit to your Homebrew tap:
   ```bash
   # If you have a separate homebrew-mailos repository
   cp Formula/mailos.rb ../homebrew-mailos/Formula/
   cd ../homebrew-mailos
   git add .
   git commit -m "Update mailos to v1.0.0"
   git push
   ```

## Installation Methods

After publishing, users can install MailOS using:

### Homebrew (macOS/Linux)
```bash
# From your tap
brew tap emailos/mailos
brew install mailos

# Or directly
brew install emailos/mailos/mailos
```

### npm (All platforms)
```bash
npm install -g mailos
```

### Direct Download
Users can download binaries directly from GitHub releases:
https://github.com/anduimagui/emailos/releases

## Version Management

- Use semantic versioning: `vMAJOR.MINOR.PATCH`
- Tag format must be `v1.0.0` (with 'v' prefix)
- npm version will automatically remove the 'v' prefix

## Troubleshooting

### npm Publishing Issues

1. **401 Unauthorized**: Check your NPM_TOKEN is valid
2. **403 Forbidden**: Ensure package name is available
3. **Package exists**: Increment version number

### Homebrew Issues

1. **SHA256 mismatch**: Recalculate hash from the actual release file
2. **Formula syntax**: Run `brew audit Formula/mailos.rb`

### GitHub Actions Issues

1. Check workflow runs at: https://github.com/anduimagui/emailos/actions
2. Ensure all secrets are properly configured
3. Verify tag format matches the workflow trigger

## Security Notes

- Never commit tokens or secrets to the repository
- Use GitHub Secrets for all sensitive information
- Keep npm tokens with minimal required permissions
- Regularly rotate access tokens
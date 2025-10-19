# Release Checklist for MailOS

Use this checklist when preparing a new release.

## Pre-Release Checklist

- [ ] All tests pass locally
- [ ] Documentation is up to date
- [ ] CHANGELOG is updated with new features/fixes
- [ ] Version numbers are consistent across files
- [ ] License validation works correctly
- [ ] Colors display correctly in terminal

## Release Steps

### 1. Prepare Release
- [ ] Run `./scripts/prepare-release.sh <version>`
- [ ] Review generated files in `dist/` folder
- [ ] Test binary on your platform

### 2. GitHub Setup (First Time Only)
- [ ] Create GitHub repository at https://github.com/anduimagui/emailos-cli
- [ ] Add NPM_TOKEN secret to GitHub repository
- [ ] Create homebrew-mailos tap repository

### 3. NPM Setup (First Time Only)
- [ ] Create npm account at https://www.npmjs.com/
- [ ] Generate access token with publish permissions
- [ ] Test npm login: `npm login`

### 4. Commit and Tag
- [ ] Commit all changes: `git add . && git commit -m "Release v<version>"`
- [ ] Create tag: `git tag v<version>`
- [ ] Push: `git push && git push --tags`

### 5. Monitor Release
- [ ] Check GitHub Actions workflow at https://github.com/anduimagui/emailos-cli/actions
- [ ] Verify GitHub release is created
- [ ] Confirm binaries are attached to release

### 6. Verify npm Publication
- [ ] Check package at https://www.npmjs.com/package/mailos
- [ ] Test installation: `npm install -g mailos`
- [ ] Verify installed version: `mailos --version`

### 7. Verify Homebrew
- [ ] Check formula is updated in tap repository
- [ ] Test installation: `brew install emailos/mailos/mailos`
- [ ] Verify installed version: `mailos --version`

## Post-Release

- [ ] Announce release on social media
- [ ] Update website with new version
- [ ] Send release notes to users
- [ ] Monitor issues for any problems

## Rollback Plan

If issues are found:

1. Delete the release tag:
   ```bash
   git tag -d v<version>
   git push origin :refs/tags/v<version>
   ```

2. Unpublish from npm (within 72 hours):
   ```bash
   npm unpublish mailos@<version>
   ```

3. Revert Homebrew formula in tap repository

## Common Issues

### npm publish fails
- Check NPM_TOKEN is valid
- Ensure you're logged in: `npm whoami`
- Verify package name is available

### Homebrew formula issues
- Verify SHA256 matches the release file
- Test formula locally first
- Check syntax with `brew audit`

### GitHub Actions fails
- Check workflow syntax
- Verify all secrets are set
- Review error logs in Actions tab
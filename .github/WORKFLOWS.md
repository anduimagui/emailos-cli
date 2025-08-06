# GitHub Actions Workflows

This repository uses several GitHub Actions workflows for CI/CD:

## 1. Go CI (`go.yml`)
- **Trigger**: Push to main, Pull requests to main
- **Purpose**: Build and test Go code
- **Actions**: 
  - Build the Go project
  - Run tests

## 2. Release (`release.yml`)
- **Trigger**: Git tags (v*)
- **Purpose**: Main release workflow
- **Actions**:
  - Build binaries for all platforms (macOS, Linux, Windows)
  - Create GitHub release with binaries
  - Publish to npm registry
  - Update Homebrew formula

## 3. Node.js Package (`npm-publish.yml`)
- **Trigger**: GitHub release created
- **Purpose**: Publish npm package
- **Actions**:
  - Build and test npm package
  - Publish to npm registry

## 4. GitHub Packages (`npm-publish-github-packages.yml`)
- **Trigger**: GitHub release created
- **Purpose**: Publish to GitHub Packages registry
- **Actions**:
  - Build and test npm package
  - Publish to GitHub Packages

## 5. SLSA Go Releaser (`go-ossf-slsa3-publish.yml`)
- **Trigger**: Manual workflow dispatch, GitHub release created
- **Purpose**: Build with SLSA3 compliance for supply chain security
- **Actions**:
  - Build Go binaries with provenance
  - Generate SLSA attestation

## Configuration Files

- `.slsa-goreleaser.yml`: Configuration for SLSA Go builds
- `Formula/mailos.rb`: Homebrew formula template

## Secrets Required

The following secrets need to be configured in repository settings:

- `NPM_TOKEN`: npm registry authentication token
- `GITHUB_TOKEN`: Automatically provided by GitHub Actions

## Version Compatibility

- Go version: 1.23
- Node.js version: 20

## Notes

- The main release workflow is `release.yml` which handles most publishing tasks
- `npm-publish.yml` and `npm-publish-github-packages.yml` provide additional npm publishing options
- SLSA workflow provides supply chain security attestation for releases
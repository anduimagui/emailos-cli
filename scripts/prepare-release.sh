#!/bin/bash

# prepare-release.sh - Prepare MailOS for release
# Usage: ./scripts/prepare-release.sh <version>
# Example: ./scripts/prepare-release.sh 1.0.0

set -e

VERSION=$1

if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 1.0.0"
    exit 1
fi

echo "Preparing MailOS release v$VERSION"
echo "=================================="

# Update npm package.json version
echo "Updating npm package version..."
cd npm
npm version $VERSION --no-git-tag-version
cd ..

# Update Homebrew formula version (placeholder)
echo "Updating Homebrew formula..."
sed -i.bak "s/v[0-9]\+\.[0-9]\+\.[0-9]\+/v$VERSION/g" Formula/mailos.rb
rm Formula/mailos.rb.bak

# Build binaries for testing
echo "Building test binaries..."
PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
)

mkdir -p dist

for PLATFORM in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$PLATFORM"
    OUTPUT="dist/mailos-${GOOS}-${GOARCH}"
    
    if [ "$GOOS" = "windows" ]; then
        OUTPUT="${OUTPUT}.exe"
    fi
    
    echo "Building for $GOOS/$GOARCH..."
    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w -X main.version=v$VERSION" -o "$OUTPUT" .
    
    # Create tar.gz archive
    tar -czf "${OUTPUT}.tar.gz" -C dist "$(basename $OUTPUT)"
done

echo ""
echo "Release preparation complete!"
echo ""
echo "Next steps:"
echo "1. Review the changes"
echo "2. Commit: git add . && git commit -m 'Release v$VERSION'"
echo "3. Tag: git tag v$VERSION"
echo "4. Push: git push && git push --tags"
echo ""
echo "This will trigger the GitHub Actions workflow to:"
echo "- Build and release binaries"
echo "- Publish to npm"
echo "- Update Homebrew formula"
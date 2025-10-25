#!/bin/bash
set -e

# Minor version release script
# Moved from Taskfile.yml for better organization

echo "📦 Starting minor release process..."

# Bump minor version
cd npm
npm version minor --no-git-tag-version
NEW_VERSION=$(node -p "require('./package.json').version")
cd ..

# Update Homebrew formula
sed -i.bak "s/v[0-9]\+\.[0-9]\+\.[0-9]\+/v$NEW_VERSION/g" Formula/mailos.rb
rm Formula/mailos.rb.bak

# Commit, tag and push
git add npm/package.json Formula/mailos.rb
git commit -m "Release v$NEW_VERSION"
git tag "v$NEW_VERSION"
git push && git push --tags

echo "✅ Minor release v$NEW_VERSION published!"
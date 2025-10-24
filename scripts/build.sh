#!/bin/bash
set -e

BINARY_NAME=mailos
BINARY_PATH=cmd/mailos/main.go

# Function to report errors to GitHub if in CI environment
report_error() {
    local error_msg="$1"
    local context="$2"
    
    echo "❌ BUILD ERROR: $error_msg"
    
    # Report to GitHub if we have gh CLI available and are in a git repo
    if command -v gh >/dev/null 2>&1 && git rev-parse --git-dir >/dev/null 2>&1; then
        local tag_name=$(git describe --tags --exact-match 2>/dev/null || echo "")
        if [ -n "$tag_name" ]; then
            echo "🔍 Detected tag build failure for $tag_name"
            echo "📝 Creating GitHub issue for build failure..."
            
            # Create an issue about the build failure
            gh issue create \
                --title "Build failure for release $tag_name" \
                --body "**Build Error**: $error_msg

**Context**: $context

**Tag**: $tag_name

**Build Environment**: 
- OS: $(uname -s) $(uname -r)
- Go Version: $(go version)
- Time: $(date)

**Error Details**: 
Build failed during $context stage.

**Logs**: 
Check the GitHub Actions logs for detailed error information.

**Auto-generated issue** - created by build script on build failure." \
                --label "bug,build-failure,auto-generated" || echo "⚠️  Failed to create GitHub issue"
        fi
    fi
    
    exit 1
}

echo "Building ${BINARY_NAME}..."

# Check Go module integrity first
if ! go mod tidy; then
    report_error "Go module tidy failed" "module-verification"
fi

# Build with error handling
if ! go build -o ${BINARY_NAME} ${BINARY_PATH}; then
    report_error "Go build compilation failed" "compilation"
fi

# Verify the binary was created and is executable
if [ ! -f "./${BINARY_NAME}" ]; then
    report_error "Binary not found after build" "post-build-verification"
fi

if [ ! -x "./${BINARY_NAME}" ]; then
    report_error "Binary is not executable" "post-build-verification"
fi

echo "Build complete: ./${BINARY_NAME}"

echo "Creating local emailOS alias..."
if [ -f "./${BINARY_NAME}" ]; then
    if ln -sf "./${BINARY_NAME}" "./emailOS"; then
        echo "✓ Created local emailOS alias -> ${BINARY_NAME}"
        echo "✓ Both './${BINARY_NAME}' and './emailOS' commands are now available"
    else
        echo "⚠️  Warning: Failed to create emailOS alias, but continuing..."
    fi
else
    echo "⚠️  Warning: ${BINARY_NAME} binary not found, skipping alias creation"
fi
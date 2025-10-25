#!/bin/bash
set -e

# Enhanced publish-patch script with automatic error reporting
# Moved from Taskfile.yml for better organization and maintainability

echo "📦 Starting patch release process..."

# Function to report errors to GitHub
report_release_error() {
    local error_msg="$1"
    local context="$2"
    local run_id="$3"
    
    echo "❌ RELEASE ERROR: $error_msg"
    
    # Report to GitHub if we have gh CLI available
    if command -v gh >/dev/null 2>&1; then
        local tag_name=$(git describe --tags --exact-match 2>/dev/null || echo "")
        if [ -n "$tag_name" ]; then
            echo "📝 Creating GitHub issue for release failure..."
            
            ISSUE_BODY="Release Failed: $tag_name

Context: $context
Error: $error_msg
Workflow Run: https://github.com/anduimagui/emailos-cli/actions/runs/$run_id
Time: $(date)

Actions Required:
1. Check detailed logs in workflow run
2. Fix build/dependency issues
3. Re-run workflow or create new release

Auto-generated issue created by release script."
            
            gh issue create \
                --title "Release failure for $tag_name" \
                --body "$ISSUE_BODY" \
                --label "bug,release-failure,auto-generated" \
                2>/dev/null || echo "⚠️  Could not create GitHub issue (may need authentication)"
        fi
    fi
}

# Step 1: Version bump in npm package
echo "🔢 Bumping patch version..."
cd npm
CURRENT_VERSION=$(node -p "require('./package.json').version")
echo "Current version: v$CURRENT_VERSION"

if ! npm version patch --no-git-tag-version; then
    cd ..
    report_release_error "Failed to bump npm version" "version-bump" ""
    exit 1
fi

NEW_VERSION=$(node -p "require('./package.json').version")
echo "New version: v$NEW_VERSION"
cd ..

# Step 2: Update Homebrew formula
echo "🍺 Updating Homebrew formula..."
if ! sed -i.bak "s/v[0-9]\+\.[0-9]\+\.[0-9]\+/v$NEW_VERSION/g" Formula/mailos.rb; then
    report_release_error "Failed to update Homebrew formula" "homebrew-update" ""
    exit 1
fi
rm -f Formula/mailos.rb.bak

# Step 3: Commit version changes
echo "📝 Committing version changes..."
git add npm/package.json Formula/mailos.rb
git commit -m "Release v$NEW_VERSION" || true

# Step 4: Create and push tag
echo "🏷️  Creating git tag..."
if ! git tag "v$NEW_VERSION"; then
    report_release_error "Failed to create git tag v$NEW_VERSION" "git-tagging" ""
    exit 1
fi
echo "✅ Tagged version v$NEW_VERSION"

# Step 5: Push to GitHub (this triggers the release workflow)
echo "🚀 Pushing to GitHub..."
if ! git push; then
    report_release_error "Failed to push commits to GitHub" "git-push" ""
    exit 1
fi

if ! git push --tags; then
    report_release_error "Failed to push tags to GitHub" "git-push-tags" ""
    exit 1
fi

# Step 6: Monitor workflow execution
echo "✅ Release v$NEW_VERSION initiated!"
echo ""
echo "📋 The GitHub Actions workflow will now:"
echo "  • Build binaries for all platforms"
echo "  • Create a GitHub release"
echo "  • Publish to npm registry"
echo "  • Update Homebrew formula"
echo ""
echo "🔍 Monitoring workflow status..."

sleep 10

# Monitor workflow for up to 15 minutes (30 attempts * 30 seconds)
for i in {1..30}; do
    echo "⏳ Checking workflow status (attempt $i/30)..."
    
    # Get the latest workflow run matching our release pattern
    WORKFLOW_STATUS=$(gh run list --limit 3 --json status,conclusion,displayTitle,databaseId --jq '.[] | select(.displayTitle | test("Release|release")) | .status + ":" + (.conclusion // "") + ":" + (.databaseId | tostring)' | head -1)
    
    if [ -n "$WORKFLOW_STATUS" ]; then
        STATUS=$(echo $WORKFLOW_STATUS | cut -d: -f1)
        CONCLUSION=$(echo $WORKFLOW_STATUS | cut -d: -f2)
        RUN_ID=$(echo $WORKFLOW_STATUS | cut -d: -f3)
        
        if [ "$STATUS" = "completed" ]; then
            if [ "$CONCLUSION" = "success" ]; then
                echo "✅ Release workflow completed successfully!"
                echo "📦 Check release at: https://github.com/anduimagui/emailos-cli/releases/tag/v$NEW_VERSION"
                break
            else
                echo "⚠️  Release workflow completed with conclusion: $CONCLUSION"
                echo "🔍 Checking if GitHub release was created anyway..."
                
                if gh release view "v$NEW_VERSION" >/dev/null 2>&1; then
                    echo "✅ GitHub release v$NEW_VERSION was created successfully!"
                    echo "📦 Check release at: https://github.com/anduimagui/emailos-cli/releases/tag/v$NEW_VERSION"
                    echo "ℹ️  Some optional jobs (npm/homebrew) may have failed - check details if needed:"
                    echo "🔗 https://github.com/anduimagui/emailos-cli/actions/runs/$RUN_ID"
                    break
                else
                    echo "❌ Release workflow failed and no release was created"
                    echo "🔍 Fetching detailed error information..."
                    echo "🔗 View details at: https://github.com/anduimagui/emailos-cli/actions/runs/$RUN_ID"
                    echo ""
                    echo "📋 Failed Job Details:"
                    
                    # Fetch detailed failure information
                    FAILED_JOBS=$(gh api repos/anduimagui/emailos-cli/actions/runs/$RUN_ID/jobs --jq '.jobs[] | select(.conclusion == "failure") | {name, conclusion, completed_at: .completed_at, steps: [.steps[] | select(.conclusion == "failure") | {name, conclusion, number}]}' 2>/dev/null || echo "Could not fetch job details")
                    echo "$FAILED_JOBS"
                    echo ""
                    
                    # Get detailed logs from the first failed job for specific error messages
                    echo "🔍 Fetching detailed error logs..."
                    FIRST_FAILED_JOB_ID=$(gh api repos/anduimagui/emailos-cli/actions/runs/$RUN_ID/jobs --jq '.jobs[] | select(.conclusion == "failure") | .id' | head -1)
                    if [ -n "$FIRST_FAILED_JOB_ID" ]; then
                        echo "📋 Error details from job $FIRST_FAILED_JOB_ID:"
                        # Extract the actual error lines from the logs
                        ERROR_LOGS=$(gh api repos/anduimagui/emailos-cli/actions/jobs/$FIRST_FAILED_JOB_ID/logs 2>/dev/null | grep -E "(error|Error|ERROR|##\[error\]|fail|FAIL|exit code)" | tail -10 || echo "Could not fetch error logs")
                        echo "$ERROR_LOGS"
                        echo ""
                    fi
                    
                    # Create GitHub issue for the failure
                    echo "📝 Creating GitHub issue for release failure..."
                    ISSUE_BODY="Release Failed: v$NEW_VERSION

Workflow Run: https://github.com/anduimagui/emailos-cli/actions/runs/$RUN_ID

Failed Jobs:
\`\`\`json
$FAILED_JOBS
\`\`\`

Error Details:
\`\`\`
${ERROR_LOGS:-No detailed error logs available}
\`\`\`

Time: $(date)

Actions Required:
1. Check detailed logs in workflow run
2. Fix build/dependency issues
3. Re-run workflow or create new release

Auto-generated issue created by release script."
                    
                    gh issue create \
                        --title "Release workflow failure for $NEW_VERSION" \
                        --body "$ISSUE_BODY" \
                        --label "bug,release-failure,auto-generated" \
                        2>/dev/null || echo "⚠️  Could not create GitHub issue (may need authentication)"
                    
                    break
                fi
            fi
        else
            echo "🔄 Workflow status: $STATUS (Run ID: $RUN_ID)"
            sleep 15
        fi
    else
        echo "⚠️  No release workflow found yet, waiting..."
        sleep 15
    fi
done

echo ""
echo "Monitor progress at: https://github.com/anduimagui/emailos-cli/actions"
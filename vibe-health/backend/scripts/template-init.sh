#!/bin/bash

set -e

echo "🚀 Initializing template-goapi..."

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Error: Not in a git repository. Please run 'git init' first."
    exit 1
fi

# Get git remote URL
REMOTE_URL=$(git remote get-url origin 2>/dev/null || echo "")

if [[ -z "$REMOTE_URL" ]]; then
    echo "❌ Error: No git remote 'origin' found. Please add a remote first:"
    echo "   git remote add origin git@github.com:username/repo-name.git"
    exit 1
fi

echo "📡 Found git remote: $REMOTE_URL"

# Extract repo name from various git URL formats
# Handle: git@github.com:foo/bar.git, https://github.com/foo/bar.git, etc.
REPO_NAME=""
if [[ "$REMOTE_URL" =~ git@[^:]+:([^/]+)/([^/]+)\.git$ ]]; then
    # SSH format: git@github.com:user/repo.git
    REPO_NAME="${BASH_REMATCH[2]}"
elif [[ "$REMOTE_URL" =~ https?://[^/]+/([^/]+)/([^/]+)\.git$ ]]; then
    # HTTPS format: https://github.com/user/repo.git
    REPO_NAME="${BASH_REMATCH[2]}"
elif [[ "$REMOTE_URL" =~ ([^/]+)/([^/]+)$ ]]; then
    # Simple format: user/repo
    REPO_NAME="${BASH_REMATCH[2]}"
else
    echo "❌ Error: Could not parse repository name from remote URL: $REMOTE_URL"
    echo "Please ensure your remote URL follows standard git conventions."
    exit 1
fi

echo "📦 Detected repository name: $REPO_NAME"

# Extract GitHub username/org for full module name
MODULE_NAME=""
if [[ "$REMOTE_URL" =~ git@[^:]+:([^/]+)/([^/]+)\.git$ ]]; then
    # SSH format: git@github.com:user/repo.git
    MODULE_NAME="github.com/${BASH_REMATCH[1]}/${BASH_REMATCH[2]}"
elif [[ "$REMOTE_URL" =~ https?://([^/]+)/([^/]+)/([^/]+)\.git$ ]]; then
    # HTTPS format: https://github.com/user/repo.git
    MODULE_NAME="${BASH_REMATCH[1]}/${BASH_REMATCH[2]}/${BASH_REMATCH[3]}"
elif [[ "$REMOTE_URL" =~ ([^/]+)/([^/]+)$ ]]; then
    # Simple format: user/repo - assume GitHub
    MODULE_NAME="github.com/${BASH_REMATCH[1]}/${BASH_REMATCH[2]}"
else
    echo "❌ Error: Could not parse module name from remote URL: $REMOTE_URL"
    exit 1
fi

echo "🔧 Module name: $MODULE_NAME"

# Convert repo name to uppercase for PROJECT_NAME
PROJECT_NAME=$(echo "$REPO_NAME" | tr '[:lower:]' '[:upper:]' | tr '-' '_')
echo "📛 Project name: $PROJECT_NAME"

# Create config directory name from repo name (lowercase, with dot prefix)
CONFIG_DIR=".$(echo "$REPO_NAME" | tr '[:upper:]' '[:lower:]')"
echo "📁 Config directory: $CONFIG_DIR"

# Prompt for additional template variables
echo ""
echo "🔧 Additional configuration needed for OpenAPI and Docker:"
echo ""

# API Title
read -p "📝 Enter API title for documentation (e.g., 'My API'): " API_TITLE
if [[ -z "$API_TITLE" ]]; then
    API_TITLE="$REPO_NAME API"
    echo "   Using default: $API_TITLE"
fi

# API Base URL
read -p "🌐 Enter production API base URL (e.g., 'https://api.example.com'): " API_BASE_URL
if [[ -z "$API_BASE_URL" ]]; then
    API_BASE_URL="https://api.example.com"
    echo "   Using default: $API_BASE_URL"
fi

# Docker Username
read -p "🐳 Enter Docker Hub username: " DOCKER_USERNAME
if [[ -z "$DOCKER_USERNAME" ]]; then
    DOCKER_USERNAME="username"
    echo "   Using placeholder: $DOCKER_USERNAME"
fi

# Docker Image Name
read -p "📦 Enter Docker image name (e.g., '$REPO_NAME'): " DOCKER_IMAGE_NAME
if [[ -z "$DOCKER_IMAGE_NAME" ]]; then
    DOCKER_IMAGE_NAME="$REPO_NAME"
    echo "   Using default: $DOCKER_IMAGE_NAME"
fi

echo ""
echo "🔄 Replacing template variables in all files..."

# Find all files (excluding .git, node_modules, vendor, etc.)
find . -type f \
    -not -path "./.git/*" \
    -not -path "./node_modules/*" \
    -not -path "./vendor/*" \
    -not -path "./.vscode/*" \
    -not -path "./.idea/*" \
    -not -name "*.log" \
    -not -name "*.tmp" \
    -not -name "template-init.sh" \
    | while read -r file; do

    # Check if file contains any of our template variables (exclude GitHub Actions and Docker metadata patterns)
    if grep -l "github.com/JerkyTreats/PHITE/vibe-health/backend\|vibe_health_BACKEND\|.vibe-health\|vibe health API\|https://api.vibe-health.internal.jerkytreats.dev\|jerkytreats\|vibe-health-api" "$file" > /dev/null 2>&1; then
        echo "  📝 Processing: $file"

        # Use sed to replace only our template variables (not GitHub Actions or Docker metadata patterns)
        sed -i.bak \
            -e "s|github.com/JerkyTreats/PHITE/vibe-health/backend|$MODULE_NAME|g" \
            -e "s|vibe_health_BACKEND|$PROJECT_NAME|g" \
            -e "s|.vibe-health|$CONFIG_DIR|g" \
            -e "s|vibe health API|$API_TITLE|g" \
            -e "s|https://api.vibe-health.internal.jerkytreats.dev|$API_BASE_URL|g" \
            -e "s|jerkytreats|$DOCKER_USERNAME|g" \
            -e "s|vibe-health-api|$DOCKER_IMAGE_NAME|g" \
            "$file"

        # Remove backup file
        rm "${file}.bak"
    fi
done

echo ""
echo "✅ Template initialization complete!"
echo ""
echo "🎯 Summary:"
echo "   Repository: $REPO_NAME"
echo "   Module: $MODULE_NAME"
echo "   Project: $PROJECT_NAME"
echo "   Config dir: ~/$CONFIG_DIR"
echo "   API Title: $API_TITLE"
echo "   API Base URL: $API_BASE_URL"
echo "   Docker User: $DOCKER_USERNAME"
echo "   Docker Image: $DOCKER_IMAGE_NAME"
echo ""
echo "🚀 Next steps:"
echo "   1. Run: go mod tidy"
echo "   2. Run: go run cmd/server/main.go"
echo "   3. Test: curl http://localhost:8080/health"
echo "   4. Generate docs: go run cmd/generate-openapi/main.go"
echo ""
echo "🐳 Docker setup:"
echo "   Set GitHub secrets: DOCKER_USERNAME, DOCKER_PASSWORD"
echo ""
echo "🗑️  You can now delete this script: rm scripts/template-init.sh"
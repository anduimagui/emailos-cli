#!/bin/bash

# EmailOS Local Build Script for macOS
# This script builds and installs mailos system-wide on your Mac

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Binary name
BINARY_NAME="mailos"
BINARY_PATH="cmd/mailos/main.go"

# Installation paths
LOCAL_BIN="/usr/local/bin"
ALTERNATIVE_BIN="$HOME/.local/bin"

echo -e "${GREEN}🔨 EmailOS Local Build Script${NC}"
echo "================================"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed. Please install Go first.${NC}"
    echo "Visit: https://go.dev/doc/install"
    exit 1
fi

echo -e "${YELLOW}📦 Installing dependencies...${NC}"
go mod download
go mod tidy

echo -e "${YELLOW}🔨 Building ${BINARY_NAME}...${NC}"
go build -ldflags="-s -w" -o ${BINARY_NAME} ${BINARY_PATH}

if [ ! -f ${BINARY_NAME} ]; then
    echo -e "${RED}❌ Build failed!${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Build successful!${NC}"

# Choose installation directory
INSTALL_DIR=""
if [ -w "$LOCAL_BIN" ]; then
    INSTALL_DIR="$LOCAL_BIN"
elif [ -d "$ALTERNATIVE_BIN" ] && [ -w "$ALTERNATIVE_BIN" ]; then
    INSTALL_DIR="$ALTERNATIVE_BIN"
else
    echo -e "${YELLOW}📁 Creating local bin directory...${NC}"
    mkdir -p "$ALTERNATIVE_BIN"
    INSTALL_DIR="$ALTERNATIVE_BIN"
    
    # Add to PATH if not already there
    if [[ ":$PATH:" != *":$ALTERNATIVE_BIN:"* ]]; then
        echo -e "${YELLOW}📝 Adding $ALTERNATIVE_BIN to PATH...${NC}"
        
        # Detect shell and update appropriate config file
        if [ -n "$ZSH_VERSION" ] || [ -f "$HOME/.zshrc" ]; then
            echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.zshrc"
            echo -e "${GREEN}✅ Added to ~/.zshrc${NC}"
            echo -e "${YELLOW}⚠️  Run 'source ~/.zshrc' or restart your terminal${NC}"
        elif [ -n "$BASH_VERSION" ] || [ -f "$HOME/.bash_profile" ]; then
            echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.bash_profile"
            echo -e "${GREEN}✅ Added to ~/.bash_profile${NC}"
            echo -e "${YELLOW}⚠️  Run 'source ~/.bash_profile' or restart your terminal${NC}"
        fi
    fi
fi

# Install the binary
echo -e "${YELLOW}📥 Installing to ${INSTALL_DIR}...${NC}"

# Backup existing binary if it exists
if [ -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
    echo -e "${YELLOW}📦 Backing up existing binary...${NC}"
    mv "${INSTALL_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}.backup"
fi

# Copy new binary
cp ${BINARY_NAME} "${INSTALL_DIR}/"
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

# Verify installation
if [ -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
    echo -e "${GREEN}✅ Installation successful!${NC}"
    echo
    echo -e "${GREEN}📍 Installed to: ${INSTALL_DIR}/${BINARY_NAME}${NC}"
    
    # Test if it's in PATH
    if command -v ${BINARY_NAME} &> /dev/null; then
        VERSION=$(${BINARY_NAME} --version 2>/dev/null | head -n1 || echo "version unknown")
        echo -e "${GREEN}✅ ${BINARY_NAME} is ready to use! ${VERSION}${NC}"
        echo
        echo "Run '${BINARY_NAME}' from anywhere to get started!"
    else
        echo -e "${YELLOW}⚠️  ${BINARY_NAME} installed but not in PATH yet${NC}"
        echo "Either:"
        echo "  1. Restart your terminal, or"
        echo "  2. Run: source ~/.zshrc (or ~/.bash_profile)"
        echo "  3. Use full path: ${INSTALL_DIR}/${BINARY_NAME}"
    fi
else
    echo -e "${RED}❌ Installation failed!${NC}"
    exit 1
fi

# Clean up build artifact
rm -f ${BINARY_NAME}

echo
echo -e "${GREEN}🎉 Done! EmailOS (mailos) is ready to use across your Mac!${NC}"
#!/bin/bash

# Quick Deploy Script for EmailOS Server
# This script provides a one-command deployment to Hetzner Cloud

set -e

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_info() {
    echo -e "${BLUE}[ℹ]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

print_header() {
    echo -e "${CYAN}$1${NC}"
}

# Check if we're in the right directory
if [ ! -f "Taskfile.yml" ]; then
    print_error "Please run this script from the emailos/deployment directory"
    exit 1
fi

echo ""
print_header "================================"
print_header "📧 EmailOS Server Quick Deploy"
print_header "================================"
echo ""

# Check dependencies
print_info "Checking dependencies..."

MISSING_DEPS=0

if ! command -v hcloud &> /dev/null; then
    print_error "hcloud CLI not found"
    echo "   Install with: brew install hcloud"
    echo "   Or visit: https://github.com/hetznercloud/cli"
    MISSING_DEPS=1
fi

if ! command -v task &> /dev/null; then
    print_error "task not found"
    echo "   Install with: brew install go-task/tap/go-task"
    echo "   Or visit: https://taskfile.dev/"
    MISSING_DEPS=1
fi

if ! command -v jq &> /dev/null; then
    print_error "jq not found"
    echo "   Install with: brew install jq"
    MISSING_DEPS=1
fi

if [ $MISSING_DEPS -eq 1 ]; then
    echo ""
    print_error "Please install missing dependencies and try again"
    exit 1
fi

print_status "Dependencies OK"

# Check hcloud configuration
print_info "Checking Hetzner Cloud configuration..."
if ! hcloud context active &> /dev/null; then
    print_warning "Hetzner Cloud not configured"
    echo ""
    echo "Let's set up your Hetzner Cloud access:"
    echo ""
    echo "1. Get your API token from: https://console.hetzner.cloud/"
    echo "   (Project → Security → API Tokens → Generate API Token)"
    echo ""
    read -p "Enter your project name (e.g., emailos-project): " PROJECT_NAME
    
    if [ -z "$PROJECT_NAME" ]; then
        PROJECT_NAME="emailos-project"
    fi
    
    hcloud context create "$PROJECT_NAME"
    
    if [ $? -ne 0 ]; then
        print_error "Failed to configure hcloud"
        exit 1
    fi
fi

print_status "Hetzner Cloud configured"

# Deployment options
echo ""
print_header "📋 Deployment Options"
echo "================================"
echo "Location: Helsinki (hel1)"
echo "Server: CPX11 (1 vCPU, 2GB RAM)"
echo "OS: Ubuntu 24.04"
echo "Cost: ~€4.15/month"
echo "================================"
echo ""

read -p "Continue with deployment? (y/N): " CONFIRM
if [[ ! "$CONFIRM" =~ ^[Yy]$ ]]; then
    print_info "Deployment cancelled"
    exit 0
fi

# Deploy server
echo ""
print_info "Deploying EmailOS server..."
task deploy

if [ $? -ne 0 ]; then
    print_error "Deployment failed"
    exit 1
fi

# Get connection info
print_info "Getting connection information..."
if [ ! -f ".server-ip" ]; then
    task get-ip
fi

IP=$(cat .server-ip)

# Install EmailOS and AI tools
echo ""
print_info "Installing AI tools on server..."
task install-claude
task install-other-ai

# Display success message
echo ""
print_header "================================"
print_header "🎉 Deployment Complete!"
print_header "================================"
echo ""
echo -e "${GREEN}Server Details:${NC}"
echo "  IP Address: $IP"
echo "  Web Interface: http://$IP:8080"
echo "  EmailOS API: http://$IP:8089 (when configured)"
echo ""
echo -e "${CYAN}📱 For Termius (mobile):${NC}"
echo "  Host: $IP"
echo "  Username: root"
echo "  SSH Key: ~/.ssh/emailos_server_key"
echo ""
echo -e "${CYAN}🖥️ Terminal Connection:${NC}"
echo "  SSH:  ssh -i ~/.ssh/emailos_server_key root@$IP"
echo "  Mosh: mosh --ssh='ssh -i ~/.ssh/emailos_server_key' root@$IP"
echo ""
echo -e "${YELLOW}⚡ Quick Commands:${NC}"
echo "  task connect         # SSH to server"
echo "  task mosh           # Mosh to server (better for mobile)"
echo "  task setup-tmux     # Start persistent session"
echo "  task configure-emailos  # Configure email account"
echo "  task status         # Show server status"
echo "  task termius-config # Show Termius setup"
echo ""
echo -e "${GREEN}📧 Next Steps:${NC}"
echo "1. Connect to server: ${CYAN}task connect${NC}"
echo "2. Configure EmailOS: ${CYAN}mailos setup${NC}"
echo "3. Start Claude Code: ${CYAN}claude-code${NC}"
echo "4. Use tmux for persistent sessions: ${CYAN}tmux${NC}"
echo ""
print_header "================================"
echo ""
print_info "Server is ready! Connect and start using EmailOS with AI assistance."
echo ""

# Offer to show Termius config
read -p "Show Termius configuration now? (y/N): " SHOW_TERMIUS
if [[ "$SHOW_TERMIUS" =~ ^[Yy]$ ]]; then
    task termius-config
fi
#!/bin/bash

# EmailOS Account Session Manager
# This script helps set the session default account for the current terminal

if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "EmailOS Account Session Manager"
    echo ""
    echo "Usage:"
    echo "  source set_account.sh                    # List accounts and set interactively"
    echo "  source set_account.sh user@example.com   # Set specific account"
    echo "  source set_account.sh --clear            # Clear session default"
    echo ""
    echo "Note: You must use 'source' or '.' to run this script for it to affect your current shell"
    exit 0
fi

if [ "$1" = "--clear" ]; then
    unset MAILOS_SESSION_ACCOUNT
    echo "✓ Cleared session default account"
    return 0 2>/dev/null || exit 0
fi

if [ -n "$1" ]; then
    # Set specific account
    export MAILOS_SESSION_ACCOUNT="$1"
    echo "✓ Set session default account to: $1"
    echo "  This will be used for all mailos commands in this terminal session"
else
    # Interactive selection
    echo "Available accounts:"
    mailos accounts 2>/dev/null | grep -E "^\s+[0-9]\." | sed 's/^  //'
    echo ""
    read -p "Enter account email or number: " selection
    
    if [[ "$selection" =~ ^[0-9]+$ ]]; then
        # User entered a number, extract the email
        account=$(mailos accounts 2>/dev/null | grep "^  $selection\." | sed -E 's/^[^.]+\. ([^ ]+).*/\1/')
    else
        # User entered an email
        account="$selection"
    fi
    
    if [ -n "$account" ]; then
        export MAILOS_SESSION_ACCOUNT="$account"
        echo "✓ Set session default account to: $account"
        echo "  This will be used for all mailos commands in this terminal session"
    else
        echo "✗ Invalid selection"
        return 1 2>/dev/null || exit 1
    fi
fi
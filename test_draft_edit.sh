#!/bin/bash

# Test script for draft editing functionality

echo "Testing draft editing functionality..."
echo "======================================"

# Create a test draft
echo ""
echo "1. Creating a test draft..."
./mailos draft -t test@example.com -s "Test Draft" -b "Initial draft content"

echo ""
echo "2. Listing drafts to see UIDs..."
./mailos draft --list

echo ""
echo "3. Instructions for editing a draft:"
echo "   To edit a draft, use: ./mailos draft --edit-uid <UID> -b 'Updated content'"
echo "   Example: ./mailos draft --edit-uid 12345 -s 'Updated Subject' -b 'New body content'"
echo ""
echo "The draft command now:"
echo "  - Shows UIDs when creating drafts (e.g., 'Saved draft to email account's Drafts folder (UID: 12345)')"
echo "  - Shows UIDs when listing drafts (e.g., 'Draft #1 (UID: 12345)')"
echo "  - Allows editing with --edit-uid flag"
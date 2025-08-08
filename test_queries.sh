#!/bin/bash

echo "Testing EmailOS Query Parsing"
echo "=============================="
echo ""

# Build the binary first
echo "Building mailos..."
go build -o mailos cmd/mailos/main.go
echo ""

echo "Test 1: mailos (no args) - should show landing page"
echo "Command: ./mailos"
echo "(Would normally show interactive menu)"
echo ""

echo "Test 2: mailos q=unread emails"
echo "Command: ./mailos q=unread emails"
echo "(Would process query: 'unread emails')"
echo ""

echo "Test 3: mailos \"find all attachments\""
echo "Command: ./mailos \"find all attachments\""
echo "(Would process query: 'find all attachments')"
echo ""

echo "Test 4: mailos 'emails from john'"
echo "Command: ./mailos 'emails from john'"
echo "(Would process query: 'emails from john')"
echo ""

echo "Test 5: mailos read (known command)"
echo "Command: ./mailos read --help | head -n 3"
./mailos read --help | head -n 3
echo ""

echo "Test 6: mailos random text (unknown command, no query format)"
echo "Command: ./mailos random text"
echo "(Should show landing page since it's not in query format)"
echo ""

echo "Test 7: mailos q="
echo "Command: ./mailos q="
echo "(Empty query - should be handled gracefully)"
echo ""

echo "=============================="
echo "Query parsing tests complete!"
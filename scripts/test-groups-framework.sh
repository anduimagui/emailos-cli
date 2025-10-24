#!/bin/bash

echo "🧪 Running groups framework tests..."

echo "=== Groups Help Tests ==="
go run test/test_framework/main.go groups help | while IFS='|' read -r name command; do
    if [ -n "$name" ] && [ -n "$command" ]; then
        echo -n "Testing: $name... "
        if eval "$command" > /dev/null 2>&1; then
            echo "✓ PASS"
        else
            echo "✗ FAIL"
        fi
    fi
done

echo "=== Groups Basic Tests ==="
go run test/test_framework/main.go groups basic | while IFS='|' read -r name command; do
    if [ -n "$name" ] && [ -n "$command" ]; then
        echo -n "Testing: $name... "
        if eval "$command" > /dev/null 2>&1; then
            echo "✓ PASS"
        else
            echo "✗ FAIL"
        fi
    fi
done

echo "=== Groups Member Management Tests ==="
go run test/test_framework/main.go groups members | while IFS='|' read -r name command; do
    if [ -n "$name" ] && [ -n "$command" ]; then
        echo -n "Testing: $name... "
        if eval "$command" > /dev/null 2>&1; then
            echo "✓ PASS"
        else
            echo "✗ FAIL"
        fi
    fi
done

echo "=== Groups Error Handling Tests ==="
go run test/test_framework/main.go groups errors | while IFS='|' read -r name command; do
    if [ -n "$name" ] && [ -n "$command" ]; then
        echo -n "Testing: $name... "
        if eval "$command" > /dev/null 2>&1; then
            # For error tests, command failing is expected
            echo "✓ PASS (error handled)"
        else
            echo "✗ FAIL (expected error)"
        fi
    fi
done

echo "✅ Groups framework tests completed"
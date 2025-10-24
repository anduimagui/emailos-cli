#!/bin/bash

echo "🧪 Running help command tests..."

go run test/test_framework/main.go help | while IFS='|' read -r name command; do
    if [ -n "$name" ] && [ -n "$command" ]; then
        echo -n "Testing: $name... "
        if eval "$command" > /dev/null 2>&1; then
            echo "✓ PASS"
        else
            echo "✗ FAIL"
        fi
    fi
done

echo "✅ Help command tests completed"
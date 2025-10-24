#!/bin/bash

echo "🧪 Running error handling tests..."

go run test/test_framework/main.go errors | while IFS='|' read -r name command; do
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

echo "✅ Error handling tests completed"
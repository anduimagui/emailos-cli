# Polar License Integration

## Overview
EmailOS now integrates with Polar for license key validation. This ensures that only users with valid licenses can use the application.

## Configuration Required

### 1. Update Organization ID
Before deploying, you must update the organization ID in `license.go`:

```go
PolarOrganizationID = "your-actual-org-id" // Replace with your Polar org ID
```

You can find your organization ID in your Polar dashboard at https://polar.sh

### 2. Create a Product and License Keys
1. Log into your Polar account
2. Create a product for EmailOS
3. Generate license keys for your customers
4. Share the checkout URL: https://email-os.com/checkout

## How It Works

### License Validation Flow
1. **On Setup**: Users must provide a valid license key during initial setup
2. **Validation**: The key is validated against Polar's API
3. **Caching**: Valid licenses are cached for 24 hours
4. **Periodic Checks**: License is re-validated periodically during:
   - Email sending
   - Email reading
5. **Offline Grace Period**: 7-day grace period for offline operation

### Files Modified
- `license.go`: Core license validation logic
- `config.go`: Added LicenseKey field to Config struct
- `setup.go`: Integrated license validation into setup flow
- `send.go`: Added periodic license check
- `read.go`: Added periodic license check
- `go.mod`: Added Polar SDK dependency

### Cache Storage
License validation results are cached in:
- Local: `.email/license_cache.json`
- Global: `~/.email/license_cache.json`

Cache includes:
- License key
- Validation timestamp
- Expiration time (24 hours)
- Customer information
- License status

### Security
- License keys are stored in `config.json` with restricted permissions (600)
- License cache is also stored with restricted permissions (600)
- Keys are validated server-side through Polar's API

## Testing

### Test License Validation
1. Run `mailos setup`
2. Enter an invalid key - should redirect to checkout
3. Enter a valid key - should proceed with setup

### Test Periodic Validation
1. Complete setup with valid license
2. Send an email with `mailos send`
3. Read emails with `mailos read`
4. Both should work without re-prompting for license

### Test Offline Mode
1. Validate a license successfully
2. Disconnect from internet
3. Commands should still work for 7 days (grace period)

## API Reference
The integration uses Polar's license validation endpoint:
- Endpoint: `POST /v1/customer-portal/license-keys/validate`
- Documentation: https://docs.polar.sh/api

## Support
For issues with license validation:
1. Verify the organization ID is correct
2. Check that the license key is active in Polar dashboard
3. Ensure internet connectivity for initial validation
4. Check cache files in `.email/` directory
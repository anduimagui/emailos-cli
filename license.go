package mailos

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	polargo "github.com/polarsource/polar-go"
	"github.com/polarsource/polar-go/models/components"
)

const (
	// Organization ID for email-os.com on Polar
	// IMPORTANT: Replace this with your actual Polar organization ID
	// You can find this in your Polar dashboard at https://polar.sh
	PolarOrganizationID = "6f94c751-4407-41bf-8397-54f09549e547" // TODO: Replace with actual org ID

	// Cache duration for license validation
	LicenseCacheDuration = 24 * time.Hour

	// Grace period for expired cache (allows offline usage)
	LicenseGracePeriod = 7 * 24 * time.Hour
)

type LicenseCache struct {
	Key           string    `json:"key"`
	ValidatedAt   time.Time `json:"validated_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	CustomerID    string    `json:"customer_id,omitempty"`
	CustomerEmail string    `json:"customer_email,omitempty"`
	Status        string    `json:"status"`
}

type LicenseManager struct {
	mu        sync.RWMutex
	cache     *LicenseCache
	cachePath string
	client    *polargo.Polar
}

var (
	licenseManager *LicenseManager
	once           sync.Once
)

// GetLicenseManager returns the singleton license manager instance
func GetLicenseManager() *LicenseManager {
	once.Do(func() {
		cachePath, _ := getLicenseCachePath()
		licenseManager = &LicenseManager{
			cachePath: cachePath,
			client:    polargo.New(), // No auth token needed for license validation
		}
		// Try to load cached license
		licenseManager.loadCache()
	})
	return licenseManager
}

func getLicenseCachePath() (string, error) {
	// Store license cache alongside config
	configPath, err := GetConfigPath()
	if err != nil {
		return "", err
	}

	dir := filepath.Dir(configPath)
	return filepath.Join(dir, "license_cache.json"), nil
}

// ValidateLicense validates a license key with Polar API
func (lm *LicenseManager) ValidateLicense(key string) error {
	// Check cache first
	if lm.isValidCached(key) {
		return nil
	}

	// Validate with Polar API
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	res, err := lm.client.CustomerPortal.LicenseKeys.Validate(ctx, components.LicenseKeyValidate{
		Key:            key,
		OrganizationID: PolarOrganizationID,
	})

	if err != nil {
		// Check if we have a valid cached license within grace period
		if lm.isWithinGracePeriod(key) {
			// Allow offline usage within grace period
			return nil
		}
		return fmt.Errorf("failed to validate license: %v", err)
	}

	if res.ValidatedLicenseKey == nil {
		return fmt.Errorf("invalid license key")
	}

	// Check license status
	validatedKey := res.ValidatedLicenseKey
	if validatedKey.Status != components.LicenseKeyStatusGranted {
		return fmt.Errorf("license key is not active (status: %s)", validatedKey.Status)
	}

	// Cache the validated license
	lm.cacheValidation(key, validatedKey)

	return nil
}

// isValidCached checks if we have a valid cached license
func (lm *LicenseManager) isValidCached(key string) bool {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	if lm.cache == nil || lm.cache.Key != key {
		return false
	}

	// Check if cache is still valid
	return time.Now().Before(lm.cache.ExpiresAt)
}

// isWithinGracePeriod checks if we're within the grace period for offline usage
func (lm *LicenseManager) isWithinGracePeriod(key string) bool {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	if lm.cache == nil || lm.cache.Key != key {
		return false
	}

	// Allow usage up to grace period after cache expiry
	graceExpiry := lm.cache.ExpiresAt.Add(LicenseGracePeriod)
	return time.Now().Before(graceExpiry)
}

// cacheValidation stores the validated license in cache
func (lm *LicenseManager) cacheValidation(key string, validated *components.ValidatedLicenseKey) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	cache := &LicenseCache{
		Key:         key,
		ValidatedAt: time.Now(),
		ExpiresAt:   time.Now().Add(LicenseCacheDuration),
		Status:      string(validated.Status),
	}

	// Store customer info if available
	// Customer is always present as a value type
	cache.CustomerID = validated.Customer.ID
	cache.CustomerEmail = validated.Customer.Email

	lm.cache = cache
	lm.saveCache()
}

// loadCache loads the license cache from disk
func (lm *LicenseManager) loadCache() {
	if lm.cachePath == "" {
		return
	}

	data, err := os.ReadFile(lm.cachePath)
	if err != nil {
		return // Cache doesn't exist or can't be read
	}

	var cache LicenseCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return // Invalid cache format
	}

	lm.mu.Lock()
	lm.cache = &cache
	lm.mu.Unlock()
}

// saveCache saves the license cache to disk
func (lm *LicenseManager) saveCache() {
	if lm.cachePath == "" || lm.cache == nil {
		return
	}

	data, err := json.MarshalIndent(lm.cache, "", "  ")
	if err != nil {
		return
	}

	// Save with restricted permissions
	os.WriteFile(lm.cachePath, data, 0600)
}

// GetCachedLicense returns the cached license info if available
func (lm *LicenseManager) GetCachedLicense() *LicenseCache {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	if lm.cache == nil {
		return nil
	}

	// Return a copy to prevent external modification
	cacheCopy := *lm.cache
	return &cacheCopy
}

// ClearCache removes the cached license
func (lm *LicenseManager) ClearCache() {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.cache = nil
	if lm.cachePath != "" {
		os.Remove(lm.cachePath)
	}
}

// IsInGracePeriod checks if we're within the grace period for offline operation
func (lm *LicenseManager) IsInGracePeriod() bool {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	if lm.cache == nil {
		return false
	}

	// Check if we're within the grace period (7 days)
	gracePeriod := 7 * 24 * time.Hour
	return time.Since(lm.cache.ValidatedAt) < gracePeriod
}

// ShouldCheckLicense determines if we should perform a license check
// This is used for periodic validation during function calls
func (lm *LicenseManager) ShouldCheckLicense() bool {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	if lm.cache == nil {
		return true // No cache, should check
	}

	// Check every 24 hours
	nextCheck := lm.cache.ValidatedAt.Add(24 * time.Hour)
	return time.Now().After(nextCheck)
}

// QuickValidate performs a quick validation using cache when possible
// This is used for periodic checks during function calls
func (lm *LicenseManager) QuickValidate(key string) error {
	// If we have a valid cache and it's recent, skip API call
	if lm.isValidCached(key) && !lm.ShouldCheckLicense() {
		return nil
	}

	// Perform full validation
	return lm.ValidateLicense(key)
}

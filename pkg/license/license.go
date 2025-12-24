package license

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
)

// License tiers
const (
	TierBasic      = "BASIC"
	TierPro        = "PRO"
	TierEnterprise = "ENTERPRISE"
)

// License errors
var (
	ErrInvalidFormat    = errors.New("invalid license format")
	ErrInvalidSignature = errors.New("invalid license signature")
	ErrExpired          = errors.New("license has expired")
	ErrModuleMismatch   = errors.New("license module mismatch")
)

// License represents a parsed license key
type License struct {
	IModule     string
	Tier       string
	ExpiryDate time.Time
	Signature  string
}

// Parse extracts license components from license key
// Format: PROMENADE-MODULE-TIER-YYYYMMDD-SIGNATURE
func Parse(licenseKey string) (*License, error) {
	if licenseKey == "" {
		return nil, ErrInvalidFormat
	}

	parts := strings.Split(licenseKey, "-")
	if len(parts) < 5 {
		return nil, ErrInvalidFormat
	}

	// Validate prefix
	if parts[0] != "PROMENADE" {
		return nil, ErrInvalidFormat
	}

	// Parse expiry date
	expiryDate, err := time.Parse("20060102", parts[3])
	if err != nil {
		return nil, fmt.Errorf("%w: invalid expiry date", ErrInvalidFormat)
	}

	return &License{
		IModule:     parts[1],
		Tier:       parts[2],
		ExpiryDate: expiryDate,
		Signature:  strings.Join(parts[4:], "-"), // Join remaining parts as signature
	}, nil
}

// Validate checks license validity against secret and module name
func (l *License) Validate(secret, expectedModule string, gracePeriodDays int) error {
	// Check module match
	if !strings.EqualFold(l.IModule, expectedModule) {
		return ErrModuleMismatch
	}

	// Verify signature
	expectedSignature := l.generateSignature(secret)
	if !hmac.Equal([]byte(expectedSignature), []byte(l.Signature)) {
		return ErrInvalidSignature
	}

	// Check expiry with grace period
	if l.IsExpired() {
		daysExpired := -l.DaysUntilExpiry()
		if daysExpired > gracePeriodDays {
			return ErrExpired
		}
	}

	return nil
}

// IsExpired returns true if license has expired
func (l *License) IsExpired() bool {
	return time.Now().After(l.ExpiryDate)
}

// DaysUntilExpiry returns days until expiry (negative if expired)
func (l *License) DaysUntilExpiry() int {
	duration := time.Until(l.ExpiryDate)
	return int(duration.Hours() / 24)
}

// generateSignature creates HMAC-SHA256 signature
func (l *License) generateSignature(secret string) string {
	data := fmt.Sprintf("PROMENADE-%s-%s-%s",
		l.IModule,
		l.Tier,
		l.ExpiryDate.Format("20060102"))

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))

	// Remove padding
	return strings.TrimRight(signature, "=")
}

// Generate creates a new license key (for testing/tooling)
func Generate(secret, module, tier string, expiryDate time.Time) string {
	license := &License{
		IModule:     strings.ToUpper(module),
		Tier:       strings.ToUpper(tier),
		ExpiryDate: expiryDate,
	}

	signature := license.generateSignature(secret)
	license.Signature = signature

	return fmt.Sprintf("PROMENADE-%s-%s-%s-%s",
		license.IModule,
		license.Tier,
		expiryDate.Format("20060102"),
		signature)
}

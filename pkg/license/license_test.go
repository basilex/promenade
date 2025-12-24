package license

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-key-12345"

func TestGenerate(t *testing.T) {
	expiresAt := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	key := Generate(testSecret, "audit", TierPro, expiresAt)

	assert.NotEmpty(t, key)
	assert.Contains(t, key, "PROMENADE-AUDIT-PRO-20261231")
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "valid license",
			key:     Generate(testSecret, "audit", TierPro, time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)),
			wantErr: false,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
		},
		{
			name:    "invalid format - too few parts",
			key:     "PROMENADE-AUDIT-PRO",
			wantErr: true,
		},
		{
			name:    "invalid format - wrong prefix",
			key:     "INVALID-AUDIT-PRO-20261231-signature",
			wantErr: true,
		},
		{
			name:    "invalid date format",
			key:     "PROMENADE-AUDIT-PRO-2026-signature",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			license, err := Parse(tt.key)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, license)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, license)
			}
		})
	}
}

func TestLicense_Validate(t *testing.T) {
	tests := []struct {
		name            string
		license         func() *License
		moduleName      string
		gracePeriodDays int
		wantErr         error
	}{
		{
			name: "valid license",
			license: func() *License {
				key := Generate(testSecret, "audit", TierPro, time.Now().Add(365*24*time.Hour))
				l, _ := Parse(key)
				return l
			},
			moduleName:      "audit",
			gracePeriodDays: 7,
			wantErr:         nil,
		},
		{
			name: "module mismatch",
			license: func() *License {
				key := Generate(testSecret, "audit", TierPro, time.Now().Add(365*24*time.Hour))
				l, _ := Parse(key)
				return l
			},
			moduleName:      "warehouse",
			gracePeriodDays: 7,
			wantErr:         ErrModuleMismatch,
		},
		{
			name: "expired without grace period",
			license: func() *License {
				key := Generate(testSecret, "audit", TierPro, time.Now().Add(-10*24*time.Hour))
				l, _ := Parse(key)
				return l
			},
			moduleName:      "audit",
			gracePeriodDays: 0,
			wantErr:         ErrExpired,
		},
		{
			name: "expired but within grace period",
			license: func() *License {
				key := Generate(testSecret, "audit", TierPro, time.Now().Add(-3*24*time.Hour))
				l, _ := Parse(key)
				return l
			},
			moduleName:      "audit",
			gracePeriodDays: 7,
			wantErr:         nil,
		},
		{
			name: "invalid signature",
			license: func() *License {
				key := Generate(testSecret, "audit", TierPro, time.Now().Add(365*24*time.Hour))
				l, _ := Parse(key)
				// Tamper with signature
				l.Signature = "invalid-signature"
				return l
			},
			moduleName:      "audit",
			gracePeriodDays: 7,
			wantErr:         ErrInvalidSignature,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			license := tt.license()
			require.NotNil(t, license)

			err := license.Validate(testSecret, tt.moduleName, tt.gracePeriodDays)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLicense_IsExpired(t *testing.T) {
	tests := []struct {
		name    string
		license func() *License
		want    bool
	}{
		{
			name: "not expired",
			license: func() *License {
				key := Generate(testSecret, "audit", TierPro, time.Now().Add(365*24*time.Hour))
				l, _ := Parse(key)
				return l
			},
			want: false,
		},
		{
			name: "expired",
			license: func() *License {
				key := Generate(testSecret, "audit", TierPro, time.Now().Add(-1*24*time.Hour))
				l, _ := Parse(key)
				return l
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			license := tt.license()
			assert.Equal(t, tt.want, license.IsExpired())
		})
	}
}

func TestLicense_DaysUntilExpiry(t *testing.T) {
	tests := []struct {
		name    string
		license func() *License
		wantMin int
		wantMax int
	}{
		{
			name: "expires in 30 days",
			license: func() *License {
				key := Generate(testSecret, "audit", TierPro, time.Now().Add(30*24*time.Hour))
				l, _ := Parse(key)
				return l
			},
			wantMin: 29,
			wantMax: 30,
		},
		{
			name: "expired 5 days ago",
			license: func() *License {
				key := Generate(testSecret, "audit", TierPro, time.Now().Add(-5*24*time.Hour))
				l, _ := Parse(key)
				return l
			},
			wantMin: -6,
			wantMax: -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			license := tt.license()
			days := license.DaysUntilExpiry()
			assert.GreaterOrEqual(t, days, tt.wantMin)
			assert.LessOrEqual(t, days, tt.wantMax)
		})
	}
}

package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	module := flag.String("module", "", "IModule name (analytics, warehouse, etc.)")
	tier := flag.String("tier", "BASIC", "License tier (BASIC, PRO, ENTERPRISE)")
	expiry := flag.String("expiry", "", "Expiry date (YYYYMMDD)")
	secret := flag.String("secret", "", "Secret key for signing")

	flag.Parse()

	// Validate required flags
	if *module == "" || *expiry == "" || *secret == "" {
		fmt.Fprintf(os.Stderr, "Error: module, expiry, and secret are required\n")
		flag.Usage()
		os.Exit(1)
	}

	// Validate tier
	upperTier := strings.ToUpper(*tier)
	if upperTier != "BASIC" && upperTier != "PRO" && upperTier != "ENTERPRISE" {
		fmt.Fprintf(os.Stderr, "Error: tier must be BASIC, PRO, or ENTERPRISE\n")
		os.Exit(1)
	}

	// Validate expiry date format
	if _, err := time.Parse("20060102", *expiry); err != nil {
		fmt.Fprintf(os.Stderr, "Error: expiry must be in YYYYMMDD format\n")
		os.Exit(1)
	}

	// Generate license
	license := generateLicense(strings.ToUpper(*module), upperTier, *expiry, *secret)
	fmt.Println(license)
}

func generateLicense(module, tier, expiry, secret string) string {
	// Format: PROMENADE-MODULE-TIER-EXPIRY-SIGNATURE
	parts := []string{"PROMENADE", module, tier, expiry}
	data := strings.Join(parts, "-")

	// Generate HMAC-SHA256 signature
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))

	// Remove padding
	signature = strings.TrimRight(signature, "=")

	return fmt.Sprintf("%s-%s", data, signature)
}

package checkbox

import "fmt"

// ConvertPaymentType converts Promenade payment type to Checkbox format
func ConvertPaymentType(paymentType string) (string, error) {
	switch paymentType {
	case "cash":
		return "CASH", nil
	case "card":
		return "CARD", nil
	case "cashless":
		return "CASHLESS", nil
	default:
		return "", fmt.Errorf("unsupported payment type: %s", paymentType)
	}
}

// ConvertToKopiyky converts float amount to kopiyky (cents)
func ConvertToKopiyky(amount float64) int {
	return int(amount * 100)
}

// ConvertFromKopiyky converts kopiyky to float amount
func ConvertFromKopiyky(kopiyky int) float64 {
	return float64(kopiyky) / 100.0
}

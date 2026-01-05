package shared

import (
	"context"
	"log/slog"

	"github.com/basilex/promenade/internal/contexts/shared/currency"
	"github.com/basilex/promenade/pkg/logger"
)

// newCurrency is a helper to create a currency with all fields
func newCurrency(code, numericCode, name, symbol string, decimalPlaces, rounding int, isCrypto, isActive bool) *currency.Currency {
	c, _ := currency.NewCurrency(code, name, symbol, decimalPlaces)
	c.NumericCode = numericCode
	c.Rounding = rounding
	c.IsCrypto = isCrypto
	c.IsActive = isActive
	return c
}

// CurrenciesData returns reference currency data for seeding (50+ fiat + crypto currencies)
func CurrenciesData() []*currency.Currency {
	return []*currency.Currency{
		// Major World Currencies
		newCurrency("USD", "840", "US Dollar", "$", 2, 0, false, true),
		newCurrency("EUR", "978", "Euro", "€", 2, 0, false, true),
		newCurrency("GBP", "826", "British Pound", "£", 2, 0, false, true),
		newCurrency("JPY", "392", "Japanese Yen", "¥", 0, 0, false, true),
		newCurrency("CNY", "156", "Chinese Yuan", "¥", 2, 0, false, true),
		newCurrency("CHF", "756", "Swiss Franc", "CHF", 2, 0, false, true),
		newCurrency("CAD", "124", "Canadian Dollar", "CA$", 2, 0, false, true),
		newCurrency("AUD", "036", "Australian Dollar", "A$", 2, 0, false, true),

		// European Currencies
		newCurrency("UAH", "980", "Ukrainian Hryvnia", "₴", 2, 0, false, true),
		newCurrency("PLN", "985", "Polish Zloty", "zł", 2, 0, false, true),
		newCurrency("SEK", "752", "Swedish Krona", "kr", 2, 0, false, true),
		newCurrency("NOK", "578", "Norwegian Krone", "kr", 2, 0, false, true),
		newCurrency("DKK", "208", "Danish Krone", "kr", 2, 0, false, true),
		newCurrency("CZK", "203", "Czech Koruna", "Kč", 2, 0, false, true),
		newCurrency("HUF", "348", "Hungarian Forint", "Ft", 2, 0, false, true),
		newCurrency("RON", "946", "Romanian Leu", "lei", 2, 0, false, true),
		newCurrency("BGN", "975", "Bulgarian Lev", "лв", 2, 0, false, true),

		// Asian Currencies
		newCurrency("INR", "356", "Indian Rupee", "₹", 2, 0, false, true),
		newCurrency("KRW", "410", "South Korean Won", "₩", 0, 0, false, true),
		newCurrency("SGD", "702", "Singapore Dollar", "S$", 2, 0, false, true),
		newCurrency("HKD", "344", "Hong Kong Dollar", "HK$", 2, 0, false, true),
		newCurrency("THB", "764", "Thai Baht", "฿", 2, 0, false, true),
		newCurrency("MYR", "458", "Malaysian Ringgit", "RM", 2, 0, false, true),
		newCurrency("IDR", "360", "Indonesian Rupiah", "Rp", 2, 0, false, true),
		newCurrency("PHP", "608", "Philippine Peso", "₱", 2, 0, false, true),
		newCurrency("VND", "704", "Vietnamese Dong", "₫", 0, 0, false, true),

		// Middle East
		newCurrency("AED", "784", "UAE Dirham", "د.إ", 2, 0, false, true),
		newCurrency("SAR", "682", "Saudi Riyal", "", 2, 0, false, true),
		newCurrency("ILS", "376", "Israeli Shekel", "₪", 2, 0, false, true),
		newCurrency("TRY", "949", "Turkish Lira", "₺", 2, 0, false, true),

		// Americas
		newCurrency("MXN", "484", "Mexican Peso", "Mex$", 2, 0, false, true),
		newCurrency("BRL", "986", "Brazilian Real", "R$", 2, 0, false, true),
		newCurrency("ARS", "032", "Argentine Peso", "AR$", 2, 0, false, true),
		newCurrency("CLP", "152", "Chilean Peso", "CL$", 0, 0, false, true),
		newCurrency("COP", "170", "Colombian Peso", "COL$", 2, 0, false, true),

		// Africa
		newCurrency("ZAR", "710", "South African Rand", "R", 2, 0, false, true),
		newCurrency("EGP", "818", "Egyptian Pound", "E£", 2, 0, false, true),
		newCurrency("NGN", "566", "Nigerian Naira", "₦", 2, 0, false, true),
		newCurrency("KES", "404", "Kenyan Shilling", "KSh", 2, 0, false, true),

		// Oceania
		newCurrency("NZD", "554", "New Zealand Dollar", "NZ$", 2, 0, false, true),

		// Other Notable
		newCurrency("RUB", "643", "Russian Ruble", "₽", 2, 0, false, true),

		// Cryptocurrencies
		newCurrency("BTC", "900", "Bitcoin", "₿", 8, 2, true, true),
		newCurrency("ETH", "901", "Ethereum", "Ξ", 8, 2, true, true),
		newCurrency("USDT", "902", "Tether", "₮", 6, 0, true, true),
		newCurrency("USDC", "903", "USD Coin", "USDC", 6, 0, true, true),
		newCurrency("BNB", "904", "Binance Coin", "BNB", 8, 2, true, true),
		newCurrency("XRP", "905", "Ripple", "XRP", 6, 2, true, true),
		newCurrency("ADA", "906", "Cardano", "₳", 6, 2, true, true),
		newCurrency("SOL", "907", "Solana", "SOL", 9, 2, true, true),
		newCurrency("DOT", "908", "Polkadot", "DOT", 10, 2, true, true),
		newCurrency("DOGE", "909", "Dogecoin", "Ð", 8, 2, true, true),
	}
}

// SeedCurrencies inserts currency data into database
func SeedCurrencies(ctx context.Context, repo currency.IRepository) error {
	currencies := CurrenciesData()
	logger.FromContext(ctx).Info("Seeding currencies...", slog.Int("count", len(currencies)))

	for _, c := range currencies {
		if err := repo.Create(ctx, c); err != nil {
			return err
		}
	}

	logger.FromContext(ctx).Info("Currencies seeded successfully", slog.Int("count", len(currencies)))
	return nil
}

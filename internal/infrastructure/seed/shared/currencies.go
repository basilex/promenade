package shared

import (
	"context"
	"log/slog"

	"github.com/basilex/promenade/internal/contexts/shared/currency"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// CurrenciesData returns reference currency data for seeding (50+ fiat + crypto currencies)
func CurrenciesData() []*currency.Currency {
	return []*currency.Currency{
		// Major World Currencies
		{ID: uuidv7.New(), Code: "USD", NumericCode: "840", Name: "US Dollar", Symbol: "$", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "EUR", NumericCode: "978", Name: "Euro", Symbol: "€", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "GBP", NumericCode: "826", Name: "British Pound", Symbol: "£", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "JPY", NumericCode: "392", Name: "Japanese Yen", Symbol: "¥", DecimalPlaces: 0, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "CNY", NumericCode: "156", Name: "Chinese Yuan", Symbol: "¥", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "CHF", NumericCode: "756", Name: "Swiss Franc", Symbol: "CHF", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "CAD", NumericCode: "124", Name: "Canadian Dollar", Symbol: "CA$", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "AUD", NumericCode: "036", Name: "Australian Dollar", Symbol: "A$", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},

		// European Currencies
		{ID: uuidv7.New(), Code: "UAH", NumericCode: "980", Name: "Ukrainian Hryvnia", Symbol: "₴", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "PLN", NumericCode: "985", Name: "Polish Zloty", Symbol: "zł", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "SEK", NumericCode: "752", Name: "Swedish Krona", Symbol: "kr", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "NOK", NumericCode: "578", Name: "Norwegian Krone", Symbol: "kr", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "DKK", NumericCode: "208", Name: "Danish Krone", Symbol: "kr", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "CZK", NumericCode: "203", Name: "Czech Koruna", Symbol: "Kč", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "HUF", NumericCode: "348", Name: "Hungarian Forint", Symbol: "Ft", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "RON", NumericCode: "946", Name: "Romanian Leu", Symbol: "lei", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "BGN", NumericCode: "975", Name: "Bulgarian Lev", Symbol: "лв", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},

		// Asian Currencies
		{ID: uuidv7.New(), Code: "INR", NumericCode: "356", Name: "Indian Rupee", Symbol: "₹", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "KRW", NumericCode: "410", Name: "South Korean Won", Symbol: "₩", DecimalPlaces: 0, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "SGD", NumericCode: "702", Name: "Singapore Dollar", Symbol: "S$", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "HKD", NumericCode: "344", Name: "Hong Kong Dollar", Symbol: "HK$", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "THB", NumericCode: "764", Name: "Thai Baht", Symbol: "฿", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "MYR", NumericCode: "458", Name: "Malaysian Ringgit", Symbol: "RM", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "IDR", NumericCode: "360", Name: "Indonesian Rupiah", Symbol: "Rp", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "PHP", NumericCode: "608", Name: "Philippine Peso", Symbol: "₱", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "VND", NumericCode: "704", Name: "Vietnamese Dong", Symbol: "₫", DecimalPlaces: 0, Rounding: 0, IsCrypto: false, IsActive: true},

		// Middle East
		{ID: uuidv7.New(), Code: "AED", NumericCode: "784", Name: "UAE Dirham", Symbol: "د.إ", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "SAR", NumericCode: "682", Name: "Saudi Riyal", Symbol: "﷼", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "ILS", NumericCode: "376", Name: "Israeli Shekel", Symbol: "₪", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "TRY", NumericCode: "949", Name: "Turkish Lira", Symbol: "₺", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},

		// Americas
		{ID: uuidv7.New(), Code: "MXN", NumericCode: "484", Name: "Mexican Peso", Symbol: "Mex$", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "BRL", NumericCode: "986", Name: "Brazilian Real", Symbol: "R$", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "ARS", NumericCode: "032", Name: "Argentine Peso", Symbol: "AR$", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "CLP", NumericCode: "152", Name: "Chilean Peso", Symbol: "CL$", DecimalPlaces: 0, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "COP", NumericCode: "170", Name: "Colombian Peso", Symbol: "COL$", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},

		// Africa
		{ID: uuidv7.New(), Code: "ZAR", NumericCode: "710", Name: "South African Rand", Symbol: "R", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "EGP", NumericCode: "818", Name: "Egyptian Pound", Symbol: "E£", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "NGN", NumericCode: "566", Name: "Nigerian Naira", Symbol: "₦", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},
		{ID: uuidv7.New(), Code: "KES", NumericCode: "404", Name: "Kenyan Shilling", Symbol: "KSh", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},

		// Oceania
		{ID: uuidv7.New(), Code: "NZD", NumericCode: "554", Name: "New Zealand Dollar", Symbol: "NZ$", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},

		// Other Notable
		{ID: uuidv7.New(), Code: "RUB", NumericCode: "643", Name: "Russian Ruble", Symbol: "₽", DecimalPlaces: 2, Rounding: 0, IsCrypto: false, IsActive: true},

		// Cryptocurrencies
		{ID: uuidv7.New(), Code: "BTC", NumericCode: "900", Name: "Bitcoin", Symbol: "₿", DecimalPlaces: 8, Rounding: 2, IsCrypto: true, IsActive: true},
		{ID: uuidv7.New(), Code: "ETH", NumericCode: "901", Name: "Ethereum", Symbol: "Ξ", DecimalPlaces: 8, Rounding: 2, IsCrypto: true, IsActive: true},
		{ID: uuidv7.New(), Code: "USDT", NumericCode: "902", Name: "Tether", Symbol: "₮", DecimalPlaces: 6, Rounding: 0, IsCrypto: true, IsActive: true},
		{ID: uuidv7.New(), Code: "USDC", NumericCode: "903", Name: "USD Coin", Symbol: "USDC", DecimalPlaces: 6, Rounding: 0, IsCrypto: true, IsActive: true},
		{ID: uuidv7.New(), Code: "BNB", NumericCode: "904", Name: "Binance Coin", Symbol: "BNB", DecimalPlaces: 8, Rounding: 2, IsCrypto: true, IsActive: true},
		{ID: uuidv7.New(), Code: "XRP", NumericCode: "905", Name: "Ripple", Symbol: "XRP", DecimalPlaces: 6, Rounding: 2, IsCrypto: true, IsActive: true},
		{ID: uuidv7.New(), Code: "ADA", NumericCode: "906", Name: "Cardano", Symbol: "₳", DecimalPlaces: 6, Rounding: 2, IsCrypto: true, IsActive: true},
		{ID: uuidv7.New(), Code: "SOL", NumericCode: "907", Name: "Solana", Symbol: "SOL", DecimalPlaces: 9, Rounding: 2, IsCrypto: true, IsActive: true},
		{ID: uuidv7.New(), Code: "DOT", NumericCode: "908", Name: "Polkadot", Symbol: "DOT", DecimalPlaces: 10, Rounding: 2, IsCrypto: true, IsActive: true},
		{ID: uuidv7.New(), Code: "DOGE", NumericCode: "909", Name: "Dogecoin", Symbol: "Ð", DecimalPlaces: 8, Rounding: 2, IsCrypto: true, IsActive: true},
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

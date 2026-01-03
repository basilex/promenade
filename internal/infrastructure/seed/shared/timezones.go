package shared

import (
	"context"
	"log/slog"

	"github.com/basilex/promenade/internal/contexts/shared/timezone"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/ref"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// TimezonesData returns reference timezone data for seeding (100+ IANA timezones)
func TimezonesData() []*timezone.Timezone {
	return []*timezone.Timezone{
		// UTC & Special
		{ID: uuidv7.New(), Name: "UTC", Abbreviation: "UTC", UTCOffset: 0, CountryCode: "", DSTOffset: ref.Int(0), DisplayName: "Coordinated Universal Time", IsActive: true},
		{ID: uuidv7.New(), Name: "GMT", Abbreviation: "GMT", UTCOffset: 0, CountryCode: "GB", DSTOffset: ref.Int(0), DisplayName: "Greenwich Mean Time", IsActive: true},

		// Americas - North America
		{ID: uuidv7.New(), Name: "America/New_York", Abbreviation: "EST", UTCOffset: -18000, CountryCode: "US", DSTOffset: ref.Int(3600), DisplayName: "Eastern Time", IsActive: true},
		{ID: uuidv7.New(), Name: "America/Chicago", Abbreviation: "CST", UTCOffset: -21600, CountryCode: "US", DSTOffset: ref.Int(3600), DisplayName: "Central Time", IsActive: true},
		{ID: uuidv7.New(), Name: "America/Denver", Abbreviation: "MST", UTCOffset: -25200, CountryCode: "US", DSTOffset: ref.Int(3600), DisplayName: "Mountain Time", IsActive: true},
		{ID: uuidv7.New(), Name: "America/Los_Angeles", Abbreviation: "PST", UTCOffset: -28800, CountryCode: "US", DSTOffset: ref.Int(3600), DisplayName: "Pacific Time", IsActive: true},
		{ID: uuidv7.New(), Name: "America/Anchorage", Abbreviation: "AKST", UTCOffset: -32400, CountryCode: "US", DSTOffset: ref.Int(3600), DisplayName: "Alaska Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Pacific/Honolulu", Abbreviation: "HST", UTCOffset: -36000, CountryCode: "US", DSTOffset: ref.Int(0), DisplayName: "Hawaii Time", IsActive: true},
		{ID: uuidv7.New(), Name: "America/Toronto", Abbreviation: "EST", UTCOffset: -18000, CountryCode: "CA", DSTOffset: ref.Int(3600), DisplayName: "Eastern Time - Toronto", IsActive: true},
		{ID: uuidv7.New(), Name: "America/Vancouver", Abbreviation: "PST", UTCOffset: -28800, CountryCode: "CA", DSTOffset: ref.Int(3600), DisplayName: "Pacific Time - Vancouver", IsActive: true},
		{ID: uuidv7.New(), Name: "America/Mexico_City", Abbreviation: "CST", UTCOffset: -21600, CountryCode: "MX", DSTOffset: ref.Int(0), DisplayName: "Central Time - Mexico City", IsActive: true},

		// Americas - South America
		{ID: uuidv7.New(), Name: "America/Sao_Paulo", Abbreviation: "BRT", UTCOffset: -10800, CountryCode: "BR", DSTOffset: ref.Int(0), DisplayName: "Brasilia Time", IsActive: true},
		{ID: uuidv7.New(), Name: "America/Buenos_Aires", Abbreviation: "ART", UTCOffset: -10800, CountryCode: "AR", DSTOffset: ref.Int(0), DisplayName: "Argentina Time", IsActive: true},
		{ID: uuidv7.New(), Name: "America/Santiago", Abbreviation: "CLT", UTCOffset: -14400, CountryCode: "CL", DSTOffset: ref.Int(3600), DisplayName: "Chile Time", IsActive: true},
		{ID: uuidv7.New(), Name: "America/Bogota", Abbreviation: "COT", UTCOffset: -18000, CountryCode: "CO", DSTOffset: ref.Int(0), DisplayName: "Colombia Time", IsActive: true},
		{ID: uuidv7.New(), Name: "America/Lima", Abbreviation: "PET", UTCOffset: -18000, CountryCode: "PE", DSTOffset: ref.Int(0), DisplayName: "Peru Time", IsActive: true},
		{ID: uuidv7.New(), Name: "America/Caracas", Abbreviation: "VET", UTCOffset: -14400, CountryCode: "VE", DSTOffset: ref.Int(0), DisplayName: "Venezuela Time", IsActive: true},

		// Europe - Western
		{ID: uuidv7.New(), Name: "Europe/London", Abbreviation: "GMT", UTCOffset: 0, CountryCode: "GB", DSTOffset: ref.Int(3600), DisplayName: "British Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Europe/Dublin", Abbreviation: "GMT", UTCOffset: 0, CountryCode: "IE", DSTOffset: ref.Int(3600), DisplayName: "Irish Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Europe/Lisbon", Abbreviation: "WET", UTCOffset: 0, CountryCode: "PT", DSTOffset: ref.Int(3600), DisplayName: "Western European Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Atlantic/Reykjavik", Abbreviation: "GMT", UTCOffset: 0, CountryCode: "IS", DSTOffset: ref.Int(0), DisplayName: "Iceland Time", IsActive: true},

		// Europe - Central
		{ID: uuidv7.New(), Name: "Europe/Paris", Abbreviation: "CET", UTCOffset: 3600, CountryCode: "FR", DSTOffset: ref.Int(3600), DisplayName: "Central European Time - Paris", IsActive: true},
		{ID: uuidv7.New(), Name: "Europe/Berlin", Abbreviation: "CET", UTCOffset: 3600, CountryCode: "DE", DSTOffset: ref.Int(3600), DisplayName: "Central European Time - Berlin", IsActive: true},
		{ID: uuidv7.New(), Name: "Europe/Rome", Abbreviation: "CET", UTCOffset: 3600, CountryCode: "IT", DSTOffset: ref.Int(3600), DisplayName: "Central European Time - Rome", IsActive: true},
		{ID: uuidv7.New(), Name: "Europe/Madrid", Abbreviation: "CET", UTCOffset: 3600, CountryCode: "ES", DSTOffset: ref.Int(3600), DisplayName: "Central European Time - Madrid", IsActive: true},
		{ID: uuidv7.New(), Name: "Europe/Amsterdam", Abbreviation: "CET", UTCOffset: 3600, CountryCode: "NL", DSTOffset: ref.Int(3600), DisplayName: "Central European Time - Amsterdam", IsActive: true},
		{ID: uuidv7.New(), Name: "Europe/Brussels", Abbreviation: "CET", UTCOffset: 3600, CountryCode: "BE", DSTOffset: ref.Int(3600), DisplayName: "Central European Time - Brussels", IsActive: true},
		{ID: uuidv7.New(), Name: "Europe/Vienna", Abbreviation: "CET", UTCOffset: 3600, CountryCode: "AT", DSTOffset: ref.Int(3600), DisplayName: "Central European Time - Vienna", IsActive: true},
		{ID: uuidv7.New(), Name: "Europe/Zurich", Abbreviation: "CET", UTCOffset: 3600, CountryCode: "CH", DSTOffset: ref.Int(3600), DisplayName: "Central European Time - Zurich", IsActive: true},
		{ID: uuidv7.New(), Name: "Europe/Prague", Abbreviation: "CET", UTCOffset: 3600, CountryCode: "CZ", DSTOffset: ref.Int(3600), DisplayName: "Central European Time - Prague", IsActive: true},
		{ID: uuidv7.New(), Name: "Europe/Warsaw", Abbreviation: "CET", UTCOffset: 3600, CountryCode: "PL", DSTOffset: ref.Int(3600), DisplayName: "Central European Time - Warsaw", IsActive: true},
		{ID: uuidv7.New(), Name: "Europe/Budapest", Abbreviation: "CET", UTCOffset: 3600, CountryCode: "HU", DSTOffset: ref.Int(3600), DisplayName: "Central European Time - Budapest", IsActive: true},

		// Europe - Northern
		{ID: uuidv7.New(), Name: "Europe/Stockholm", Abbreviation: "CET", UTCOffset: 3600, CountryCode: "SE", DSTOffset: ref.Int(3600), DisplayName: "Central European Time - Stockholm", IsActive: true},
		{ID: uuidv7.New(), Name: "Europe/Oslo", Abbreviation: "CET", UTCOffset: 3600, CountryCode: "NO", DSTOffset: ref.Int(3600), DisplayName: "Central European Time - Oslo", IsActive: true},
		{ID: uuidv7.New(), Name: "Europe/Copenhagen", Abbreviation: "CET", UTCOffset: 3600, CountryCode: "DK", DSTOffset: ref.Int(3600), DisplayName: "Central European Time - Copenhagen", IsActive: true},
		{ID: uuidv7.New(), Name: "Europe/Helsinki", Abbreviation: "EET", UTCOffset: 7200, CountryCode: "FI", DSTOffset: ref.Int(3600), DisplayName: "Eastern European Time - Helsinki", IsActive: true},

		// Europe - Eastern
		{ID: uuidv7.New(), Name: "Europe/Kyiv", Abbreviation: "EET", UTCOffset: 7200, CountryCode: "UA", DSTOffset: ref.Int(3600), DisplayName: "Eastern European Time - Kyiv", IsActive: true},
		{ID: uuidv7.New(), Name: "Europe/Bucharest", Abbreviation: "EET", UTCOffset: 7200, CountryCode: "RO", DSTOffset: ref.Int(3600), DisplayName: "Eastern European Time - Bucharest", IsActive: true},
		{ID: uuidv7.New(), Name: "Europe/Sofia", Abbreviation: "EET", UTCOffset: 7200, CountryCode: "BG", DSTOffset: ref.Int(3600), DisplayName: "Eastern European Time - Sofia", IsActive: true},
		{ID: uuidv7.New(), Name: "Europe/Athens", Abbreviation: "EET", UTCOffset: 7200, CountryCode: "GR", DSTOffset: ref.Int(3600), DisplayName: "Eastern European Time - Athens", IsActive: true},
		{ID: uuidv7.New(), Name: "Europe/Riga", Abbreviation: "EET", UTCOffset: 7200, CountryCode: "LV", DSTOffset: ref.Int(3600), DisplayName: "Eastern European Time - Riga", IsActive: true},
		{ID: uuidv7.New(), Name: "Europe/Vilnius", Abbreviation: "EET", UTCOffset: 7200, CountryCode: "LT", DSTOffset: ref.Int(3600), DisplayName: "Eastern European Time - Vilnius", IsActive: true},
		{ID: uuidv7.New(), Name: "Europe/Tallinn", Abbreviation: "EET", UTCOffset: 7200, CountryCode: "EE", DSTOffset: ref.Int(3600), DisplayName: "Eastern European Time - Tallinn", IsActive: true},
		{ID: uuidv7.New(), Name: "Europe/Moscow", Abbreviation: "MSK", UTCOffset: 10800, CountryCode: "RU", DSTOffset: ref.Int(0), DisplayName: "Moscow Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Europe/Istanbul", Abbreviation: "TRT", UTCOffset: 10800, CountryCode: "TR", DSTOffset: ref.Int(0), DisplayName: "Turkey Time", IsActive: true},

		// Asia - Middle East
		{ID: uuidv7.New(), Name: "Asia/Dubai", Abbreviation: "GST", UTCOffset: 14400, CountryCode: "AE", DSTOffset: ref.Int(0), DisplayName: "Gulf Standard Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Asia/Riyadh", Abbreviation: "AST", UTCOffset: 10800, CountryCode: "SA", DSTOffset: ref.Int(0), DisplayName: "Arabia Standard Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Asia/Jerusalem", Abbreviation: "IST", UTCOffset: 7200, CountryCode: "IL", DSTOffset: ref.Int(3600), DisplayName: "Israel Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Asia/Tehran", Abbreviation: "IRST", UTCOffset: 12600, CountryCode: "IR", DSTOffset: ref.Int(3600), DisplayName: "Iran Time", IsActive: true},

		// Asia - South
		{ID: uuidv7.New(), Name: "Asia/Kolkata", Abbreviation: "IST", UTCOffset: 19800, CountryCode: "IN", DSTOffset: ref.Int(0), DisplayName: "India Standard Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Asia/Karachi", Abbreviation: "PKT", UTCOffset: 18000, CountryCode: "PK", DSTOffset: ref.Int(0), DisplayName: "Pakistan Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Asia/Dhaka", Abbreviation: "BST", UTCOffset: 21600, CountryCode: "BD", DSTOffset: ref.Int(0), DisplayName: "Bangladesh Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Asia/Colombo", Abbreviation: "IST", UTCOffset: 19800, CountryCode: "LK", DSTOffset: ref.Int(0), DisplayName: "Sri Lanka Time", IsActive: true},

		// Asia - East
		{ID: uuidv7.New(), Name: "Asia/Shanghai", Abbreviation: "CST", UTCOffset: 28800, CountryCode: "CN", DSTOffset: ref.Int(0), DisplayName: "China Standard Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Asia/Hong_Kong", Abbreviation: "HKT", UTCOffset: 28800, CountryCode: "HK", DSTOffset: ref.Int(0), DisplayName: "Hong Kong Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Asia/Tokyo", Abbreviation: "JST", UTCOffset: 32400, CountryCode: "JP", DSTOffset: ref.Int(0), DisplayName: "Japan Standard Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Asia/Seoul", Abbreviation: "KST", UTCOffset: 32400, CountryCode: "KR", DSTOffset: ref.Int(0), DisplayName: "Korea Standard Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Asia/Taipei", Abbreviation: "CST", UTCOffset: 28800, CountryCode: "TW", DSTOffset: ref.Int(0), DisplayName: "Taiwan Time", IsActive: true},

		// Asia - Southeast
		{ID: uuidv7.New(), Name: "Asia/Singapore", Abbreviation: "SGT", UTCOffset: 28800, CountryCode: "SG", DSTOffset: ref.Int(0), DisplayName: "Singapore Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Asia/Bangkok", Abbreviation: "ICT", UTCOffset: 25200, CountryCode: "TH", DSTOffset: ref.Int(0), DisplayName: "Indochina Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Asia/Ho_Chi_Minh", Abbreviation: "ICT", UTCOffset: 25200, CountryCode: "VN", DSTOffset: ref.Int(0), DisplayName: "Indochina Time - Vietnam", IsActive: true},
		{ID: uuidv7.New(), Name: "Asia/Jakarta", Abbreviation: "WIB", UTCOffset: 25200, CountryCode: "ID", DSTOffset: ref.Int(0), DisplayName: "Western Indonesia Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Asia/Manila", Abbreviation: "PHT", UTCOffset: 28800, CountryCode: "PH", DSTOffset: ref.Int(0), DisplayName: "Philippine Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Asia/Kuala_Lumpur", Abbreviation: "MYT", UTCOffset: 28800, CountryCode: "MY", DSTOffset: ref.Int(0), DisplayName: "Malaysia Time", IsActive: true},

		// Oceania
		{ID: uuidv7.New(), Name: "Australia/Sydney", Abbreviation: "AEDT", UTCOffset: 39600, CountryCode: "AU", DSTOffset: ref.Int(3600), DisplayName: "Australian Eastern Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Australia/Melbourne", Abbreviation: "AEDT", UTCOffset: 39600, CountryCode: "AU", DSTOffset: ref.Int(3600), DisplayName: "Australian Eastern Time - Melbourne", IsActive: true},
		{ID: uuidv7.New(), Name: "Australia/Brisbane", Abbreviation: "AEST", UTCOffset: 36000, CountryCode: "AU", DSTOffset: ref.Int(0), DisplayName: "Australian Eastern Time - Brisbane", IsActive: true},
		{ID: uuidv7.New(), Name: "Australia/Perth", Abbreviation: "AWST", UTCOffset: 28800, CountryCode: "AU", DSTOffset: ref.Int(0), DisplayName: "Australian Western Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Australia/Adelaide", Abbreviation: "ACDT", UTCOffset: 37800, CountryCode: "AU", DSTOffset: ref.Int(3600), DisplayName: "Australian Central Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Pacific/Auckland", Abbreviation: "NZDT", UTCOffset: 46800, CountryCode: "NZ", DSTOffset: ref.Int(3600), DisplayName: "New Zealand Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Pacific/Fiji", Abbreviation: "FJT", UTCOffset: 43200, CountryCode: "FJ", DSTOffset: ref.Int(3600), DisplayName: "Fiji Time", IsActive: true},

		// Africa
		{ID: uuidv7.New(), Name: "Africa/Cairo", Abbreviation: "EET", UTCOffset: 7200, CountryCode: "EG", DSTOffset: ref.Int(0), DisplayName: "Eastern European Time - Cairo", IsActive: true},
		{ID: uuidv7.New(), Name: "Africa/Johannesburg", Abbreviation: "SAST", UTCOffset: 7200, CountryCode: "ZA", DSTOffset: ref.Int(0), DisplayName: "South Africa Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Africa/Lagos", Abbreviation: "WAT", UTCOffset: 3600, CountryCode: "NG", DSTOffset: ref.Int(0), DisplayName: "West Africa Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Africa/Nairobi", Abbreviation: "EAT", UTCOffset: 10800, CountryCode: "KE", DSTOffset: ref.Int(0), DisplayName: "East Africa Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Africa/Casablanca", Abbreviation: "WET", UTCOffset: 0, CountryCode: "MA", DSTOffset: ref.Int(3600), DisplayName: "Western European Time - Casablanca", IsActive: true},
		{ID: uuidv7.New(), Name: "Africa/Algiers", Abbreviation: "CET", UTCOffset: 3600, CountryCode: "DZ", DSTOffset: ref.Int(0), DisplayName: "Central European Time - Algiers", IsActive: true},

		// Atlantic
		{ID: uuidv7.New(), Name: "Atlantic/Azores", Abbreviation: "AZOT", UTCOffset: -3600, CountryCode: "PT", DSTOffset: ref.Int(3600), DisplayName: "Azores Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Atlantic/Cape_Verde", Abbreviation: "CVT", UTCOffset: -3600, CountryCode: "CV", DSTOffset: ref.Int(0), DisplayName: "Cape Verde Time", IsActive: true},

		// Indian Ocean
		{ID: uuidv7.New(), Name: "Indian/Mauritius", Abbreviation: "MUT", UTCOffset: 14400, CountryCode: "MU", DSTOffset: ref.Int(0), DisplayName: "Mauritius Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Indian/Maldives", Abbreviation: "MVT", UTCOffset: 18000, CountryCode: "MV", DSTOffset: ref.Int(0), DisplayName: "Maldives Time", IsActive: true},

		// Pacific Islands
		{ID: uuidv7.New(), Name: "Pacific/Guam", Abbreviation: "ChST", UTCOffset: 36000, CountryCode: "GU", DSTOffset: ref.Int(0), DisplayName: "Chamorro Standard Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Pacific/Tahiti", Abbreviation: "TAHT", UTCOffset: -36000, CountryCode: "PF", DSTOffset: ref.Int(0), DisplayName: "Tahiti Time", IsActive: true},
		{ID: uuidv7.New(), Name: "Pacific/Samoa", Abbreviation: "SST", UTCOffset: -39600, CountryCode: "WS", DSTOffset: ref.Int(0), DisplayName: "Samoa Standard Time", IsActive: true},

		// Antarctica
		{ID: uuidv7.New(), Name: "Antarctica/McMurdo", Abbreviation: "NZDT", UTCOffset: 46800, CountryCode: "AQ", DSTOffset: ref.Int(3600), DisplayName: "New Zealand Time - McMurdo", IsActive: true},
		{ID: uuidv7.New(), Name: "Antarctica/Troll", Abbreviation: "UTC", UTCOffset: 0, CountryCode: "AQ", DSTOffset: ref.Int(7200), DisplayName: "Troll Time", IsActive: true},
	}
}

// SeedTimezones inserts timezone data into database
func SeedTimezones(ctx context.Context, repo timezone.IRepository) error {
	timezones := TimezonesData()
	logger.FromContext(ctx).Info("Seeding timezones...", slog.Int("count", len(timezones)))

	for _, tz := range timezones {
		if err := repo.Create(ctx, tz); err != nil {
			return err
		}
	}

	logger.FromContext(ctx).Info("Timezones seeded successfully", slog.Int("count", len(timezones)))
	return nil
}

package shared

import (
	"context"
	"log/slog"

	"github.com/basilex/promenade/internal/contexts/shared/timezone/aggregate"
	"github.com/basilex/promenade/internal/contexts/shared/timezone/repository"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/ref"
)

// newTimezone is a helper to create a timezone with all fields
func newTimezone(name, abbreviation string, utcOffsetSeconds int, countryCode string, dstOffset *int, displayName string, isActive bool) *aggregate.Timezone {
	tz, _ := aggregate.NewTimezone(name, abbreviation, utcOffsetSeconds)
	tz.CountryCode = countryCode
	tz.DSTOffset = dstOffset
	tz.DisplayName = displayName
	tz.IsActive = isActive
	return tz
}

// TimezonesData returns reference timezone data for seeding (100+ IANA timezones)
func TimezonesData() []*aggregate.Timezone {
	return []*aggregate.Timezone{
		// UTC & Special
		newTimezone("UTC", "UTC", 0, "", ref.Int(0), "Coordinated Universal Time", true),
		newTimezone("GMT", "GMT", 0, "GB", ref.Int(0), "Greenwich Mean Time", true),

		// Americas - North America
		newTimezone("America/New_York", "EST", -18000, "US", ref.Int(3600), "Eastern Time", true),
		newTimezone("America/Chicago", "CST", -21600, "US", ref.Int(3600), "Central Time", true),
		newTimezone("America/Denver", "MST", -25200, "US", ref.Int(3600), "Mountain Time", true),
		newTimezone("America/Los_Angeles", "PST", -28800, "US", ref.Int(3600), "Pacific Time", true),
		newTimezone("America/Anchorage", "AKST", -32400, "US", ref.Int(3600), "Alaska Time", true),
		newTimezone("Pacific/Honolulu", "HST", -36000, "US", ref.Int(0), "Hawaii Time", true),
		newTimezone("America/Toronto", "EST", -18000, "CA", ref.Int(3600), "Eastern Time - Toronto", true),
		newTimezone("America/Vancouver", "PST", -28800, "CA", ref.Int(3600), "Pacific Time - Vancouver", true),
		newTimezone("America/Mexico_City", "CST", -21600, "MX", ref.Int(0), "Central Time - Mexico City", true),

		// Americas - South America
		newTimezone("America/Sao_Paulo", "BRT", -10800, "BR", ref.Int(0), "Brasilia Time", true),
		newTimezone("America/Buenos_Aires", "ART", -10800, "AR", ref.Int(0), "Argentina Time", true),
		newTimezone("America/Santiago", "CLT", -14400, "CL", ref.Int(3600), "Chile Time", true),
		newTimezone("America/Bogota", "COT", -18000, "CO", ref.Int(0), "Colombia Time", true),
		newTimezone("America/Lima", "PET", -18000, "PE", ref.Int(0), "Peru Time", true),
		newTimezone("America/Caracas", "VET", -14400, "VE", ref.Int(0), "Venezuela Time", true),

		// Europe - Western
		newTimezone("Europe/London", "GMT", 0, "GB", ref.Int(3600), "British Time", true),
		newTimezone("Europe/Dublin", "GMT", 0, "IE", ref.Int(3600), "Irish Time", true),
		newTimezone("Europe/Lisbon", "WET", 0, "PT", ref.Int(3600), "Western European Time", true),
		newTimezone("Atlantic/Reykjavik", "GMT", 0, "IS", ref.Int(0), "Iceland Time", true),

		// Europe - Central
		newTimezone("Europe/Paris", "CET", 3600, "FR", ref.Int(3600), "Central European Time - Paris", true),
		newTimezone("Europe/Berlin", "CET", 3600, "DE", ref.Int(3600), "Central European Time - Berlin", true),
		newTimezone("Europe/Rome", "CET", 3600, "IT", ref.Int(3600), "Central European Time - Rome", true),
		newTimezone("Europe/Madrid", "CET", 3600, "ES", ref.Int(3600), "Central European Time - Madrid", true),
		newTimezone("Europe/Amsterdam", "CET", 3600, "NL", ref.Int(3600), "Central European Time - Amsterdam", true),
		newTimezone("Europe/Brussels", "CET", 3600, "BE", ref.Int(3600), "Central European Time - Brussels", true),
		newTimezone("Europe/Vienna", "CET", 3600, "AT", ref.Int(3600), "Central European Time - Vienna", true),
		newTimezone("Europe/Zurich", "CET", 3600, "CH", ref.Int(3600), "Central European Time - Zurich", true),
		newTimezone("Europe/Prague", "CET", 3600, "CZ", ref.Int(3600), "Central European Time - Prague", true),
		newTimezone("Europe/Warsaw", "CET", 3600, "PL", ref.Int(3600), "Central European Time - Warsaw", true),
		newTimezone("Europe/Budapest", "CET", 3600, "HU", ref.Int(3600), "Central European Time - Budapest", true),

		// Europe - Northern
		newTimezone("Europe/Stockholm", "CET", 3600, "SE", ref.Int(3600), "Central European Time - Stockholm", true),
		newTimezone("Europe/Oslo", "CET", 3600, "NO", ref.Int(3600), "Central European Time - Oslo", true),
		newTimezone("Europe/Copenhagen", "CET", 3600, "DK", ref.Int(3600), "Central European Time - Copenhagen", true),
		newTimezone("Europe/Helsinki", "EET", 7200, "FI", ref.Int(3600), "Eastern European Time - Helsinki", true),

		// Europe - Eastern
		newTimezone("Europe/Kyiv", "EET", 7200, "UA", ref.Int(3600), "Eastern European Time - Kyiv", true),
		newTimezone("Europe/Bucharest", "EET", 7200, "RO", ref.Int(3600), "Eastern European Time - Bucharest", true),
		newTimezone("Europe/Sofia", "EET", 7200, "BG", ref.Int(3600), "Eastern European Time - Sofia", true),
		newTimezone("Europe/Athens", "EET", 7200, "GR", ref.Int(3600), "Eastern European Time - Athens", true),
		newTimezone("Europe/Riga", "EET", 7200, "LV", ref.Int(3600), "Eastern European Time - Riga", true),
		newTimezone("Europe/Vilnius", "EET", 7200, "LT", ref.Int(3600), "Eastern European Time - Vilnius", true),
		newTimezone("Europe/Tallinn", "EET", 7200, "EE", ref.Int(3600), "Eastern European Time - Tallinn", true),
		newTimezone("Europe/Moscow", "MSK", 10800, "RU", ref.Int(0), "Moscow Time", true),
		newTimezone("Europe/Istanbul", "TRT", 10800, "TR", ref.Int(0), "Turkey Time", true),

		// Asia - Middle East
		newTimezone("Asia/Dubai", "GST", 14400, "AE", ref.Int(0), "Gulf Standard Time", true),
		newTimezone("Asia/Riyadh", "AST", 10800, "SA", ref.Int(0), "Arabia Standard Time", true),
		newTimezone("Asia/Jerusalem", "IST", 7200, "IL", ref.Int(3600), "Israel Time", true),
		newTimezone("Asia/Tehran", "IRST", 12600, "IR", ref.Int(3600), "Iran Time", true),

		// Asia - South
		newTimezone("Asia/Kolkata", "IST", 19800, "IN", ref.Int(0), "India Standard Time", true),
		newTimezone("Asia/Karachi", "PKT", 18000, "PK", ref.Int(0), "Pakistan Time", true),
		newTimezone("Asia/Dhaka", "BST", 21600, "BD", ref.Int(0), "Bangladesh Time", true),
		newTimezone("Asia/Colombo", "IST", 19800, "LK", ref.Int(0), "Sri Lanka Time", true),

		// Asia - East
		newTimezone("Asia/Shanghai", "CST", 28800, "CN", ref.Int(0), "China Standard Time", true),
		newTimezone("Asia/Hong_Kong", "HKT", 28800, "HK", ref.Int(0), "Hong Kong Time", true),
		newTimezone("Asia/Tokyo", "JST", 32400, "JP", ref.Int(0), "Japan Standard Time", true),
		newTimezone("Asia/Seoul", "KST", 32400, "KR", ref.Int(0), "Korea Standard Time", true),
		newTimezone("Asia/Taipei", "CST", 28800, "TW", ref.Int(0), "Taiwan Time", true),

		// Asia - Southeast
		newTimezone("Asia/Singapore", "SGT", 28800, "SG", ref.Int(0), "Singapore Time", true),
		newTimezone("Asia/Bangkok", "ICT", 25200, "TH", ref.Int(0), "Indochina Time", true),
		newTimezone("Asia/Ho_Chi_Minh", "ICT", 25200, "VN", ref.Int(0), "Indochina Time - Vietnam", true),
		newTimezone("Asia/Jakarta", "WIB", 25200, "ID", ref.Int(0), "Western Indonesia Time", true),
		newTimezone("Asia/Manila", "PHT", 28800, "PH", ref.Int(0), "Philippine Time", true),
		newTimezone("Asia/Kuala_Lumpur", "MYT", 28800, "MY", ref.Int(0), "Malaysia Time", true),

		// Oceania
		newTimezone("Australia/Sydney", "AEDT", 39600, "AU", ref.Int(3600), "Australian Eastern Time", true),
		newTimezone("Australia/Melbourne", "AEDT", 39600, "AU", ref.Int(3600), "Australian Eastern Time - Melbourne", true),
		newTimezone("Australia/Brisbane", "AEST", 36000, "AU", ref.Int(0), "Australian Eastern Time - Brisbane", true),
		newTimezone("Australia/Perth", "AWST", 28800, "AU", ref.Int(0), "Australian Western Time", true),
		newTimezone("Australia/Adelaide", "ACDT", 37800, "AU", ref.Int(3600), "Australian Central Time", true),
		newTimezone("Pacific/Auckland", "NZDT", 46800, "NZ", ref.Int(3600), "New Zealand Time", true),
		newTimezone("Pacific/Fiji", "FJT", 43200, "FJ", ref.Int(3600), "Fiji Time", true),

		// Africa
		newTimezone("Africa/Cairo", "EET", 7200, "EG", ref.Int(0), "Eastern European Time - Cairo", true),
		newTimezone("Africa/Johannesburg", "SAST", 7200, "ZA", ref.Int(0), "South Africa Time", true),
		newTimezone("Africa/Lagos", "WAT", 3600, "NG", ref.Int(0), "West Africa Time", true),
		newTimezone("Africa/Nairobi", "EAT", 10800, "KE", ref.Int(0), "East Africa Time", true),
		newTimezone("Africa/Casablanca", "WET", 0, "MA", ref.Int(3600), "Western European Time - Casablanca", true),
		newTimezone("Africa/Algiers", "CET", 3600, "DZ", ref.Int(0), "Central European Time - Algiers", true),

		// Atlantic
		newTimezone("Atlantic/Azores", "AZOT", -3600, "PT", ref.Int(3600), "Azores Time", true),
		newTimezone("Atlantic/Cape_Verde", "CVT", -3600, "CV", ref.Int(0), "Cape Verde Time", true),

		// Indian Ocean
		newTimezone("Indian/Mauritius", "MUT", 14400, "MU", ref.Int(0), "Mauritius Time", true),
		newTimezone("Indian/Maldives", "MVT", 18000, "MV", ref.Int(0), "Maldives Time", true),

		// Pacific Islands
		newTimezone("Pacific/Guam", "ChST", 36000, "GU", ref.Int(0), "Chamorro Standard Time", true),
		newTimezone("Pacific/Tahiti", "TAHT", -36000, "PF", ref.Int(0), "Tahiti Time", true),
		newTimezone("Pacific/Samoa", "SST", -39600, "WS", ref.Int(0), "Samoa Standard Time", true),

		// Antarctica
		newTimezone("Antarctica/McMurdo", "NZDT", 46800, "AQ", ref.Int(3600), "New Zealand Time - McMurdo", true),
		newTimezone("Antarctica/Troll", "UTC", 0, "AQ", ref.Int(7200), "Troll Time", true),
	}
}

// SeedTimezones inserts timezone data into database
func SeedTimezones(ctx context.Context, repo repository.ITimezoneRepository) error {
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

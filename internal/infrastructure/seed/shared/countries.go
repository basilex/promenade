package shared

import (
	"context"
	"log/slog"

	"github.com/basilex/promenade/internal/contexts/shared/country"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/ref"
)

// newCountry is a helper to create a country with all fields
func newCountry(code, code3, numericCode, name, nameLocal, phoneCode, capital, region, subregion, flagEmoji string,
	latitude, longitude *float64, areaKm2 *int, population *int64, translations map[string]any, isActive bool) *country.Country {
	c, _ := country.NewCountry(code, name, phoneCode)
	c.Code3 = code3
	c.NumericCode = numericCode
	c.NameLocal = nameLocal
	c.Capital = capital
	c.Region = region
	c.Subregion = subregion
	c.FlagEmoji = flagEmoji
	c.Latitude = latitude
	c.Longitude = longitude
	c.AreaKm2 = areaKm2
	c.Population = population
	c.Translations = translations
	c.IsActive = isActive
	return c
}

// CountriesData returns reference country data for seeding (30 countries with full metadata)
func CountriesData() []*country.Country {

	return []*country.Country{
		// North America
		newCountry("US", "USA", "840", "United States", "United States", "+1", "Washington, D.C.", "Americas", "Northern America", "",
			ref.Float64(38.8951), ref.To(-77.0364), ref.Int(9833520), ref.Int64(331002651),
			map[string]any{"uk": "Сполучені Штати", "de": "Vereinigte Staaten", "fr": "États-Unis"}, true),
		newCountry("CA", "CAN", "124", "Canada", "Canada", "+1", "Ottawa", "Americas", "Northern America", "",
			ref.Float64(45.4215), ref.To(-75.6972), ref.Int(9984670), ref.Int64(37742154),
			map[string]any{"uk": "Канада", "de": "Kanada", "fr": "Canada"}, true),
		newCountry("MX", "MEX", "484", "Mexico", "México", "+52", "Mexico City", "Americas", "Central America", "",
			ref.Float64(19.4326), ref.To(-99.1332), ref.Int(1964375), ref.Int64(128932753),
			map[string]any{"uk": "Мексика", "de": "Mexiko", "fr": "Mexique"}, true),

		// Western Europe
		newCountry("GB", "GBR", "826", "United Kingdom", "United Kingdom", "+44", "London", "Europe", "Northern Europe", "",
			ref.Float64(51.5074), ref.To(-0.1278), ref.Int(242495), ref.Int64(67886011),
			map[string]any{"uk": "Велика Британія", "de": "Vereinigtes Königreich", "fr": "Royaume-Uni"}, true),
		newCountry("DE", "DEU", "276", "Germany", "Deutschland", "+49", "Berlin", "Europe", "Western Europe", "",
			ref.Float64(52.5200), ref.Float64(13.4050), ref.Int(357114), ref.Int64(83783942),
			map[string]any{"uk": "Німеччина", "de": "Deutschland", "fr": "Allemagne"}, true),
		newCountry("FR", "FRA", "250", "France", "France", "+33", "Paris", "Europe", "Western Europe", "",
			ref.Float64(48.8566), ref.Float64(2.3522), ref.Int(551695), ref.Int64(65273511),
			map[string]any{"uk": "Франція", "de": "Frankreich", "fr": "France"}, true),
		newCountry("ES", "ESP", "724", "Spain", "España", "+34", "Madrid", "Europe", "Southern Europe", "",
			ref.Float64(40.4168), ref.To(-3.7038), ref.Int(505992), ref.Int64(46754778),
			map[string]any{"uk": "Іспанія", "de": "Spanien", "fr": "Espagne"}, true),
		newCountry("IT", "ITA", "380", "Italy", "Italia", "+39", "Rome", "Europe", "Southern Europe", "",
			ref.Float64(41.9028), ref.Float64(12.4964), ref.Int(301340), ref.Int64(60461826),
			map[string]any{"uk": "Італія", "de": "Italien", "fr": "Italie"}, true),
		newCountry("NL", "NLD", "528", "Netherlands", "Nederland", "+31", "Amsterdam", "Europe", "Western Europe", "",
			ref.Float64(52.3676), ref.Float64(4.9041), ref.Int(41543), ref.Int64(17134872),
			map[string]any{"uk": "Нідерланди", "de": "Niederlande", "fr": "Pays-Bas"}, true),
		newCountry("BE", "BEL", "056", "Belgium", "België", "+32", "Brussels", "Europe", "Western Europe", "",
			ref.Float64(50.8503), ref.Float64(4.3517), ref.Int(30528), ref.Int64(11589623),
			map[string]any{"uk": "Бельгія", "de": "Belgien", "fr": "Belgique"}, true),
		newCountry("CH", "CHE", "756", "Switzerland", "Schweiz", "+41", "Bern", "Europe", "Western Europe", "",
			ref.Float64(46.9480), ref.Float64(7.4474), ref.Int(41290), ref.Int64(8654622),
			map[string]any{"uk": "Швейцарія", "de": "Schweiz", "fr": "Suisse"}, true),
		newCountry("AT", "AUT", "040", "Austria", "Österreich", "+43", "Vienna", "Europe", "Western Europe", "",
			ref.Float64(48.2082), ref.Float64(16.3738), ref.Int(83871), ref.Int64(9006398),
			map[string]any{"uk": "Австрія", "de": "Österreich", "fr": "Autriche"}, true),

		// Northern Europe
		newCountry("SE", "SWE", "752", "Sweden", "Sverige", "+46", "Stockholm", "Europe", "Northern Europe", "",
			ref.Float64(59.3293), ref.Float64(18.0686), ref.Int(450295), ref.Int64(10099265),
			map[string]any{"uk": "Швеція", "de": "Schweden", "fr": "Suède"}, true),
		newCountry("NO", "NOR", "578", "Norway", "Norge", "+47", "Oslo", "Europe", "Northern Europe", "",
			ref.Float64(59.9139), ref.Float64(10.7522), ref.Int(323802), ref.Int64(5421241),
			map[string]any{"uk": "Норвегія", "de": "Norwegen", "fr": "Norvège"}, true),
		newCountry("DK", "DNK", "208", "Denmark", "Danmark", "+45", "Copenhagen", "Europe", "Northern Europe", "",
			ref.Float64(55.6761), ref.Float64(12.5683), ref.Int(43094), ref.Int64(5792202),
			map[string]any{"uk": "Данія", "de": "Dänemark", "fr": "Danemark"}, true),
		newCountry("FI", "FIN", "246", "Finland", "Suomi", "+358", "Helsinki", "Europe", "Northern Europe", "",
			ref.Float64(60.1699), ref.Float64(24.9384), ref.Int(338424), ref.Int64(5540720),
			map[string]any{"uk": "Фінляндія", "de": "Finnland", "fr": "Finlande"}, true),

		// Eastern Europe
		newCountry("UA", "UKR", "804", "Ukraine", "Україна", "+380", "Kyiv", "Europe", "Eastern Europe", "",
			ref.Float64(50.4501), ref.Float64(30.5234), ref.Int(603500), ref.Int64(43733762),
			map[string]any{"en": "Ukraine", "de": "Ukraine", "fr": "Ukraine"}, true),
		newCountry("PL", "POL", "616", "Poland", "Polska", "+48", "Warsaw", "Europe", "Eastern Europe", "",
			ref.Float64(52.2297), ref.Float64(21.0122), ref.Int(312696), ref.Int64(37846611),
			map[string]any{"uk": "Польща", "de": "Polen", "fr": "Pologne"}, true),
		newCountry("CZ", "CZE", "203", "Czech Republic", "Česko", "+420", "Prague", "Europe", "Eastern Europe", "",
			ref.Float64(50.0755), ref.Float64(14.4378), ref.Int(78865), ref.Int64(10708981),
			map[string]any{"uk": "Чехія", "de": "Tschechien", "fr": "République tchèque"}, true),

		// Asia
		newCountry("JP", "JPN", "392", "Japan", "", "+81", "Tokyo", "Asia", "Eastern Asia", "",
			ref.Float64(35.6762), ref.Float64(139.6503), ref.Int(377930), ref.Int64(126476461),
			map[string]any{"uk": "Японія", "de": "Japan", "fr": "Japon"}, true),
		newCountry("CN", "CHN", "156", "China", "", "+86", "Beijing", "Asia", "Eastern Asia", "",
			ref.Float64(39.9042), ref.Float64(116.4074), ref.Int(9596961), ref.Int64(1439323776),
			map[string]any{"uk": "Китай", "de": "China", "fr": "Chine"}, true),
		newCountry("IN", "IND", "356", "India", "भारत", "+91", "New Delhi", "Asia", "Southern Asia", "",
			ref.Float64(28.6139), ref.Float64(77.2090), ref.Int(3287590), ref.Int64(1380004385),
			map[string]any{"uk": "Індія", "de": "Indien", "fr": "Inde"}, true),
		newCountry("KR", "KOR", "410", "South Korea", "", "+82", "Seoul", "Asia", "Eastern Asia", "",
			ref.Float64(37.5665), ref.Float64(126.9780), ref.Int(100210), ref.Int64(51269185),
			map[string]any{"uk": "Південна Корея", "de": "Südkorea", "fr": "Corée du Sud"}, true),
		newCountry("SG", "SGP", "702", "Singapore", "Singapore", "+65", "Singapore", "Asia", "South-Eastern Asia", "",
			ref.Float64(1.3521), ref.Float64(103.8198), ref.Int(719), ref.Int64(5850342),
			map[string]any{"uk": "Сінгапур", "de": "Singapur", "fr": "Singapour"}, true),

		// Oceania
		newCountry("AU", "AUS", "036", "Australia", "Australia", "+61", "Canberra", "Oceania", "Australia and New Zealand", "",
			ref.To(-35.2809), ref.Float64(149.1300), ref.Int(7692024), ref.Int64(25499884),
			map[string]any{"uk": "Австралія", "de": "Australien", "fr": "Australie"}, true),
		newCountry("NZ", "NZL", "554", "New Zealand", "New Zealand", "+64", "Wellington", "Oceania", "Australia and New Zealand", "",
			ref.To(-41.2865), ref.Float64(174.7762), ref.Int(270467), ref.Int64(4822233),
			map[string]any{"uk": "Нова Зеландія", "de": "Neuseeland", "fr": "Nouvelle-Zélande"}, true),

		// South America
		newCountry("BR", "BRA", "076", "Brazil", "Brasil", "+55", "Brasília", "Americas", "South America", "",
			ref.To(-15.8267), ref.To(-47.9218), ref.Int(8515767), ref.Int64(212559417),
			map[string]any{"uk": "Бразилія", "de": "Brasilien", "fr": "Brésil"}, true),
		newCountry("AR", "ARG", "032", "Argentina", "Argentina", "+54", "Buenos Aires", "Americas", "South America", "",
			ref.To(-34.6037), ref.To(-58.3816), ref.Int(2780400), ref.Int64(45195774),
			map[string]any{"uk": "Аргентина", "de": "Argentinien", "fr": "Argentine"}, true),

		// Middle East
		newCountry("AE", "ARE", "784", "United Arab Emirates", "الإمارات العربية المتحدة", "+971", "Abu Dhabi", "Asia", "Western Asia", "",
			ref.Float64(24.4539), ref.Float64(54.3773), ref.Int(83600), ref.Int64(9890402),
			map[string]any{"uk": "Об'єднані Арабські Емірати", "de": "Vereinigte Arabische Emirate", "fr": "Émirats arabes unis"}, true),
		newCountry("IL", "ISR", "376", "Israel", "ישראל", "+972", "Jerusalem", "Asia", "Western Asia", "",
			ref.Float64(31.7683), ref.Float64(35.2137), ref.Int(20770), ref.Int64(8655535),
			map[string]any{"uk": "Ізраїль", "de": "Israel", "fr": "Israël"}, true),

		// Africa
		newCountry("ZA", "ZAF", "710", "South Africa", "South Africa", "+27", "Pretoria", "Africa", "Southern Africa", "",
			ref.To(-25.7479), ref.Float64(28.2293), ref.Int(1221037), ref.Int64(59308690),
			map[string]any{"uk": "Південна Африка", "de": "Südafrika", "fr": "Afrique du Sud"}, true),
	}
}

// SeedCountries inserts country data into database
func SeedCountries(ctx context.Context, repo country.IRepository) error {
	countries := CountriesData()
	logger.FromContext(ctx).Info("Seeding countries...", slog.Int("count", len(countries)))

	for _, c := range countries {
		if err := repo.Create(ctx, c); err != nil {
			return err
		}
	}

	logger.FromContext(ctx).Info("Countries seeded successfully", slog.Int("count", len(countries)))
	return nil
}

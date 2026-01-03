package shared

import (
	"context"
	"log/slog"

	"github.com/basilex/promenade/internal/contexts/shared/language"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/ref"
)

// newLanguage is a helper to create a language with all fields
func newLanguage(code, code3, name, nativeName, direction string, nativeSpeakers *int64, isActive bool) *language.Language {
	l, _ := language.NewLanguage(code, name, nativeName)
	l.Code3 = code3
	l.Direction = direction
	l.NativeSpeakers = nativeSpeakers
	l.IsActive = isActive
	return l
}

// LanguagesData returns reference language data for seeding (50+ major world languages)
func LanguagesData() []*language.Language {
	return []*language.Language{
		// Top 10 Most Spoken Languages
		newLanguage("en", "eng", "English", "English", "ltr", ref.Int64(380000000), true),
		newLanguage("zh", "zho", "Chinese", "中文", "ltr", ref.Int64(1300000000), true),
		newLanguage("hi", "hin", "Hindi", "हिन्दी", "ltr", ref.Int64(602000000), true),
		newLanguage("es", "spa", "Spanish", "Español", "ltr", ref.Int64(548000000), true),
		newLanguage("fr", "fra", "French", "Français", "ltr", ref.Int64(274000000), true),
		newLanguage("ar", "ara", "Arabic", "العربية", "rtl", ref.Int64(274000000), true),
		newLanguage("bn", "ben", "Bengali", "বাংলা", "ltr", ref.Int64(272000000), true),
		newLanguage("pt", "por", "Portuguese", "Português", "ltr", ref.Int64(264000000), true),
		newLanguage("ru", "rus", "Russian", "Русский", "ltr", ref.Int64(258000000), true),
		newLanguage("ur", "urd", "Urdu", "اردو", "rtl", ref.Int64(231000000), true),

		// European Languages
		newLanguage("de", "deu", "German", "Deutsch", "ltr", ref.Int64(134000000), true),
		newLanguage("uk", "ukr", "Ukrainian", "Українська", "ltr", ref.Int64(40000000), true),
		newLanguage("pl", "pol", "Polish", "Polski", "ltr", ref.Int64(45000000), true),
		newLanguage("it", "ita", "Italian", "Italiano", "ltr", ref.Int64(85000000), true),
		newLanguage("nl", "nld", "Dutch", "Nederlands", "ltr", ref.Int64(25000000), true),
		newLanguage("sv", "swe", "Swedish", "Svenska", "ltr", ref.Int64(13000000), true),
		newLanguage("no", "nor", "Norwegian", "Norsk", "ltr", ref.Int64(5000000), true),
		newLanguage("da", "dan", "Danish", "Dansk", "ltr", ref.Int64(6000000), true),
		newLanguage("fi", "fin", "Finnish", "Suomi", "ltr", ref.Int64(5000000), true),
		newLanguage("cs", "ces", "Czech", "Čeština", "ltr", ref.Int64(13000000), true),
		newLanguage("hu", "hun", "Hungarian", "Magyar", "ltr", ref.Int64(13000000), true),
		newLanguage("ro", "ron", "Romanian", "Română", "ltr", ref.Int64(26000000), true),
		newLanguage("bg", "bul", "Bulgarian", "Български", "ltr", ref.Int64(8000000), true),
		newLanguage("el", "ell", "Greek", "Ελληνικά", "ltr", ref.Int64(13000000), true),
		newLanguage("sr", "srp", "Serbian", "Српски", "ltr", ref.Int64(12000000), true),
		newLanguage("hr", "hrv", "Croatian", "Hrvatski", "ltr", ref.Int64(7000000), true),

		// Asian Languages
		newLanguage("ja", "jpn", "Japanese", "日本語", "ltr", ref.Int64(125000000), true),
		newLanguage("ko", "kor", "Korean", "한국어", "ltr", ref.Int64(81000000), true),
		newLanguage("vi", "vie", "Vietnamese", "Tiếng Việt", "ltr", ref.Int64(85000000), true),
		newLanguage("th", "tha", "Thai", "ไทย", "ltr", ref.Int64(69000000), true),
		newLanguage("id", "ind", "Indonesian", "Bahasa Indonesia", "ltr", ref.Int64(199000000), true),
		newLanguage("ms", "msa", "Malay", "Bahasa Melayu", "ltr", ref.Int64(290000000), true),
		newLanguage("tl", "tgl", "Tagalog", "Tagalog", "ltr", ref.Int64(82000000), true),

		// Middle Eastern Languages
		newLanguage("he", "heb", "Hebrew", "עברית", "rtl", ref.Int64(9000000), true),
		newLanguage("tr", "tur", "Turkish", "Türkçe", "ltr", ref.Int64(88000000), true),
		newLanguage("fa", "fas", "Persian", "فارسی", "rtl", ref.Int64(110000000), true),

		// Indian Subcontinent
		newLanguage("ta", "tam", "Tamil", "தமிழ்", "ltr", ref.Int64(81000000), true),
		newLanguage("te", "tel", "Telugu", "తెలుగు", "ltr", ref.Int64(93000000), true),
		newLanguage("mr", "mar", "Marathi", "मराठी", "ltr", ref.Int64(95000000), true),
		newLanguage("gu", "guj", "Gujarati", "ગુજરાતી", "ltr", ref.Int64(60000000), true),

		// African Languages
		newLanguage("sw", "swa", "Swahili", "Kiswahili", "ltr", ref.Int64(200000000), true),
		newLanguage("am", "amh", "Amharic", "አማርኛ", "ltr", ref.Int64(57000000), true),
		newLanguage("yo", "yor", "Yoruba", "Yorùbá", "ltr", ref.Int64(50000000), true),
		newLanguage("ha", "hau", "Hausa", "Hausa", "ltr", ref.Int64(85000000), true),

		// Other Major Languages
		newLanguage("jv", "jav", "Javanese", "Basa Jawa", "ltr", ref.Int64(82000000), true),
		newLanguage("pa", "pan", "Punjabi", "ਪੰਜਾਬੀ", "ltr", ref.Int64(125000000), true),
		newLanguage("kn", "kan", "Kannada", "ಕನ್ನಡ", "ltr", ref.Int64(56000000), true),
		newLanguage("ml", "mal", "Malayalam", "മലയാളം", "ltr", ref.Int64(38000000), true),

		// Additional European
		newLanguage("sk", "slk", "Slovak", "Slovenčina", "ltr", ref.Int64(5000000), true),
		newLanguage("lt", "lit", "Lithuanian", "Lietuvių", "ltr", ref.Int64(3000000), true),
		newLanguage("lv", "lav", "Latvian", "Latviešu", "ltr", ref.Int64(2000000), true),
		newLanguage("et", "est", "Estonian", "Eesti", "ltr", ref.Int64(1000000), true),
	}
}

// SeedLanguages inserts language data into database
func SeedLanguages(ctx context.Context, repo language.IRepository) error {
	languages := LanguagesData()
	logger.FromContext(ctx).Info("Seeding languages...", slog.Int("count", len(languages)))

	for _, l := range languages {
		if err := repo.Create(ctx, l); err != nil {
			return err
		}
	}

	logger.FromContext(ctx).Info("Languages seeded successfully", slog.Int("count", len(languages)))
	return nil
}

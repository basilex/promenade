package shared

import (
	"context"
	"log/slog"

	"github.com/basilex/promenade/internal/contexts/shared/language"
	"github.com/basilex/promenade/pkg/logger"
	"github.com/basilex/promenade/pkg/ref"
	"github.com/basilex/promenade/pkg/uuidv7"
)

// LanguagesData returns reference language data for seeding (50+ major world languages)
func LanguagesData() []*language.Language {
	return []*language.Language{
		// Top 10 Most Spoken Languages
		{ID: uuidv7.New(), Code: "en", Code3: "eng", Name: "English", NativeName: "English", Direction: "ltr", NativeSpeakers: ref.Int64(380000000), IsActive: true},
		{ID: uuidv7.New(), Code: "zh", Code3: "zho", Name: "Chinese", NativeName: "中文", Direction: "ltr", NativeSpeakers: ref.Int64(1300000000), IsActive: true},
		{ID: uuidv7.New(), Code: "hi", Code3: "hin", Name: "Hindi", NativeName: "हिन्दी", Direction: "ltr", NativeSpeakers: ref.Int64(602000000), IsActive: true},
		{ID: uuidv7.New(), Code: "es", Code3: "spa", Name: "Spanish", NativeName: "Español", Direction: "ltr", NativeSpeakers: ref.Int64(548000000), IsActive: true},
		{ID: uuidv7.New(), Code: "fr", Code3: "fra", Name: "French", NativeName: "Français", Direction: "ltr", NativeSpeakers: ref.Int64(274000000), IsActive: true},
		{ID: uuidv7.New(), Code: "ar", Code3: "ara", Name: "Arabic", NativeName: "العربية", Direction: "rtl", NativeSpeakers: ref.Int64(274000000), IsActive: true},
		{ID: uuidv7.New(), Code: "bn", Code3: "ben", Name: "Bengali", NativeName: "বাংলা", Direction: "ltr", NativeSpeakers: ref.Int64(272000000), IsActive: true},
		{ID: uuidv7.New(), Code: "pt", Code3: "por", Name: "Portuguese", NativeName: "Português", Direction: "ltr", NativeSpeakers: ref.Int64(264000000), IsActive: true},
		{ID: uuidv7.New(), Code: "ru", Code3: "rus", Name: "Russian", NativeName: "Русский", Direction: "ltr", NativeSpeakers: ref.Int64(258000000), IsActive: true},
		{ID: uuidv7.New(), Code: "ur", Code3: "urd", Name: "Urdu", NativeName: "اردو", Direction: "rtl", NativeSpeakers: ref.Int64(231000000), IsActive: true},

		// European Languages
		{ID: uuidv7.New(), Code: "de", Code3: "deu", Name: "German", NativeName: "Deutsch", Direction: "ltr", NativeSpeakers: ref.Int64(134000000), IsActive: true},
		{ID: uuidv7.New(), Code: "uk", Code3: "ukr", Name: "Ukrainian", NativeName: "Українська", Direction: "ltr", NativeSpeakers: ref.Int64(40000000), IsActive: true},
		{ID: uuidv7.New(), Code: "pl", Code3: "pol", Name: "Polish", NativeName: "Polski", Direction: "ltr", NativeSpeakers: ref.Int64(45000000), IsActive: true},
		{ID: uuidv7.New(), Code: "it", Code3: "ita", Name: "Italian", NativeName: "Italiano", Direction: "ltr", NativeSpeakers: ref.Int64(85000000), IsActive: true},
		{ID: uuidv7.New(), Code: "nl", Code3: "nld", Name: "Dutch", NativeName: "Nederlands", Direction: "ltr", NativeSpeakers: ref.Int64(25000000), IsActive: true},
		{ID: uuidv7.New(), Code: "sv", Code3: "swe", Name: "Swedish", NativeName: "Svenska", Direction: "ltr", NativeSpeakers: ref.Int64(13000000), IsActive: true},
		{ID: uuidv7.New(), Code: "no", Code3: "nor", Name: "Norwegian", NativeName: "Norsk", Direction: "ltr", NativeSpeakers: ref.Int64(5000000), IsActive: true},
		{ID: uuidv7.New(), Code: "da", Code3: "dan", Name: "Danish", NativeName: "Dansk", Direction: "ltr", NativeSpeakers: ref.Int64(6000000), IsActive: true},
		{ID: uuidv7.New(), Code: "fi", Code3: "fin", Name: "Finnish", NativeName: "Suomi", Direction: "ltr", NativeSpeakers: ref.Int64(5000000), IsActive: true},
		{ID: uuidv7.New(), Code: "cs", Code3: "ces", Name: "Czech", NativeName: "Čeština", Direction: "ltr", NativeSpeakers: ref.Int64(13000000), IsActive: true},
		{ID: uuidv7.New(), Code: "hu", Code3: "hun", Name: "Hungarian", NativeName: "Magyar", Direction: "ltr", NativeSpeakers: ref.Int64(13000000), IsActive: true},
		{ID: uuidv7.New(), Code: "ro", Code3: "ron", Name: "Romanian", NativeName: "Română", Direction: "ltr", NativeSpeakers: ref.Int64(26000000), IsActive: true},
		{ID: uuidv7.New(), Code: "bg", Code3: "bul", Name: "Bulgarian", NativeName: "Български", Direction: "ltr", NativeSpeakers: ref.Int64(8000000), IsActive: true},
		{ID: uuidv7.New(), Code: "el", Code3: "ell", Name: "Greek", NativeName: "Ελληνικά", Direction: "ltr", NativeSpeakers: ref.Int64(13000000), IsActive: true},
		{ID: uuidv7.New(), Code: "sr", Code3: "srp", Name: "Serbian", NativeName: "Српски", Direction: "ltr", NativeSpeakers: ref.Int64(12000000), IsActive: true},
		{ID: uuidv7.New(), Code: "hr", Code3: "hrv", Name: "Croatian", NativeName: "Hrvatski", Direction: "ltr", NativeSpeakers: ref.Int64(7000000), IsActive: true},

		// Asian Languages
		{ID: uuidv7.New(), Code: "ja", Code3: "jpn", Name: "Japanese", NativeName: "日本語", Direction: "ltr", NativeSpeakers: ref.Int64(125000000), IsActive: true},
		{ID: uuidv7.New(), Code: "ko", Code3: "kor", Name: "Korean", NativeName: "한국어", Direction: "ltr", NativeSpeakers: ref.Int64(81000000), IsActive: true},
		{ID: uuidv7.New(), Code: "vi", Code3: "vie", Name: "Vietnamese", NativeName: "Tiếng Việt", Direction: "ltr", NativeSpeakers: ref.Int64(85000000), IsActive: true},
		{ID: uuidv7.New(), Code: "th", Code3: "tha", Name: "Thai", NativeName: "ไทย", Direction: "ltr", NativeSpeakers: ref.Int64(69000000), IsActive: true},
		{ID: uuidv7.New(), Code: "id", Code3: "ind", Name: "Indonesian", NativeName: "Bahasa Indonesia", Direction: "ltr", NativeSpeakers: ref.Int64(199000000), IsActive: true},
		{ID: uuidv7.New(), Code: "ms", Code3: "msa", Name: "Malay", NativeName: "Bahasa Melayu", Direction: "ltr", NativeSpeakers: ref.Int64(290000000), IsActive: true},
		{ID: uuidv7.New(), Code: "tl", Code3: "tgl", Name: "Tagalog", NativeName: "Tagalog", Direction: "ltr", NativeSpeakers: ref.Int64(82000000), IsActive: true},

		// Middle Eastern Languages
		{ID: uuidv7.New(), Code: "he", Code3: "heb", Name: "Hebrew", NativeName: "עברית", Direction: "rtl", NativeSpeakers: ref.Int64(9000000), IsActive: true},
		{ID: uuidv7.New(), Code: "tr", Code3: "tur", Name: "Turkish", NativeName: "Türkçe", Direction: "ltr", NativeSpeakers: ref.Int64(88000000), IsActive: true},
		{ID: uuidv7.New(), Code: "fa", Code3: "fas", Name: "Persian", NativeName: "فارسی", Direction: "rtl", NativeSpeakers: ref.Int64(110000000), IsActive: true},

		// Indian Subcontinent
		{ID: uuidv7.New(), Code: "ta", Code3: "tam", Name: "Tamil", NativeName: "தமிழ்", Direction: "ltr", NativeSpeakers: ref.Int64(81000000), IsActive: true},
		{ID: uuidv7.New(), Code: "te", Code3: "tel", Name: "Telugu", NativeName: "తెలుగు", Direction: "ltr", NativeSpeakers: ref.Int64(93000000), IsActive: true},
		{ID: uuidv7.New(), Code: "mr", Code3: "mar", Name: "Marathi", NativeName: "मराठी", Direction: "ltr", NativeSpeakers: ref.Int64(95000000), IsActive: true},
		{ID: uuidv7.New(), Code: "gu", Code3: "guj", Name: "Gujarati", NativeName: "ગુજરાતી", Direction: "ltr", NativeSpeakers: ref.Int64(60000000), IsActive: true},

		// African Languages
		{ID: uuidv7.New(), Code: "sw", Code3: "swa", Name: "Swahili", NativeName: "Kiswahili", Direction: "ltr", NativeSpeakers: ref.Int64(200000000), IsActive: true},
		{ID: uuidv7.New(), Code: "am", Code3: "amh", Name: "Amharic", NativeName: "አማርኛ", Direction: "ltr", NativeSpeakers: ref.Int64(57000000), IsActive: true},
		{ID: uuidv7.New(), Code: "yo", Code3: "yor", Name: "Yoruba", NativeName: "Yorùbá", Direction: "ltr", NativeSpeakers: ref.Int64(50000000), IsActive: true},
		{ID: uuidv7.New(), Code: "ha", Code3: "hau", Name: "Hausa", NativeName: "Hausa", Direction: "ltr", NativeSpeakers: ref.Int64(85000000), IsActive: true},

		// Other Major Languages
		{ID: uuidv7.New(), Code: "jv", Code3: "jav", Name: "Javanese", NativeName: "Basa Jawa", Direction: "ltr", NativeSpeakers: ref.Int64(82000000), IsActive: true},
		{ID: uuidv7.New(), Code: "pa", Code3: "pan", Name: "Punjabi", NativeName: "ਪੰਜਾਬੀ", Direction: "ltr", NativeSpeakers: ref.Int64(125000000), IsActive: true},
		{ID: uuidv7.New(), Code: "kn", Code3: "kan", Name: "Kannada", NativeName: "ಕನ್ನಡ", Direction: "ltr", NativeSpeakers: ref.Int64(56000000), IsActive: true},
		{ID: uuidv7.New(), Code: "ml", Code3: "mal", Name: "Malayalam", NativeName: "മലയാളം", Direction: "ltr", NativeSpeakers: ref.Int64(38000000), IsActive: true},

		// Additional European
		{ID: uuidv7.New(), Code: "sk", Code3: "slk", Name: "Slovak", NativeName: "Slovenčina", Direction: "ltr", NativeSpeakers: ref.Int64(5000000), IsActive: true},
		{ID: uuidv7.New(), Code: "lt", Code3: "lit", Name: "Lithuanian", NativeName: "Lietuvių", Direction: "ltr", NativeSpeakers: ref.Int64(3000000), IsActive: true},
		{ID: uuidv7.New(), Code: "lv", Code3: "lav", Name: "Latvian", NativeName: "Latviešu", Direction: "ltr", NativeSpeakers: ref.Int64(2000000), IsActive: true},
		{ID: uuidv7.New(), Code: "et", Code3: "est", Name: "Estonian", NativeName: "Eesti", Direction: "ltr", NativeSpeakers: ref.Int64(1000000), IsActive: true},
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

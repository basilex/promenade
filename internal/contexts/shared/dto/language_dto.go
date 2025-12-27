package dto

import "github.com/basilex/promenade/pkg/reference"

// LanguageResponse represents a language in API responses
type LanguageResponse struct {
	ID         string `json:"id"`
	Code       string `json:"code"`
	Code3      string `json:"code3"`
	Name       string `json:"name"`
	NativeName string `json:"native_name"`
	IsActive   bool   `json:"is_active"`
}

// ToLanguageResponse converts reference.Language to LanguageResponse
func ToLanguageResponse(language *reference.Language) LanguageResponse {
	return LanguageResponse{
		ID:         language.ID.String(),
		Code:       language.Code,
		Code3:      language.Code3,
		Name:       language.Name,
		NativeName: language.NativeName,
		IsActive:   language.IsActive,
	}
}

// ToLanguageResponses converts slice of languages to response DTOs
func ToLanguageResponses(languages []reference.Language) []LanguageResponse {
	responses := make([]LanguageResponse, len(languages))
	for i, language := range languages {
		responses[i] = ToLanguageResponse(&language)
	}
	return responses
}

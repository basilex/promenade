package dto

import (
	"github.com/basilex/promenade/internal/contexts/shared/language/aggregate"
)

// LanguageResponse represents a language in API responses
type LanguageResponse struct {
	ID         string `json:"id"`
	Code       string `json:"code"`
	Code3      string `json:"code3"`
	Name       string `json:"name"`
	NativeName string `json:"native_name"`
	IsActive   bool   `json:"is_active"`
}

// ToLanguageResponse converts language entity to LanguageResponse
func ToLanguageResponse(l *aggregate.Language) LanguageResponse {
	return LanguageResponse{
		ID:         l.GetID().String(),
		Code:       l.Code,
		Code3:      l.Code3,
		Name:       l.Name,
		NativeName: l.NativeName,
		IsActive:   l.IsActive,
	}
}

// ToLanguageResponses converts slice of languages to response DTOs
func ToLanguageResponses(languages []*aggregate.Language) []LanguageResponse {
	responses := make([]LanguageResponse, len(languages))
	for i, l := range languages {
		responses[i] = ToLanguageResponse(l)
	}
	return responses
}

// CreateLanguageRequest represents request to create a language
type CreateLanguageRequest struct {
	Code       string `json:"code" binding:"required,len=2"`
	Code3      string `json:"code3" binding:"omitempty,len=3"`
	Name       string `json:"name" binding:"required,min=2,max=100"`
	NativeName string `json:"native_name" binding:"required,min=2,max=100"`
}

// UpdateLanguageRequest represents request to update a language
type UpdateLanguageRequest struct {
	Code3      string `json:"code3" binding:"omitempty,len=3"`
	Name       string `json:"name" binding:"required,min=2,max=100"`
	NativeName string `json:"native_name" binding:"required,min=2,max=100"`
	IsActive   bool   `json:"is_active"`
}

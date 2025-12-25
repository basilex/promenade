package mocks

import (
	"context"

	"github.com/basilex/promenade/internal/modules/notifications/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type MockUserPreferenceRepository struct {
	preferences map[string]*entity.UserPreference
	createErr   error
	updateErr   error
	getErr      error
}

func NewMockUserPreferenceRepository() *MockUserPreferenceRepository {
	return &MockUserPreferenceRepository{
		preferences: make(map[string]*entity.UserPreference),
	}
}

func (m *MockUserPreferenceRepository) Create(ctx context.Context, pref *entity.UserPreference) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.preferences[pref.UserID.String()] = pref
	return nil
}

func (m *MockUserPreferenceRepository) Update(ctx context.Context, pref *entity.UserPreference) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.preferences[pref.UserID.String()] = pref
	return nil
}

func (m *MockUserPreferenceRepository) GetByUserID(ctx context.Context, userID uuidv7.UUID) (*entity.UserPreference, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	pref, exists := m.preferences[userID.String()]
	if !exists {
		return nil, nil
	}
	return pref, nil
}

func (m *MockUserPreferenceRepository) Exists(ctx context.Context, userID uuidv7.UUID) (bool, error) {
	if m.getErr != nil {
		return false, m.getErr
	}
	_, exists := m.preferences[userID.String()]
	return exists, nil
}

func (m *MockUserPreferenceRepository) Delete(ctx context.Context, userID uuidv7.UUID) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	delete(m.preferences, userID.String())
	return nil
}

func (m *MockUserPreferenceRepository) Reset() {
	m.preferences = make(map[string]*entity.UserPreference)
	m.createErr = nil
	m.updateErr = nil
	m.getErr = nil
}

package mocks

import (
	"context"

	"github.com/basilex/promenade/internal/modules/notifications/domain/entity"
	"github.com/basilex/promenade/pkg/uuidv7"
)

type MockNotificationRepository struct {
	notifications map[string]*entity.Notification
	createErr     error
	updateErr     error
	getErr        error
	listErr       error
	countErr      error
}

func NewMockNotificationRepository() *MockNotificationRepository {
	return &MockNotificationRepository{
		notifications: make(map[string]*entity.Notification),
	}
}

func (m *MockNotificationRepository) Create(ctx context.Context, notification *entity.Notification) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.notifications[notification.ID.String()] = notification
	return nil
}

func (m *MockNotificationRepository) Update(ctx context.Context, notification *entity.Notification) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.notifications[notification.ID.String()] = notification
	return nil
}

func (m *MockNotificationRepository) GetByID(ctx context.Context, id uuidv7.UUID) (*entity.Notification, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	notification, exists := m.notifications[id.String()]
	if !exists {
		return nil, nil
	}
	return notification, nil
}

func (m *MockNotificationRepository) GetByUserID(ctx context.Context, userID uuidv7.UUID, limit, offset int) ([]*entity.Notification, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*entity.Notification
	for _, n := range m.notifications {
		if n.UserID == userID {
			result = append(result, n)
		}
	}
	return result, nil
}

func (m *MockNotificationRepository) CountByUserID(ctx context.Context, userID uuidv7.UUID) (int, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	var count int
	for _, n := range m.notifications {
		if n.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (m *MockNotificationRepository) GetUnreadCount(ctx context.Context, userID uuidv7.UUID) (int, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	var count int
	for _, n := range m.notifications {
		if n.UserID == userID && n.Status == entity.NotificationStatusPending {
			count++
		}
	}
	return count, nil
}

func (m *MockNotificationRepository) GetPendingNotifications(ctx context.Context, limit int) ([]*entity.Notification, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*entity.Notification
	for _, n := range m.notifications {
		if n.Status == entity.NotificationStatusPending {
			result = append(result, n)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *MockNotificationRepository) Delete(ctx context.Context, id uuidv7.UUID) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	delete(m.notifications, id.String())
	return nil
}

func (m *MockNotificationRepository) Reset() {
	m.notifications = make(map[string]*entity.Notification)
	m.createErr = nil
	m.updateErr = nil
	m.getErr = nil
	m.listErr = nil
	m.countErr = nil
}

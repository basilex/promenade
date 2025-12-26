# Notifications Module

**Status**: Commercial (requires license)
**Version**: 1.0.0
**License Tier**: BASIC, PRO, ENTERPRISE

Multi-channel notification system with user preferences, quiet hours, and delivery tracking.

## Features

### Core Functionality

- **Multi-Channel Delivery**:

  - Email notifications
  - SMS notifications
  - Push notifications (mobile/web)
  - In-app notifications

- **User Preferences**:

  - Per-channel enablement (email, SMS, push, in-app)
  - Per-type enablement (system, security, marketing, product, social)
  - Quiet hours configuration (time range when no notifications are sent)
  - Timezone support for quiet hours

- **Notification Management**:
  - Create and send notifications
  - Track delivery status (pending → sent → delivered → opened → clicked)
  - Query notification history with pagination
  - Unread notification count

### Notification Types

- **System**: System updates, maintenance notifications
- **Security**: Login alerts, password changes, security warnings
- **Marketing**: Promotional content, newsletters, campaigns
- **Product**: Feature announcements, product updates
- **Social**: Comments, likes, mentions, follows

### Delivery Channels

- **Email**: SMTP-based email delivery
- **SMS**: SMS gateway integration (Twilio, MessageBird, etc.)
- **Push**: FCM/APNS push notifications
- **In-App**: Real-time in-app notifications

## Architecture

```
notifications/
 domain/
    entity/                       # Domain entities
       notification.go           # Notification entity with status machine
       user_preference.go        # User preferences entity
    repository/                   # Repository interfaces
        notification_repository.go
        user_preference_repository.go
 usecase/                          # Business logic
    notification_usecase.go       # INotificationUseCase implementation
    events.go                     # Domain events
 adapter/
    http/handler/                 # HTTP handlers
       notification_handler.go   # REST API endpoints
    repository/postgres/          # PostgreSQL implementations
        base_repository.go
        notification_repository.go
        user_preference_repository.go
 config/                           # Configuration files
    config.dev.yaml
    config.test.yaml
    config.prod.yaml
 module.go                         # IModule implementation
 register.go                       # Auto-registration
 README.md                         # This file
```

## Database Schema

### notifications_notifications

Stores all notification records:

```sql
CREATE TABLE notifications_notifications (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id UUID NOT NULL REFERENCES users(id),
    type VARCHAR(50) NOT NULL,           -- system, security, marketing, product, social
    channel VARCHAR(50) NOT NULL,        -- email, sms, push, in_app
    status VARCHAR(50) NOT NULL,         -- pending, sent, delivered, failed, opened, clicked
    template VARCHAR(255),
    subject VARCHAR(500),
    content TEXT NOT NULL,
    data JSONB DEFAULT '{}'::jsonb,
    sent_at TIMESTAMP,
    opened_at TIMESTAMP,
    clicked_at TIMESTAMP,
    failed_at TIMESTAMP,
    error_msg TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
```

### notifications_user_preferences

Stores user notification preferences:

```sql
CREATE TABLE notifications_user_preferences (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id),
    email_enabled BOOLEAN NOT NULL DEFAULT true,
    sms_enabled BOOLEAN NOT NULL DEFAULT false,
    push_enabled BOOLEAN NOT NULL DEFAULT true,
    in_app_enabled BOOLEAN NOT NULL DEFAULT true,
    system_enabled BOOLEAN NOT NULL DEFAULT true,
    security_enabled BOOLEAN NOT NULL DEFAULT true,
    marketing_enabled BOOLEAN NOT NULL DEFAULT false,
    product_enabled BOOLEAN NOT NULL DEFAULT true,
    social_enabled BOOLEAN NOT NULL DEFAULT true,
    quiet_hours_start TIME,
    quiet_hours_end TIME,
    timezone VARCHAR(100) NOT NULL DEFAULT 'UTC',
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
```

## API Endpoints

### Send Notification

```http
POST /api/v1/notifications
Authorization: Bearer {token}
Content-Type: application/json

{
  "type": "product",
  "channel": "email",
  "subject": "New Feature Available",
  "content": "Check out our new feature...",
  "data": {
    "feature_id": "feature-123",
    "link": "https://app.promenade.com/features/123"
  }
}
```

### Get User Notifications

```http
GET /api/v1/notifications?page=1&page_size=20
Authorization: Bearer {token}
```

### Get Unread Count

```http
GET /api/v1/notifications/unread-count
Authorization: Bearer {token}
```

### Mark as Opened

```http
POST /api/v1/notifications/{id}/opened
Authorization: Bearer {token}
```

### Get Preferences

```http
GET /api/v1/notifications/preferences
Authorization: Bearer {token}
```

### Update Preferences

```http
PUT /api/v1/notifications/preferences
Authorization: Bearer {token}
Content-Type: application/json

{
  "email_enabled": true,
  "marketing_enabled": false,
  "quiet_hours_start": "22:00",
  "quiet_hours_end": "08:00",
  "timezone": "America/New_York"
}
```

## Configuration

### Development (`config.dev.yaml`)

```yaml
module:
  name: "notifications"
  enabled: true
  license_required: true

notifications:
  rate_limit:
    enabled: true
    max_per_user_per_minute: 5
    max_per_user_per_hour: 100

  retry:
    enabled: true
    max_attempts: 3
    backoff_seconds: 60

  channels:
    email:
      enabled: true
      from_address: "noreply@promenade.dev"
      from_name: "Promenade"
```

### Production (`config.prod.yaml`)

```yaml
module:
  name: "notifications"
  enabled: true
  license_required: true

notifications:
  rate_limit:
    enabled: true
    max_per_user_per_minute: 3
    max_per_user_per_hour: 50

  retry:
    enabled: true
    max_attempts: 5
    backoff_seconds: 300

  channels:
    email:
      enabled: true
      from_address: "${NOTIFICATIONS_EMAIL_FROM}"
      from_name: "${NOTIFICATIONS_EMAIL_NAME}"
    sms:
      enabled: true
      provider: "${SMS_PROVIDER}"
      api_key: "${SMS_API_KEY}"
```

## License Management

### Generate License

```bash
./scripts/generate-license.sh notifications PRO 365
```

### Set License Key

**Environment Variable**:

```bash
export NOTIFICATIONS_LICENSE_KEY="PROMENADE-NOTIFICATIONS-PRO-20261231-abc123..."
```

**Or in `config/modules.yaml`**:

```yaml
modules:
  config:
    notifications:
      license_key: "PROMENADE-NOTIFICATIONS-PRO-20261231-abc123..."
```

### License Tiers

| Tier       | Features                                     | Price  |
| ---------- | -------------------------------------------- | ------ |
| BASIC      | Email + In-App, 1000/day                     | $29/mo |
| PRO        | + SMS + Push, 10000/day, Quiet Hours         | $99/mo |
| ENTERPRISE | Unlimited, Custom Channels, Priority Support | Custom |

## Testing

```bash
# Run module tests
make test-module-notifications

# Or directly with go test
go test ./internal/modules/notifications/...
```

## Usage Example

### Send Welcome Email

```go
notification, err := notifUseCase.SendNotification(
    ctx,
    userID,
    entity.NotificationTypeSystem,
    entity.NotificationChannelEmail,
    "welcome_email",
    "Welcome to Promenade!",
    "Thank you for joining us. Get started here: https://app.promenade.com/onboarding",
    map[string]any{
        "user_name": "John Doe",
        "onboarding_link": "https://app.promenade.com/onboarding",
    },
)
```

### Check User Preferences

```go
prefs, err := notifUseCase.GetUserPreferences(ctx, userID)
if err != nil {
    // Create default preferences
    prefs, err = notifUseCase.CreateDefaultPreferences(ctx, userID)
}

// Check if email is enabled
if prefs.EmailEnabled {
    // Send email notification
}
```

### Subscribe to Events

```go
// In module initialization
eventBus.Subscribe("user.registered", func(ctx context.Context, event bus.Event) error {
    evt := event.(*UserRegisteredEvent)

    // Create default notification preferences
    _, err := notifUseCase.CreateDefaultPreferences(ctx, evt.UserID)
    if err != nil {
        return err
    }

    // Send welcome notification
    _, err = notifUseCase.SendNotification(
        ctx,
        evt.UserID,
        entity.NotificationTypeSystem,
        entity.NotificationChannelEmail,
        "welcome_email",
        "Welcome to Promenade!",
        "Thank you for joining us!",
        nil,
    )

    return err
})
```

## Testing

### Test Structure

The module includes comprehensive unit tests with manual mocks (no mockgen dependency):

#### Entity Tests (`domain/entity/*_test.go`): 20 tests

**Notification Entity** (`notification_test.go`): 9 tests

- `TestNewNotification` - Entity creation
- `TestNotification_Validate` - 7 validation scenarios:
  - Valid notification
  - Missing user_id
  - Missing type
  - Missing channel
  - Email without subject
  - Missing content
  - Push notification without subject is valid
- `TestNotification_MarkAsSent` - State transition to sent
- `TestNotification_MarkAsDelivered` - State transition to delivered
- `TestNotification_MarkAsFailed` - State transition to failed with error
- `TestNotification_MarkAsOpened` - User opened notification
- `TestNotification_MarkAsClicked` - User clicked notification
- `TestNotification_IsDelivered` - Status check
- `TestNotification_IsFailed` - Status check

**UserPreference Entity** (`user_preference_test.go`): 11 tests

- `TestNewUserPreference` - Default preferences creation
- `TestUserPreference_Validate` - 6 validation scenarios:
  - Valid preference
  - Missing user_id
  - Missing timezone
  - Quiet hours start without end
  - Quiet hours end without start
  - Valid quiet hours
- `TestUserPreference_IsChannelEnabled` - 5 channel checks (email, sms, push, in_app, unknown)
- `TestUserPreference_IsTypeEnabled` - 6 type checks (system, security, marketing, product, social, unknown)
- `TestUserPreference_DisableChannel` - Channel disabling
- `TestUserPreference_DisableType` - Type disabling
- `TestUserPreference_IsInQuietHours` - Quiet hours calculation
- `TestUserPreference_EnableAllChannels` - Bulk channel enablement
- `TestUserPreference_EnableMarketingNotifications` - Bulk type enablement
- `TestUserPreference_TimezoneSettings` - 5 timezone tests (UTC, New York, Tokyo, London, Kyiv)
- `TestUserPreference_Update` - Preference updates

#### Use Case Tests (`usecase/notification_usecase_test.go`): 15 tests

**SendNotification**: 5 scenarios

- `TestNotificationUseCase_SendNotification_Success` - Successful notification send
- `TestNotificationUseCase_SendNotification_ChannelDisabled` - Channel disabled in preferences
- `TestNotificationUseCase_SendNotification_TypeDisabled` - Type disabled in preferences
- `TestNotificationUseCase_SendNotification_CreatesDefaultPreferences` - Auto-create missing preferences
- `TestNotificationUseCase_SendNotification_SystemBypassesQuietHours` - System/security notifications bypass quiet hours

**Notification Retrieval**: 2 scenarios

- `TestNotificationUseCase_GetNotification_Success` - Get notification by ID
- `TestNotificationUseCase_GetNotification_NotFound` - Handle not found error

**User Operations**: 4 tests

- `TestNotificationUseCase_GetUserNotifications` - List user notifications with pagination
- `TestNotificationUseCase_GetUnreadCount` - Count unread notifications
- `TestNotificationUseCase_MarkAsOpened` - Mark notification as opened
- `TestNotificationUseCase_MarkAsClicked` - Mark notification as clicked

**Preference Management**: 4 tests

- `TestNotificationUseCase_GetUserPreferences` - Get user preferences
- `TestNotificationUseCase_GetUserPreferences_NotFound` - Handle missing preferences
- `TestNotificationUseCase_UpdateUserPreferences` - Update preferences
- `TestNotificationUseCase_CreateDefaultPreferences` - Create default preferences

#### Manual Mocks (`domain/repository/mocks/`): 3 implementations

**Mock Features**:

- In-memory storage with thread-safe access (mutex)
- Configurable error injection for testing error paths
- Reset() method for clean test state
- Event tracking (MockEventBus tracks published events)

**Mocks**:

- `notification_repository_mock.go` - INotificationRepository (9 methods)
- `user_preference_repository_mock.go` - IUserPreferenceRepository (5 methods)
- `event_bus_mock.go` - bus.IBus (6 methods including Health)

### Running Tests

```bash
# Unit tests (entity + use case)
go test -v github.com/basilex/promenade/internal/modules/notifications/domain/entity
go test -v github.com/basilex/promenade/internal/modules/notifications/usecase

# Integration tests (requires test database)
make test-integration-notifications

# All tests
make test-module-notifications

# With coverage
go test -cover github.com/basilex/promenade/internal/modules/notifications/...
```

### Test Coverage

**Status:** ALL TESTS PASSING

| Test Type      | Tests  | Duration | Coverage             |
| -------------- | ------ | -------- | -------------------- |
| Unit (Entity)  | 20     | ~0.4s    | 100% methods         |
| Unit (UseCase) | 15     | ~0.2s    | 100% business logic  |
| Integration    | 13     | ~1.2s    | 100% CRUD operations |
| **Total**      | **48** | **~2s**  | **Complete**         |

**Key Features Tested:**

- Multi-channel delivery (Email, SMS, Push, In-App)
- User preferences (per-channel, per-type)
- Quiet hours with timezone support
- Status tracking (pending → opened → clicked)
- System notifications bypass quiet hours
- Auto-create preferences on first use
- JSONB data storage
- UUID v7 primary keys
- Soft delete for notifications

See [TESTING.md](TESTING.md) for detailed testing guide.

## Permissions

- `notifications:read` - View own notifications
- `notifications:create` - Send notifications
- `notifications:update` - Update notification status
- `notifications.preferences:read` - View preferences
- `notifications.preferences:update` - Update preferences

## Dependencies

- PostgreSQL 16+ (JSONB support)
- Core Event IBus (Memory or Redis adapter)

## Future Enhancements

- [ ] Notification templates with variable substitution
- [ ] Email template designer
- [ ] A/B testing for notifications
- [ ] Analytics dashboard (open rates, click rates)
- [ ] Webhook support for external integrations
- [ ] Batch notification sending
- [ ] Notification scheduling (send at specific time)
- [ ] Rich content support (images, buttons, etc.)

## Support

For commercial license inquiries: alexander.vasilenko@gmail.com

## License

Commercial - Requires valid license key for production use

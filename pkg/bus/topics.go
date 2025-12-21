package bus

// Topic constants define standard event topics/channels.
// Using constants ensures consistency and prevents typos.
const (
	// User Management Topics
	TopicUserRegistered      = "user.registered"
	TopicUserActivated       = "user.activated"
	TopicUserDeactivated     = "user.deactivated"
	TopicUserSuspended       = "user.suspended"
	TopicUserBanned          = "user.banned"
	TopicUserPasswordChanged = "user.password.changed"
	TopicUserEmailVerified   = "user.email.verified"

	// Profile Topics
	TopicProfileCreated = "profile.created"
	TopicProfileUpdated = "profile.updated"
	TopicProfileDeleted = "profile.deleted"

	// Content Topics
	TopicPostCreated   = "post.created"
	TopicPostPublished = "post.published"
	TopicPostUpdated   = "post.updated"
	TopicPostDeleted   = "post.deleted"

	// Comment Topics
	TopicCommentAdded   = "comment.added"
	TopicCommentUpdated = "comment.updated"
	TopicCommentDeleted = "comment.deleted"

	// RBAC Topics
	TopicRoleAssigned = "role.assigned"
	TopicRoleRevoked  = "role.revoked"

	// Purge Topics (data cleanup)
	TopicPurgeCompleted = "purge.completed"
	TopicPurgeFailed    = "purge.failed"

	// Notification Topics (for consuming events and triggering notifications)
	TopicNotificationEmail = "notification.email"
	TopicNotificationSMS   = "notification.sms"
	TopicNotificationPush  = "notification.push"
)

// BoundedContext represents a domain context that events belong to.
type BoundedContext string

const (
	ContextUserManagement    BoundedContext = "user_management"
	ContextProfileManagement BoundedContext = "profile_management"
	ContextContentManagement BoundedContext = "content_management"
	ContextRBAC              BoundedContext = "rbac"
	ContextPurge             BoundedContext = "purge"
	ContextNotification      BoundedContext = "notification"
	ContextReferenceData     BoundedContext = "reference_data"
)

// GetContext returns the bounded context for a given topic.
func GetContext(topic string) BoundedContext {
	switch {
	case topic[:5] == "user.":
		return ContextUserManagement
	case topic[:8] == "profile.":
		return ContextProfileManagement
	case topic[:5] == "post.", topic[:8] == "comment.":
		return ContextContentManagement
	case topic[:5] == "role.":
		return ContextRBAC
	case topic[:6] == "purge.":
		return ContextPurge
	case topic[:13] == "notification.":
		return ContextNotification
	default:
		return ContextReferenceData
	}
}

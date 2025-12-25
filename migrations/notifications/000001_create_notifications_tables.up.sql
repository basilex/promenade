-- Create notifications table
CREATE TABLE IF NOT EXISTS notifications_notifications (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('system', 'security', 'marketing', 'product', 'social')),
    channel VARCHAR(50) NOT NULL CHECK (channel IN ('email', 'sms', 'push', 'in_app')),
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'delivered', 'failed', 'bounced', 'opened', 'clicked')),
    template VARCHAR(255) NOT NULL,
    subject VARCHAR(500),
    content TEXT NOT NULL,
    data JSONB,
    sent_at TIMESTAMP,
    opened_at TIMESTAMP,
    clicked_at TIMESTAMP,
    failed_at TIMESTAMP,
    error_msg TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_notifications_user FOREIGN KEY (user_id) REFERENCES core_users(id) ON DELETE CASCADE
);

-- Create indexes for notifications
CREATE INDEX idx_notifications_user_id ON notifications_notifications(user_id);
CREATE INDEX idx_notifications_status ON notifications_notifications(status);
CREATE INDEX idx_notifications_type ON notifications_notifications(type);
CREATE INDEX idx_notifications_channel ON notifications_notifications(channel);
CREATE INDEX idx_notifications_created_at ON notifications_notifications(created_at DESC);
CREATE INDEX idx_notifications_user_created ON notifications_notifications(user_id, created_at DESC);

-- Create user preferences table
CREATE TABLE IF NOT EXISTS notifications_user_preferences (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id UUID NOT NULL UNIQUE,
    email_enabled BOOLEAN NOT NULL DEFAULT true,
    sms_enabled BOOLEAN NOT NULL DEFAULT false,
    push_enabled BOOLEAN NOT NULL DEFAULT true,
    in_app_enabled BOOLEAN NOT NULL DEFAULT true,
    system_enabled BOOLEAN NOT NULL DEFAULT true,
    security_enabled BOOLEAN NOT NULL DEFAULT true,
    marketing_enabled BOOLEAN NOT NULL DEFAULT false,
    product_enabled BOOLEAN NOT NULL DEFAULT true,
    social_enabled BOOLEAN NOT NULL DEFAULT true,
    quiet_hours_start VARCHAR(5),
    quiet_hours_end VARCHAR(5),
    quiet_hours_enabled BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_preferences_user FOREIGN KEY (user_id) REFERENCES core_users(id) ON DELETE CASCADE,
    CONSTRAINT chk_quiet_hours_format CHECK (
        (quiet_hours_start IS NULL OR quiet_hours_start ~ '^([0-1][0-9]|2[0-3]):[0-5][0-9]$') AND
        (quiet_hours_end IS NULL OR quiet_hours_end ~ '^([0-1][0-9]|2[0-3]):[0-5][0-9]$')
    )
);

-- Create index for user preferences
CREATE INDEX idx_notifications_preferences_user_id ON notifications_user_preferences(user_id);

-- Add comments for documentation
COMMENT ON TABLE notifications_notifications IS 'Stores all notification records with delivery tracking';
COMMENT ON TABLE notifications_user_preferences IS 'User notification preferences and channel settings';
COMMENT ON COLUMN notifications_notifications.type IS 'Notification type: system, security, marketing, product, social';
COMMENT ON COLUMN notifications_notifications.channel IS 'Delivery channel: email, sms, push, in_app';
COMMENT ON COLUMN notifications_notifications.status IS 'Delivery status: pending, sent, delivered, failed, bounced, opened, clicked';
COMMENT ON COLUMN notifications_notifications.data IS 'Additional notification data in JSONB format';

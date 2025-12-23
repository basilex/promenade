-- Create analytics_metrics table
CREATE TABLE IF NOT EXISTS analytics_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('counter', 'gauge', 'histogram')),
    scope VARCHAR(50) NOT NULL CHECK (scope IN ('user', 'post', 'comment', 'system')),
    value DOUBLE PRECISION NOT NULL,
    entity_id UUID,
    tags JSONB DEFAULT '{}',
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for metrics
CREATE INDEX idx_analytics_metrics_name ON analytics_metrics(name);
CREATE INDEX idx_analytics_metrics_scope ON analytics_metrics(scope);
CREATE INDEX idx_analytics_metrics_entity_id ON analytics_metrics(entity_id);
CREATE INDEX idx_analytics_metrics_timestamp ON analytics_metrics(timestamp DESC);
CREATE INDEX idx_analytics_metrics_tags ON analytics_metrics USING GIN(tags);

-- Create analytics_metric_aggregates table
CREATE TABLE IF NOT EXISTS analytics_metric_aggregates (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    scope VARCHAR(50) NOT NULL,
    period VARCHAR(10) NOT NULL CHECK (period IN ('1h', '1d', '1w', '1M')),
    count BIGINT NOT NULL DEFAULT 0,
    sum DOUBLE PRECISION NOT NULL DEFAULT 0,
    avg DOUBLE PRECISION NOT NULL DEFAULT 0,
    min DOUBLE PRECISION NOT NULL DEFAULT 0,
    max DOUBLE PRECISION NOT NULL DEFAULT 0,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for aggregates
CREATE INDEX idx_analytics_aggregates_name ON analytics_metric_aggregates(name);
CREATE INDEX idx_analytics_aggregates_period ON analytics_metric_aggregates(period);
CREATE INDEX idx_analytics_aggregates_start_time ON analytics_metric_aggregates(start_time DESC);
CREATE UNIQUE INDEX idx_analytics_aggregates_unique ON analytics_metric_aggregates(name, period, start_time);

-- Create analytics_reports table
CREATE TABLE IF NOT EXISTS analytics_reports (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'processing', 'ready', 'failed')),
    format VARCHAR(20) NOT NULL CHECK (format IN ('json', 'csv', 'pdf', 'xlsx')),
    config JSONB NOT NULL DEFAULT '{}',
    result JSONB,
    error TEXT,
    file_url TEXT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for reports
CREATE INDEX idx_analytics_reports_user_id ON analytics_reports(user_id);
CREATE INDEX idx_analytics_reports_status ON analytics_reports(status);
CREATE INDEX idx_analytics_reports_type ON analytics_reports(type);
CREATE INDEX idx_analytics_reports_created_at ON analytics_reports(created_at DESC);

-- Create analytics_report_schedules table
CREATE TABLE IF NOT EXISTS analytics_report_schedules (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    report_type VARCHAR(100) NOT NULL,
    format VARCHAR(20) NOT NULL CHECK (format IN ('json', 'csv', 'pdf', 'xlsx')),
    config JSONB NOT NULL DEFAULT '{}',
    schedule VARCHAR(100) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    recipients JSONB NOT NULL DEFAULT '[]',
    last_run_at TIMESTAMP,
    next_run_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for schedules
CREATE INDEX idx_analytics_schedules_user_id ON analytics_report_schedules(user_id);
CREATE INDEX idx_analytics_schedules_enabled ON analytics_report_schedules(enabled);
CREATE INDEX idx_analytics_schedules_next_run ON analytics_report_schedules(next_run_at);

-- Create analytics_dashboards table
CREATE TABLE IF NOT EXISTS analytics_dashboards (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    layout JSONB NOT NULL DEFAULT '{}',
    is_default BOOLEAN NOT NULL DEFAULT false,
    is_public BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for dashboards
CREATE INDEX idx_analytics_dashboards_user_id ON analytics_dashboards(user_id);
CREATE INDEX idx_analytics_dashboards_is_default ON analytics_dashboards(is_default) WHERE is_default = true;
CREATE INDEX idx_analytics_dashboards_is_public ON analytics_dashboards(is_public) WHERE is_public = true;

-- Create analytics_dashboard_widgets table
CREATE TABLE IF NOT EXISTS analytics_dashboard_widgets (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    dashboard_id UUID NOT NULL REFERENCES analytics_dashboards(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL CHECK (type IN ('chart', 'table', 'counter', 'trend', 'gauge')),
    title VARCHAR(255) NOT NULL,
    config JSONB NOT NULL DEFAULT '{}',
    position INT NOT NULL DEFAULT 0,
    width INT NOT NULL DEFAULT 6 CHECK (width >= 1 AND width <= 12),
    height INT NOT NULL DEFAULT 4 CHECK (height >= 1),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes for widgets
CREATE INDEX idx_analytics_widgets_dashboard_id ON analytics_dashboard_widgets(dashboard_id);
CREATE INDEX idx_analytics_widgets_position ON analytics_dashboard_widgets(position);

-- Create update trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_analytics_reports_updated_at BEFORE UPDATE ON analytics_reports
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_analytics_report_schedules_updated_at BEFORE UPDATE ON analytics_report_schedules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_analytics_dashboards_updated_at BEFORE UPDATE ON analytics_dashboards
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_analytics_dashboard_widgets_updated_at BEFORE UPDATE ON analytics_dashboard_widgets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

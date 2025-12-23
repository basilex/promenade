-- Drop triggers
DROP TRIGGER IF EXISTS update_analytics_dashboard_widgets_updated_at ON analytics_dashboard_widgets;
DROP TRIGGER IF EXISTS update_analytics_dashboards_updated_at ON analytics_dashboards;
DROP TRIGGER IF EXISTS update_analytics_report_schedules_updated_at ON analytics_report_schedules;
DROP TRIGGER IF EXISTS update_analytics_reports_updated_at ON analytics_reports;

-- Drop tables (in reverse order of dependencies)
DROP TABLE IF EXISTS analytics_dashboard_widgets;
DROP TABLE IF EXISTS analytics_dashboards;
DROP TABLE IF EXISTS analytics_report_schedules;
DROP TABLE IF EXISTS analytics_reports;
DROP TABLE IF EXISTS analytics_metric_aggregates;
DROP TABLE IF EXISTS analytics_metrics;

-- Note: Function update_updated_at_column() is shared across modules, don't drop it

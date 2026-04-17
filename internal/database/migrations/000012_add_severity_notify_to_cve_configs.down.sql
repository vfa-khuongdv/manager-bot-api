-- Remove severity notification fields from cve_configs
ALTER TABLE cve_configs DROP COLUMN notify_on_low;
ALTER TABLE cve_configs DROP COLUMN notify_on_medium;
ALTER TABLE cve_configs DROP COLUMN notify_on_high;
ALTER TABLE cve_configs DROP COLUMN notify_on_critical;
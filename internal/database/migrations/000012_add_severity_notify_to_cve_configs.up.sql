-- Add severity notification fields to cve_configs
ALTER TABLE cve_configs ADD COLUMN notify_on_critical BOOLEAN DEFAULT TRUE AFTER notify_on_failure;
ALTER TABLE cve_configs ADD COLUMN notify_on_high BOOLEAN DEFAULT TRUE AFTER notify_on_critical;
ALTER TABLE cve_configs ADD COLUMN notify_on_medium BOOLEAN DEFAULT FALSE AFTER notify_on_high;
ALTER TABLE cve_configs ADD COLUMN notify_on_low BOOLEAN DEFAULT FALSE AFTER notify_on_medium;
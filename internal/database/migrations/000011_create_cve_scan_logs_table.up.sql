-- Create cve_scan_logs table
CREATE TABLE IF NOT EXISTS cve_scan_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    config_id VARCHAR(36) NOT NULL,
    project_id BIGINT UNSIGNED NOT NULL,
    status VARCHAR(20) NOT NULL,
    vuln_found_count INT DEFAULT 0,
    error_message TEXT,
    started_at DATETIME NOT NULL,
    finished_at DATETIME,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME,
    INDEX idx_log_config_id (config_id),
    INDEX idx_log_project_id (project_id)
);
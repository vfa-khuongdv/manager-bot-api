-- Create cve_configs table
CREATE TABLE IF NOT EXISTS cve_configs (
    id VARCHAR(36) PRIMARY KEY,
    project_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    repo_url TEXT,
    languages TEXT NOT NULL,
    cron VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    api_key VARCHAR(255),
    bot_id INT,
    last_scan DATETIME NULL,
    last_status VARCHAR(20) DEFAULT 'no_scan',
    vulnerabilities_found INT DEFAULT 0,
    notify_on_success BOOLEAN DEFAULT FALSE,
    notify_on_failure BOOLEAN DEFAULT TRUE,
    notify_room_id VARCHAR(255),
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME,
    INDEX idx_project_id (project_id),
    INDEX idx_deleted_at (deleted_at)
);
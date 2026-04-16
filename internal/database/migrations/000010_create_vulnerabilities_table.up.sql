-- Create vulnerabilities table
CREATE TABLE IF NOT EXISTS vulnerabilities (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    config_id VARCHAR(36) NOT NULL,
    cve_id VARCHAR(50),
    severity VARCHAR(20) NOT NULL,
    package VARCHAR(255) NOT NULL,
    version VARCHAR(100) NOT NULL,
    summary TEXT,
    description TEXT,
    fixed_version VARCHAR(100),
    score DECIMAL(5,2) DEFAULT NULL,
    reference_url VARCHAR(500),
    scan_log_id BIGINT UNSIGNED,
    created_at DATETIME NOT NULL,
    INDEX idx_config_id (config_id),
    INDEX idx_vuln_scan_log_id (scan_log_id),
    INDEX idx_cve_id (cve_id),
    FOREIGN KEY (config_id) REFERENCES cve_configs(id) ON DELETE CASCADE
);
CREATE TABLE `chatwork_bots` (
    `id`          INT UNSIGNED NOT NULL AUTO_INCREMENT,
    `api_token`   VARCHAR(255) NOT NULL,
    `email`       VARCHAR(255) NULL DEFAULT NULL COMMENT 'Used for Chatwork room invite flow',
    `description` TEXT        COMMENT 'Custom description managed by admin',
    `created_at`  DATETIME NOT NULL,
    `updated_at`  DATETIME NOT NULL,
    `deleted_at`  DATETIME NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    INDEX `idx_chatwork_bots_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

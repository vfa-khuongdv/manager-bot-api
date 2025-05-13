-- reminder_schedules table
CREATE TABLE `reminder_schedules` (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `project_id` bigint UNSIGNED NOT NULL,
  `cron_expression` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `chatwork_room_id` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `chatwork_token` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `message` text COLLATE utf8mb4_unicode_ci,
  `active` boolean DEFAULT TRUE,
  PRIMARY KEY (`id`),
  KEY `idx_reminder_project_id` (`project_id`),
  CONSTRAINT `fk_reminder_project` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
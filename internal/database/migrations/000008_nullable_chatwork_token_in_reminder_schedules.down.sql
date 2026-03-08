ALTER TABLE `reminder_schedules`
  MODIFY COLUMN `chatwork_token` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '';

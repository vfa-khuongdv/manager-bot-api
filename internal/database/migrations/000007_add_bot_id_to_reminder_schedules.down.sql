ALTER TABLE `reminder_schedules`
  DROP INDEX `idx_reminder_schedules_bot_id`,
  DROP COLUMN `bot_id`;

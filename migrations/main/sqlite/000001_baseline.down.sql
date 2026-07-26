-- Baseline down is intentionally destructive and for empty/dev DBs only.
-- Production rollback: restore backup; do not migrate down past baseline.
-- See docs/operations/db-migrations.md § rollback.

DROP TABLE IF EXISTS `user_oauth_bindings`;
DROP TABLE IF EXISTS `user_subscriptions`;
DROP TABLE IF EXISTS `users`;
DROP TABLE IF EXISTS `two_fa_backup_codes`;
DROP TABLE IF EXISTS `two_fas`;
DROP TABLE IF EXISTS `tokens`;
DROP TABLE IF EXISTS `top_ups`;
DROP TABLE IF EXISTS `tasks`;
DROP TABLE IF EXISTS `system_tasks`;
DROP TABLE IF EXISTS `system_task_locks`;
DROP TABLE IF EXISTS `system_instances`;
DROP TABLE IF EXISTS `subscription_pre_consume_records`;
DROP TABLE IF EXISTS `subscription_orders`;
DROP TABLE IF EXISTS `subscription_plans`;
DROP TABLE IF EXISTS `setups`;
DROP TABLE IF EXISTS `redemptions`;
DROP TABLE IF EXISTS `quota_data`;
DROP TABLE IF EXISTS `prefill_groups`;
DROP TABLE IF EXISTS `perf_metrics`;
DROP TABLE IF EXISTS `passkey_credentials`;
DROP TABLE IF EXISTS `options`;
DROP TABLE IF EXISTS `models`;
DROP TABLE IF EXISTS `midjourneys`;
DROP TABLE IF EXISTS `logs`;
DROP TABLE IF EXISTS `custom_oauth_providers`;
DROP TABLE IF EXISTS `checkins`;
DROP TABLE IF EXISTS `channels`;
DROP TABLE IF EXISTS `casbin_rule`;
DROP TABLE IF EXISTS `authz_roles`;
DROP TABLE IF EXISTS `abilities`;
DROP TABLE IF EXISTS `vendors`;

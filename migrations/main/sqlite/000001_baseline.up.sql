-- Phase1 WP-S S3: SQLite-compatible baseline (expand-first).
-- Source: scripts/export-sqlite-schema against current model.AutoMigrate list.
-- Dialect note: generated for SQLite. MySQL/PostgreSQL production deploys may
-- still use GORM AutoMigrate until dialect-specific baselines land (see migrations/README.md).
-- Do not re-run bin/migration_v*.sql; historical data patches live in migrations/archive/.
-- golang-migrate maintains schema_migrations separately.
-- Code-generated draft baseline from AutoMigrate (SQLite).
-- Phase1 WP-S: review before treating as multi-dialect SSOT.
-- golang-migrate will also create schema_migrations; do not include it here.

CREATE TABLE `abilities` (`group` varchar(64),`model` varchar(255),`channel_id` integer,`enabled` numeric,`priority` integer DEFAULT 0,`weight` integer DEFAULT 0,`tag` text,PRIMARY KEY (`group`,`model`,`channel_id`));

CREATE TABLE `authz_roles` (`id` integer,`key` text NOT NULL,`name` text NOT NULL,`description` text,`built_in` numeric,`enabled` numeric,`sort` integer,`created_at` integer,`updated_at` integer,PRIMARY KEY (`id`));

CREATE TABLE `casbin_rule` (`id` integer,`ptype` text,`v0` text,`v1` text,`v2` text,`v3` text,`v4` text,`v5` text,PRIMARY KEY (`id`));

CREATE TABLE `channels` (`id` integer,`type` integer DEFAULT 0,`key` text NOT NULL,`open_ai_organization` text,`test_model` text,`status` integer DEFAULT 1,`name` text,`weight` integer DEFAULT 0,`created_time` integer,`test_time` integer,`response_time` integer,`base_url` text DEFAULT "",`other` text,`balance` real,`balance_updated_time` integer,`models` text,`group` varchar(64) DEFAULT "default",`used_quota` integer DEFAULT 0,`model_mapping` text,`status_code_mapping` varchar(1024) DEFAULT "",`priority` integer DEFAULT 0,`auto_ban` integer DEFAULT 1,`other_info` text,`tag` text,`setting` text,`param_override` text,`header_override` text,`remark` varchar(255),`channel_info` json,`settings` text,PRIMARY KEY (`id`));

CREATE TABLE `checkins` (`id` integer,`user_id` integer NOT NULL,`checkin_date` varchar(10) NOT NULL,`quota_awarded` integer NOT NULL,`created_at` integer,PRIMARY KEY (`id`));

CREATE TABLE `custom_oauth_providers` (`id` integer,`name` varchar(64) NOT NULL,`slug` varchar(64) NOT NULL,`icon` varchar(128) DEFAULT "",`enabled` numeric DEFAULT false,`client_id` varchar(256),`client_secret` varchar(512),`authorization_endpoint` varchar(512),`token_endpoint` varchar(512),`user_info_endpoint` varchar(512),`scopes` varchar(256) DEFAULT "openid profile email",`user_id_field` varchar(128) DEFAULT "sub",`username_field` varchar(128) DEFAULT "preferred_username",`display_name_field` varchar(128) DEFAULT "name",`email_field` varchar(128) DEFAULT "email",`well_known` varchar(512),`auth_style` integer DEFAULT 0,`access_policy` text,`access_denied_message` varchar(512),`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`));

CREATE TABLE `logs` (`id` integer,`user_id` integer,`created_at` integer,`type` integer,`content` text,`username` text DEFAULT "",`token_name` text DEFAULT "",`model_name` text DEFAULT "",`quota` integer DEFAULT 0,`prompt_tokens` integer DEFAULT 0,`completion_tokens` integer DEFAULT 0,`use_time` integer DEFAULT 0,`is_stream` numeric,`channel_id` integer,`channel_name` text,`token_id` integer DEFAULT 0,`group` text,`ip` text DEFAULT "",`request_id` varchar(64) DEFAULT "",`upstream_request_id` varchar(128) DEFAULT "",`trace_id` varchar(128) DEFAULT "",`other` text,PRIMARY KEY (`id`));

CREATE TABLE `midjourneys` (`id` integer,`code` integer,`user_id` integer,`action` varchar(40),`mj_id` text,`prompt` text,`prompt_en` text,`description` text,`state` text,`submit_time` integer,`start_time` integer,`finish_time` integer,`image_url` text,`video_url` text,`video_urls` text,`status` varchar(20),`progress` varchar(30),`fail_reason` text,`channel_id` integer,`quota` integer,`buttons` text,`properties` text,PRIMARY KEY (`id`));

CREATE TABLE `models` (`id` integer,`model_name` text NOT NULL,`description` text,`icon` varchar(128),`tags` varchar(255),`vendor_id` integer,`endpoints` text,`status` integer DEFAULT 1,`sync_official` integer DEFAULT 1,`created_time` integer,`updated_time` integer,`deleted_at` datetime,`name_rule` integer DEFAULT 0,PRIMARY KEY (`id`));

CREATE TABLE `options` (`key` text,`value` text,PRIMARY KEY (`key`));

CREATE TABLE `passkey_credentials` (`id` integer,`user_id` integer NOT NULL,`credential_id` varchar(512) NOT NULL,`public_key` text NOT NULL,`attestation_type` varchar(255),`aa_guid` varchar(512),`sign_count` integer DEFAULT 0,`clone_warning` numeric,`user_present` numeric,`user_verified` numeric,`backup_eligible` numeric,`backup_state` numeric,`transports` text,`attachment` varchar(32),`last_used_at` datetime,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,PRIMARY KEY (`id`));

CREATE TABLE `perf_metrics` (`id` integer,`model_name` text,`group` text,`bucket_ts` integer,`request_count` integer DEFAULT 0,`success_count` integer DEFAULT 0,`total_latency_ms` integer DEFAULT 0,`ttft_sum_ms` integer DEFAULT 0,`ttft_count` integer DEFAULT 0,`output_tokens` integer DEFAULT 0,`generation_ms` integer DEFAULT 0,PRIMARY KEY (`id`));

CREATE TABLE `prefill_groups` (`id` integer,`name` text NOT NULL,`type` text NOT NULL,`items` json,`description` varchar(255),`created_time` integer,`updated_time` integer,`deleted_at` datetime,PRIMARY KEY (`id`));

CREATE TABLE `quota_data` (`id` integer,`user_id` integer,`username` text DEFAULT "",`model_name` text DEFAULT "",`created_at` integer,`use_group` text DEFAULT "",`token_id` integer DEFAULT 0,`channel_id` integer DEFAULT 0,`node_name` text DEFAULT "",`token_used` integer DEFAULT 0,`count` integer DEFAULT 0,`quota` integer DEFAULT 0,PRIMARY KEY (`id`));

CREATE TABLE `redemptions` (`id` integer,`user_id` integer,`key` char(32),`status` integer DEFAULT 1,`name` text,`quota` integer DEFAULT 100,`created_time` integer,`redeemed_time` integer,`used_user_id` integer,`deleted_at` datetime,`expired_time` integer,PRIMARY KEY (`id`));

CREATE TABLE `setups` (`id` integer,`version` varchar(50) NOT NULL,`initialized_at` bigint NOT NULL,PRIMARY KEY (`id`));

CREATE TABLE `subscription_orders` (`id` integer,`user_id` integer,`plan_id` integer,`money` real,`trade_no` varchar(255) UNIQUE,`payment_method` varchar(50),`payment_provider` varchar(50) DEFAULT "",`status` text,`create_time` integer,`complete_time` integer,`provider_payload` text,PRIMARY KEY (`id`));

CREATE TABLE `subscription_plans` (
`id` integer,
`title` varchar(128) NOT NULL,
`subtitle` varchar(255) DEFAULT '',
`price_amount` decimal(10,6) NOT NULL,
`currency` varchar(8) NOT NULL DEFAULT 'USD',
`duration_unit` varchar(16) NOT NULL DEFAULT 'month',
`duration_value` integer NOT NULL DEFAULT 1,
`custom_seconds` bigint NOT NULL DEFAULT 0,
`enabled` numeric DEFAULT 1,
`sort_order` integer DEFAULT 0,
`allow_balance_pay` numeric DEFAULT 1,
`allow_wallet_overflow` numeric DEFAULT 1,
`stripe_price_id` varchar(128) DEFAULT '',
`creem_product_id` varchar(128) DEFAULT '',
`waffo_pancake_product_id` varchar(128) DEFAULT '',
`max_purchase_per_user` integer DEFAULT 0,
`upgrade_group` varchar(64) DEFAULT '',
`downgrade_group` varchar(64) DEFAULT '',
`total_amount` bigint NOT NULL DEFAULT 0,
`quota_reset_period` varchar(16) DEFAULT 'never',
`quota_reset_custom_seconds` bigint DEFAULT 0,
`created_at` bigint,
`updated_at` bigint,
PRIMARY KEY (`id`)
);

CREATE TABLE `subscription_pre_consume_records` (`id` integer,`request_id` varchar(64),`user_id` integer,`user_subscription_id` integer,`pre_consumed` bigint NOT NULL DEFAULT 0,`status` varchar(32),`created_at` integer,`updated_at` integer,PRIMARY KEY (`id`));

CREATE TABLE `system_instances` (`node_name` varchar(128),`info` text,`started_at` integer,`last_seen_at` integer,`created_at` integer,`updated_at` integer,PRIMARY KEY (`node_name`));

CREATE TABLE `system_task_locks` (`type` varchar(64),`task_id` varchar(64),`locked_by` varchar(128),`locked_until` integer,`updated_at` integer,PRIMARY KEY (`type`));

CREATE TABLE `system_tasks` (`id` integer,`task_id` varchar(64),`type` varchar(64),`status` varchar(32),`active_key` varchar(64),`payload` text,`state` text,`result` text,`error` text,`locked_by` varchar(128),`created_at` integer,`updated_at` integer,PRIMARY KEY (`id`));

CREATE TABLE `tasks` (`id` integer,`created_at` integer,`updated_at` integer,`task_id` varchar(191),`platform` varchar(30),`user_id` integer,`group` varchar(50),`channel_id` integer,`quota` integer,`action` varchar(40),`status` varchar(20),`fail_reason` text,`submit_time` integer,`start_time` integer,`finish_time` integer,`progress` varchar(20),`properties` json,`private_data` json,`data` json,PRIMARY KEY (`id`));

CREATE TABLE `tokens` (`id` integer,`user_id` integer,`key` varchar(128),`status` integer DEFAULT 1,`name` text,`created_time` integer,`accessed_time` integer,`expired_time` integer DEFAULT -1,`remain_quota` integer DEFAULT 0,`unlimited_quota` numeric,`model_limits_enabled` numeric,`model_limits` text,`allow_ips` text DEFAULT "",`used_quota` integer DEFAULT 0,`group` text DEFAULT "",`cross_group_retry` numeric,`deleted_at` datetime,PRIMARY KEY (`id`));

CREATE TABLE `top_ups` (`id` integer,`user_id` integer,`amount` integer,`money` real,`trade_no` varchar(255) UNIQUE,`payment_method` varchar(50),`payment_provider` varchar(50) DEFAULT "",`create_time` integer,`complete_time` integer,`status` text,PRIMARY KEY (`id`));

CREATE TABLE `two_fa_backup_codes` (`id` integer,`user_id` integer NOT NULL,`code_hash` varchar(255) NOT NULL,`is_used` numeric,`used_at` datetime,`created_at` datetime,`deleted_at` datetime,PRIMARY KEY (`id`));

CREATE TABLE `two_fas` (`id` integer,`user_id` integer NOT NULL UNIQUE,`secret` varchar(255) NOT NULL,`is_enabled` numeric,`failed_attempts` integer DEFAULT 0,`locked_until` datetime,`last_used_at` datetime,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,PRIMARY KEY (`id`));

CREATE TABLE `user_oauth_bindings` (`id` integer,`user_id` integer NOT NULL,`provider_id` integer NOT NULL,`provider_user_id` varchar(256) NOT NULL,`created_at` datetime,PRIMARY KEY (`id`));

CREATE TABLE `user_subscriptions` (`id` integer,`user_id` integer,`plan_id` integer,`amount_total` bigint NOT NULL DEFAULT 0,`amount_used` bigint NOT NULL DEFAULT 0,`start_time` integer,`end_time` integer,`status` varchar(32),`source` varchar(32) DEFAULT "order",`last_reset_time` bigint DEFAULT 0,`next_reset_time` bigint DEFAULT 0,`upgrade_group` varchar(64) DEFAULT "",`prev_user_group` varchar(64) DEFAULT "",`downgrade_group` varchar(64) DEFAULT "",`allow_wallet_overflow` numeric,`created_at` integer,`updated_at` integer,PRIMARY KEY (`id`));

CREATE TABLE `users` (`id` integer,`username` text UNIQUE,`password` text NOT NULL,`display_name` text,`role` integer DEFAULT 1,`status` integer DEFAULT 1,`email` text,`github_id` text,`discord_id` text,`oidc_id` text,`wechat_id` text,`telegram_id` text,`access_token` char(32),`quota` integer DEFAULT 0,`used_quota` integer DEFAULT 0,`request_count` integer DEFAULT 0,`group` varchar(64) DEFAULT "default",`aff_code` varchar(32),`aff_count` integer DEFAULT 0,`aff_quota` integer DEFAULT 0,`aff_history` integer DEFAULT 0,`inviter_id` integer,`deleted_at` datetime,`linux_do_id` text,`setting` text,`remark` varchar(255),`stripe_customer` varchar(64),`created_at` integer,`last_login_at` integer DEFAULT 0,PRIMARY KEY (`id`));

CREATE TABLE `vendors` (`id` integer,`name` text NOT NULL,`description` text,`icon` varchar(128),`status` integer DEFAULT 1,`created_time` integer,`updated_time` integer,`deleted_at` datetime,PRIMARY KEY (`id`));

CREATE INDEX `idx_abilities_channel_id` ON `abilities`(`channel_id`);

CREATE INDEX `idx_abilities_priority` ON `abilities`(`priority`);

CREATE INDEX `idx_abilities_tag` ON `abilities`(`tag`);

CREATE INDEX `idx_abilities_weight` ON `abilities`(`weight`);

CREATE UNIQUE INDEX `idx_authz_roles_key` ON `authz_roles`(`key`);

CREATE INDEX `idx_casbin_rule` ON `casbin_rule`(`ptype`,`v0`,`v1`,`v2`,`v3`,`v4`,`v5`);

CREATE UNIQUE INDEX `idx_casbin_rule_unique` ON `casbin_rule`(`ptype`,`v0`,`v1`,`v2`,`v3`,`v4`,`v5`);

CREATE INDEX `idx_channels_name` ON `channels`(`name`);

CREATE INDEX `idx_channels_tag` ON `channels`(`tag`);

CREATE INDEX `idx_created_at_id` ON `logs`(`created_at`,`id`);

CREATE INDEX `idx_created_at_type` ON `logs`(`created_at`,`type`);

CREATE UNIQUE INDEX `idx_custom_oauth_providers_slug` ON `custom_oauth_providers`(`slug`);

CREATE INDEX `idx_logs_channel_id` ON `logs`(`channel_id`);

CREATE INDEX `idx_logs_group` ON `logs`(`group`);

CREATE INDEX `idx_logs_ip` ON `logs`(`ip`);

CREATE INDEX `idx_logs_model_name` ON `logs`(`model_name`);

CREATE INDEX `idx_logs_request_id` ON `logs`(`request_id`);

CREATE INDEX `idx_logs_token_id` ON `logs`(`token_id`);

CREATE INDEX `idx_logs_token_name` ON `logs`(`token_name`);

CREATE INDEX `idx_logs_trace_created` ON `logs`(`trace_id`,`created_at`);

CREATE INDEX `idx_logs_trace_id` ON `logs`(`trace_id`);

CREATE INDEX `idx_logs_upstream_request_id` ON `logs`(`upstream_request_id`);

CREATE INDEX `idx_logs_user_created_id` ON `logs`(`user_id`,`created_at`,`id`);

CREATE INDEX `idx_logs_user_id` ON `logs`(`user_id`);

CREATE INDEX `idx_logs_username` ON `logs`(`username`);

CREATE INDEX `idx_midjourneys_action` ON `midjourneys`(`action`);

CREATE INDEX `idx_midjourneys_finish_time` ON `midjourneys`(`finish_time`);

CREATE INDEX `idx_midjourneys_mj_id` ON `midjourneys`(`mj_id`);

CREATE INDEX `idx_midjourneys_progress` ON `midjourneys`(`progress`);

CREATE INDEX `idx_midjourneys_start_time` ON `midjourneys`(`start_time`);

CREATE INDEX `idx_midjourneys_status` ON `midjourneys`(`status`);

CREATE INDEX `idx_midjourneys_submit_time` ON `midjourneys`(`submit_time`);

CREATE INDEX `idx_midjourneys_user_id` ON `midjourneys`(`user_id`);

CREATE INDEX `idx_models_deleted_at` ON `models`(`deleted_at`);

CREATE INDEX `idx_models_vendor_id` ON `models`(`vendor_id`);

CREATE UNIQUE INDEX `idx_passkey_credentials_credential_id` ON `passkey_credentials`(`credential_id`);

CREATE INDEX `idx_passkey_credentials_deleted_at` ON `passkey_credentials`(`deleted_at`);

CREATE UNIQUE INDEX `idx_passkey_credentials_user_id` ON `passkey_credentials`(`user_id`);

CREATE INDEX `idx_perf_bucket_ts` ON `perf_metrics`(`bucket_ts`);

CREATE UNIQUE INDEX `idx_perf_model_group_bucket` ON `perf_metrics`(`model_name`,`group`,`bucket_ts`);

CREATE INDEX `idx_prefill_groups_deleted_at` ON `prefill_groups`(`deleted_at`);

CREATE INDEX `idx_prefill_groups_type` ON `prefill_groups`(`type`);

CREATE INDEX `idx_qdt_created_at` ON `quota_data`(`created_at`);

CREATE INDEX `idx_qdt_model_user_name` ON `quota_data`(`model_name`,`username`);

CREATE INDEX `idx_quota_data_channel_id` ON `quota_data`(`channel_id`);

CREATE INDEX `idx_quota_data_node_name` ON `quota_data`(`node_name`);

CREATE INDEX `idx_quota_data_token_id` ON `quota_data`(`token_id`);

CREATE INDEX `idx_quota_data_use_group` ON `quota_data`(`use_group`);

CREATE INDEX `idx_quota_data_user_id` ON `quota_data`(`user_id`);

CREATE INDEX `idx_redemptions_deleted_at` ON `redemptions`(`deleted_at`);

CREATE UNIQUE INDEX `idx_redemptions_key` ON `redemptions`(`key`);

CREATE INDEX `idx_redemptions_name` ON `redemptions`(`name`);

CREATE INDEX `idx_subscription_orders_plan_id` ON `subscription_orders`(`plan_id`);

CREATE INDEX `idx_subscription_orders_trade_no` ON `subscription_orders`(`trade_no`);

CREATE INDEX `idx_subscription_orders_user_id` ON `subscription_orders`(`user_id`);

CREATE UNIQUE INDEX `idx_subscription_pre_consume_records_request_id` ON `subscription_pre_consume_records`(`request_id`);

CREATE INDEX `idx_subscription_pre_consume_records_status` ON `subscription_pre_consume_records`(`status`);

CREATE INDEX `idx_subscription_pre_consume_records_updated_at` ON `subscription_pre_consume_records`(`updated_at`);

CREATE INDEX `idx_subscription_pre_consume_records_user_id` ON `subscription_pre_consume_records`(`user_id`);

CREATE INDEX `idx_subscription_pre_consume_records_user_subscription_id` ON `subscription_pre_consume_records`(`user_subscription_id`);

CREATE INDEX `idx_system_instances_created_at` ON `system_instances`(`created_at`);

CREATE INDEX `idx_system_instances_last_seen_at` ON `system_instances`(`last_seen_at`);

CREATE INDEX `idx_system_instances_started_at` ON `system_instances`(`started_at`);

CREATE INDEX `idx_system_instances_updated_at` ON `system_instances`(`updated_at`);

CREATE INDEX `idx_system_task_locks_locked_by` ON `system_task_locks`(`locked_by`);

CREATE INDEX `idx_system_task_locks_locked_until` ON `system_task_locks`(`locked_until`);

CREATE INDEX `idx_system_task_locks_task_id` ON `system_task_locks`(`task_id`);

CREATE INDEX `idx_system_task_locks_updated_at` ON `system_task_locks`(`updated_at`);

CREATE UNIQUE INDEX `idx_system_tasks_active_key` ON `system_tasks`(`active_key`);

CREATE INDEX `idx_system_tasks_created_at` ON `system_tasks`(`created_at`);

CREATE INDEX `idx_system_tasks_locked_by` ON `system_tasks`(`locked_by`);

CREATE INDEX `idx_system_tasks_status` ON `system_tasks`(`status`);

CREATE UNIQUE INDEX `idx_system_tasks_task_id` ON `system_tasks`(`task_id`);

CREATE INDEX `idx_system_tasks_type` ON `system_tasks`(`type`);

CREATE INDEX `idx_system_tasks_updated_at` ON `system_tasks`(`updated_at`);

CREATE INDEX `idx_tasks_action` ON `tasks`(`action`);

CREATE INDEX `idx_tasks_channel_id` ON `tasks`(`channel_id`);

CREATE INDEX `idx_tasks_created_at` ON `tasks`(`created_at`);

CREATE INDEX `idx_tasks_finish_time` ON `tasks`(`finish_time`);

CREATE INDEX `idx_tasks_platform` ON `tasks`(`platform`);

CREATE INDEX `idx_tasks_progress` ON `tasks`(`progress`);

CREATE INDEX `idx_tasks_start_time` ON `tasks`(`start_time`);

CREATE INDEX `idx_tasks_status` ON `tasks`(`status`);

CREATE INDEX `idx_tasks_submit_time` ON `tasks`(`submit_time`);

CREATE INDEX `idx_tasks_task_id` ON `tasks`(`task_id`);

CREATE INDEX `idx_tasks_user_id` ON `tasks`(`user_id`);

CREATE INDEX `idx_tokens_deleted_at` ON `tokens`(`deleted_at`);

CREATE UNIQUE INDEX `idx_tokens_key` ON `tokens`(`key`);

CREATE INDEX `idx_tokens_name` ON `tokens`(`name`);

CREATE INDEX `idx_tokens_user_id` ON `tokens`(`user_id`);

CREATE INDEX `idx_top_ups_trade_no` ON `top_ups`(`trade_no`);

CREATE INDEX `idx_top_ups_user_id` ON `top_ups`(`user_id`);

CREATE INDEX `idx_two_fa_backup_codes_deleted_at` ON `two_fa_backup_codes`(`deleted_at`);

CREATE INDEX `idx_two_fa_backup_codes_user_id` ON `two_fa_backup_codes`(`user_id`);

CREATE INDEX `idx_two_fas_deleted_at` ON `two_fas`(`deleted_at`);

CREATE INDEX `idx_two_fas_user_id` ON `two_fas`(`user_id`);

CREATE UNIQUE INDEX `idx_user_checkin_date` ON `checkins`(`user_id`,`checkin_date`);

CREATE INDEX `idx_user_id_id` ON `logs`(`user_id`,`id`);

CREATE INDEX `idx_user_sub_active` ON `user_subscriptions`(`user_id`,`status`,`end_time`);

CREATE INDEX `idx_user_subscriptions_end_time` ON `user_subscriptions`(`end_time`);

CREATE INDEX `idx_user_subscriptions_next_reset_time` ON `user_subscriptions`(`next_reset_time`);

CREATE INDEX `idx_user_subscriptions_plan_id` ON `user_subscriptions`(`plan_id`);

CREATE INDEX `idx_user_subscriptions_status` ON `user_subscriptions`(`status`);

CREATE INDEX `idx_user_subscriptions_user_id` ON `user_subscriptions`(`user_id`);

CREATE UNIQUE INDEX `idx_users_access_token` ON `users`(`access_token`);

CREATE UNIQUE INDEX `idx_users_aff_code` ON `users`(`aff_code`);

CREATE INDEX `idx_users_deleted_at` ON `users`(`deleted_at`);

CREATE INDEX `idx_users_discord_id` ON `users`(`discord_id`);

CREATE INDEX `idx_users_display_name` ON `users`(`display_name`);

CREATE INDEX `idx_users_email` ON `users`(`email`);

CREATE INDEX `idx_users_git_hub_id` ON `users`(`github_id`);

CREATE INDEX `idx_users_inviter_id` ON `users`(`inviter_id`);

CREATE INDEX `idx_users_linux_do_id` ON `users`(`linux_do_id`);

CREATE INDEX `idx_users_oidc_id` ON `users`(`oidc_id`);

CREATE INDEX `idx_users_stripe_customer` ON `users`(`stripe_customer`);

CREATE INDEX `idx_users_telegram_id` ON `users`(`telegram_id`);

CREATE INDEX `idx_users_username` ON `users`(`username`);

CREATE INDEX `idx_users_we_chat_id` ON `users`(`wechat_id`);

CREATE INDEX `idx_vendors_deleted_at` ON `vendors`(`deleted_at`);

CREATE INDEX `index_username_model_name` ON `logs`(`model_name`,`username`);

CREATE UNIQUE INDEX `uk_model_name_delete_at` ON `models`(`model_name`,`deleted_at`);

CREATE UNIQUE INDEX `uk_prefill_name` ON `prefill_groups`(`name`) WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX `uk_vendor_name_delete_at` ON `vendors`(`name`,`deleted_at`);

CREATE UNIQUE INDEX `ux_provider_userid` ON `user_oauth_bindings`(`provider_id`,`provider_user_id`);

CREATE UNIQUE INDEX `ux_user_provider` ON `user_oauth_bindings`(`user_id`,`provider_id`);


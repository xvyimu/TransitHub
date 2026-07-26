-- MySQL baseline: empty-DB schema matching GORM AutoMigrate output.
-- Generated from SQLite baseline by dialect translation.
-- Columns: id → BIGINT AUTO_INCREMENT, numeric → tinyint(1),
--          real → double, json → json, datetime → datetime.
-- Indexed text columns: varchar(255) for unique indexes, text with
-- prefix length (191) for non-unique text indexes.
-- Primary keys with composite key: no AUTO_INCREMENT.
-- No DEFAULT '' on text columns (MySQL 5.7 compat).
-- golang-migrate creates schema_migrations separately; do not include it here.

CREATE TABLE `abilities` (
  `group` varchar(64) NOT NULL,
  `model` varchar(255) NOT NULL,
  `channel_id` bigint NOT NULL,
  `enabled` tinyint(1) DEFAULT NULL,
  `priority` int DEFAULT 0,
  `weight` int DEFAULT 0,
  `tag` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`group`,`model`,`channel_id`),
  KEY `idx_abilities_channel_id` (`channel_id`),
  KEY `idx_abilities_priority` (`priority`),
  KEY `idx_abilities_tag` (`tag`),
  KEY `idx_abilities_weight` (`weight`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `authz_roles` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `key` varchar(255) NOT NULL,
  `name` text NOT NULL,
  `description` text,
  `built_in` tinyint(1) DEFAULT NULL,
  `enabled` tinyint(1) DEFAULT NULL,
  `sort` int DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  `updated_at` bigint DEFAULT NULL,
  UNIQUE KEY `idx_authz_roles_key` (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `casbin_rule` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `ptype` text,
  `v0` text,
  `v1` text,
  `v2` text,
  `v3` text,
  `v4` text,
  `v5` text,
  KEY `idx_casbin_rule` (`ptype`(191),`v0`(191),`v1`(191),`v2`(191),`v3`(191),`v4`(191),`v5`(191)),
  UNIQUE KEY `idx_casbin_rule_unique` (`ptype`(191),`v0`(191),`v1`(191),`v2`(191),`v3`(191),`v4`(191),`v5`(191))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `channels` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `type` int DEFAULT 0,
  `key` text NOT NULL,
  `open_ai_organization` text,
  `test_model` text,
  `status` int DEFAULT 1,
  `name` varchar(255) DEFAULT NULL,
  `weight` int DEFAULT 0,
  `created_time` bigint DEFAULT NULL,
  `test_time` bigint DEFAULT NULL,
  `response_time` bigint DEFAULT NULL,
  `base_url` text DEFAULT NULL,
  `other` text,
  `balance` double DEFAULT NULL,
  `balance_updated_time` bigint DEFAULT NULL,
  `models` text,
  `group` varchar(64) DEFAULT 'default',
  `used_quota` int DEFAULT 0,
  `model_mapping` text,
  `status_code_mapping` varchar(1024) DEFAULT '',
  `priority` int DEFAULT 0,
  `auto_ban` int DEFAULT 1,
  `other_info` text,
  `tag` varchar(255) DEFAULT NULL,
  `setting` text,
  `param_override` text,
  `header_override` text,
  `remark` varchar(255) DEFAULT NULL,
  `channel_info` json DEFAULT NULL,
  `settings` text,
  KEY `idx_channels_name` (`name`),
  KEY `idx_channels_tag` (`tag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `checkins` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `user_id` int NOT NULL,
  `checkin_date` varchar(10) NOT NULL,
  `quota_awarded` int NOT NULL,
  `created_at` bigint DEFAULT NULL,
  UNIQUE KEY `idx_user_checkin_date` (`user_id`,`checkin_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `custom_oauth_providers` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `name` varchar(64) NOT NULL,
  `slug` varchar(64) NOT NULL,
  `icon` varchar(128) DEFAULT '',
  `enabled` tinyint(1) DEFAULT false,
  `client_id` varchar(256) DEFAULT NULL,
  `client_secret` varchar(512) DEFAULT NULL,
  `authorization_endpoint` varchar(512) DEFAULT NULL,
  `token_endpoint` varchar(512) DEFAULT NULL,
  `user_info_endpoint` varchar(512) DEFAULT NULL,
  `scopes` varchar(256) DEFAULT 'openid profile email',
  `user_id_field` varchar(128) DEFAULT 'sub',
  `username_field` varchar(128) DEFAULT 'preferred_username',
  `display_name_field` varchar(128) DEFAULT 'name',
  `email_field` varchar(128) DEFAULT 'email',
  `well_known` varchar(512) DEFAULT NULL,
  `auth_style` int DEFAULT 0,
  `access_policy` text,
  `access_denied_message` varchar(512) DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  UNIQUE KEY `idx_custom_oauth_providers_slug` (`slug`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `logs` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `user_id` int DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  `type` int DEFAULT NULL,
  `content` text,
  `username` varchar(255) DEFAULT '',
  `token_name` varchar(255) DEFAULT '',
  `model_name` varchar(255) DEFAULT '',
  `quota` int DEFAULT 0,
  `prompt_tokens` int DEFAULT 0,
  `completion_tokens` int DEFAULT 0,
  `use_time` int DEFAULT 0,
  `is_stream` tinyint(1) DEFAULT NULL,
  `channel_id` int DEFAULT NULL,
  `channel_name` text,
  `token_id` int DEFAULT 0,
  `group` varchar(255) DEFAULT NULL,
  `ip` varchar(255) DEFAULT '',
  `request_id` varchar(64) DEFAULT '',
  `upstream_request_id` varchar(128) DEFAULT '',
  `trace_id` varchar(128) DEFAULT '',
  `other` text,
  KEY `idx_created_at_id` (`created_at`,`id`),
  KEY `idx_created_at_type` (`created_at`,`type`),
  KEY `idx_logs_channel_id` (`channel_id`),
  KEY `idx_logs_group` (`group`),
  KEY `idx_logs_ip` (`ip`),
  KEY `idx_logs_model_name` (`model_name`),
  KEY `idx_logs_request_id` (`request_id`),
  KEY `idx_logs_token_id` (`token_id`),
  KEY `idx_logs_token_name` (`token_name`),
  KEY `idx_logs_trace_created` (`trace_id`,`created_at`),
  KEY `idx_logs_trace_id` (`trace_id`),
  KEY `idx_logs_upstream_request_id` (`upstream_request_id`),
  KEY `idx_logs_user_created_id` (`user_id`,`created_at`,`id`),
  KEY `idx_logs_user_id` (`user_id`),
  KEY `idx_logs_username` (`username`),
  KEY `idx_user_id_id` (`user_id`,`id`),
  KEY `index_username_model_name` (`model_name`,`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `midjourneys` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `code` int DEFAULT NULL,
  `user_id` int DEFAULT NULL,
  `action` varchar(40) DEFAULT NULL,
  `mj_id` text,
  `prompt` text,
  `prompt_en` text,
  `description` text,
  `state` text,
  `submit_time` bigint DEFAULT NULL,
  `start_time` bigint DEFAULT NULL,
  `finish_time` bigint DEFAULT NULL,
  `image_url` text,
  `video_url` text,
  `video_urls` text,
  `status` varchar(20) DEFAULT NULL,
  `progress` varchar(30) DEFAULT NULL,
  `fail_reason` text,
  `channel_id` int DEFAULT NULL,
  `quota` int DEFAULT NULL,
  `buttons` text,
  `properties` text,
  KEY `idx_midjourneys_action` (`action`),
  KEY `idx_midjourneys_finish_time` (`finish_time`),
  KEY `idx_midjourneys_mj_id` (`mj_id`(191)),
  KEY `idx_midjourneys_progress` (`progress`),
  KEY `idx_midjourneys_start_time` (`start_time`),
  KEY `idx_midjourneys_status` (`status`),
  KEY `idx_midjourneys_submit_time` (`submit_time`),
  KEY `idx_midjourneys_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `models` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `model_name` varchar(255) NOT NULL,
  `description` text,
  `icon` varchar(128) DEFAULT NULL,
  `tags` varchar(255) DEFAULT NULL,
  `vendor_id` int DEFAULT NULL,
  `endpoints` text,
  `status` int DEFAULT 1,
  `sync_official` int DEFAULT 1,
  `created_time` bigint DEFAULT NULL,
  `updated_time` bigint DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  `name_rule` int DEFAULT 0,
  KEY `idx_models_deleted_at` (`deleted_at`),
  KEY `idx_models_vendor_id` (`vendor_id`),
  UNIQUE KEY `uk_model_name_delete_at` (`model_name`,`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `options` (
  `key` text NOT NULL,
  `value` text,
  PRIMARY KEY (`key`(191))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `passkey_credentials` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `user_id` int NOT NULL,
  `credential_id` varchar(512) NOT NULL,
  `public_key` text NOT NULL,
  `attestation_type` varchar(255) DEFAULT NULL,
  `aa_guid` varchar(512) DEFAULT NULL,
  `sign_count` int DEFAULT 0,
  `clone_warning` tinyint(1) DEFAULT NULL,
  `user_present` tinyint(1) DEFAULT NULL,
  `user_verified` tinyint(1) DEFAULT NULL,
  `backup_eligible` tinyint(1) DEFAULT NULL,
  `backup_state` tinyint(1) DEFAULT NULL,
  `transports` text,
  `attachment` varchar(32) DEFAULT NULL,
  `last_used_at` datetime DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  UNIQUE KEY `idx_passkey_credentials_credential_id` (`credential_id`),
  KEY `idx_passkey_credentials_deleted_at` (`deleted_at`),
  UNIQUE KEY `idx_passkey_credentials_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `perf_metrics` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `model_name` text,
  `group` text,
  `bucket_ts` bigint DEFAULT NULL,
  `request_count` int DEFAULT 0,
  `success_count` int DEFAULT 0,
  `total_latency_ms` int DEFAULT 0,
  `ttft_sum_ms` int DEFAULT 0,
  `ttft_count` int DEFAULT 0,
  `output_tokens` int DEFAULT 0,
  `generation_ms` int DEFAULT 0,
  KEY `idx_perf_bucket_ts` (`bucket_ts`),
  UNIQUE KEY `idx_perf_model_group_bucket` (`model_name`(191),`group`(191),`bucket_ts`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `prefill_groups` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `name` varchar(255) NOT NULL,
  `type` text NOT NULL,
  `items` json DEFAULT NULL,
  `description` varchar(255) DEFAULT NULL,
  `created_time` bigint DEFAULT NULL,
  `updated_time` bigint DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  KEY `idx_prefill_groups_deleted_at` (`deleted_at`),
  KEY `idx_prefill_groups_type` (`type`(191)),
  UNIQUE KEY `uk_prefill_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `quota_data` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `user_id` int DEFAULT NULL,
  `username` varchar(255) DEFAULT '',
  `model_name` varchar(255) DEFAULT '',
  `created_at` bigint DEFAULT NULL,
  `use_group` varchar(255) DEFAULT '',
  `token_id` int DEFAULT 0,
  `channel_id` int DEFAULT 0,
  `node_name` varchar(255) DEFAULT '',
  `token_used` int DEFAULT 0,
  `count` int DEFAULT 0,
  `quota` int DEFAULT 0,
  KEY `idx_qdt_created_at` (`created_at`),
  KEY `idx_qdt_model_user_name` (`model_name`,`username`),
  KEY `idx_quota_data_channel_id` (`channel_id`),
  KEY `idx_quota_data_node_name` (`node_name`),
  KEY `idx_quota_data_token_id` (`token_id`),
  KEY `idx_quota_data_use_group` (`use_group`),
  KEY `idx_quota_data_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `redemptions` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `user_id` int DEFAULT NULL,
  `key` char(32) DEFAULT NULL,
  `status` int DEFAULT 1,
  `name` varchar(255) DEFAULT NULL,
  `quota` int DEFAULT 100,
  `created_time` bigint DEFAULT NULL,
  `redeemed_time` bigint DEFAULT NULL,
  `used_user_id` int DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  `expired_time` bigint DEFAULT NULL,
  UNIQUE KEY `idx_redemptions_key` (`key`),
  KEY `idx_redemptions_deleted_at` (`deleted_at`),
  KEY `idx_redemptions_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `setups` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `version` varchar(50) NOT NULL,
  `initialized_at` bigint NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `subscription_orders` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `user_id` int DEFAULT NULL,
  `plan_id` int DEFAULT NULL,
  `money` double DEFAULT NULL,
  `trade_no` varchar(255) DEFAULT NULL,
  `payment_method` varchar(50) DEFAULT NULL,
  `payment_provider` varchar(50) DEFAULT '',
  `status` text,
  `create_time` bigint DEFAULT NULL,
  `complete_time` bigint DEFAULT NULL,
  `provider_payload` text,
  KEY `idx_subscription_orders_plan_id` (`plan_id`),
  UNIQUE KEY `idx_subscription_orders_trade_no` (`trade_no`),
  KEY `idx_subscription_orders_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `subscription_plans` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `title` varchar(128) NOT NULL,
  `subtitle` varchar(255) DEFAULT '',
  `price_amount` decimal(10,6) NOT NULL,
  `currency` varchar(8) NOT NULL DEFAULT 'USD',
  `duration_unit` varchar(16) NOT NULL DEFAULT 'month',
  `duration_value` int NOT NULL DEFAULT 1,
  `custom_seconds` bigint NOT NULL DEFAULT 0,
  `enabled` tinyint(1) DEFAULT 1,
  `sort_order` int DEFAULT 0,
  `allow_balance_pay` tinyint(1) DEFAULT 1,
  `allow_wallet_overflow` tinyint(1) DEFAULT 1,
  `stripe_price_id` varchar(128) DEFAULT '',
  `creem_product_id` varchar(128) DEFAULT '',
  `waffo_pancake_product_id` varchar(128) DEFAULT '',
  `max_purchase_per_user` int DEFAULT 0,
  `upgrade_group` varchar(64) DEFAULT '',
  `downgrade_group` varchar(64) DEFAULT '',
  `total_amount` bigint NOT NULL DEFAULT 0,
  `quota_reset_period` varchar(16) DEFAULT 'never',
  `quota_reset_custom_seconds` bigint DEFAULT 0,
  `created_at` bigint DEFAULT NULL,
  `updated_at` bigint DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `subscription_pre_consume_records` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `request_id` varchar(64) DEFAULT NULL,
  `user_id` int DEFAULT NULL,
  `user_subscription_id` int DEFAULT NULL,
  `pre_consumed` bigint NOT NULL DEFAULT 0,
  `status` varchar(32) DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  `updated_at` bigint DEFAULT NULL,
  UNIQUE KEY `idx_subscription_pre_consume_records_request_id` (`request_id`),
  KEY `idx_subscription_pre_consume_records_status` (`status`),
  KEY `idx_subscription_pre_consume_records_updated_at` (`updated_at`),
  KEY `idx_subscription_pre_consume_records_user_id` (`user_id`),
  KEY `idx_subscription_pre_consume_records_user_subscription_id` (`user_subscription_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `system_instances` (
  `node_name` varchar(128) NOT NULL,
  `info` text,
  `started_at` bigint DEFAULT NULL,
  `last_seen_at` bigint DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  `updated_at` bigint DEFAULT NULL,
  PRIMARY KEY (`node_name`),
  KEY `idx_system_instances_created_at` (`created_at`),
  KEY `idx_system_instances_last_seen_at` (`last_seen_at`),
  KEY `idx_system_instances_started_at` (`started_at`),
  KEY `idx_system_instances_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `system_task_locks` (
  `type` varchar(64) NOT NULL,
  `task_id` varchar(64) DEFAULT NULL,
  `locked_by` varchar(128) DEFAULT NULL,
  `locked_until` bigint DEFAULT NULL,
  `updated_at` bigint DEFAULT NULL,
  PRIMARY KEY (`type`),
  KEY `idx_system_task_locks_locked_by` (`locked_by`),
  KEY `idx_system_task_locks_locked_until` (`locked_until`),
  KEY `idx_system_task_locks_task_id` (`task_id`),
  KEY `idx_system_task_locks_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `system_tasks` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `task_id` varchar(64) DEFAULT NULL,
  `type` varchar(64) DEFAULT NULL,
  `status` varchar(32) DEFAULT NULL,
  `active_key` varchar(64) DEFAULT NULL,
  `payload` text,
  `state` text,
  `result` text,
  `error` text,
  `locked_by` varchar(128) DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  `updated_at` bigint DEFAULT NULL,
  UNIQUE KEY `idx_system_tasks_active_key` (`active_key`),
  KEY `idx_system_tasks_created_at` (`created_at`),
  KEY `idx_system_tasks_locked_by` (`locked_by`),
  KEY `idx_system_tasks_status` (`status`),
  UNIQUE KEY `idx_system_tasks_task_id` (`task_id`),
  KEY `idx_system_tasks_type` (`type`),
  KEY `idx_system_tasks_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `tasks` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `created_at` bigint DEFAULT NULL,
  `updated_at` bigint DEFAULT NULL,
  `task_id` varchar(191) DEFAULT NULL,
  `platform` varchar(30) DEFAULT NULL,
  `user_id` int DEFAULT NULL,
  `group` varchar(50) DEFAULT NULL,
  `channel_id` int DEFAULT NULL,
  `quota` int DEFAULT NULL,
  `action` varchar(40) DEFAULT NULL,
  `status` varchar(20) DEFAULT NULL,
  `fail_reason` text,
  `submit_time` bigint DEFAULT NULL,
  `start_time` bigint DEFAULT NULL,
  `finish_time` bigint DEFAULT NULL,
  `progress` varchar(20) DEFAULT NULL,
  `properties` json DEFAULT NULL,
  `private_data` json DEFAULT NULL,
  `data` json DEFAULT NULL,
  KEY `idx_tasks_action` (`action`),
  KEY `idx_tasks_channel_id` (`channel_id`),
  KEY `idx_tasks_created_at` (`created_at`),
  KEY `idx_tasks_finish_time` (`finish_time`),
  KEY `idx_tasks_platform` (`platform`),
  KEY `idx_tasks_progress` (`progress`),
  KEY `idx_tasks_start_time` (`start_time`),
  KEY `idx_tasks_status` (`status`),
  KEY `idx_tasks_submit_time` (`submit_time`),
  KEY `idx_tasks_task_id` (`task_id`),
  KEY `idx_tasks_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `tokens` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `user_id` int DEFAULT NULL,
  `key` varchar(128) DEFAULT NULL,
  `status` int DEFAULT 1,
  `name` varchar(255) DEFAULT NULL,
  `created_time` bigint DEFAULT NULL,
  `accessed_time` bigint DEFAULT NULL,
  `expired_time` bigint DEFAULT -1,
  `remain_quota` int DEFAULT 0,
  `unlimited_quota` tinyint(1) DEFAULT NULL,
  `model_limits_enabled` tinyint(1) DEFAULT NULL,
  `model_limits` text,
  `allow_ips` varchar(255) DEFAULT '',
  `used_quota` int DEFAULT 0,
  `group` varchar(255) DEFAULT '',
  `cross_group_retry` tinyint(1) DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  KEY `idx_tokens_deleted_at` (`deleted_at`),
  UNIQUE KEY `idx_tokens_key` (`key`),
  KEY `idx_tokens_name` (`name`),
  KEY `idx_tokens_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `top_ups` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `user_id` int DEFAULT NULL,
  `amount` int DEFAULT NULL,
  `money` double DEFAULT NULL,
  `trade_no` varchar(255) DEFAULT NULL,
  `payment_method` varchar(50) DEFAULT NULL,
  `payment_provider` varchar(50) DEFAULT '',
  `create_time` bigint DEFAULT NULL,
  `complete_time` bigint DEFAULT NULL,
  `status` text,
  KEY `idx_top_ups_trade_no` (`trade_no`),
  KEY `idx_top_ups_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `two_fa_backup_codes` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `user_id` int NOT NULL,
  `code_hash` varchar(255) NOT NULL,
  `is_used` tinyint(1) DEFAULT NULL,
  `used_at` datetime DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  KEY `idx_two_fa_backup_codes_deleted_at` (`deleted_at`),
  KEY `idx_two_fa_backup_codes_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `two_fas` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `user_id` int NOT NULL,
  `secret` varchar(255) NOT NULL,
  `is_enabled` tinyint(1) DEFAULT NULL,
  `failed_attempts` int DEFAULT 0,
  `locked_until` datetime DEFAULT NULL,
  `last_used_at` datetime DEFAULT NULL,
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  UNIQUE KEY `idx_two_fas_user_id` (`user_id`),
  KEY `idx_two_fas_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `user_oauth_bindings` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `user_id` int NOT NULL,
  `provider_id` int NOT NULL,
  `provider_user_id` varchar(256) NOT NULL,
  `created_at` datetime DEFAULT NULL,
  UNIQUE KEY `ux_provider_userid` (`provider_id`,`provider_user_id`),
  UNIQUE KEY `ux_user_provider` (`user_id`,`provider_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `user_subscriptions` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `user_id` int DEFAULT NULL,
  `plan_id` int DEFAULT NULL,
  `amount_total` bigint NOT NULL DEFAULT 0,
  `amount_used` bigint NOT NULL DEFAULT 0,
  `start_time` bigint DEFAULT NULL,
  `end_time` bigint DEFAULT NULL,
  `status` varchar(32) DEFAULT NULL,
  `source` varchar(32) DEFAULT 'order',
  `last_reset_time` bigint DEFAULT 0,
  `next_reset_time` bigint DEFAULT 0,
  `upgrade_group` varchar(64) DEFAULT '',
  `prev_user_group` varchar(64) DEFAULT '',
  `downgrade_group` varchar(64) DEFAULT '',
  `allow_wallet_overflow` tinyint(1) DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  `updated_at` bigint DEFAULT NULL,
  KEY `idx_user_sub_active` (`user_id`,`status`,`end_time`),
  KEY `idx_user_subscriptions_end_time` (`end_time`),
  KEY `idx_user_subscriptions_next_reset_time` (`next_reset_time`),
  KEY `idx_user_subscriptions_plan_id` (`plan_id`),
  KEY `idx_user_subscriptions_status` (`status`),
  KEY `idx_user_subscriptions_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `users` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `username` varchar(255) NOT NULL,
  `password` text NOT NULL,
  `display_name` varchar(255) DEFAULT NULL,
  `role` int DEFAULT 1,
  `status` int DEFAULT 1,
  `email` varchar(255) DEFAULT NULL,
  `github_id` text,
  `discord_id` text,
  `oidc_id` text,
  `wechat_id` text,
  `telegram_id` text,
  `access_token` char(32) DEFAULT NULL,
  `quota` int DEFAULT 0,
  `used_quota` int DEFAULT 0,
  `request_count` int DEFAULT 0,
  `group` varchar(64) DEFAULT 'default',
  `aff_code` varchar(32) DEFAULT NULL,
  `aff_count` int DEFAULT 0,
  `aff_quota` int DEFAULT 0,
  `aff_history` int DEFAULT 0,
  `inviter_id` int DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  `linux_do_id` text,
  `setting` text,
  `remark` varchar(255) DEFAULT NULL,
  `stripe_customer` varchar(64) DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  `last_login_at` bigint DEFAULT 0,
  UNIQUE KEY `idx_users_username` (`username`),
  UNIQUE KEY `idx_users_access_token` (`access_token`),
  UNIQUE KEY `idx_users_aff_code` (`aff_code`),
  KEY `idx_users_deleted_at` (`deleted_at`),
  KEY `idx_users_discord_id` (`discord_id`),
  KEY `idx_users_display_name` (`display_name`),
  KEY `idx_users_email` (`email`),
  KEY `idx_users_git_hub_id` (`github_id`),
  KEY `idx_users_inviter_id` (`inviter_id`),
  KEY `idx_users_linux_do_id` (`linux_do_id`),
  KEY `idx_users_oidc_id` (`oidc_id`),
  KEY `idx_users_stripe_customer` (`stripe_customer`),
  KEY `idx_users_telegram_id` (`telegram_id`),
  KEY `idx_users_we_chat_id` (`wechat_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `vendors` (
  `id` bigint AUTO_INCREMENT PRIMARY KEY,
  `name` varchar(255) NOT NULL,
  `description` text,
  `icon` varchar(128) DEFAULT NULL,
  `status` int DEFAULT 1,
  `created_time` bigint DEFAULT NULL,
  `updated_time` bigint DEFAULT NULL,
  `deleted_at` datetime DEFAULT NULL,
  KEY `idx_vendors_deleted_at` (`deleted_at`),
  UNIQUE KEY `uk_vendor_name_delete_at` (`name`,`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
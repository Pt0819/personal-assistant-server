-- 微信密钥配置种子数据
-- 用法: 填入真实值后执行此 SQL，然后清空 config.yaml 中对应字段

INSERT INTO system_configs (`key`, `value`, `created_at`, `updated_at`) VALUES
('wechat.app_id',                   '', NOW(), NOW()),
('wechat.app_secret',               '', NOW(), NOW()),
('wechat.open_platform_app_id',     '', NOW(), NOW()),
('wechat.open_platform_app_secret', '', NOW(), NOW())
ON DUPLICATE KEY UPDATE `value` = VALUES(`value`), `updated_at` = NOW();

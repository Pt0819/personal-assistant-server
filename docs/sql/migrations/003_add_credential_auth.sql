-- ============================================
-- 迁移 003: 邮箱/手机号注册登录支持
-- 日期: 2026-06-23
-- ============================================

-- 1. users 表改造
ALTER TABLE users
  MODIFY openid VARCHAR(64) NULL COMMENT '微信OpenID（注册用户为空）',
  ADD COLUMN username VARCHAR(12) NOT NULL DEFAULT '' COMMENT '随机生成唯一用户名',
  ADD COLUMN email VARCHAR(128) NULL COMMENT '邮箱',
  ADD COLUMN password_hash VARCHAR(256) NULL COMMENT 'bcrypt密码哈希',
  ADD COLUMN auth_method ENUM('wechat','email','phone') NOT NULL DEFAULT 'wechat' COMMENT '注册方式',
  ADD UNIQUE INDEX idx_username (username),
  ADD UNIQUE INDEX idx_email (email),
  ADD INDEX idx_auth_method (auth_method);

-- 2. 补齐可能缺失的字段
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS default_reminder_minutes INT NOT NULL DEFAULT 30,
  ADD COLUMN IF NOT EXISTS week_start_day INT NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS onboarding_completed TINYINT NOT NULL DEFAULT 0;

-- 3. 为存量微信用户生成用户名
UPDATE users SET username = CONCAT('小助手_', LOWER(SUBSTRING(MD5(CONCAT(id, RAND())), 1, 6)))
WHERE username = '' OR username IS NULL;

-- 4. 邮箱验证码表
CREATE TABLE IF NOT EXISTS email_verifications (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  email      VARCHAR(128) NOT NULL,
  code       VARCHAR(6) NOT NULL,
  purpose    ENUM('register') NOT NULL DEFAULT 'register',
  expires_at DATETIME NOT NULL,
  verified   TINYINT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_email_code (email, code),
  INDEX idx_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='邮箱验证码表';

-- 5. 短信验证码表
CREATE TABLE IF NOT EXISTS sms_verifications (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  phone      VARCHAR(20) NOT NULL,
  code       VARCHAR(6) NOT NULL,
  purpose    ENUM('register') NOT NULL DEFAULT 'register',
  expires_at DATETIME NOT NULL,
  verified   TINYINT NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_phone_code (phone, code),
  INDEX idx_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='短信验证码表';

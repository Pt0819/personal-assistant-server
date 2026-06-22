package initialize

import (
	"personal-assistant-server/global"
	"personal-assistant-server/model"

	"go.uber.org/zap"
)

// ConfigFromDBFallback 从数据库加载配置作为兜底。
// 若 config 文件中对应的值为空，则尝试从 system_configs 表读取。
func ConfigFromDBFallback() {
	if global.GVA_DB == nil {
		return
	}

	// 确保 system_configs 表存在（即使全局 auto-migrate 关闭）
	if err := global.GVA_DB.AutoMigrate(&model.SystemConfig{}); err != nil {
		global.GVA_LOG.Warn("创建system_configs表失败: " + err.Error())
		return
	}

	fallbacks := map[string]*string{
		"wechat.app_id":                   &global.GVA_CONFIG.Wechat.AppID,
		"wechat.app_secret":               &global.GVA_CONFIG.Wechat.AppSecret,
		"wechat.open_platform_app_id":     &global.GVA_CONFIG.Wechat.OpenPlatformAppID,
		"wechat.open_platform_app_secret": &global.GVA_CONFIG.Wechat.OpenPlatformAppSecret,
	}

	loaded := 0
	for key, target := range fallbacks {
		if *target != "" {
			continue // 配置文件已有值，不覆盖
		}

		var cfg model.SystemConfig
		if err := global.GVA_DB.Where("`key` = ?", key).First(&cfg).Error; err != nil {
			continue // 数据库中也没有，保持为空
		}
		*target = cfg.Value
		loaded++
	}

	if loaded > 0 {
		global.GVA_LOG.Info("已从数据库加载兜底配置", zap.Int("count", loaded))
	}
}

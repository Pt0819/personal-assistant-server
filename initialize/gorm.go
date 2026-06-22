package initialize

import (
	"os"

	"personal-assistant-server/global"
	"personal-assistant-server/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Gorm() *gorm.DB {
	return GormMysql()
}

func RegisterTables() {
	if global.GVA_CONFIG.System.DisableAutoMigrate {
		global.GVA_LOG.Info("auto-migrate is disabled, skipping table registration")
		return
	}

	db := global.GVA_DB
	err := db.AutoMigrate(
		&model.User{},
		&model.Schedule{},
		&model.Conversation{},
		&model.Message{},
		&model.PushSubscription{},
		&model.JwtBlacklist{},
		&model.UserSession{},
		&model.SystemConfig{},
	)
	if err != nil {
		global.GVA_LOG.Error("register table failed", zap.Error(err))
		os.Exit(0)
	}
	global.GVA_LOG.Info("register table success")
}

package global

import (
	"personal-assistant-server/config"
	"personal-assistant-server/utils/storage"
	"personal-assistant-server/utils/timer"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	GVA_DB      *gorm.DB
	GVA_REDIS   redis.UniversalClient
	GVA_CONFIG  config.Server
	GVA_VP      *viper.Viper
	GVA_LOG     *zap.Logger
	GVA_Timer   timer.Timer = timer.NewTimerTask()
	GVA_STORAGE storage.FileStorage
)

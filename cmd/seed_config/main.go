package main

import (
	"fmt"
	"personal-assistant-server/core"
	"personal-assistant-server/global"
	"personal-assistant-server/initialize"
	"personal-assistant-server/model"
)

func main() {
	global.GVA_VP = core.Viper()
	global.GVA_LOG = core.Zap()
	global.GVA_DB = initialize.Gorm()
	if global.GVA_DB == nil {
		fmt.Println("DB 连接失败")
		return
	}

	// 确保表存在
	global.GVA_DB.AutoMigrate(&model.SystemConfig{})

	entries := []model.SystemConfig{
		{Key: "wechat.app_id", Value: "wxafa4d4ecf19d8898"},
		{Key: "wechat.app_secret", Value: "665accc347d50cf0bc300aa891ea11a4"},
		{Key: "wechat.open_platform_app_id", Value: "wxafa4d4ecf19d8898"},
		{Key: "wechat.open_platform_app_secret", Value: "665accc347d50cf0bc300aa891ea11a4"},
	}

	for _, e := range entries {
		global.GVA_DB.Where("`key` = ?", e.Key).Assign(e).FirstOrCreate(&e)
		fmt.Printf("写入: %s\n", e.Key)
	}

	fmt.Println("配置种子数据写入完成")
}

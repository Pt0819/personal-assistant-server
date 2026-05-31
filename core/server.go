package core

import (
	"fmt"
	"personal-assistant-server/global"
	"personal-assistant-server/initialize"
	"personal-assistant-server/service/system"
	"time"

	"go.uber.org/zap"
)

func RunServer() {
	if global.GVA_CONFIG.System.UseRedis {
		// 初始化redis服务
		initialize.Redis()
		if global.GVA_CONFIG.System.UseMultipoint {
			initialize.RedisList()
		}
	}

	if global.GVA_CONFIG.System.UseMongo {
		err := initialize.Mongo.Initialization()
		if err != nil {
			zap.L().Error(fmt.Sprintf("%+v", err))
		}
	}
	// 从db加载jwt数据
	if global.GVA_DB != nil {
		system.LoadAll()
	}

	Router := initialize.Routers()

	address := fmt.Sprintf(":%d", global.GVA_CONFIG.System.Addr)

	fmt.Printf(`
	欢迎使用 个人ai小助手
	当前版本:%s
	默认前端文件运行地址:http://127.0.0.1:8080
`, global.Version)
	initServer(address, Router, 10*time.Minute, 10*time.Minute)
}

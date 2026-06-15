package core

import (
	"fmt"
	"time"

	"personal-assistant-server/global"
	"personal-assistant-server/initialize"

	"go.uber.org/zap"
)

func RunServer() {
	if global.GVA_CONFIG.System.UseRedis {
		initialize.Redis()
	}

	// 初始化 gRPC Agent 客户端
	// TODO: 当 Agent Server 就绪时取消注释
	// rpc.InitAgentClient(global.GVA_CONFIG.Grpc.AgentAddr)

	Router := initialize.Routers()

	address := fmt.Sprintf(":%d", global.GVA_CONFIG.System.Addr)

	fmt.Printf(`
  欢迎使用 个人AI小助手 API Server
  当前版本:%s
  运行地址: http://127.0.0.1%s
`, global.Version, address)
	zap.L().Info("服务器启动中...", zap.String("address", address))
	initServer(address, Router, 10*time.Minute, 10*time.Minute)
}

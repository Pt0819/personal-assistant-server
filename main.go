package main

import (
	"personal-assistant-server/core"
	"personal-assistant-server/global"
	"personal-assistant-server/initialize"

	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
)

func main() {
	initializeSystem()
	core.RunServer()
}

func initializeSystem() {
	global.GVA_VP = core.Viper()
	global.GVA_LOG = core.Zap()
	zap.ReplaceGlobals(global.GVA_LOG)
	global.GVA_DB = initialize.Gorm()
	initialize.Timer()
	if global.GVA_DB != nil {
		initialize.RegisterTables()
		initialize.ConfigFromDBFallback()
	}
}

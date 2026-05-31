package initialize

import (
	"context"
	model "personal-assistant-server/model/system"
	"personal-assistant-server/plugin/plugin-tool/utils"
)

func Dictionary(ctx context.Context) {
	entities := []model.SysDictionary{}
	utils.RegisterDictionaries(entities...)
}

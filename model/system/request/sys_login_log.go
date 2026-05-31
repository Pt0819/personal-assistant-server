package request

import (
	"personal-assistant-server/model/common/request"
	"personal-assistant-server/model/system"
)

type SysLoginLogSearch struct {
	system.SysLoginLog
	request.PageInfo
}

package request

import (
	"personal-assistant-server/model/common/request"
	"personal-assistant-server/model/system"
)

type SysOperationRecordSearch struct {
	system.SysOperationRecord
	request.PageInfo
}

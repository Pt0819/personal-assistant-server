package request

import (
	"personal-assistant-server/model/common/request"
	"personal-assistant-server/model/system"
)

type SysApiTokenSearch struct {
	system.SysApiToken
	request.PageInfo
    Status *bool `json:"status" form:"status"`
}

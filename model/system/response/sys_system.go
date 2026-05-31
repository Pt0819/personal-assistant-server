package response

import "personal-assistant-server/config"

type SysConfigResponse struct {
	Config config.Server `json:"config"`
}

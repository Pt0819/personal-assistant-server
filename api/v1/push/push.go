package push

import (
	"github.com/gin-gonic/gin"

	"personal-assistant-server/model/common/response"
	"personal-assistant-server/service"
	"personal-assistant-server/utils"
)

type PushApi struct{}

// Subscribe 订阅消息推送
func (a *PushApi) Subscribe(c *gin.Context) {
	var req struct {
		TemplateID string `json:"template_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("请提供模板ID", c)
		return
	}

	userID := utils.GetUserID(c)
	openID := utils.GetOpenID(c)

	sub, err := service.ServiceGroupApp.PushService.Subscribe(c.Request.Context(), userID, openID, req.TemplateID)
	if err != nil {
		response.FailWithMessage("订阅失败: "+err.Error(), c)
		return
	}
	response.OkWithData(sub, c)
}

// Unsubscribe 取消订阅
func (a *PushApi) Unsubscribe(c *gin.Context) {
	var req struct {
		TemplateID string `json:"template_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("请提供模板ID", c)
		return
	}

	userID := utils.GetUserID(c)
	if err := service.ServiceGroupApp.PushService.Unsubscribe(c.Request.Context(), userID, req.TemplateID); err != nil {
		response.FailWithMessage("取消订阅失败: "+err.Error(), c)
		return
	}
	response.Ok(c)
}

// ListSubscriptions 获取订阅列表
func (a *PushApi) ListSubscriptions(c *gin.Context) {
	userID := utils.GetUserID(c)
	subs, err := service.ServiceGroupApp.PushService.GetSubscriptions(c.Request.Context(), userID)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(subs, c)
}

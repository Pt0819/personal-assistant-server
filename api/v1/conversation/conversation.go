package conversation

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"personal-assistant-server/model/common/response"
	"personal-assistant-server/service"
	"personal-assistant-server/utils"
)

type ConversationApi struct{}

// SendMessage 发送对话消息
func (a *ConversationApi) SendMessage(c *gin.Context) {
	var req struct {
		ConversationID uint   `json:"conversation_id"`
		Content        string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("请提供消息内容", c)
		return
	}

	userID := utils.GetUserID(c)
	msg, conv, err := service.ServiceGroupApp.ConversationService.ProcessMessage(
		c.Request.Context(), userID, req.ConversationID, req.Content,
	)
	if err != nil {
		response.FailWithMessage("处理失败: "+err.Error(), c)
		return
	}

	response.OkWithData(gin.H{
		"message":      msg,
		"conversation": conv,
	}, c)
}

// Create 创建新会话
func (a *ConversationApi) Create(c *gin.Context) {
	var req struct {
		Title string `json:"title"`
	}
	c.ShouldBindJSON(&req)

	userID := utils.GetUserID(c)
	conv, err := service.ServiceGroupApp.ConversationService.CreateConversation(
		c.Request.Context(), userID, req.Title,
	)
	if err != nil {
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	response.OkWithData(conv, c)
}

// List 获取会话列表
func (a *ConversationApi) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	userID := utils.GetUserID(c)
	convs, total, err := service.ServiceGroupApp.ConversationService.ListConversations(
		c.Request.Context(), userID, page, pageSize,
	)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{
		"list":  convs,
		"total": total,
	}, c)
}

// Get 获取会话详情
func (a *ConversationApi) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("无效的会话ID", c)
		return
	}

	userID := utils.GetUserID(c)
	conv, err := service.ServiceGroupApp.ConversationService.GetConversation(
		c.Request.Context(), userID, uint(id),
	)
	if err != nil {
		response.FailWithMessage("未找到该会话", c)
		return
	}
	response.OkWithData(conv, c)
}

// ListMessages 获取会话消息列表
func (a *ConversationApi) ListMessages(c *gin.Context) {
	convID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("无效的会话ID", c)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	userID := utils.GetUserID(c)
	messages, total, err := service.ServiceGroupApp.ConversationService.ListMessages(
		c.Request.Context(), userID, uint(convID), page, pageSize,
	)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{
		"list":  messages,
		"total": total,
	}, c)
}

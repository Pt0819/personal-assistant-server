package conversation

import (
	"context"
	"fmt"
	"time"

	"personal-assistant-server/global"
	"personal-assistant-server/model"
	"personal-assistant-server/rpc"
)

type ConversationService struct{}

// ProcessMessage 处理用户对话消息
func (s *ConversationService) ProcessMessage(ctx context.Context, userID uint, convID uint, content string) (*model.Message, *model.Conversation, error) {
	// 1. 如果未指定会话，创建新会话
	var conv *model.Conversation
	if convID == 0 {
		conv = &model.Conversation{
			UserID: userID,
			Title:  truncateContent(content, 100),
			Status: "active",
		}
		if err := global.GVA_DB.WithContext(ctx).Create(conv).Error; err != nil {
			return nil, nil, err
		}
		convID = conv.ID
	} else {
		var c model.Conversation
		if err := global.GVA_DB.WithContext(ctx).Where("id = ? AND user_id = ?", convID, userID).First(&c).Error; err != nil {
			return nil, nil, err
		}
		conv = &c
	}

	// 2. 保存用户消息
	userMsg := &model.Message{
		ConversationID: convID,
		UserID:         userID,
		Role:           "user",
		Content:        content,
	}
	if err := global.GVA_DB.WithContext(ctx).Create(userMsg).Error; err != nil {
		return nil, nil, err
	}

	// 3. 调用 Agent Server 进行 NLU 处理
	agentResp, err := rpc.GetAgentClient().CallAgent(ctx, userID, content, fmt.Sprintf("%d", convID))
	if err != nil {
		// Agent 不可用时，返回降级响应
		fallbackMsg := &model.Message{
			ConversationID: convID,
			UserID:         userID,
			Role:           "assistant",
			Content:        "抱歉，AI服务暂时不可用，请稍后再试。",
			Intent:         "error",
		}
		global.GVA_DB.WithContext(ctx).Create(fallbackMsg)
		return fallbackMsg, conv, nil
	}

	// 4. 保存助手回复
	assistantMsg := &model.Message{
		ConversationID: convID,
		UserID:         userID,
		Role:           "assistant",
		Content:        agentResp.ReplyText,
		Intent:         agentResp.Intent,
		ParsedJSON:     agentResp.ParsedJSON,
		ModelUsed:      agentResp.ModelUsed,
		LatencyMs:      agentResp.LatencyMs,
	}
	if err := global.GVA_DB.WithContext(ctx).Create(assistantMsg).Error; err != nil {
		return nil, nil, err
	}

	// 5. 更新会话的 updated_at
	global.GVA_DB.WithContext(ctx).Model(conv).Update("updated_at", time.Now())

	return assistantMsg, conv, nil
}

// CreateConversation 创建会话
func (s *ConversationService) CreateConversation(ctx context.Context, userID uint, title string) (*model.Conversation, error) {
	if title == "" {
		title = "新对话"
	}
	conv := &model.Conversation{
		UserID: userID,
		Title:  title,
		Status: "active",
	}
	if err := global.GVA_DB.WithContext(ctx).Create(conv).Error; err != nil {
		return nil, err
	}
	return conv, nil
}

// ListConversations 获取会话列表
func (s *ConversationService) ListConversations(ctx context.Context, userID uint, page, pageSize int) ([]model.Conversation, int64, error) {
	var convs []model.Conversation
	var total int64

	query := global.GVA_DB.WithContext(ctx).Model(&model.Conversation{}).Where("user_id = ?", userID)
	query.Count(&total)

	if err := query.Order("updated_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&convs).Error; err != nil {
		return nil, 0, err
	}
	return convs, total, nil
}

// GetConversation 获取会话详情
func (s *ConversationService) GetConversation(ctx context.Context, userID uint, convID uint) (*model.Conversation, error) {
	var conv model.Conversation
	if err := global.GVA_DB.WithContext(ctx).Where("id = ? AND user_id = ?", convID, userID).First(&conv).Error; err != nil {
		return nil, err
	}
	return &conv, nil
}

// ListMessages 获取会话消息列表
func (s *ConversationService) ListMessages(ctx context.Context, userID uint, convID uint, page, pageSize int) ([]model.Message, int64, error) {
	var messages []model.Message
	var total int64

	query := global.GVA_DB.WithContext(ctx).Model(&model.Message{}).Where("conversation_id = ? AND user_id = ?", convID, userID)
	query.Count(&total)

	if err := query.Order("created_at ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&messages).Error; err != nil {
		return nil, 0, err
	}
	return messages, total, nil
}

func truncateContent(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen])
}

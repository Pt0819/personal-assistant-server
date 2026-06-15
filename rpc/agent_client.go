package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// AgentResponse represents the structured response from the Agent Server
type AgentResponse struct {
	ReplyText    string `json:"reply_text"`
	Intent       string `json:"intent"`
	ParsedJSON   string `json:"parsed_json"`
	ModelUsed    string `json:"model_used"`
	LatencyMs    int    `json:"latency_ms"`
	NeedsConfirm bool   `json:"needs_confirmation"`
}

// AgentClient is a gRPC client to the Agent Server
// TODO: Replace stub with real gRPC client when Agent Server is ready
type AgentClient struct {
	addr string
}

var defaultClient *AgentClient

// InitAgentClient initializes the global agent client
func InitAgentClient(addr string) {
	defaultClient = &AgentClient{addr: addr}
}

// GetAgentClient returns the global agent client instance
func GetAgentClient() *AgentClient {
	if defaultClient == nil {
		defaultClient = &AgentClient{addr: "127.0.0.1:50051"}
	}
	return defaultClient
}

// CallAgent sends a message to the Agent Server for NLU processing
// Currently returns a stub response for development without the Agent Server
func (c *AgentClient) CallAgent(ctx context.Context, userID uint, content string, conversationID string) (*AgentResponse, error) {
	// TODO: Replace with actual gRPC call when Agent Server is ready
	// For now, return a stub response so the API Server can work standalone
	startTime := time.Now()

	// Stub: simulate processing delay
	select {
	case <-time.After(10 * time.Millisecond):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	latency := int(time.Since(startTime).Milliseconds())

	// Stub response — the real Agent Server will provide actual NLU parsing
	stubResponse := &AgentResponse{
		ReplyText:    fmt.Sprintf("收到您的消息：「%s」。AI解析功能正在开发中，当前为模拟响应。", truncateString(content, 50)),
		Intent:       "chat",
		ParsedJSON:   "{}",
		ModelUsed:    "stub",
		LatencyMs:    latency,
		NeedsConfirm: false,
	}

	return stubResponse, nil
}

// HealthCheck checks if the Agent Server is available
func (c *AgentClient) HealthCheck(ctx context.Context) error {
	// TODO: Implement real health check via gRPC
	return nil
}

func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// ParseAgentJSON parses the JSON string from agent response into a map
func ParseAgentJSON(jsonStr string) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, err
	}
	return result, nil
}

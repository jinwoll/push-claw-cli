package api

import (
	"encoding/json"
	"time"
)

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// Health 调用 GET /health，返回服务状态和请求延迟
func (c *Client) Health() (*HealthResponse, time.Duration, error) {
	data, statusCode, latency, err := c.GetRaw("/health")
	if err != nil {
		return nil, latency, err
	}
	if statusCode >= 400 {
		return &HealthResponse{Status: "error"}, latency, nil
	}
	var result HealthResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, latency, err
	}
	return &result, latency, nil
}

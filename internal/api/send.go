package api

import "encoding/json"

// SendRequest 发送消息的请求体，对应 POST /api/send
type SendRequest struct {
	Apikey    string `json:"apikey"`
	Role      string `json:"role"`
	Type      string `json:"type,omitempty"`
	Content   string `json:"content"`
	Level     string `json:"level,omitempty"`
	Group     string `json:"group,omitempty"`
	URL       string `json:"url,omitempty"`
	MessageID string `json:"message_id,omitempty"`
}

// SendResponse 发送消息的响应
type SendResponse struct {
	MessageID string `json:"message_id"`
	Status    string `json:"status"` // ok / duplicate / error
}

// Send 调用 POST /api/send 发送消息
func (c *Client) Send(req *SendRequest) (*SendResponse, error) {
	resp, err := c.Post("/api/send", req)
	if err != nil {
		return nil, err
	}
	var result SendResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

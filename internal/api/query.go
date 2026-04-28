package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// Command 服务端返回的单条指令
type Command struct {
	ClientCmdID string `json:"clientCmdId"`
	Apikey      string `json:"apikey"`
	Role        string `json:"role"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	UserID      string `json:"userId"`
	Timestamp   int64  `json:"timestamp"`
	Cursor      int64  `json:"cursor,omitempty"`
}

// QueryResponse 轮询指令的响应
type QueryResponse struct {
	Commands   []Command `json:"commands"`
	NextCursor string    `json:"nextCursor"`
}

// AckRequest 确认指令的请求体
type AckRequest struct {
	Apikey       string   `json:"apikey"`
	Role         string   `json:"role"`
	ClientCmdIDs []string `json:"client_cmd_ids"`
}

// AckResponse 确认指令的响应
type AckResponse struct {
	Acked int `json:"acked"`
}

// Query 调用 GET /api/query/:apikey 拉取指令
func (c *Client) Query(apikey, role, cursor string, limit int) (*QueryResponse, error) {
	params := url.Values{}
	if role != "" {
		params.Set("role", role)
	}
	if cursor != "" {
		params.Set("cursor", cursor)
	}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	path := fmt.Sprintf("/api/query/%s?%s", url.PathEscape(apikey), params.Encode())
	resp, err := c.Get(path)
	if err != nil {
		return nil, err
	}

	var result QueryResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Ack 调用 POST /api/query/ack 确认指令已处理
func (c *Client) Ack(apikey, role string, cmdIDs []string) (*AckResponse, error) {
	req := &AckRequest{
		Apikey:       apikey,
		Role:         role,
		ClientCmdIDs: cmdIDs,
	}
	resp, err := c.Post("/api/query/ack", req)
	if err != nil {
		return nil, err
	}
	var result AckResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

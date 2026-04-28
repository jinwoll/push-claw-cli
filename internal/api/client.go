package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client 封装与推送虾服务端的 HTTP 通信
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Verbose    bool
}

// NewClient 创建 API 客户端，默认超时 30 秒
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// APIResponse 服务端标准响应信封：{ code, message, data }
type APIResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// APIError 服务端返回的业务错误
type APIError struct {
	StatusCode int
	Code       string // 错误码，如 ERR_INVALID_APIKEY
	Message    string
	RawBody    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("[%d] %s: %s", e.StatusCode, e.Code, e.Message)
}

// doRequest 执行 HTTP 请求并解析标准响应信封
func (c *Client) doRequest(method, path string, body interface{}) (*APIResponse, error) {
	url := c.BaseURL + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("User-Agent", "push-claw-cli")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("网络请求失败: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 健康检查等非标准接口直接返回原始 JSON
	var apiResp APIResponse
	if err := json.Unmarshal(respData, &apiResp); err != nil {
		// 无法解析为标准信封时，包装成 APIResponse
		apiResp = APIResponse{
			Code:    resp.StatusCode,
			Message: resp.Status,
			Data:    respData,
		}
	}

	// HTTP 状态码 >= 400 视为错误，尝试提取错误码
	if resp.StatusCode >= 400 {
		apiErr := &APIError{
			StatusCode: resp.StatusCode,
			Code:       apiResp.Message,
			RawBody:    string(respData),
		}
		// data.error 字段存放人类可读信息
		var errData struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(apiResp.Data, &errData) == nil && errData.Error != "" {
			apiErr.Message = errData.Error
		} else {
			apiErr.Message = apiResp.Message
		}
		return nil, apiErr
	}

	return &apiResp, nil
}

// Get 发起 GET 请求
func (c *Client) Get(path string) (*APIResponse, error) {
	return c.doRequest(http.MethodGet, path, nil)
}

// Post 发起 POST 请求
func (c *Client) Post(path string, body interface{}) (*APIResponse, error) {
	return c.doRequest(http.MethodPost, path, body)
}

// Put 发起 PUT 请求
func (c *Client) Put(path string, body interface{}) (*APIResponse, error) {
	return c.doRequest(http.MethodPut, path, body)
}

// Delete 发起 DELETE 请求
func (c *Client) Delete(path string, body interface{}) (*APIResponse, error) {
	return c.doRequest(http.MethodDelete, path, body)
}

// GetRaw 发起 GET 请求并返回原始响应（用于健康检查等非标准接口）
func (c *Client) GetRaw(path string) ([]byte, int, time.Duration, error) {
	url := c.BaseURL + path
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, 0, err
	}
	req.Header.Set("User-Agent", "push-claw-cli")

	start := time.Now()
	resp, err := c.HTTPClient.Do(req)
	latency := time.Since(start)
	if err != nil {
		return nil, 0, latency, fmt.Errorf("网络请求失败: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	return data, resp.StatusCode, latency, err
}

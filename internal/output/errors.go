package output

import (
	"fmt"
	"os"
	"strings"

	"github.com/jinwoll/push-claw-cli/internal/api"
)

// 退出码定义（与 PRD 对应）
const (
	ExitOK          = 0
	ExitGeneral     = 1
	ExitUsage       = 2
	ExitAuthFailed  = 3
	ExitForbidden   = 4
	ExitNetwork     = 5
	ExitQuota       = 6
	ExitRateLimit   = 7
)

// 错误码到退出码的映射
var errorCodeExitMap = map[string]int{
	"ERR_INVALID_APIKEY": ExitAuthFailed,
	"ERR_UNAUTHORIZED":   ExitAuthFailed,
	"ERR_TOKEN_EXPIRED":  ExitAuthFailed,
	"ERR_FORBIDDEN":      ExitForbidden,
	"ERR_PLAN_LIMIT":     ExitForbidden,
	"ERR_ROLE_LIMIT":     ExitForbidden,
	"ERR_QUOTA_EXCEEDED": ExitQuota,
	"ERR_RATE_LIMIT":     ExitRateLimit,
}

// ExitCodeForError 根据错误类型返回对应的退出码
func ExitCodeForError(err error) int {
	if apiErr, ok := err.(*api.APIError); ok {
		if code, exists := errorCodeExitMap[apiErr.Code]; exists {
			return code
		}
		return ExitGeneral
	}
	// 网络相关错误
	msg := err.Error()
	if strings.Contains(msg, "网络请求失败") || strings.Contains(msg, "connection refused") {
		return ExitNetwork
	}
	return ExitGeneral
}

// HandleError 格式化输出错误并退出。附带修复建议和文档链接
func HandleError(err error) {
	if apiErr, ok := err.(*api.APIError); ok {
		PrintError(fmt.Sprintf("%s (%s)", apiErr.Message, apiErr.Code))
		fmt.Fprintln(os.Stderr)
		printSuggestion(apiErr.Code)
	} else {
		PrintError(err.Error())
	}
	os.Exit(ExitCodeForError(err))
}

// HandleErrorSilent 静默模式下只输出 JSON 错误
func HandleErrorSilent(err error) {
	if apiErr, ok := err.(*api.APIError); ok {
		PrintJSON(map[string]interface{}{
			"code":    apiErr.StatusCode,
			"message": apiErr.Code,
			"error":   apiErr.Message,
		})
	} else {
		PrintJSON(map[string]interface{}{
			"code":    ExitGeneral,
			"message": "ERR_INTERNAL",
			"error":   err.Error(),
		})
	}
	os.Exit(ExitCodeForError(err))
}

func printSuggestion(code string) {
	switch code {
	case "ERR_INVALID_APIKEY":
		fmt.Fprintln(os.Stderr, "   请检查你的 API Key 是否正确：")
		fmt.Fprintln(os.Stderr, "   → 运行 push-claw config show 查看当前配置")
		fmt.Fprintln(os.Stderr, "   → 运行 push-claw init 重新配置")
	case "ERR_QUOTA_EXCEEDED":
		fmt.Fprintln(os.Stderr, "   当日配额已用尽，请升级套餐或等待次日重置：")
		fmt.Fprintln(os.Stderr, "   → https://push-claw.app/upgrade")
	case "ERR_PLAN_LIMIT":
		fmt.Fprintln(os.Stderr, "   当前套餐不支持此功能，请升级：")
		fmt.Fprintln(os.Stderr, "   → https://push-claw.app/upgrade")
	case "ERR_RATE_LIMIT":
		fmt.Fprintln(os.Stderr, "   请求过于频繁，请稍后重试。")
	case "ERR_ROLE_REQUIRED":
		fmt.Fprintln(os.Stderr, "   缺少 role 参数，请通过 --role 指定或在 profile 中配置。")
	}
	fmt.Fprintln(os.Stderr, "   → 文档：https://docs.push-claw.app/cli/troubleshooting")
}

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/jinwoll/push-claw-cli/internal/api"
	"github.com/jinwoll/push-claw-cli/internal/media"
	"github.com/jinwoll/push-claw-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	sendType      string
	sendLevel     string
	sendGroup     string
	sendURL       string
	sendMessageID string
	sendSilent    bool
	sendDryRun    bool
)

var sendCmd = &cobra.Command{
	Use:   "send [flags] <content>",
	Short: "发送消息到推送虾",
	Long: `发送文本或图片消息。

示例：
  push-claw send "Hello, World!"
  push-claw send --type image ./screenshot.png
  echo "部署完成" | push-claw send -`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if resolvedCfg.Apikey == "" {
			return fmt.Errorf("缺少 API Key，请运行 push-claw init 或通过 --apikey 指定")
		}

		// 获取消息内容：参数 > stdin
		content, err := resolveContent(args)
		if err != nil {
			return err
		}
		if content == "" {
			return fmt.Errorf("消息内容不能为空")
		}

		// 根据消息类型处理 content（图片/音频需要 base64 编码）
		processedContent, err := processContentByType(sendType, content)
		if err != nil {
			return err
		}

		req := &api.SendRequest{
			Apikey:    resolvedCfg.Apikey,
			Role:      resolvedCfg.Role,
			Type:      sendType,
			Content:   processedContent,
			Level:     sendLevel,
			Group:     sendGroup,
			URL:       sendURL,
			MessageID: sendMessageID,
		}

		// --dry-run：仅打印请求 JSON，不发送
		if sendDryRun {
			data, _ := json.MarshalIndent(req, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		resp, err := apiClient.Send(req)
		if err != nil {
			if sendSilent {
				output.HandleErrorSilent(err)
			}
			output.HandleError(err)
			return nil
		}

		if sendSilent {
			output.PrintJSON(resp)
		} else {
			if resp.Status == "duplicate" {
				output.PrintWarning(fmt.Sprintf("消息已存在（重复）(message_id: %s)", resp.MessageID))
			} else {
				output.PrintSuccess(fmt.Sprintf("消息已发送 (message_id: %s)", resp.MessageID))
			}
		}
		return nil
	},
}

func init() {
	sendCmd.Flags().StringVarP(&sendType, "type", "t", "text", "消息类型：text / image / urgent")
	sendCmd.Flags().StringVarP(&sendLevel, "level", "l", "info", "消息级别：info / warning / error")
	sendCmd.Flags().StringVarP(&sendGroup, "group", "g", "", "消息分组")
	sendCmd.Flags().StringVarP(&sendURL, "url", "u", "", "关联外部链接")
	sendCmd.Flags().StringVarP(&sendMessageID, "message-id", "m", "", "幂等键（默认自动生成 UUID）")
	sendCmd.Flags().BoolVarP(&sendSilent, "silent", "s", false, "静默模式，仅输出 JSON")
	sendCmd.Flags().BoolVar(&sendDryRun, "dry-run", false, "仅打印请求，不实际发送")
	rootCmd.AddCommand(sendCmd)
}

// resolveContent 解析消息内容来源：位置参数、@file 读取文件、- 读取 stdin
func resolveContent(args []string) (string, error) {
	if len(args) == 0 {
		// 无参数时尝试从 stdin 读取
		stat, _ := os.Stdin.Stat()
		if stat.Mode()&os.ModeCharDevice == 0 {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return "", fmt.Errorf("读取 stdin 失败: %w", err)
			}
			return strings.TrimSpace(string(data)), nil
		}
		return "", nil
	}

	content := args[0]

	// "-" 表示从 stdin 读取
	if content == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("读取 stdin 失败: %w", err)
		}
		return strings.TrimSpace(string(data)), nil
	}

	// "@file.txt" 表示读取文件内容作为文本
	if strings.HasPrefix(content, "@") {
		path := content[1:]
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("读取文件失败 \"%s\": %w", path, err)
		}
		return string(data), nil
	}

	return content, nil
}

// processContentByType 根据消息类型对内容做相应处理
func processContentByType(msgType, content string) (string, error) {
	switch msgType {
	case "text":
		return content, nil

	case "urgent":
		return content, nil

	case "image":
		// 图片必须是文件路径，读取并 base64 编码
		if !media.FileExists(content) {
			return "", fmt.Errorf("图片文件不存在: %s", content)
		}
		if !media.IsImageFile(content) {
			return "", fmt.Errorf("不支持的图片格式: %s（支持 png/jpg/jpeg/gif/webp）", content)
		}
		encoded, err := media.EncodeFileBase64(content)
		if err != nil {
			return "", err
		}
		return media.BuildDataURI(content, encoded), nil

	default:
		return "", fmt.Errorf("不支持的消息类型: %s（可用：text, image）", msgType)
	}
}

// 自动生成 message_id（如果未指定）
func init() {
	cobra.OnInitialize(func() {
		if sendMessageID == "" && sendCmd.CalledAs() != "" {
			sendMessageID = uuid.New().String()
		}
	})
}

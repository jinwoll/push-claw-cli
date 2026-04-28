package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/jinwoll/push-claw-cli/internal/api"
	"github.com/jinwoll/push-claw-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	queryLimit    int
	queryWatch    bool
	queryInterval int
	queryAutoAck  bool
	queryExec     string
	queryOutput   string
	queryRole     string
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "轮询拉取指令",
	Long: `从服务器拉取用户下发的指令。

示例：
  push-claw query
  push-claw query -r worker
  push-claw query --watch --interval 5
  push-claw query --watch --auto-ack --exec 'echo "$CONTENT"'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if resolvedCfg.Apikey == "" {
			return fmt.Errorf("缺少 API Key，请运行 push-claw init 或通过 --apikey 指定")
		}

		if queryWatch {
			return runWatchMode()
		}
		return runSingleQuery()
	},
}

func init() {
	queryCmd.Flags().IntVarP(&queryLimit, "limit", "n", 20, "单次拉取条数")
	queryCmd.Flags().BoolVarP(&queryWatch, "watch", "w", false, "持续轮询模式")
	queryCmd.Flags().IntVarP(&queryInterval, "interval", "i", 30, "轮询间隔（秒）")
	queryCmd.Flags().BoolVar(&queryAutoAck, "auto-ack", true, "拉取后自动确认")
	queryCmd.Flags().StringVarP(&queryExec, "exec", "e", "$CONTENT", "对每条指令执行的 shell 命令（可用 $CONTENT, $TYPE, $CMD_ID）")
	queryCmd.Flags().StringVarP(&queryOutput, "output", "o", "table", "输出格式：table / json / raw")
	queryCmd.Flags().StringVarP(&queryRole, "role", "r", "", "角色（未指定时同 push-claw -r / profile 解析结果）")
	rootCmd.AddCommand(queryCmd)
}

// runSingleQuery 单次拉取并展示
func runSingleQuery() error {
	role := effectiveRole(queryRole)
	resp, err := apiClient.Query(resolvedCfg.Apikey, role, "", queryLimit)
	if err != nil {
		output.HandleError(err)
		return nil
	}

	displayCommands(resp.Commands)

	// 执行 --exec 并收集需要 ack 的 ID
	var cmdIDs []string
	for _, c := range resp.Commands {
		if queryExec != "" {
			executeCommand(queryExec, c.Content, c.Type, c.ClientCmdID)
		}
		cmdIDs = append(cmdIDs, c.ClientCmdID)
	}

	if queryAutoAck && len(cmdIDs) > 0 {
		if _, err := apiClient.Ack(resolvedCfg.Apikey, role, cmdIDs); err != nil {
			output.PrintError(fmt.Sprintf("自动确认失败: %v", err))
		} else {
			output.PrintSuccess(fmt.Sprintf("已自动确认 %d 条指令", len(cmdIDs)))
		}
	}

	if len(resp.Commands) > 0 {
		fmt.Printf("📋 共 %d 条指令。使用 push-claw ack <CMD_ID> 确认已处理。\n", len(resp.Commands))
	} else {
		output.PrintInfo("暂无待处理指令。")
	}
	return nil
}

// runWatchMode 持续轮询模式，Ctrl+C 优雅退出
func runWatchMode() error {
	output.PrintInfo(fmt.Sprintf("开始持续监听（间隔 %d 秒）…按 Ctrl+C 停止", queryInterval))

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	cursor := ""
	role := effectiveRole(queryRole)
	for {
		select {
		case <-sigCh:
			fmt.Println()
			output.PrintInfo("已停止监听")
			return nil
		default:
		}

		resp, err := apiClient.Query(resolvedCfg.Apikey, role, cursor, queryLimit)
		if err != nil {
			output.PrintError(fmt.Sprintf("拉取失败: %v", err))
			time.Sleep(time.Duration(queryInterval) * time.Second)
			continue
		}

		if len(resp.Commands) > 0 {
			var cmdIDs []string
			for _, c := range resp.Commands {
				if queryOutput == "json" {
					output.PrintJSON(c)
				} else {
					fmt.Printf("[%s] %s (%s): %s\n",
						output.FormatTimestamp(c.Timestamp),
						c.ClientCmdID, c.Type, c.Content)
				}
				if queryExec != "" {
					executeCommand(queryExec, c.Content, c.Type, c.ClientCmdID)
				}
				cmdIDs = append(cmdIDs, c.ClientCmdID)
			}

			if queryAutoAck && len(cmdIDs) > 0 {
				if _, err := apiClient.Ack(resolvedCfg.Apikey, role, cmdIDs); err != nil {
					output.PrintError(fmt.Sprintf("自动确认失败: %v", err))
				}
			}
			cursor = resp.NextCursor
		}

		time.Sleep(time.Duration(queryInterval) * time.Second)
	}
}

// displayCommands 根据 --output 格式输出指令列表
func displayCommands(commands []api.Command) {
	switch queryOutput {
	case "json":
		output.PrintJSON(commands)
	case "raw":
		for _, c := range commands {
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n",
				c.ClientCmdID, c.Role, c.Type, c.Content,
				output.FormatTimestamp(c.Timestamp))
		}
	default: // table
		if len(commands) == 0 {
			return
		}
		headers := []string{"CMD_ID", "ROLE", "TYPE", "CONTENT", "TIME"}
		var rows [][]string
		for _, c := range commands {
			// 截断过长的 content
			content := c.Content
			if len(content) > 40 {
				content = content[:37] + "..."
			}
			rows = append(rows, []string{
				c.ClientCmdID, c.Role, c.Type, content,
				output.FormatTimestamp(c.Timestamp),
			})
		}
		output.PrintTable(headers, rows)
	}
}

// executeCommand 执行用户指定的 shell 命令，注入环境变量
func executeCommand(template, content, cmdType, cmdID string) {
	cmdStr := strings.ReplaceAll(template, "$CONTENT", content)
	cmdStr = strings.ReplaceAll(cmdStr, "$TYPE", cmdType)
	cmdStr = strings.ReplaceAll(cmdStr, "$CMD_ID", cmdID)

	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.Command("cmd", "/C", cmdStr)
	} else {
		c = exec.Command("sh", "-c", cmdStr)
	}
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = append(os.Environ(),
		"CONTENT="+content,
		"TYPE="+cmdType,
		"CMD_ID="+cmdID,
	)

	if err := c.Run(); err != nil {
		output.PrintError(fmt.Sprintf("执行命令失败: %v", err))
	}
}

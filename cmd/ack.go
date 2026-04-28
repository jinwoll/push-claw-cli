package cmd

import (
	"fmt"

	"github.com/jinwoll/push-claw-cli/internal/output"
	"github.com/spf13/cobra"
)

var ackAll bool

var ackCmd = &cobra.Command{
	Use:   "ack <cmd_id> [cmd_id...]",
	Short: "确认指令已处理",
	Long: `确认一条或多条指令已处理，服务器将不再返回这些指令。

示例：
  push-claw ack cmd-uuid-001
  push-claw ack cmd-uuid-001 cmd-uuid-002 cmd-uuid-003
  push-claw ack --all`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if resolvedCfg.Apikey == "" {
			return fmt.Errorf("缺少 API Key，请运行 push-claw init 或通过 --apikey 指定")
		}

		var cmdIDs []string

		if ackAll {
			// --all 模式：先拉取所有待处理指令的 ID，再批量确认
			resp, err := apiClient.Query(resolvedCfg.Apikey, resolvedCfg.Role, "", 1000)
			if err != nil {
				output.HandleError(err)
				return nil
			}
			if len(resp.Commands) == 0 {
				output.PrintInfo("没有待确认的指令。")
				return nil
			}
			for _, c := range resp.Commands {
				cmdIDs = append(cmdIDs, c.ClientCmdID)
			}
		} else {
			if len(args) == 0 {
				return fmt.Errorf("请指定至少一个 cmd_id，或使用 --all 确认全部")
			}
			cmdIDs = args
		}

		result, err := apiClient.Ack(resolvedCfg.Apikey, resolvedCfg.Role, cmdIDs)
		if err != nil {
			output.HandleError(err)
			return nil
		}

		output.PrintSuccess(fmt.Sprintf("已确认 %d 条指令", result.Acked))
		return nil
	},
}

func init() {
	ackCmd.Flags().BoolVar(&ackAll, "all", false, "确认全部已拉取的指令")
	rootCmd.AddCommand(ackCmd)
}

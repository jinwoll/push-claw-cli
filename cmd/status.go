package cmd

import (
	"fmt"

	apiPkg "github.com/jinwoll/push-claw-cli/internal/api"
	"github.com/jinwoll/push-claw-cli/internal/config"
	"github.com/jinwoll/push-claw-cli/internal/output"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "检查推送虾服务状态",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Resolve(config.ResolveOptions{
			FlagApikey:  flagApikey,
			FlagRole:    flagRole,
			FlagBaseURL: flagBaseURL,
			FlagProfile: flagProfile,
		})
		if err != nil {
			return err
		}

		client := apiClient
		if client == nil {
			client = apiPkg.NewClient(cfg.BaseURL)
		}

		fmt.Println("🦐 推送虾服务状态")
		fmt.Println()

		health, latency, err := client.Health()
		if err != nil {
			output.PrintKeyValue([][]string{
				{"服务器", cfg.BaseURL},
				{"状态", "❌ 不可用"},
				{"错误", err.Error()},
			})
			return nil
		}

		statusText := "✅ 在线"
		if health.Status != "ok" {
			statusText = "⚠ 异常"
		}

		apikeyDisplay := "（未配置）"
		if cfg.Apikey != "" {
			apikeyDisplay = output.MaskApikey(cfg.Apikey)
		}

		output.PrintKeyValue([][]string{
			{"服务器", cfg.BaseURL},
			{"状态", fmt.Sprintf("%s（延迟 %dms）", statusText, latency.Milliseconds())},
			{"版本", fmt.Sprintf("CLI v%s", version)},
			{"Profile", cfg.Profile},
			{"API Key", apikeyDisplay},
			{"角色", cfg.Role},
		})
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

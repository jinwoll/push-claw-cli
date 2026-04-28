package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jinwoll/push-claw-cli/internal/config"
	"github.com/jinwoll/push-claw-cli/internal/output"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "交互式初始化配置",
	Long:  "引导你输入 API Key、角色名称和服务器地址，保存为默认 profile。",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🦐 欢迎使用推送虾 CLI！")
		fmt.Println()

		scanner := bufio.NewScanner(os.Stdin)

		apikey := prompt(scanner, "请输入你的 API Key", "")
		if apikey == "" {
			return fmt.Errorf("API Key 不能为空")
		}

		role := prompt(scanner, "请输入角色名称（默认 bot）", config.DefaultRole)
		baseURL := prompt(scanner, "请输入服务器地址（默认 "+config.DefaultBaseURL+"）", config.DefaultBaseURL)
		profileName := prompt(scanner, "Profile 名称（默认 default）", "default")

		// 保存 profile
		profile := &config.Profile{
			Apikey:  apikey,
			Role:    role,
			BaseURL: baseURL,
		}
		if err := config.SaveProfile(profileName, profile); err != nil {
			return fmt.Errorf("保存 profile 失败: %w", err)
		}

		// 设为当前 profile
		global, _ := config.LoadGlobal()
		if global == nil {
			global = &config.GlobalConfig{AutoUpgradeCheck: true}
		}
		global.CurrentProfile = profileName
		if err := config.SaveGlobal(global); err != nil {
			return fmt.Errorf("保存全局配置失败: %w", err)
		}

		fmt.Println()
		output.PrintSuccess(fmt.Sprintf("配置已保存至 %s", config.ConfigDir()))
		fmt.Printf("🚀 快速开始：push-claw send \"Hello, World!\"\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

// prompt 显示提示并读取用户输入，为空时返回默认值
func prompt(scanner *bufio.Scanner, label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("? %s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("? %s: ", label)
	}
	scanner.Scan()
	text := strings.TrimSpace(scanner.Text())
	if text == "" {
		return defaultVal
	}
	return text
}

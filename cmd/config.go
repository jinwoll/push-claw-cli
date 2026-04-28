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

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "管理配置与 profile",
}

// ---- config list ----
var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有 profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		global, err := config.LoadGlobal()
		if err != nil {
			return err
		}
		names, err := config.ListProfiles()
		if err != nil {
			return err
		}
		if len(names) == 0 {
			output.PrintInfo("暂无 profile，运行 push-claw init 创建。")
			return nil
		}
		for _, name := range names {
			marker := "  "
			if name == global.CurrentProfile {
				marker = "* " // 当前激活的 profile
			}
			fmt.Printf("%s%s\n", marker, name)
		}
		return nil
	},
}

// ---- config show ----
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "显示当前 profile 详情",
	RunE: func(cmd *cobra.Command, args []string) error {
		global, err := config.LoadGlobal()
		if err != nil {
			return err
		}
		p, err := config.LoadProfile(global.CurrentProfile)
		if err != nil {
			return err
		}
		output.PrintKeyValue([][]string{
			{"Profile", global.CurrentProfile},
			{"API Key", output.MaskApikey(p.Apikey)},
			{"角色", p.Role},
			{"服务器", p.BaseURL},
			{"配置目录", config.ConfigDir()},
		})
		return nil
	},
}

// ---- config set ----
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "修改当前 profile 的配置项",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]
		global, err := config.LoadGlobal()
		if err != nil {
			return err
		}
		p, err := config.LoadProfile(global.CurrentProfile)
		if err != nil {
			return err
		}
		switch strings.ToLower(key) {
		case "apikey":
			p.Apikey = value
		case "role":
			p.Role = value
		case "base_url", "baseurl", "base-url":
			p.BaseURL = value
		default:
			return fmt.Errorf("不支持的配置项：%s（可用：apikey, role, base_url）", key)
		}
		if err := config.SaveProfile(global.CurrentProfile, p); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("已更新 %s.%s = %s", global.CurrentProfile, key, value))
		return nil
	},
}

// ---- config use ----
var configUseCmd = &cobra.Command{
	Use:   "use <profile_name>",
	Short: "切换当前使用的 profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if !config.ProfileExists(name) {
			return fmt.Errorf("profile \"%s\" 不存在，运行 push-claw config create %s 创建", name, name)
		}
		global, _ := config.LoadGlobal()
		if global == nil {
			global = &config.GlobalConfig{AutoUpgradeCheck: true}
		}
		global.CurrentProfile = name
		if err := config.SaveGlobal(global); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("已切换到 profile: %s", name))
		return nil
	},
}

// ---- config create ----
var configCreateCmd = &cobra.Command{
	Use:   "create <profile_name>",
	Short: "交互式创建新 profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if config.ProfileExists(name) {
			return fmt.Errorf("profile \"%s\" 已存在", name)
		}
		scanner := bufio.NewScanner(os.Stdin)
		apikey := prompt(scanner, "请输入 API Key", "")
		role := prompt(scanner, "请输入角色名称（默认 bot）", config.DefaultRole)
		baseURL := prompt(scanner, "请输入服务器地址（默认 "+config.DefaultBaseURL+"）", config.DefaultBaseURL)

		p := &config.Profile{Apikey: apikey, Role: role, BaseURL: baseURL}
		if err := config.SaveProfile(name, p); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Profile \"%s\" 已创建", name))
		return nil
	},
}

// ---- config delete ----
var configDeleteCmd = &cobra.Command{
	Use:   "delete <profile_name>",
	Short: "删除指定 profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		global, _ := config.LoadGlobal()
		if global != nil && global.CurrentProfile == name {
			return fmt.Errorf("不能删除当前正在使用的 profile \"%s\"，请先 push-claw config use <other>", name)
		}
		if err := config.DeleteProfile(name); err != nil {
			return fmt.Errorf("删除 profile \"%s\" 失败: %w", name, err)
		}
		output.PrintSuccess(fmt.Sprintf("Profile \"%s\" 已删除", name))
		return nil
	},
}

func init() {
	configCmd.AddCommand(configListCmd, configShowCmd, configSetCmd, configUseCmd, configCreateCmd, configDeleteCmd)
	rootCmd.AddCommand(configCmd)
}

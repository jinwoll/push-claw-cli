package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jinwoll/push-claw-cli/internal/config"
	"github.com/jinwoll/push-claw-cli/internal/output"
	"github.com/spf13/cobra"
)

var uninstallForce bool

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "卸载 push-claw CLI",
	Long:  "删除 push-claw 二进制文件和配置目录。",
	RunE: func(cmd *cobra.Command, args []string) error {
		execPath, err := os.Executable()
		if err != nil {
			return err
		}
		execPath, _ = filepath.EvalSymlinks(execPath)
		configDir := config.ConfigDir()

		if !uninstallForce {
			fmt.Printf("即将删除：\n")
			fmt.Printf("  · 二进制文件: %s\n", execPath)
			fmt.Printf("  · 配置目录:   %s\n", configDir)
			fmt.Print("\n确认卸载？(y/N): ")

			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
			if answer != "y" && answer != "yes" {
				output.PrintInfo("已取消卸载")
				return nil
			}
		}

		// 删除配置目录
		if err := os.RemoveAll(configDir); err != nil {
			output.PrintWarning(fmt.Sprintf("删除配置目录失败: %v", err))
		} else {
			output.PrintSuccess("配置目录已删除")
		}

		// 删除自身二进制（Windows 下可能需要延迟删除）
		if err := os.Remove(execPath); err != nil {
			output.PrintWarning(fmt.Sprintf("删除二进制文件失败: %v（可手动删除 %s）", err, execPath))
		} else {
			output.PrintSuccess("二进制文件已删除")
		}

		fmt.Println("\n👋 推送虾 CLI 已卸载。")
		return nil
	},
}

func init() {
	uninstallCmd.Flags().BoolVar(&uninstallForce, "force", false, "跳过确认直接删除")
	rootCmd.AddCommand(uninstallCmd)
}

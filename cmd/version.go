package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本号",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("push-claw v%s (commit: %s, built: %s)\n", version, commit, date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	// 同时支持 --version flag
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("push-claw v{{.Version}}\n")
}

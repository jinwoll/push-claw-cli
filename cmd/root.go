package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jinwoll/push-claw-cli/internal/api"
	"github.com/jinwoll/push-claw-cli/internal/config"
	"github.com/spf13/cobra"
)

// 构建时通过 ldflags 注入
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// 全局 flag 值，在 PersistentPreRun 中解析为 resolvedCfg
var (
	flagApikey  string
	flagRole    string
	flagBaseURL string
	flagProfile string
	flagVerbose bool
	flagDebug   bool

	resolvedCfg *config.ResolvedConfig
	apiClient   *api.Client
)

var rootCmd = &cobra.Command{
	Use:   "push-claw",
	Short: "推送虾 CLI — 跨平台命令行工具",
	Long: `🦐 推送虾 CLI (push-claw)

一行命令即可安装的跨平台工具，封装推送虾全部 HTTP API：
  · 发送消息（文本/图片/语音/语音通话）
  · WebSocket 实时接收指令与轮询拉取
  · 配置持久化与多 profile 管理

快速开始: push-claw init`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// init / config create / version / help 等命令不需要解析配置
		if skipConfigResolve(cmd) {
			return nil
		}
		cfg, err := config.Resolve(config.ResolveOptions{
			FlagApikey:  flagApikey,
			FlagRole:    flagRole,
			FlagBaseURL: flagBaseURL,
			FlagProfile: flagProfile,
		})
		if err != nil {
			return err
		}
		resolvedCfg = cfg
		apiClient = api.NewClient(cfg.BaseURL)
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// 注册全局持久化 flags
	pf := rootCmd.PersistentFlags()
	pf.StringVarP(&flagApikey, "apikey", "k", "", "API 密钥")
	pf.StringVarP(&flagRole, "role", "r", "", "角色名称")
	pf.StringVar(&flagBaseURL, "base-url", "", "服务器地址")
	pf.StringVar(&flagProfile, "profile", "", "使用指定 profile")
	pf.BoolVar(&flagVerbose, "verbose", false, "输出详细日志")
	pf.BoolVar(&flagDebug, "debug", false, "输出调试信息")
}

// skipConfigResolve 判断当前命令是否无需加载 profile 配置
func skipConfigResolve(cmd *cobra.Command) bool {
	name := cmd.Name()
	switch name {
	case "init", "version", "help", "completion", "uninstall":
		return true
	}
	// config create / config delete 等子命令也跳过
	if cmd.Parent() != nil && cmd.Parent().Name() == "config" {
		return true
	}
	return false
}

// effectiveRole 子命令本地 -r 优先，否则使用 PersistentPreRun 解析后的 resolvedCfg.Role
func effectiveRole(localOverride string) string {
	if s := strings.TrimSpace(localOverride); s != "" {
		return s
	}
	if resolvedCfg == nil {
		return ""
	}
	return resolvedCfg.Role
}

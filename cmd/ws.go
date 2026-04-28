package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/jinwoll/push-claw-cli/internal/api"
	"github.com/jinwoll/push-claw-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	wsExec    string
	wsAutoAck bool
	wsRole    string
)

var wsCmd = &cobra.Command{
	Use:   "ws",
	Short: "WebSocket 实时接收指令",
	Long: `连接服务端的 CLI WebSocket，接收下行指令（与轮询队列语义一致）。

示例：
  push-claw ws
  push-claw ws -r worker
  push-claw ws --exec 'echo "$CONTENT"'
  push-claw ws --auto-ack=false`,
	RunE: runWs,
}

func init() {
	wsCmd.Flags().StringVarP(&wsExec, "exec", "e", "$CONTENT", "收到指令后执行的 shell（$CONTENT / $TYPE / $CMD_ID）")
	wsCmd.Flags().BoolVar(&wsAutoAck, "auto-ack", true, "处理完后自动 HTTP 确认")
	wsCmd.Flags().StringVarP(&wsRole, "role", "r", "", "角色（未指定时同 push-claw -r / profile 解析结果）")
	rootCmd.AddCommand(wsCmd)
}

func cliWebSocketURL(baseURL, apikey, role string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return "", err
	}
	if u.Host == "" {
		return "", fmt.Errorf("无效的 base URL")
	}
	scheme := "ws"
	if u.Scheme == "https" {
		scheme = "wss"
	}
	path := "/api/cli/ws"
	q := url.Values{}
	q.Set("apikey", apikey)
	q.Set("role", role)
	return fmt.Sprintf("%s://%s%s?%s", scheme, u.Host, path, q.Encode()), nil
}

type wsEnvelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func runWs(_ *cobra.Command, _ []string) error {
	if resolvedCfg.Apikey == "" {
		return fmt.Errorf("缺少 API Key，请运行 push-claw init 或通过 --apikey 指定")
	}
	role := effectiveRole(wsRole)
	if role == "" {
		return fmt.Errorf("缺少 role，请通过 -r / --role 或 profile 指定")
	}

	wsURL, err := cliWebSocketURL(resolvedCfg.BaseURL, resolvedCfg.Apikey, role)
	if err != nil {
		return err
	}

	output.PrintInfo(fmt.Sprintf("连接 %s …", wsURL))

	dialer := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: apiClient.HTTPClient.Timeout,
	}

	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		output.HandleError(err)
		return nil
	}
	defer conn.Close()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		for {
			_, msg, rerr := conn.ReadMessage()
			if rerr != nil {
				errCh <- rerr
				return
			}
			handleWsFrame(role, msg)
		}
	}()

	select {
	case <-sigCh:
		output.PrintInfo("已断开。")
		_ = conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	case err := <-errCh:
		if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
			output.HandleError(err)
		}
	}
	return nil
}

func handleWsFrame(role string, raw []byte) {
	var env wsEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		output.PrintError(fmt.Sprintf("无法解析消息: %v", err))
		return
	}
	switch env.Type {
	case "command":
		var c api.Command
		if err := json.Unmarshal(env.Data, &c); err != nil {
			output.PrintError(fmt.Sprintf("指令解析失败: %v", err))
			return
		}
		displayCommands([]api.Command{c})
		if wsExec != "" {
			executeCommand(wsExec, c.Content, c.Type, c.ClientCmdID)
		}
		if wsAutoAck {
			if _, err := apiClient.Ack(resolvedCfg.Apikey, role, []string{c.ClientCmdID}); err != nil {
				output.PrintError(fmt.Sprintf("自动确认失败: %v", err))
			} else {
				output.PrintSuccess(fmt.Sprintf("已确认 %s", c.ClientCmdID))
			}
		} else {
			fmt.Printf("📋 待确认: %s （push-claw ack %s）\n", c.ClientCmdID, c.ClientCmdID)
		}
	default:
		output.PrintInfo(string(raw))
	}
}

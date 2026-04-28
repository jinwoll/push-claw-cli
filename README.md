# 推送虾 CLI（push-claw）

跨平台命令行工具，用于连接 [推送虾](https://push-claw.com) 服务，支持：

- 发送消息（文本、图片）
- 轮询拉取指令并确认（`query` / `ack`）
- WebSocket 实时接收指令（`ws`）
- 多环境配置（profile）管理
- 自更新与卸载（`upgrade` / `uninstall`）

仓库地址：`github.com/jinwoll/push-claw-cli`  
模块路径：`github.com/jinwoll/push-claw-cli`

---

## 功能特性

- **一键安装**：支持 Linux / macOS / Windows（PowerShell）
- **开箱即用**：`push-claw init` 完成交互式初始化
- **脚本友好**：支持 stdin、`@file`、JSON 输出、`--exec` 自动执行
- **实时与轮询并存**：可按场景选择 `ws` 或 `query --watch`
- **配置优先级清晰**：`flag > env > profile > default`

---

## 环境要求

- 从源码编译：Go `1.26.1+`（以 `go.mod` 为准）
- 一键安装脚本：
  - Linux / macOS：`curl` 或 `wget`
  - Windows：PowerShell `5.1+`

---

## 安装

### 方式一：一键安装（推荐）

Linux / macOS / WSL / Git Bash：

```bash
curl -fsSL https://raw.githubusercontent.com/jinwoll/push-claw-cli/main/install.sh | sh
```

Windows PowerShell：

```powershell
irm https://raw.githubusercontent.com/jinwoll/push-claw-cli/main/install.ps1 | iex
```

### 方式二：Go 安装

```bash
go install github.com/jinwoll/push-claw-cli@latest
```

> 安装后请确保 `$(go env GOPATH)/bin` 在 `PATH` 中。

---

## 快速开始

```bash
# 1) 初始化配置（输入 apikey / role / base_url）
push-claw init

# 2) 发送消息
push-claw send "Hello, World!"

# 3) 查看服务状态
push-claw status
```

---

## 常用命令

| 命令 | 说明 |
|------|------|
| `push-claw init` | 交互式初始化配置 |
| `push-claw send` | 发送消息（文本 / 图片） |
| `push-claw query` | 拉取指令（支持 `--watch` 持续轮询） |
| `push-claw ws` | WebSocket 实时接收指令 |
| `push-claw ack` | 确认一条或多条指令 |
| `push-claw config` | 管理 profile（list/show/set/use/create/delete） |
| `push-claw status` | 检查服务连通性与延迟 |
| `push-claw upgrade` | 检查并升级到最新版本 |
| `push-claw uninstall` | 卸载 CLI 与配置 |
| `push-claw version` | 查看版本信息 |

---

## 使用示例

### 发送消息

```bash
# 文本
push-claw send "部署完成"

# 图片
push-claw send --type image ./screenshot.png

# 从管道读取内容
echo "来自脚本的消息" | push-claw send -
```

### 拉取指令

```bash
# 单次拉取
push-claw query

# 持续轮询（每 5 秒）
push-claw query --watch --interval 5

# 拉取后自动执行命令并自动确认
push-claw query --watch --auto-ack --exec 'echo "$CONTENT"'
```

### WebSocket 实时接收

```bash
push-claw ws
push-claw ws --exec 'echo "$CONTENT"'
push-claw ws --auto-ack=false
```

### 指令确认

```bash
# 确认单条
push-claw ack cmd-001

# 确认多条
push-claw ack cmd-001 cmd-002

# 一次确认全部待处理
push-claw ack --all
```

---

## 配置说明

### 配置目录

- Windows：`%LOCALAPPDATA%\push-claw\`
- Linux / macOS：`~/.push-claw/`

目录下包含：

- `config.toml`（全局配置）
- `profiles/<name>.toml`（多环境配置）

### 环境变量

| 变量 | 说明 |
|------|------|
| `MINIXIA_APIKEY` | API Key |
| `MINIXIA_ROLE` | 角色名 |
| `MINIXIA_BASE_URL` | 服务端地址 |
| `MINIXIA_PROFILE` | 指定 profile |
| `MINIXIA_GITHUB_OWNER` | （可选）升级检查用仓库 Owner |
| `MINIXIA_GITHUB_REPO` | （可选）升级检查用仓库名 |

配置合并优先级：**命令行参数 > 环境变量 > profile 文件 > 默认值**

---

## 开发与发布

- 本地构建：

```bash
go build -o push-claw .
```

- 查看全部命令帮助：

```bash
push-claw --help
push-claw <command> --help
```

更完整的安装、构建、GoReleaser 打包与 GitHub Release 发布说明见：`INSTALL.md`。

---

## 许可证

如仓库根目录存在 `LICENSE` 文件，以该文件为准。

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/jinwoll/push-claw-cli/internal/output"
	"github.com/spf13/cobra"
)

// 与发布仓库一致；可通过环境变量覆盖（fork 自建 Release 时）
const (
	defaultGithubOwner = "jinwoll"
	defaultGithubRepo  = "push-claw-cli"
)

var upgradeCheckOnly bool

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "检查并升级到最新版本",
	RunE: func(cmd *cobra.Command, args []string) error {
		output.PrintInfo("正在检查最新版本…")

		latest, downloadURL, err := fetchLatestVersion()
		if err != nil {
			return fmt.Errorf("检查更新失败: %w", err)
		}

		current := version
		if normalizeSemver(current) == normalizeSemver(latest) {
			output.PrintSuccess(fmt.Sprintf("已是最新版本 v%s", current))
			return nil
		}

		output.PrintInfo(fmt.Sprintf("发现新版本 v%s（当前 v%s）", latest, current))

		if upgradeCheckOnly {
			fmt.Printf("运行 push-claw upgrade 执行升级。\n")
			return nil
		}

		output.PrintInfo("正在下载…")
		if err := downloadAndReplace(downloadURL); err != nil {
			return fmt.Errorf("升级失败: %w", err)
		}

		output.PrintSuccess(fmt.Sprintf("已升级到 v%s", latest))
		return nil
	},
}

func init() {
	upgradeCmd.Flags().BoolVar(&upgradeCheckOnly, "check", false, "仅检查，不执行升级")
	rootCmd.AddCommand(upgradeCmd)
}

func githubOwnerRepo() (owner, repo string) {
	owner = os.Getenv("MINIXIA_GITHUB_OWNER")
	repo = os.Getenv("MINIXIA_GITHUB_REPO")
	if owner == "" {
		owner = defaultGithubOwner
	}
	if repo == "" {
		repo = defaultGithubRepo
	}
	return owner, repo
}

// fetchLatestVersion 通过 GitHub Releases API 获取最新 tag 与当前平台安装包下载地址
func fetchLatestVersion() (string, string, error) {
	owner, repo := githubOwnerRepo()
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "push-claw-cli")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", "", fmt.Errorf("GitHub API %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var rel struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", "", fmt.Errorf("解析 Release 信息失败: %w", err)
	}
	if rel.TagName == "" {
		return "", "", fmt.Errorf("响应中无 tag_name")
	}

	latest := strings.TrimPrefix(rel.TagName, "v")

	osName := runtime.GOOS
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x86_64"
	}
	ext := ""
	if osName == "windows" {
		ext = ".exe"
	}
	filename := fmt.Sprintf("push-claw-%s-%s%s", osName, arch, ext)
	dl := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", owner, repo, rel.TagName, filename)
	return latest, dl, nil
}

func normalizeSemver(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

// downloadAndReplace 下载新二进制并替换当前可执行文件
func downloadAndReplace(url string) error {
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("下载失败，HTTP %d", resp.StatusCode)
	}

	// 写入临时文件
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execPath, _ = filepath.EvalSymlinks(execPath)

	tmpPath := execPath + ".new"
	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return err
	}
	f.Close()

	// 赋予执行权限（非 Windows）
	if !strings.HasSuffix(tmpPath, ".exe") {
		os.Chmod(tmpPath, 0755)
	}

	// 替换旧文件：先重命名旧文件，再重命名新文件
	oldPath := execPath + ".old"
	os.Remove(oldPath)
	if err := os.Rename(execPath, oldPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("备份旧版本失败: %w", err)
	}
	if err := os.Rename(tmpPath, execPath); err != nil {
		os.Rename(oldPath, execPath) // 回滚
		return fmt.Errorf("替换新版本失败: %w", err)
	}
	os.Remove(oldPath) // 清理旧版本备份

	return nil
}

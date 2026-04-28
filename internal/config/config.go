package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// GlobalConfig 全局配置，存储在 config.toml
type GlobalConfig struct {
	CurrentProfile   string `toml:"current_profile"`
	AutoUpgradeCheck bool   `toml:"auto_upgrade_check"`
}

// Profile 单个配置档，存储在 profiles/<name>.toml
type Profile struct {
	Apikey  string `toml:"apikey"`
	Role    string `toml:"role"`
	BaseURL string `toml:"base_url"`
}

// ConfigDir 返回配置目录路径，Windows 用 %LOCALAPPDATA%\push-claw，其他用 ~/.push-claw
func ConfigDir() string {
	if runtime.GOOS == "windows" {
		if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
			return filepath.Join(dir, "push-claw")
		}
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".push-claw")
}

// EnsureDirs 确保配置目录和 profiles 子目录存在
func EnsureDirs() error {
	return os.MkdirAll(filepath.Join(ConfigDir(), "profiles"), 0700)
}

// LoadGlobal 加载全局配置（config.toml）
func LoadGlobal() (*GlobalConfig, error) {
	path := filepath.Join(ConfigDir(), "config.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在时返回默认配置
			return &GlobalConfig{CurrentProfile: "default", AutoUpgradeCheck: true}, nil
		}
		return nil, err
	}
	var cfg GlobalConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析 config.toml 失败: %w", err)
	}
	if cfg.CurrentProfile == "" {
		cfg.CurrentProfile = "default"
	}
	return &cfg, nil
}

// SaveGlobal 保存全局配置到 config.toml
func SaveGlobal(cfg *GlobalConfig) error {
	if err := EnsureDirs(); err != nil {
		return err
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(ConfigDir(), "config.toml"), data, 0600)
}

// LoadProfile 加载指定名称的 profile
func LoadProfile(name string) (*Profile, error) {
	path := filepath.Join(ConfigDir(), "profiles", name+".toml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("profile \"%s\" 不存在", name)
		}
		return nil, err
	}
	var p Profile
	if err := toml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("解析 profile \"%s\" 失败: %w", name, err)
	}
	return &p, nil
}

// SaveProfile 保存 profile 到 profiles/<name>.toml
func SaveProfile(name string, p *Profile) error {
	if err := EnsureDirs(); err != nil {
		return err
	}
	data, err := toml.Marshal(p)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(ConfigDir(), "profiles", name+".toml"), data, 0600)
}

// ListProfiles 列出所有 profile 名称
func ListProfiles() ([]string, error) {
	dir := filepath.Join(ConfigDir(), "profiles")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".toml") {
			names = append(names, strings.TrimSuffix(e.Name(), ".toml"))
		}
	}
	return names, nil
}

// DeleteProfile 删除指定 profile
func DeleteProfile(name string) error {
	path := filepath.Join(ConfigDir(), "profiles", name+".toml")
	return os.Remove(path)
}

// ProfileExists 检查 profile 是否存在
func ProfileExists(name string) bool {
	path := filepath.Join(ConfigDir(), "profiles", name+".toml")
	_, err := os.Stat(path)
	return err == nil
}

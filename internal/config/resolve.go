package config

import "os"

const (
	DefaultBaseURL = "http://192.168.2.116:3000" // "https://api.push-claw.app"
	DefaultRole    = "bot"
)

// ResolvedConfig 合并后的最终配置，来源优先级：CLI flag > 环境变量 > profile > 默认值
type ResolvedConfig struct {
	Apikey  string
	Role    string
	BaseURL string
	Profile string
}

// ResolveOptions 用于传入各层配置源的值
type ResolveOptions struct {
	FlagApikey  string
	FlagRole    string
	FlagBaseURL string
	FlagProfile string
}

// Resolve 按照 CLI flag > 环境变量 > profile 文件 > 默认值 的优先级合并配置
func Resolve(opts ResolveOptions) (*ResolvedConfig, error) {
	// 确定当前使用的 profile 名称
	profileName := firstNonEmpty(opts.FlagProfile, os.Getenv("MINIXIA_PROFILE"))
	if profileName == "" {
		g, err := LoadGlobal()
		if err != nil {
			return nil, err
		}
		profileName = g.CurrentProfile
	}

	// 尝试加载 profile；不存在也不报错，只是没有 profile 层的值
	var profileApikey, profileRole, profileBaseURL string
	if p, err := LoadProfile(profileName); err == nil {
		profileApikey = p.Apikey
		profileRole = p.Role
		profileBaseURL = p.BaseURL
	}

	return &ResolvedConfig{
		Apikey:  firstNonEmpty(opts.FlagApikey, os.Getenv("MINIXIA_APIKEY"), profileApikey),
		Role:    firstNonEmpty(opts.FlagRole, os.Getenv("MINIXIA_ROLE"), profileRole, DefaultRole),
		BaseURL: firstNonEmpty(opts.FlagBaseURL, os.Getenv("MINIXIA_BASE_URL"), profileBaseURL, DefaultBaseURL),
		Profile: profileName,
	}, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

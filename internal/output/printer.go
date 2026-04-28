package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
)

var (
	green  = color.New(color.FgGreen).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
	bold   = color.New(color.Bold).SprintFunc()
)

// PrintSuccess 输出成功信息
func PrintSuccess(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", green("✅"), msg)
}

// PrintInfo 输出提示信息
func PrintInfo(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", cyan("ℹ"), msg)
}

// PrintWarning 输出警告信息
func PrintWarning(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", yellow("⚠"), msg)
}

// PrintError 输出错误信息到 stderr
func PrintError(msg string) {
	fmt.Fprintf(os.Stderr, "%s %s\n", red("❌"), msg)
}

// PrintJSON 以 JSON 格式输出到 stdout（机器友好模式）
func PrintJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		PrintError("JSON 序列化失败: " + err.Error())
		return
	}
	fmt.Println(string(data))
}

// PrintRaw 原样输出到 stdout
func PrintRaw(s string) {
	fmt.Print(s)
}

// PrintTable 以表格形式输出数据（使用标准库 text/tabwriter）
func PrintTable(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// 表头
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	// 分隔线
	seps := make([]string, len(headers))
	for i, h := range headers {
		seps[i] = strings.Repeat("─", max(len(h), 8))
	}
	fmt.Fprintln(w, strings.Join(seps, "\t"))
	// 数据行
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	w.Flush()
}

// PrintKeyValue 以 key-value 格式输出状态信息
func PrintKeyValue(pairs [][]string) {
	maxKeyLen := 0
	for _, p := range pairs {
		w := runeWidth(p[0])
		if w > maxKeyLen {
			maxKeyLen = w
		}
	}
	for _, p := range pairs {
		padding := maxKeyLen - runeWidth(p[0])
		if padding < 0 {
			padding = 0
		}
		fmt.Printf("  %s%s：%s\n", strings.Repeat(" ", padding), bold(p[0]), p[1])
	}
}

// FormatTimestamp 将毫秒时间戳格式化为可读字符串
func FormatTimestamp(ms int64) string {
	t := time.UnixMilli(ms)
	return t.Format("2006-01-02 15:04:05")
}

// MaskApikey 脱敏展示 apikey，只显示前 6 位和后 4 位
func MaskApikey(key string) string {
	if len(key) <= 10 {
		return key
	}
	return key[:6] + "****" + key[len(key)-4:]
}

func runeWidth(s string) int {
	n := 0
	for _, r := range s {
		if r > 0x7F {
			n += 2
		} else {
			n++
		}
	}
	return n
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

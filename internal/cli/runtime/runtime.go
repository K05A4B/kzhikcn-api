// Package runtime 为 CLI 子命令提供统一的 App 引导、输出格式化与交互式输入能力。
package runtime

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"kzhikcn/pkg/data"
	"kzhikcn/pkg/data/cache"
	"kzhikcn/pkg/log"
	"kzhikcn/server/app"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

// 命令分类，用于 cli help 中的分组展示。
const (
	CategoryBasic  = "基础命令"
	CategoryAdmin  = "管理员"
	CategoryConfig = "配置与数据"
)

// Format 输出格式。
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

type consoleOutputKey struct{}

var consoleOutputFlag consoleOutputKey

func WithLogOutput(ctx *cli.Context) *cli.Context {
	ctx.Context = context.WithValue(ctx.Context, consoleOutputFlag, true)
	return ctx
}

// FormatFlag 返回命令级输出格式选项，使 --format 可以跟在子命令之后使用。
func FormatFlag() cli.Flag {
	return &cli.StringFlag{
		Name:  "format",
		Usage: "输出格式 (text|json)",
		Value: string(FormatText),
	}
}

// OutputFormat 读取全局 --format 选项，非法值回退为 text。
func OutputFormat(ctx *cli.Context) Format {
	if strings.EqualFold(ctx.String("format"), string(FormatJSON)) {
		return FormatJSON
	}
	return FormatText
}

// PrintResult 按全局 --format 输出结果。
// JSON 模式下序列化 v；文本模式下调用 text，text 为 nil 时打印 v 本身。
func PrintResult(ctx *cli.Context, v any, text func(w io.Writer) error) error {
	w := ctx.App.Writer

	if OutputFormat(ctx) == FormatJSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		return enc.Encode(v)
	}

	if text == nil {
		_, err := fmt.Fprintln(w, v)
		return err
	}

	return text(w)
}

// Bootstrap 加载配置并初始化基础设施（数据库 / 缓存 / 存储）。
// createIfMissing 为 true 时，配置文件不存在则写入默认配置。
// 返回的 cleanup 可安全重复调用。
func Bootstrap(ctx *cli.Context, createIfMissing bool) (*app.App, func(), error) {
	a := app.New()
	configFile := ctx.String("config")

	consoleOutput := ctx.Context.Value(consoleOutputFlag) != nil

	var err error
	if createIfMissing {
		err = a.BootstrapOrCreate(configFile, consoleOutput)
	} else {
		err = a.Bootstrap(configFile, consoleOutput)
	}
	if err != nil {
		return nil, func() {}, err
	}

	if err := applyLogOptions(ctx); err != nil {
		return nil, func() {}, err
	}

	if err := a.Initialize(); err != nil {
		_ = cache.CloseCache()
		_ = data.CloseDB()
		return nil, func() {}, err
	}

	cleanup := func() {
		_ = cache.CloseCache()
		_ = data.CloseDB()
	}

	return a, cleanup, nil
}

// WithApp 完成配置加载、基础设施初始化与数据库迁移后执行 action，结束时释放资源。
func WithApp(ctx *cli.Context, action func(*app.App) error) error {
	a, cleanup, err := Bootstrap(ctx, false)
	if err != nil {
		return err
	}
	defer cleanup()

	if err := a.Migrate(); err != nil {
		return err
	}

	return action(a)
}

// applyLogOptions 在基础设施初始化前应用命令行日志选项。
// 日志输出目标已由 App.Bootstrap 依据配置建好，这里只覆盖级别。
func applyLogOptions(ctx *cli.Context) error {
	var level string
	if ctx.IsSet("log-level") {
		level = ctx.String("log-level")
	} else if ctx.Bool("quiet") {
		level = "error"
	}

	if level == "" {
		return nil
	}

	return log.SetLevel(level)
}

// PromptPassword 从终端隐藏读取并二次确认密码。非交互环境返回错误。
func PromptPassword(ctx *cli.Context, label string) (string, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return "", errors.New("当前环境不支持交互式密码输入，请使用 --password 选项")
	}

	fmt.Fprintf(ctx.App.Writer, "%s: ", label)
	first, err := term.ReadPassword(fd)
	fmt.Fprintln(ctx.App.Writer)
	if err != nil {
		return "", err
	}

	fmt.Fprintf(ctx.App.Writer, "请再次输入%s: ", label)
	second, err := term.ReadPassword(fd)
	fmt.Fprintln(ctx.App.Writer)
	if err != nil {
		return "", err
	}

	if string(first) != string(second) {
		return "", errors.New("两次输入的密码不一致")
	}

	return string(first), nil
}

// Confirm 请求用户确认；带 --yes 时直接返回 true。
// 非交互环境下未提供 --yes 会返回错误。
func Confirm(ctx *cli.Context, prompt string) (bool, error) {
	if ctx.Bool("yes") {
		return true, nil
	}

	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return false, errors.New("非交互环境下请使用 --yes 确认操作")
	}

	fmt.Fprintf(ctx.App.Writer, "%s [y/N]: ", prompt)

	answer, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}

	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes", nil
}

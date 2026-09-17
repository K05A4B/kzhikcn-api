package cli

import (
	"encoding/json"
	"fmt"
	"kzhikcn/internal/cli/runtime"
	pkgconfig "kzhikcn/pkg/config"
	"slices"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

const maskedValue = "******"

var ConfigCommands = &cli.Command{
	Name:     "config",
	Usage:    "配置相关命令",
	Category: runtime.CategoryConfig,
	Subcommands: []*cli.Command{
		{
			Name:   "check",
			Usage:  "校验配置文件",
			Action: checkConfig,
		},
		{
			Name:   "show",
			Usage:  "输出生效的配置（敏感字段已脱敏）",
			Flags:  []cli.Flag{runtime.FormatFlag()},
			Action: showConfig,
		},
	},
}

func checkConfig(ctx *cli.Context) error {
	file := ctx.String("config")

	conf, err := pkgconfig.LoadConfigFromFile(file)
	if err != nil {
		return errors.Wrapf(err, "配置文件 %s 解析失败", file)
	}

	problems := validateConfig(conf)
	if len(problems) > 0 {
		for _, p := range problems {
			fmt.Fprintf(ctx.App.Writer, "配置错误: %s\n", p)
		}
		return errors.Errorf("配置文件 %s 校验未通过，共 %d 项错误", file, len(problems))
	}

	fmt.Fprintf(ctx.App.Writer, "配置文件 %s 校验通过\n", file)
	return nil
}

func showConfig(ctx *cli.Context) error {
	conf, err := pkgconfig.LoadConfigFromFile(ctx.String("config"))
	if err != nil {
		return errors.Wrap(err, "加载配置失败")
	}

	out, err := yaml.Marshal(maskConfig(conf))
	if err != nil {
		return err
	}

	if runtime.OutputFormat(ctx) == runtime.FormatJSON {
		var generic any
		if err := yaml.Unmarshal(out, &generic); err != nil {
			return err
		}

		enc := json.NewEncoder(ctx.App.Writer)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		return enc.Encode(generic)
	}

	_, err = ctx.App.Writer.Write(out)
	return err
}

// maskConfig 返回配置副本，敏感字段以掩码替换。
func maskConfig(conf *pkgconfig.Config) *pkgconfig.Config {
	masked := *conf

	if masked.Auth.JWT.Secret != "" {
		masked.Auth.JWT.Secret = maskedValue
	}

	if masked.Cache.Redis.Password != "" {
		masked.Cache.Redis.Password = maskedValue
	}

	return &masked
}

// validateConfig 校验配置的必填项与取值合法性。
func validateConfig(conf *pkgconfig.Config) []string {
	var problems []string

	if conf.Database.Driver == "" {
		problems = append(problems, "db.driver 不能为空")
	} else if !slices.Contains([]string{"sqlite", "sqlite3", "mysql"}, conf.Database.Driver) {
		problems = append(problems, fmt.Sprintf("db.driver 不受支持: %s", conf.Database.Driver))
	}

	if conf.Database.Dsn == "" {
		problems = append(problems, "db.dsn 不能为空")
	}

	if conf.Storage.Provider == "" {
		problems = append(problems, "storage.provider 不能为空")
	}
	if conf.Storage.Articles.BasePath == "" {
		problems = append(problems, "storage.articles.base_path 不能为空")
	}

	switch strings.ToLower(conf.Cache.Provider) {
	case "":
		problems = append(problems, "cache.provider 不能为空")
	case "local":
		if conf.Cache.Local.Dir == "" {
			problems = append(problems, "cache.local.dir 不能为空")
		}
	case "redis":
		if conf.Cache.Redis.Addr == "" {
			problems = append(problems, "cache.redis.addr 不能为空")
		}
	default:
		problems = append(problems, fmt.Sprintf("cache.provider 不受支持: %s", conf.Cache.Provider))
	}

	switch {
	case conf.Auth.JWT.Secret == "":
		problems = append(problems, "auth.jwt.secret 不能为空")
	case strings.HasPrefix(conf.Auth.JWT.Secret, "${"):
		problems = append(problems, "auth.jwt.secret 存在未解析的环境变量占位符")
	}

	if conf.MRR.BaseUrl == "" {
		if conf.MRR.Rss.Enable || conf.MRR.Sitemap.Enable {
			problems = append(problems, "machine_readable_resources.base_url 不能为空 (RSS/Sitemap 已启用)")
		}
	} else if strings.HasPrefix(conf.MRR.BaseUrl, "${") {
		problems = append(problems, "machine_readable_resources.base_url 存在未解析的环境变量占位符")
	}

	if conf.Log.LogLevel != "" {
		if _, err := zerolog.ParseLevel(conf.Log.LogLevel); err != nil {
			problems = append(problems, fmt.Sprintf("log.log_level 非法: %s", conf.Log.LogLevel))
		}
	}

	return problems
}

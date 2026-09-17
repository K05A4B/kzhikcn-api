package cli

import (
	"bufio"
	"fmt"
	"io"
	"kzhikcn/pkg/assets"
	"kzhikcn/pkg/utils"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/valyala/fasttemplate"
)

func genConfig(ctx *cli.Context) error {
	file := ctx.String("config")

	if ctx.Bool("default") {
		if err := assets.ExportDefaultConfig(file); err != nil {
			return err
		}

		fmt.Fprintf(ctx.App.Writer, "已生成默认配置文件: %s\n", file)
		return nil
	}

	fp, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer fp.Close()

	keyToPrompt := map[string]string{
		"WEBSITE_URL":         "输入您的网站URL[例如: https://example.com]",
		"WEBSITE_NAME":        "输入您的网站名称",
		"WEBSITE_DESCRIPTION": "输入您的网站描述",
		"JWT_SECRET":          "输入您的JWT密钥(留空随机生成)",
	}

	stdin := bufio.NewReader(os.Stdin)

	_, err = fasttemplate.ExecuteFunc(assets.DefaultConfig, "${", "}", fp, func(w io.Writer, tag string) (int, error) {
		prompt, ok := keyToPrompt[tag]
		if !ok {
			return fmt.Fprintf(w, "${%s}", tag)
		}

		fmt.Fprintf(ctx.App.Writer, "%s: ", prompt)

		input, readErr := stdin.ReadString('\n')
		if readErr != nil && readErr != io.EOF {
			return 0, readErr
		}
		input = strings.TrimRight(input, "\r\n")

		if input == "" {
			if tag == "JWT_SECRET" {
				fmt.Fprintln(ctx.App.Writer, "已随机生成 JWT 密钥")
				return w.Write([]byte(utils.RandomString(32)))
			}
			return 0, nil
		}

		return w.Write([]byte(input))
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(ctx.App.Writer, "已生成配置文件: %s\n", file)
	return nil
}

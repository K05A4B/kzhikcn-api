package cli

import (
	"fmt"
	"io"
	"kzhikcn/pkg/assets"
	"kzhikcn/pkg/utils"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/valyala/fasttemplate"
)

func genConfig(ctx *cli.Context) error {
	file := ctx.String("config")

	fp, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}

	keyToPrompt := map[string]string{
		"WEBSITE_URL":         "输入您的网站URL[例如: https://example.com]",
		"WEBSITE_NAME":        "输入您的网站名称",
		"WEBSITE_DESCRIPTION": "输入您的网站描述",
		"JWT_SECRET":          "输入您的JWT密钥(留空随机生成)",
	}

	defer fp.Close()

	fasttemplate.ExecuteFunc(assets.DefaultConfig, "${", "}", fp, func(w io.Writer, tag string) (int, error) {
		prompt, ok := keyToPrompt[tag]
		if !ok {
			return fmt.Fprintf(w, "${%s}", tag)
		}

		fmt.Printf("%s: ", prompt)
		var input string
		fmt.Scanln(&input)

		if tag == "JWT_SECRET" && input == "" {
			return w.Write([]byte(utils.RandomString(32)))
		}

		if input == "" {
			return 0, nil
		}

		return w.Write([]byte(input))
	})
	return nil
}

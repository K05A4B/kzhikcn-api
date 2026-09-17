package main

import (
	"errors"
	"fmt"
	"kzhikcn/internal/cli"
	"os"

	urfave "github.com/urfave/cli/v2"
)

func main() {
	cmdline := cli.AppCli
	err := cmdline.Run(os.Args)

	if err == nil {
		return
	}

	fmt.Println(err)

	var exitCoder urfave.ExitCoder
	if errors.As(err, &exitCoder) {
		os.Exit(exitCoder.ExitCode())
	}

	os.Exit(1)
}

package main

import (
	"fmt"
	"kzhikcn/internal/cli"
	"os"
)

func main() {
	cmdline := cli.AppCli
	err := cmdline.Run(os.Args)

	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}

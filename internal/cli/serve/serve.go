package cmdserve

import (
	"kzhikcn/pkg/config"
	"kzhikcn/pkg/data/cache"
	"kzhikcn/server"

	"github.com/urfave/cli/v2"
)

func Serve() cli.ActionFunc {
	return func(ctx *cli.Context) error {
		signal := make(chan error)

		err := cache.InitCache(config.Conf())
		if err != nil {
			return err
		}

		go func() {
			err := server.Serve(ctx.String("address"))
			if err != nil {
				signal <- err
			}
		}()

		return <-signal
	}
}

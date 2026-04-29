package cliutils

import (
	"kzhikcn/pkg/assets"
	"kzhikcn/pkg/config"
	"kzhikcn/pkg/data"
	"os"

	"github.com/urfave/cli/v2"
)

func LoadConfig(action cli.ActionFunc) cli.ActionFunc {
	return func(ctx *cli.Context) error {

		configFile := ctx.String("config")
		_, err := config.LoadConfig(configFile)
		if os.IsNotExist(err) && !ctx.IsSet("config") {
			err = assets.ExportDefaultConfig(configFile)
			if err != nil {
				return err
			}

			_, err = config.LoadConfig(configFile)
			if err != nil {
				return err
			}

			return action(ctx)
		}
		if err != nil {
			return err
		}

		return action(ctx)
	}
}

func ConnectDatabase(action cli.ActionFunc) cli.ActionFunc {
	return LoadConfig(func(ctx *cli.Context) error {
		conf := config.Conf()

		dbDriver := conf.Database.Driver
		dsn := conf.Database.Dsn.String()

		err := data.ConnectDatabase(dbDriver, dsn)
		if err != nil {
			return err
		}

		if !data.ExistSchemaState() {
			err = data.DB().AutoMigrate(&data.SchemaState{})
			if err != nil {
				return err
			}

			_, err = data.InitDatabase()
			if err != nil {
				return err
			}

		} else {
			err = data.AutoMigrates()
			if err != nil {
				return err
			}
		}

		return action(ctx)
	})
}

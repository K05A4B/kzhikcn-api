package assets

import (
	_ "embed"
	"os"

	"github.com/pkg/errors"
)

//go:embed files/config.yml
var DefaultConfig string

func ExportDefaultConfig(file string) error {
	conf := DefaultConfig

	fp, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return errors.Errorf("Failed to create config file. Error: %s", err)
	}

	_, err = fp.WriteString(conf)
	return err
}

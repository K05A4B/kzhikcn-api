package assets

import (
	_ "embed"
	"os"

	"github.com/pkg/errors"
)

//go:embed files/config.yml
var DefaultConfig string

func ExportDefaultConfig(file string) error {
	fp, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return errors.Errorf("Failed to create config file. Error: %s", err)
	}
	defer fp.Close()

	if _, err = fp.WriteString(DefaultConfig); err != nil {
		return errors.Errorf("Failed to write config file. Error: %s", err)
	}

	return nil
}

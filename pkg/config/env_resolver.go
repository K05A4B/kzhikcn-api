package config

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/valyala/fasttemplate"
	"gopkg.in/yaml.v3"
)

func resolveEnv(values []*yaml.Node) error {
	for _, item := range values {
		if item.Kind == yaml.DocumentNode {
			resolveEnv(item.Content)

			continue
		}

		if item.Kind == yaml.MappingNode {
			children := []*yaml.Node{}

			for i := 0; i < len(item.Content); i += 2 {
				children = append(children, item.Content[i+1])

			}

			resolveEnv(children)
			continue
		}

		if item.Kind == yaml.SequenceNode {
			resolveEnv(item.Content)
			continue
		}

		if item.Kind != yaml.ScalarNode || item.Tag != "!!str" {
			continue
		}

		result := bytes.NewBufferString("")

		_, err := fasttemplate.ExecuteFunc(item.Value, "${", "}", result, func(w io.Writer, tag string) (int, error) {
			val := os.Getenv(tag)
			if val == "" {
				return fmt.Fprintf(w, "${%s}", tag)
			}

			return fmt.Fprint(w, val)
		})

		if err != nil {
			return err
		}

		item.Value = result.String()

	}

	return nil
}

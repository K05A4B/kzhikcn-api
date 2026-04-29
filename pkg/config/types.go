package config

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/valyala/fasttemplate"
	"gopkg.in/yaml.v3"
)

type Duration time.Duration

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	str := ""
	value.Decode(&str)

	dur, err := time.ParseDuration(str)
	if err != nil {
		return err
	}

	*d = Duration(dur)
	return nil
}

func (s Duration) Duration() time.Duration {
	return time.Duration(s)
}

type TemplateString string

func (e *TemplateString) Parse(transformer func(string) any) (string, error) {
	result := bytes.NewBuffer([]byte{})

	_, err := fasttemplate.ExecuteFunc(e.String(), "{{", "}}", result, func(w io.Writer, tag string) (int, error) {
		return fmt.Fprint(w, transformer(tag))
	})

	if err != nil {
		return "", err
	}

	return result.String(), nil
}

func (e *TemplateString) String() string {
	return string(*e)
}

type EnvVarResolver string

func (e *EnvVarResolver) UnmarshalYAML(value *yaml.Node) error {
	str := ""
	err := value.Decode(&str)
	if err != nil {
		return err
	}

	result := bytes.NewBuffer([]byte{})

	_, err = fasttemplate.ExecuteFunc(str, "${", "}", result, func(w io.Writer, tag string) (int, error) {
		return fmt.Fprint(w, os.Getenv(tag))
	})

	if err != nil {
		return err
	}

	*e = EnvVarResolver(result.String())
	return nil
}

func (e *EnvVarResolver) String() string {
	return string(*e)
}

type Size uint64

func (s *Size) UnmarshalYAML(value *yaml.Node) error {
	var str string
	if err := value.Decode(&str); err != nil {
		return err
	}

	rStr := []rune(str)
	numberPart := []rune{}
	unitPart := []rune{}

	for idx, char := range rStr {
		if char >= '0' && char <= '9' {
			numberPart = append(numberPart, char)
			continue
		}
		unitPart = rStr[idx:]
		break
	}

	if len(numberPart) == 0 {
		return errors.New("missing number in size string")
	}

	n, err := strconv.ParseInt(string(numberPart), 10, 64)
	if err != nil {
		return errors.Errorf("invalid number format: %q", numberPart)
	}

	units := map[string]int64{
		"bytes": 1, "byte": 1, "b": 1,
		"kb": 1024, "k": 1024,
		"mb": 1024 * 1024, "m": 1024 * 1024,
		"gb": 1024 * 1024 * 1024, "g": 1024 * 1024 * 1024,
	}

	unitStr := strings.ToLower(strings.TrimSpace(string(unitPart)))
	unit, ok := units[unitStr]
	if !ok {
		return errors.Errorf("unknown unit: %q", unitPart)
	}

	*s = Size(unit * n)
	return nil
}

func (s Size) Int64() int64 {
	return int64(s)
}

package config

import (
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type RateLimit struct {
	Window time.Duration
	Max    int
}

func (q *RateLimit) UnmarshalYAML(value *yaml.Node) error {
	str := ""
	err := value.Decode(&str)
	if err != nil {
		return err
	}

	unitMap := map[string]time.Duration{
		"ms": time.Millisecond,
		"s":  time.Second,
		"m":  time.Minute,
		"h":  time.Hour,
	}

	unit := "s"
	max := int64(0)

	parts := strings.Split(str, "/")
	if len(parts) <= 0 {
		return errors.Errorf("invalid value: empty string")
	}

	if len(parts) > 2 {
		return errors.Errorf("invalid value: %s", str)
	}

	if len(parts) >= 1 {
		max, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return err
		}
	}

	if len(parts) == 2 {
		unit = parts[1]
	}

	d, ok := unitMap[unit]
	if !ok {
		return errors.Errorf("invalid unit: %s", unit)
	}

	q.Window = d * time.Duration(max)
	q.Max = int(max)

	return nil
}

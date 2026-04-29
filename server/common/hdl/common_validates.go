package hdl

import (
	"net/http"
)

func When[T any](cond func(payload T) bool, err error) HandlerValidator[T] {
	return func(r *http.Request, meta map[string]any, payload T) error {
		if cond(payload) {
			return err
		}

		return nil
	}
}

func MissingFields[T any](f func(payload T) []string) HandlerValidator[T] {
	return func(r *http.Request, meta map[string]any, payload T) error {
		fields := f(payload)
		if len(fields) == 0 {
			return nil
		}

		meta["missingFields"] = fields

		return Error(400, "缺少必填字段", nil, "system.missing_required_fields")
	}
}

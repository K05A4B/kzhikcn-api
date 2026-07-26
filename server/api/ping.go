package api

import (
	"kzhikcn/internal/appinfo"
	"kzhikcn/pkg/hdl"
	"net/http"
	"time"
)

var PingHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	resp.Message = "pong"
	resp.Data = map[string]any{
		"timestamp":      time.Now().Unix(),
		"api_version":    "v1",
		"server_version": appinfo.Version,
		"ip":             r.RemoteAddr,
	}
	return nil
})

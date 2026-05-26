package api

import (
	"kzhikcn/server/common/hdl"
	"net/http"
	"time"
)

var PingHandler = hdl.NewSimpleHandler(func(r *http.Request, resp *hdl.Response) error {
	resp.Message = "pong"
	resp.Data = map[string]any{
		"timestamp":   time.Now().Unix(),
		"api_version": "v1",
	}
	return nil
})

package hdl

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

type Response struct {
	Message string         `json:"message"`
	Data    any            `json:"data"`
	Meta    map[string]any `json:"meta"`
}

type RawResponse struct {
	Success bool `json:"success"`
	Code    int  `json:"code"`
	*Response
	ErrorCode string `json:"errorCode,omitempty"`
	TraceID   string `json:"traceId,omitempty"`
}

func NewRawResponse() *RawResponse {
	resp := &Response{
		Meta: make(map[string]any),
	}

	rawResp := &RawResponse{
		Code:     200,
		Success:  true,
		Response: resp,
	}

	return rawResp
}

func ResponseJson(w http.ResponseWriter, r *http.Request, resp *RawResponse) error {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(resp.Code)

	enc := json.NewEncoder(w)
	err := enc.Encode(resp)

	if err != nil {
		err = errors.Wrap(err, "json marshal error")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"success\":false,\"code\":500,\"message\":\"序列化失败\",\"errorCode\":\"system.json_marshal_error\"}"))
		return err
	}

	return nil
}

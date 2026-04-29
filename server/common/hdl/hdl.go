package hdl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"kzhikcn/pkg/log"
	"kzhikcn/pkg/traceid"
	"kzhikcn/pkg/utils"
	"net/http"
	"strings"
)

type responseWriter struct{}

var respWriterKey = responseWriter{}

// 在不希望自动响应的地方返回这个错误 则不会自动响应
var NoRespond = noRespond{}

type noRespond struct{}

func (n noRespond) Error() string {
	return ""
}

func ResponseWriter(r *http.Request) http.ResponseWriter {
	ctx := r.Context()
	w := ctx.Value(respWriterKey)
	if w == nil {
		return nil
	}

	return w.(http.ResponseWriter)
}

func New[PT any](h Handler[PT]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		resp := NewRawResponse()

		ctx := r.Context()
		ctx = context.WithValue(ctx, respWriterKey, w)
		r = r.WithContext(ctx)

		defer func() { responseResult(err, resp, w, r) }()

		var payload PT

		if h.IsParsePayload() {
			payload, err = parsePayload[PT](w, r)
			if err != nil {
				err = Error(400, "请求体格式错误", err, "system.parse_payload_error").wrap(err, 1)
				return
			}
		}

		err = h.Validate(r, resp.Response, payload)
		if err != nil {
			return
		}

		err = h.Handler(r, resp.Response, payload)
		if err != nil {
			return
		}
	}
}

func responseResult(err error, resp *RawResponse, w http.ResponseWriter, r *http.Request) {
	if err == NoRespond {
		return
	}

	defer ResponseJson(w, r, resp)

	ctx := r.Context()
	traceID := traceid.GetTraceID(ctx)

	if err == nil {
		return
	}

	resp.Success = false
	resp.TraceID = traceID.String()
	resp.Message = err.Error()

	handleErr, ok := err.(*HandlerError)
	if !ok {
		resp.Code = 500
		return
	}

	if handleErr.statusCode != 0 {
		resp.Code = handleErr.statusCode
	}

	logger := log.WithTraceID(traceID.String())

	if strings.TrimSpace(handleErr.errorCode) != "" {
		resp.ErrorCode = handleErr.errorCode
		logger = logger.With("errCode", resp.ErrorCode)
	}

	if handleErr.internal == nil {
		return
	}

	for key, value := range handleErr.fields {
		logger = logger.With(key, value)
	}

	logger.
		With("caller", fmt.Sprintf("%s:%d", handleErr.file, handleErr.line)).
		Error(handleErr.internal)
}

func parsePayload[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	reader := http.MaxBytesReader(
		w,
		r.Body,
		1024*1024*15,
	)

	var reqData T

	dec := json.NewDecoder(reader)
	err := dec.Decode(&reqData)

	return reqData, err
}

// 接管自动响应，响应reader的数据
// 如果contentType为空字符串则自动检测mime
func WriteRaw(r *http.Request, reader io.Reader, contentType string) error {
	var buf []byte = nil
	if utils.IsEmptyString(contentType) {
		buf = make([]byte, 512)
		n, err := reader.Read(buf)
		if err != nil {
			return Error(500, "响应原始数据失败", err, "system.write_raw_failed").wrap(err, 3)
		}

		buf = buf[:n]

		contentType = http.DetectContentType(buf)
	}

	w := ResponseWriter(r)
	w.Header().Set("Content-Type", contentType)

	w.Write(buf)
	_, err := io.Copy(w, reader)
	if err != nil {
		log.Error("[hdl.WriteRaw] write raw failed: ", err.Error())
		return NoRespond
	}

	return NoRespond
}

// 接管自动响应，响应 data 的数据
func WriteRawData(r *http.Request, data []byte, contentType string) error {
	return WriteRaw(r, bytes.NewBuffer(data), contentType)
}

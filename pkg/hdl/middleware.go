package hdl

import "net/http"

func Middleware(handle func(w http.ResponseWriter, r *http.Request, meta map[string]any) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := NewRawResponse()

		err := handle(w, r, resp.Meta)
		if err != nil {
			responseResult(err, resp, w, r)
			return
		}
	})
}

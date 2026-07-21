package hdl

import "net/http"

type HandlerFunc[T any] func(r *http.Request, resp *Response, payload T) error
type SimpleHandlerFunc func(r *http.Request, resp *Response) error

type HandlerValidator[T any] func(r *http.Request, meta map[string]any, payload T) error

type Handler[T any] interface {
	Validate(r *http.Request, resp *Response, payload T) error
	Handler(r *http.Request, resp *Response, payload T) error
	IsParsePayload() bool
}

type handler[PT any] struct {
	hdl            HandlerFunc[PT]
	validators     []HandlerValidator[PT]
	isParsePayload bool
}

func (h *handler[T]) IsParsePayload() bool {
	return h.isParsePayload
}

func (h *handler[T]) Handler(r *http.Request, resp *Response, payload T) error {
	return h.hdl(r, resp, payload)
}

func (h *handler[T]) Validate(r *http.Request, resp *Response, payload T) error {
	for _, f := range h.validators {
		err := f(r, resp.Meta, payload)
		if err != nil {
			return err
		}
	}

	return nil
}

func newHandler[T any](hdlFunc HandlerFunc[T], isParsePayload bool, validators ...HandlerValidator[T]) Handler[T] {
	h := handler[T]{}
	h.isParsePayload = isParsePayload
	h.hdl = hdlFunc
	h.validators = validators
	return &h
}

func NewHandler[T any](hdlFunc HandlerFunc[T], validators ...HandlerValidator[T]) Handler[T] {
	return newHandler(hdlFunc, true, validators...)
}

func NewSimpleExHandler(hdlFunc SimpleHandlerFunc, validators ...HandlerValidator[any]) Handler[any] {
	return newHandler(simpleHandlerAdapter(hdlFunc), false, validators...)
}

func NewSimpleHandler(hdlFunc SimpleHandlerFunc) Handler[any] {
	return newHandler(simpleHandlerAdapter(hdlFunc), false)
}

func simpleHandlerAdapter(f SimpleHandlerFunc) HandlerFunc[any] {
	return func(r *http.Request, resp *Response, _ any) error { return f(r, resp) }
}

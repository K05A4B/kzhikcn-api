package hdl

import (
	"runtime"
)

type HandlerError struct {
	msg        string
	internal   error
	statusCode int

	fields map[string]any

	errorCode string

	file string
	line int
}

func (j *HandlerError) With(key string, val any) *HandlerError {
	if j.fields == nil {
		j.fields = make(map[string]any)
	}

	fields := cloneFields(j.fields)

	fields[key] = val

	return &HandlerError{
		msg:        j.msg,
		statusCode: j.statusCode,
		errorCode:  j.errorCode,
		fields:     cloneFields(j.fields),
		internal:   j.internal,
		file:       j.file,
		line:       j.line,
	}
}

func (j *HandlerError) Error() string {
	return j.msg
}

func (j *HandlerError) wrap(internalErr error, skip int) *HandlerError {
	_, file, line, _ := runtime.Caller(skip)

	return &HandlerError{
		msg:        j.msg,
		statusCode: j.statusCode,
		errorCode:  j.errorCode,
		fields:     cloneFields(j.fields),
		internal:   internalErr,
		file:       file,
		line:       line,
	}
}

func (j *HandlerError) Wrap(internalErr error) *HandlerError {
	return j.wrap(internalErr, 2)
}

func Error(statusCode int, msg string, internalErr error, errCode string) *HandlerError {
	return (&HandlerError{
		statusCode: statusCode,
		errorCode:  errCode,
		msg:        msg,
	}).wrap(internalErr, 3)
}

func DefineError(statusCode int, msg string, errCode string) *HandlerError {
	return &HandlerError{
		statusCode: statusCode,
		errorCode:  errCode,
		msg:        msg,
	}
}

func cloneFields(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

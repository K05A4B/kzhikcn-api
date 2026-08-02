package traceid

import "context"

type traceIdCtx struct{}

var traceIdCtxKey = traceIdCtx{}

func WithTraceID(ctx context.Context, id TraceID) context.Context {
	return context.WithValue(ctx, traceIdCtxKey, id)
}

func GetTraceID(ctx context.Context) TraceID {
	traceId := ctx.Value(traceIdCtxKey)
	if traceId != nil {
		id, ok := traceId.(TraceID)
		if ok {
			return id
		}
	}

	return nil
}

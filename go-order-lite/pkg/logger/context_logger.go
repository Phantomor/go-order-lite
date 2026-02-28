package logger

import (
	"context"

	"go.uber.org/zap"
)

const requestIDKey = "request_id"

// 从 context 里取 request_id
func WithContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return Log
	}

	rid, ok := ctx.Value(requestIDKey).(string)
	if !ok || rid == "" {
		return Log
	}

	return Log.With(zap.String("request_id", rid))
}

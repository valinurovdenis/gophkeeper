// Package logger for logging incoming requests.
package logger

import (
	"context"
	"time"
	"unsafe"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Initialize zap logger.
var Log *zap.Logger = zap.NewNop()

// Initializes logging.
// Returns error if level cant be determined.
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl
	return nil
}

// Interceptor for logging incoming requests and reponses.
// Writes processing metrics such as method, path, status, size, duration of request.
func RequestLoggerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)
		respErr := ""
		if err != nil {
			respErr = err.Error()
		}
		Log.Info("got incoming GRPC request",
			zap.String("path", info.FullMethod),
			zap.String("error", respErr),
			zap.Int("size", int(unsafe.Sizeof(resp))),
			zap.Duration("duration", duration),
		)
		return resp, err
	}
}

// Same RequestLoggerInterceptor for streaming.
func RequestStreamLoggerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		err := handler(srv, stream)
		duration := time.Since(start)
		respErr := ""
		if err != nil {
			respErr = err.Error()
		}
		Log.Info("got incoming GRPC request",
			zap.String("path", info.FullMethod),
			zap.String("error", respErr),
			zap.Duration("duration", duration),
		)
		return err
	}
}

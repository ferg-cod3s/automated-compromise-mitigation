// Package logging provides gRPC middleware for request tracing and logging.
package logging

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns a gRPC unary server interceptor that:
// - Extracts or generates request IDs
// - Logs request start and completion
// - Adds request ID to context
// - Logs errors and panics
func UnaryServerInterceptor(logger *Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Get or generate request ID
		ctx, requestID := GetOrGenerateRequestID(ctx)

		// Add request ID to outgoing metadata so clients can see it
		if err := grpc.SetHeader(ctx, metadata.Pairs(RequestIDMetadataKey, requestID)); err != nil {
			logger.Warn("failed to set request ID header", "error", err)
		}

		// Create request-scoped logger
		reqLogger := logger.WithContext(ctx)

		// Log request start
		reqLogger.Debug("gRPC request started",
			"method", info.FullMethod,
			"request_id", requestID,
		)

		// Handle panics
		defer func() {
			if r := recover(); r != nil {
				reqLogger.Error("gRPC request panicked",
					"method", info.FullMethod,
					"panic", r,
					"duration_ms", time.Since(start).Milliseconds(),
				)
				panic(r) // Re-panic after logging
			}
		}()

		// Call the handler
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		// Log completion
		if err != nil {
			st, _ := status.FromError(err)
			reqLogger.Error("gRPC request failed",
				"method", info.FullMethod,
				"status_code", st.Code().String(),
				"error", err.Error(),
				"duration_ms", duration.Milliseconds(),
			)
		} else {
			level := "info"
			// Warn on slow requests (>500ms for unary calls)
			if duration > 500*time.Millisecond {
				level = "warn"
			}

			if level == "warn" {
				reqLogger.Warn("gRPC request completed (slow)",
					"method", info.FullMethod,
					"duration_ms", duration.Milliseconds(),
				)
			} else {
				reqLogger.Info("gRPC request completed",
					"method", info.FullMethod,
					"duration_ms", duration.Milliseconds(),
				)
			}
		}

		return resp, err
	}
}

// StreamServerInterceptor returns a gRPC stream server interceptor that:
// - Extracts or generates request IDs
// - Logs stream start and completion
// - Adds request ID to context
func StreamServerInterceptor(logger *Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()
		ctx := ss.Context()

		// Get or generate request ID
		ctx, requestID := GetOrGenerateRequestID(ctx)

		// Add request ID to outgoing metadata
		if err := ss.SetHeader(metadata.Pairs(RequestIDMetadataKey, requestID)); err != nil {
			logger.Warn("failed to set request ID header for stream", "error", err)
		}

		// Create request-scoped logger
		reqLogger := logger.WithContext(ctx)

		// Log stream start
		reqLogger.Debug("gRPC stream started",
			"method", info.FullMethod,
			"request_id", requestID,
			"is_client_stream", info.IsClientStream,
			"is_server_stream", info.IsServerStream,
		)

		// Wrap the server stream to inject context with request ID
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		// Handle panics
		defer func() {
			if r := recover(); r != nil {
				reqLogger.Error("gRPC stream panicked",
					"method", info.FullMethod,
					"panic", r,
					"duration_ms", time.Since(start).Milliseconds(),
				)
				panic(r)
			}
		}()

		// Call the handler
		err := handler(srv, wrappedStream)
		duration := time.Since(start)

		// Log completion
		if err != nil {
			st, _ := status.FromError(err)
			reqLogger.Error("gRPC stream failed",
				"method", info.FullMethod,
				"status_code", st.Code().String(),
				"error", err.Error(),
				"duration_ms", duration.Milliseconds(),
			)
		} else {
			reqLogger.Info("gRPC stream completed",
				"method", info.FullMethod,
				"duration_ms", duration.Milliseconds(),
			)
		}

		return err
	}
}

// wrappedServerStream wraps a grpc.ServerStream to inject a custom context.
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the wrapped context with request ID.
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// UnaryClientInterceptor returns a gRPC unary client interceptor that:
// - Propagates request IDs to outbound calls
// - Logs outbound requests
func UnaryClientInterceptor(logger *Logger) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		start := time.Now()

		// Propagate request ID if present in context
		requestID := getRequestIDFromContext(ctx)
		if requestID != "" {
			ctx = InjectRequestID(ctx, requestID)
		}

		// Create logger with context
		reqLogger := logger.WithContext(ctx)

		reqLogger.Debug("gRPC client call started",
			"method", method,
			"target", cc.Target(),
		)

		// Call the invoker
		err := invoker(ctx, method, req, reply, cc, opts...)
		duration := time.Since(start)

		// Log completion
		if err != nil {
			st, _ := status.FromError(err)
			reqLogger.Error("gRPC client call failed",
				"method", method,
				"target", cc.Target(),
				"status_code", st.Code().String(),
				"error", err.Error(),
				"duration_ms", duration.Milliseconds(),
			)
		} else {
			reqLogger.Debug("gRPC client call completed",
				"method", method,
				"target", cc.Target(),
				"duration_ms", duration.Milliseconds(),
			)
		}

		return err
	}
}

// StreamClientInterceptor returns a gRPC stream client interceptor that:
// - Propagates request IDs to outbound streams
// - Logs outbound stream calls
func StreamClientInterceptor(logger *Logger) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		start := time.Now()

		// Propagate request ID if present in context
		requestID := getRequestIDFromContext(ctx)
		if requestID != "" {
			ctx = InjectRequestID(ctx, requestID)
		}

		// Create logger with context
		reqLogger := logger.WithContext(ctx)

		reqLogger.Debug("gRPC client stream started",
			"method", method,
			"target", cc.Target(),
			"is_client_stream", desc.ClientStreams,
			"is_server_stream", desc.ServerStreams,
		)

		// Call the streamer
		cs, err := streamer(ctx, desc, cc, method, opts...)
		duration := time.Since(start)

		// Log result
		if err != nil {
			st, _ := status.FromError(err)
			reqLogger.Error("gRPC client stream failed",
				"method", method,
				"target", cc.Target(),
				"status_code", st.Code().String(),
				"error", err.Error(),
				"duration_ms", duration.Milliseconds(),
			)
		} else {
			reqLogger.Debug("gRPC client stream established",
				"method", method,
				"target", cc.Target(),
				"duration_ms", duration.Milliseconds(),
			)
		}

		return cs, err
	}
}

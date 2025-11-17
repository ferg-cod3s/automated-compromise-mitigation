package logging

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestUnaryServerInterceptor(t *testing.T) {
	logger := NewLogger("test")

	tests := []struct {
		name           string
		ctx            context.Context
		handler        grpc.UnaryHandler
		expectError    bool
		validateResult func(t *testing.T, ctx context.Context, resp interface{}, err error)
	}{
		{
			name: "successful request",
			ctx:  context.Background(),
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				// Verify request ID is in context
				requestID := getRequestIDFromContext(ctx)
				if requestID == "" {
					t.Error("Request ID not in context")
				}
				return "success", nil
			},
			expectError: false,
			validateResult: func(t *testing.T, ctx context.Context, resp interface{}, err error) {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if resp != "success" {
					t.Errorf("Expected 'success', got %v", resp)
				}
			},
		},
		{
			name: "request with error",
			ctx:  context.Background(),
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, errors.New("test error")
			},
			expectError: true,
			validateResult: func(t *testing.T, ctx context.Context, resp interface{}, err error) {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if err.Error() != "test error" {
					t.Errorf("Expected 'test error', got %v", err)
				}
			},
		},
		{
			name: "request with existing request ID",
			ctx: metadata.NewIncomingContext(
				context.Background(),
				metadata.Pairs(RequestIDMetadataKey, "existing-id"),
			),
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				requestID := getRequestIDFromContext(ctx)
				if requestID != "existing-id" {
					t.Errorf("Expected 'existing-id', got %q", requestID)
				}
				return "ok", nil
			},
			expectError: false,
			validateResult: func(t *testing.T, ctx context.Context, resp interface{}, err error) {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := UnaryServerInterceptor(logger)

			info := &grpc.UnaryServerInfo{
				FullMethod: "/test.Service/Method",
			}

			resp, err := interceptor(tt.ctx, nil, info, tt.handler)
			tt.validateResult(t, tt.ctx, resp, err)
		})
	}
}

func TestStreamServerInterceptor(t *testing.T) {
	logger := NewLogger("test")

	tests := []struct {
		name        string
		ctx         context.Context
		handler     grpc.StreamHandler
		expectError bool
	}{
		{
			name: "successful stream",
			ctx:  context.Background(),
			handler: func(srv interface{}, stream grpc.ServerStream) error {
				// Verify request ID is in context
				ctx := stream.Context()
				requestID := getRequestIDFromContext(ctx)
				if requestID == "" {
					t.Error("Request ID not in stream context")
				}
				return nil
			},
			expectError: false,
		},
		{
			name: "stream with error",
			ctx:  context.Background(),
			handler: func(srv interface{}, stream grpc.ServerStream) error {
				return errors.New("stream error")
			},
			expectError: true,
		},
		{
			name: "stream with existing request ID",
			ctx: metadata.NewIncomingContext(
				context.Background(),
				metadata.Pairs(RequestIDMetadataKey, "stream-id"),
			),
			handler: func(srv interface{}, stream grpc.ServerStream) error {
				ctx := stream.Context()
				requestID := getRequestIDFromContext(ctx)
				if requestID != "stream-id" {
					t.Errorf("Expected 'stream-id', got %q", requestID)
				}
				return nil
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := StreamServerInterceptor(logger)

			info := &grpc.StreamServerInfo{
				FullMethod:     "/test.Service/StreamMethod",
				IsClientStream: true,
				IsServerStream: true,
			}

			mockStream := &mockServerStream{ctx: tt.ctx}
			err := interceptor(nil, mockStream, info, tt.handler)

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestUnaryClientInterceptor(t *testing.T) {
	logger := NewLogger("test")

	tests := []struct {
		name        string
		ctx         context.Context
		invoker     grpc.UnaryInvoker
		expectError bool
		validateCtx func(t *testing.T, ctx context.Context)
	}{
		{
			name: "successful client call",
			ctx:  context.Background(),
			invoker: func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
				return nil
			},
			expectError: false,
			validateCtx: func(t *testing.T, ctx context.Context) {},
		},
		{
			name: "client call with error",
			ctx:  context.Background(),
			invoker: func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
				return errors.New("client error")
			},
			expectError: true,
			validateCtx: func(t *testing.T, ctx context.Context) {},
		},
		{
			name: "propagate request ID",
			ctx:  SetRequestIDInContext(context.Background(), "propagated-id"),
			invoker: func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
				// Verify request ID is in outgoing metadata
				md, ok := metadata.FromOutgoingContext(ctx)
				if !ok {
					t.Error("No outgoing metadata")
					return nil
				}
				values := md.Get(RequestIDMetadataKey)
				if len(values) == 0 || values[0] != "propagated-id" {
					t.Errorf("Request ID not propagated, got %v", values)
				}
				return nil
			},
			expectError: false,
			validateCtx: func(t *testing.T, ctx context.Context) {
				requestID := getRequestIDFromContext(ctx)
				if requestID != "propagated-id" {
					t.Errorf("Expected 'propagated-id', got %q", requestID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := UnaryClientInterceptor(logger)

			err := interceptor(
				tt.ctx,
				"/test.Service/Method",
				nil,
				nil,
				&grpc.ClientConn{},
				tt.invoker,
			)

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			tt.validateCtx(t, tt.ctx)
		})
	}
}

func TestStreamClientInterceptor(t *testing.T) {
	logger := NewLogger("test")

	tests := []struct {
		name        string
		ctx         context.Context
		streamer    grpc.Streamer
		expectError bool
	}{
		{
			name: "successful stream client call",
			ctx:  context.Background(),
			streamer: func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
				return &mockClientStream{}, nil
			},
			expectError: false,
		},
		{
			name: "stream client call with error",
			ctx:  context.Background(),
			streamer: func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
				return nil, errors.New("stream error")
			},
			expectError: true,
		},
		{
			name: "propagate request ID in stream",
			ctx:  SetRequestIDInContext(context.Background(), "stream-propagated-id"),
			streamer: func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
				md, ok := metadata.FromOutgoingContext(ctx)
				if !ok {
					t.Error("No outgoing metadata in stream")
					return &mockClientStream{}, nil
				}
				values := md.Get(RequestIDMetadataKey)
				if len(values) == 0 || values[0] != "stream-propagated-id" {
					t.Errorf("Request ID not propagated in stream, got %v", values)
				}
				return &mockClientStream{}, nil
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := StreamClientInterceptor(logger)

			desc := &grpc.StreamDesc{
				StreamName:    "TestStream",
				ClientStreams: true,
				ServerStreams: true,
			}

			_, err := interceptor(
				tt.ctx,
				desc,
				&grpc.ClientConn{},
				"/test.Service/StreamMethod",
				tt.streamer,
			)

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestWrappedServerStream(t *testing.T) {
	ctx := SetRequestIDInContext(context.Background(), "wrapped-id")
	mockStream := &mockServerStream{ctx: context.Background()}

	wrapped := &wrappedServerStream{
		ServerStream: mockStream,
		ctx:          ctx,
	}

	// Verify context is overridden
	resultCtx := wrapped.Context()
	requestID := getRequestIDFromContext(resultCtx)
	if requestID != "wrapped-id" {
		t.Errorf("Expected 'wrapped-id', got %q", requestID)
	}
}

// Mock implementations

type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}

func (m *mockServerStream) SetHeader(md metadata.MD) error {
	return nil
}

func (m *mockServerStream) SendHeader(md metadata.MD) error {
	return nil
}

func (m *mockServerStream) SetTrailer(md metadata.MD) {}

type mockClientStream struct {
	grpc.ClientStream
}

func (m *mockClientStream) Header() (metadata.MD, error) {
	return metadata.MD{}, nil
}

func (m *mockClientStream) Trailer() metadata.MD {
	return metadata.MD{}
}

func (m *mockClientStream) CloseSend() error {
	return nil
}

func (m *mockClientStream) Context() context.Context {
	return context.Background()
}

func (m *mockClientStream) SendMsg(msg interface{}) error {
	return nil
}

func (m *mockClientStream) RecvMsg(msg interface{}) error {
	return nil
}

func TestRequestIDPropagationEndToEnd(t *testing.T) {
	logger := NewLogger("test")

	// Simulate server receiving request with request ID
	incomingCtx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs(RequestIDMetadataKey, "e2e-test-id"),
	)

	serverInterceptor := UnaryServerInterceptor(logger)
	serverInfo := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	// Server handler that makes outbound call
	serverHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		requestID := getRequestIDFromContext(ctx)
		if requestID != "e2e-test-id" {
			t.Errorf("Server context missing request ID, got %q", requestID)
		}

		// Simulate outbound client call with same context
		clientInterceptor := UnaryClientInterceptor(logger)
		clientInvoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			// Verify request ID propagated to outbound call
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				t.Error("Outbound call missing metadata")
				return nil
			}
			values := md.Get(RequestIDMetadataKey)
			if len(values) == 0 || values[0] != "e2e-test-id" {
				t.Errorf("Request ID not propagated to client call, got %v", values)
			}
			return nil
		}

		return nil, clientInterceptor(ctx, "/downstream.Service/Method", nil, nil, &grpc.ClientConn{}, clientInvoker)
	}

	_, err := serverInterceptor(incomingCtx, nil, serverInfo, serverHandler)
	if err != nil {
		t.Errorf("End-to-end test failed: %v", err)
	}
}

func BenchmarkUnaryServerInterceptor(b *testing.B) {
	logger := NewLogger("bench")
	interceptor := UnaryServerInterceptor(logger)
	ctx := context.Background()
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = interceptor(ctx, nil, info, handler)
	}
}

func BenchmarkGenerateRequestID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GenerateRequestID()
	}
}

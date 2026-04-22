// Copyright 2026 Verda Cloud Oy
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grpc

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type fakeServerTransportStream struct {
	headers metadata.MD
}

func (f *fakeServerTransportStream) Method() string { return "/test.Service/Method" }
func (f *fakeServerTransportStream) SetHeader(md metadata.MD) error {
	f.headers = metadata.Join(f.headers, md)
	return nil
}
func (f *fakeServerTransportStream) SendHeader(md metadata.MD) error { return nil }
func (f *fakeServerTransportStream) SetTrailer(md metadata.MD) error { return nil }

func ctxWithSpan(ctx context.Context) (context.Context, trace.SpanContext) {
	traceID := trace.TraceID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	spanID := trace.SpanID{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID, SpanID: spanID, TraceFlags: trace.FlagsSampled, Remote: true,
	})
	return trace.ContextWithRemoteSpanContext(ctx, sc), sc
}

func TestObservability_InjectTraceIDOnly(t *testing.T) {
	interceptor := Observability(WithTraceInjection(InjectTraceIDOnly))

	ctx, _ := ctxWithSpan(context.Background())
	stream := &fakeServerTransportStream{}
	ctx = grpc.NewContextWithServerTransportStream(ctx, stream)
	ctx = metadata.NewIncomingContext(ctx, metadata.MD{})

	handler := func(ctx context.Context, req any) (any, error) {
		return "response", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}
	resp, err := interceptor(ctx, "request", info, handler)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "response" {
		t.Fatalf("expected 'response', got %v", resp)
	}

	traceIDVal := stream.headers.Get(TraceIDHeaderKey)
	if len(traceIDVal) == 0 {
		t.Fatal("expected X-Trace-Id in response headers")
	}
}

func TestObservability_InjectW3C(t *testing.T) {
	interceptor := Observability(WithTraceInjection(InjectW3CTraceContext))

	ctx, _ := ctxWithSpan(context.Background())
	stream := &fakeServerTransportStream{}
	ctx = grpc.NewContextWithServerTransportStream(ctx, stream)
	ctx = metadata.NewIncomingContext(ctx, metadata.MD{})

	handler := func(ctx context.Context, req any) (any, error) { return "ok", nil }
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	_, err := interceptor(ctx, "request", info, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	traceparent := stream.headers.Get(TraceParentHeaderKey)
	if len(traceparent) == 0 {
		t.Fatal("expected traceparent in response headers")
	}
}

func TestObservability_InjectBoth(t *testing.T) {
	interceptor := Observability(WithTraceInjection(InjectBoth))

	ctx, _ := ctxWithSpan(context.Background())
	stream := &fakeServerTransportStream{}
	ctx = grpc.NewContextWithServerTransportStream(ctx, stream)
	ctx = metadata.NewIncomingContext(ctx, metadata.MD{})

	handler := func(ctx context.Context, req any) (any, error) { return "ok", nil }
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	_, err := interceptor(ctx, "request", info, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stream.headers.Get(TraceParentHeaderKey)) == 0 {
		t.Fatal("expected traceparent header")
	}
	if len(stream.headers.Get(TraceIDHeaderKey)) == 0 {
		t.Fatal("expected X-Trace-Id header")
	}
}

func TestObservability_InjectNone(t *testing.T) {
	interceptor := Observability(WithTraceInjection(InjectNone))

	ctx, _ := ctxWithSpan(context.Background())
	stream := &fakeServerTransportStream{}
	ctx = grpc.NewContextWithServerTransportStream(ctx, stream)
	ctx = metadata.NewIncomingContext(ctx, metadata.MD{})

	handler := func(ctx context.Context, req any) (any, error) { return "ok", nil }
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	_, err := interceptor(ctx, "request", info, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stream.headers.Get(TraceParentHeaderKey)) != 0 {
		t.Fatal("expected no traceparent for InjectNone")
	}
	if len(stream.headers.Get(TraceIDHeaderKey)) != 0 {
		t.Fatal("expected no X-Trace-Id for InjectNone")
	}
}

func TestObservability_CustomTraceHeader(t *testing.T) {
	interceptor := Observability(WithTraceInjection(InjectTraceIDOnly), WithCustomTraceHeader("X-My-Trace"))

	ctx, _ := ctxWithSpan(context.Background())
	stream := &fakeServerTransportStream{}
	ctx = grpc.NewContextWithServerTransportStream(ctx, stream)
	ctx = metadata.NewIncomingContext(ctx, metadata.MD{})

	handler := func(ctx context.Context, req any) (any, error) { return "ok", nil }
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	_, err := interceptor(ctx, "request", info, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	customVal := stream.headers.Get("X-My-Trace")
	if len(customVal) == 0 {
		t.Fatal("expected custom trace header X-My-Trace")
	}
}

func TestObservability_HandlerError(t *testing.T) {
	interceptor := Observability()

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{})
	stream := &fakeServerTransportStream{}
	ctx = grpc.NewContextWithServerTransportStream(ctx, stream)

	handler := func(ctx context.Context, req any) (any, error) {
		return nil, context.DeadlineExceeded
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	_, err := interceptor(ctx, "request", info, handler)
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func TestConvenienceConstructors(t *testing.T) {
	if ObservabilityWithW3CTraceContext() == nil {
		t.Fatal("ObservabilityWithW3CTraceContext returned nil")
	}
	if ObservabilityWithTraceID() == nil {
		t.Fatal("ObservabilityWithTraceID returned nil")
	}
	if ObservabilityWithCustomHeader("X-Test") == nil {
		t.Fatal("ObservabilityWithCustomHeader returned nil")
	}
}

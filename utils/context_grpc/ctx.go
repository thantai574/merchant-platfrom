package context_grpc

import (
	"context"
	"github.com/micro/go-micro/v2/metadata"
	"orders-system/utils/helpers"
	"time"

	"github.com/micro/go-micro/v2/debug/trace"
)

type CtxGrpc struct {
	context.Context
	TraceId string
	*trace.Span
}

func (this *CtxGrpc) StartSpan(ctx context.Context, name string) {
	this.Span = new(trace.Span)
	trace_id, span_id, _ := trace.FromContext(ctx)

	span := &trace.Span{
		Name:    name,
		Id:      "Span" + time.Now().Format("20060102150405"),
		Started: helpers.GetCurrentTime(),
	}

	span.Parent = span.Id
	span.Trace = "Trace" + time.Now().Format("20060102150405")

	if trace_id != "" {
		span.Trace = trace_id
	}
	if span_id != "" {
		span.Parent = span_id
	}

	this.Span = span
	this.Context = trace.ToContext(ctx, span.Trace, span.Parent)

}

func (this *CtxGrpc) Duration() {
	this.Span.Duration = time.Duration(this.Started.Unix() - time.Now().Unix())
}

func (c *CtxGrpc) SetOrderId(orderID string) {
	c.Context = metadata.Set(c.Context, "Micro-Order-Id", orderID)
}

func NewOrderSystemContextGRPC(ctx context.Context) *CtxGrpc {
	return &CtxGrpc{
		ctx,
		"",
		nil,
	}
}

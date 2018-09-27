package oczap // import "github.com/census-ecosystem/opencensus-experiments/go/oczap"

import (
	"context"
	"encoding/base64"

	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// IncludeTagKeys is a list of tag keys to include in log lines.
// By default it includes ocgrpc and ochttp keys.
var IncludeTagKeys = []tag.Key{
	ocgrpc.KeyServerMethod,
}

// Logger returns a new
func Logger(ctx context.Context, l *zap.Logger) *zap.Logger {
	var fields []zapcore.Field
	if span := trace.FromContext(ctx); span != nil {
		sc := span.SpanContext()
		tid := base64.StdEncoding.EncodeToString(sc.TraceID[:])
		sid := base64.StdEncoding.EncodeToString(sc.SpanID[:])
		fields = append(fields, zap.String("trace_id", tid), zap.String("span_id", sid))
	}
	if tags := tag.FromContext(ctx); tags != nil {
		for _, k := range IncludeTagKeys {
			if v, ok := tags.Value(k); ok {
				fields = append(fields, zap.String(k.Name(), v))
			}
		}
	}
	if len(fields) == 0 {
		return l
	}
	return l.WithOptions(zap.Fields(fields...))
}

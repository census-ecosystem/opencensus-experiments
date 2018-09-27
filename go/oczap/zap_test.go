package oczap_test

import (
	"context"

	"github.com/census-ecosystem/opencensus-experiments/go/oczap"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
)

var keyUserID, _ = tag.NewKey("user_id")

func init() {
	oczap.IncludeTagKeys = append(oczap.IncludeTagKeys, keyUserID)
}

func ExampleLogger() {
	l := zap.NewExample()
	ctx := context.Background()
	ctx, span := trace.StartSpan(ctx, "TestSpan")
	defer span.End()
	ctx, _ = tag.New(ctx, tag.Upsert(keyUserID, "1234"))
	l = oczap.Logger(ctx, l)
	l.Info("test")
}

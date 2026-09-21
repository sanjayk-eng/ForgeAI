package logger

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewZapLogsStructuredFields(t *testing.T) {
	var output bytes.Buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&output),
		zap.InfoLevel,
	)

	log := NewZap(zap.New(core))
	log.With("request_id", "req-123").Info(context.Background(), "tool executed", "tool", "git_status")

	message := output.String()
	for _, field := range []string{`"msg":"tool executed"`, `"request_id":"req-123"`, `"tool":"git_status"`} {
		if !strings.Contains(message, field) {
			t.Fatalf("log output missing %s: %s", field, message)
		}
	}
}

func TestNewZapNilReturnsNil(t *testing.T) {
	if NewZap(nil) != nil {
		t.Fatal("expected nil logger")
	}
}

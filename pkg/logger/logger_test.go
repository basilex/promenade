package logger

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(Config{Level: "info", Format: "text", Output: buf})
	assert.NotNil(t, l)
	l.Info("test")
	assert.Contains(t, buf.String(), "test")
}

func TestInit(t *testing.T) {
	buf := &bytes.Buffer{}
	Init(Config{Level: "info", Format: "text", Output: buf})
	Info("test")
	assert.Contains(t, buf.String(), "test")
}

func TestContext(t *testing.T) {
	buf := &bytes.Buffer{}
	Init(Config{Level: "info", Format: "text", Output: buf})
	ctx := context.WithValue(context.Background(), RequestIDKey, "req123")
	l := FromContext(ctx)
	l.Info("ctx test")
	assert.Contains(t, buf.String(), "req123")
}

func TestLevels(t *testing.T) {
	buf := &bytes.Buffer{}
	Init(Config{Level: "debug", Format: "text", Output: buf})
	Debug("debug")
	assert.Contains(t, buf.String(), "debug")
	buf.Reset()
	Warn("warn")
	assert.Contains(t, buf.String(), "warn")
	buf.Reset()
	Error("error")
	assert.Contains(t, buf.String(), "error")
}

func TestJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(Config{Level: "info", Format: "json", Output: buf})
	l.Info("json")
	assert.Contains(t, buf.String(), `"msg":"json"`)
}

func TestWithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(Config{Level: "info", Format: "text", Output: buf})
	l.WithFields(map[string]any{"k": "v"}).Info("fields")
	assert.Contains(t, buf.String(), "k")
}

func TestDefault(t *testing.T) {
	l := Default()
	assert.NotNil(t, l)
}

func TestContextFuncs(t *testing.T) {
	buf := &bytes.Buffer{}
	Init(Config{Level: "debug", Format: "text", Output: buf})
	ctx := context.WithValue(context.Background(), UserIDKey, "user1")

	DebugContext(ctx, "debug")
	assert.Contains(t, buf.String(), "user1")

	buf.Reset()
	InfoContext(ctx, "info")
	assert.Contains(t, buf.String(), "info")

	buf.Reset()
	WarnContext(ctx, "warn")
	assert.Contains(t, buf.String(), "warn")

	buf.Reset()
	ErrorContext(ctx, "error")
	assert.Contains(t, buf.String(), "error")
}

func TestWithContext(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(Config{Level: "info", Format: "text", Output: buf})
	ctx := context.WithValue(context.Background(), TraceIDKey, "trace1")
	l2 := l.WithContext(ctx)
	l2.Info("trace")
	assert.Contains(t, buf.String(), "trace1")
}

func TestAddSource(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New(Config{Level: "info", Format: "text", Output: buf, AddSource: true})
	l.Info("source")
	assert.Contains(t, buf.String(), "source")
}

func TestMultiContext(t *testing.T) {
	buf := &bytes.Buffer{}
	Init(Config{Level: "info", Format: "text", Output: buf})
	ctx := context.WithValue(context.Background(), RequestIDKey, "r1")
	ctx = context.WithValue(ctx, UserIDKey, "u1")
	ctx = context.WithValue(ctx, TraceIDKey, "t1")
	l := FromContext(ctx)
	l.Info("multi")
	output := buf.String()
	assert.Contains(t, output, "r1")
	assert.Contains(t, output, "u1")
	assert.Contains(t, output, "t1")
}

func TestLevelParsing(t *testing.T) {
	tests := []string{"debug", "info", "warn", "error", "unknown"}
	for _, level := range tests {
		buf := &bytes.Buffer{}
		l := New(Config{Level: level, Format: "text", Output: buf})
		assert.NotNil(t, l)
	}
}

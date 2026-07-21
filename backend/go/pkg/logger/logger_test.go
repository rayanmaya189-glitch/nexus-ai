package logger_test

import (
	"testing"

	"github.com/aeroxe/nexus-backend/pkg/logger"
)

func TestNew(t *testing.T) {
	l := logger.New("test")
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
	defer l.Sync()
}

func TestLogger_Info(t *testing.T) {
	l := logger.New("test")
	defer l.Sync()
	l.Info("test message", "key", "value")
}

func TestLogger_Error(t *testing.T) {
	l := logger.New("test")
	defer l.Sync()
	l.Error("test error", "key", "value")
}

func TestLogger_Debug(t *testing.T) {
	l := logger.New("test")
	defer l.Sync()
	l.Debug("test debug", "key", "value")
}

func TestLogger_WithFields(t *testing.T) {
	l := logger.New("test")
	defer l.Sync()
	l.WithFields("field1", "value1").Info("with fields")
}

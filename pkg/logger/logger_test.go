package logger

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
)

func TestLoggerConstructionDoesNotChangeGlobalLevel(t *testing.T) {
	before := zerolog.GlobalLevel()
	NewLoggerConfig(LoggerConfig{Level: "warning", IsDev: "development"})
	if zerolog.GlobalLevel() != before {
		t.Fatal("logger changed process-wide log level")
	}
}
func TestPrettyWriterReportsInputByteCount(t *testing.T) {
	var output bytes.Buffer
	writer := &PrettyJSONWriter{Writer: &output}
	input := []byte(`{"message":"ok"}`)
	n, err := writer.Write(input)
	if err != nil || n != len(input) {
		t.Fatalf("Write returned %d, %v", n, err)
	}
}

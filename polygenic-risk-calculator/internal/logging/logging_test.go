package logging

import (
	"os"
	"testing"
	"phite.io/polygenic-risk-calculator/internal/config"
)

func TestLogging_RespectsLogLevelFromConfig(t *testing.T) {
	resetLogger()
	config.ResetForTest()
	f, err := os.CreateTemp("", "phite-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(`{"log_level": "DEBUG"}`)
	f.Close()
	config.SetConfigPath(f.Name())
	Debug("debug message should appear")
	Info("info message should appear")
	Warn("warn message should appear")
	Error("error message should appear")
}

func TestLogging_DefaultsToInfoIfMissing(t *testing.T) {
	resetLogger()
	config.ResetForTest()
	config.SetConfigPath("/tmp/nonexistent.json")
	Debug("debug message should NOT appear")
	Info("info message should appear")
}

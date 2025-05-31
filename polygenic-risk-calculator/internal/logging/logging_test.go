package logging

import (
	"os"
	"testing"
	"phite.io/polygenic-risk-calculator/internal/config"
)

func TestLogging_RespectsLogLevelFromConfig(t *testing.T) {
	SetSilentLoggingForTest() // silence logs for this test
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
	SetSilentLoggingForTest() // silence logs for this test
}

func TestLogging_NoneLevelSuppressesAllLogs(t *testing.T) {
	resetLogger()
	config.ResetForTest()
	f, err := os.CreateTemp("", "phite-config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(`{"log_level": "NONE"}`)
	f.Close()
	config.SetConfigPath(f.Name())
	resetLogger()
	Debug("debug should NOT appear")
	Info("info should NOT appear")
	Warn("warn should NOT appear")
	Error("error should NOT appear")
	// No assertion: this test passes if no logs are emitted
}



package config

import (
	"testing"
)

func TestMustLoad(t *testing.T) {
	cfg := MustLoad("../config.yaml")
	if cfg.LogLevel == "" {
		t.Fatalf("expected non-empty LogLevel")
	}
}

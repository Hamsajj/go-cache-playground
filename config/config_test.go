package config

import (
	"os"
	"testing"
)

func TestNewWithName(t *testing.T) {
	_ = os.Setenv("TEST_SERVICE_DEBUG", "true")
	_ = os.Setenv("PORT", "80")
	_ = os.Setenv("HOST", "localhost")
	_ = os.Setenv("TTL_SECONDS", "100")

	conf, err := NewWithName("test_service")
	if err != nil {
		t.Fatalf("error loading conf %v", err)
	}

	if !conf.Debug {
		t.Errorf("expected confg.Debug to be true")
	}
	if conf.Port != "80" {
		t.Errorf("expected conf.Port to equal %s, got %s", "8080", conf.Port)
	}
	if conf.Host != "localhost" {
		t.Errorf("expected conf.Host to equal %s, got %s", "localhost", conf.Host)
	}

	if conf.TTLSec != 100 {
		t.Errorf("expected conf.TTLSec to equal %d, got %d", 100, conf.TTLSec)
	}
}

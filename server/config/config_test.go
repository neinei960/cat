package config

import "testing"

func TestServerAddressDefaultsToAllInterfacesWhenHostEmpty(t *testing.T) {
	cfg := ServerConfig{Port: 8080}
	if got := cfg.Address(); got != ":8080" {
		t.Fatalf("expected :8080, got %q", got)
	}
}

func TestServerAddressUsesConfiguredHost(t *testing.T) {
	cfg := ServerConfig{Host: "127.0.0.1", Port: 8080}
	if got := cfg.Address(); got != "127.0.0.1:8080" {
		t.Fatalf("expected 127.0.0.1:8080, got %q", got)
	}
}

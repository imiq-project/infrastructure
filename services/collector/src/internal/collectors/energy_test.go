package collectors

import (
	"testing"
)

// 1. Verify the collector name is correct.
func TestEnergyCollector_Name(t *testing.T) {
	collectors, err := NewEnergyCollector()
	if err != nil {
		t.Fatalf("Failed to create EnergyCollector: %v", err)
	}

	expectedName := "Building"
	if collectors.Name() != expectedName {
		t.Errorf("Expected collector name '%s', got '%s'", expectedName, collectors.Name())
	}
}

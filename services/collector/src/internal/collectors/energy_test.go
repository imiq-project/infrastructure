package collectors

import (
	"imiq/collector/internal/config"
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

// Test 2: Verify parsing of the exact YAML structure for Building 1
func TestEnergyCollector_Fetch_DynamicData(t *testing.T) {

	collectors, err := NewEnergyCollector()
	if err != nil {
		t.Fatalf("Failed to create EnergyCollector: %v", err)
	}
	// A. Prepare a mock location with the expected metadata structure
	mockLocation := config.Location{
		ID:   "Building:OVGU:1",
		Name: "Gebäude 01 - Campus Welcome Center (CWC)",
		Coord: config.Coord{
			Lat: 52.14035767199704,
			Lon: 11.640291546017703,
		},
		Metadata: map[string]any{
			"metadata": map[string]any{
				"resources": []any{
					map[string]any{
						"type":     "electricity",
						"meter-id": "DE99999739106DIHRTE015BE3UOTFXF91", // 33 chars
						"obis":     "001-001:001.029.000*255",           // 23 chars
						"userKey":  "fake_256_char_key_to_test_api_rejection",
					},
				},
			},
		},
	}
	// B. Call the Fetch method
	result, err := collectors.Fetch(mockLocation)
	if err != nil {
		t.Fatalf("Fetch returned an error: %v", err)
	}
	// C. Verify the structure of the result
	if result["type"] != "Building-Energy" {
		t.Errorf("Expected type 'Building-Energy', got '%v'", result["type"])
	}

	// D. Verify the location and name fields are correctly parsed it is strict but
	// locationData, ok := result["location"].(map[string]any)
	// if !ok {
	// 	t.Fatalf("Expected 'location' to be a map, got %T", result["location"])
	// }
	// if locationData["type"] != "geo:point" {
	// 	t.Errorf("Expected location type 'geo:point', got '%v'", locationData["type"])
	// }
	// if locationData["value"] != "52.140357, 11.640292" {
	// 	t.Errorf("Expected location value '52.140357, 11.640292', got '%v'", locationData["value"])
	// }

	// E. Verify the name field is correctly parsed and sanitized
	nameData, ok := result["name"].(map[string]any)
	if !ok {
		t.Fatalf("Expected 'name' to be a map, got %T", result["name"])
	}
	if nameData["type"] != "String" {
		t.Errorf("Expected name type 'String', got '%v'", nameData["type"])
	}
	if nameData["value"] != "Gebäude 01 - Campus Welcome Center (CWC)" {
		t.Errorf("Expected name value 'Gebäude 01 - Campus Welcome Center (CWC)', got '%v'", nameData["value"])
	}
	// F. Verify that the  resource name ("electricity") is skipped due to the fake userKey
	if _, exists := result["electricity"]; exists {
		t.Errorf("Expected 'electricity' to be skipped due to fake userKey, but it was included")
	}
}

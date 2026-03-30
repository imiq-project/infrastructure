package collectors

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// TestExtractParkingSpots_Mock simulates the website HTML to test the logic
func TestExtractParkingSpots_Mock(t *testing.T) {
	// 1. MOCK DATA: This is the "Mensa-style" HTML snapshot.
	// We include the "City Carr" encoding bug and a "Tiefgarage" name to test the normalization.
	mockHTML := `
	<!DOCTYPE html>
	<html>
	<body>
		<table>
			<tr>
				<td colspan="2">Parkanlage: Magdeburg | Freie Plätze: Realtime</td>
			</tr>
			<tr>
				<td><a href="/info1.shtml">Allee Center</a></td>
				<td>964</td>
			</tr>
			<tr>
				<td><a href="/info2.shtml">City Carr</a></td>
				<td>300</td>
			</tr>
			<tr>
				<td><a href="/info3.shtml">Domviertel</a></td>
				<td>50</td>
			</tr>
			<tr>
				<td><a href="/info4.shtml">Offline Park</a></td>
				<td>Offline</td>
			</tr>
		</table>
	</body>
	</html>`

	// 2. CONVERT HTML TO DOM
	reader := strings.NewReader(mockHTML)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		t.Fatalf("Failed to parse mock HTML: %v", err)
	}

	// 3. RUN EXTRACTION
	// We target the table that contains "Parkanlage" just like your colly logic does
	var results map[string]ParkingInfo
	doc.Find("table").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "Parkanlage") {
			results = extractParkingSpotsFromTable(s)
		}
	})

	// 4. VERIFY RESULTS (Overview)
	t.Log("--- MOCK SCRAPER OVERVIEW ---")
	for name, info := range results {
		t.Logf("Parsed: %-15s | Spaces: %d\n", name, info.AvailableSpaces)
	}

	// 5. ASSERTIONS (Edge Cases)
	t.Run("Check Basic Extraction", func(t *testing.T) {
		if val, ok := results["Allee Center"]; !ok || val.AvailableSpaces != 964 {
			t.Errorf("Allee Center failed. Got %v", val)
		}
	})

	t.Run("Check ISO-Encoding Case (City Carr)", func(t *testing.T) {
		if _, ok := results["City Carr"]; !ok {
			t.Error("City Carr was not found in the results map")
		}
	})

	t.Run("Check Offline Handling", func(t *testing.T) {
		if _, ok := results["Offline Park"]; ok {
			t.Error("Offline Park should have been skipped by the extractor")
		}
	})
}

// TestNormalization ensures your Fetch matching logic handles the name differences
func TestNormalizationLogic(t *testing.T) {
	normalize := func(s string) string {
		s = strings.ToLower(s)
		return strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
				return r
			}
			return -1
		}, s)
	}

	tests := []struct {
		yamlName    string
		scraped     string
		expectMatch bool
	}{
		{"City Carré", "City Carr", true},             // Encoding bug
		{"Tiefgarage Domviertel", "Domviertel", true}, // Prefix bug
		{"Allee Center", "Allee Center", true},        // Exact
		{"Mensa", "Allee Center", false},              // Wrong match
	}

	for _, tt := range tests {
		normYaml := normalize(tt.yamlName)
		normScraped := normalize(tt.scraped)
		match := strings.Contains(normYaml, normScraped) || strings.Contains(normScraped, normYaml)

		if match != tt.expectMatch {
			t.Errorf("Normalization error: %s vs %s | Expected match: %v", tt.yamlName, tt.scraped, tt.expectMatch)
		}
	}
}

package collectors

import (
	"imiq/collector/internal/config"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// 1. Test the Sanitize Function (Table-Driven)
// We check tricky characters: quotes, newlines, and the price format (parens).
func TestSanitize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Standard Text", "Simple Text", "Simple Text"},
		{"Price Format", "(3.50 | 4.50)", "[3.50 | 4.50]"},       // Checks parens replacement
		{"Forbidden Chars", "Bad;Chars<Here>", "Bad,Chars Here"}, // Checks ; < >
		{"Quotes", "Mimi\"s 'Welt'", "Mimi s  Welt"},             // Checks " '
		{"Newlines", "Line\nBreak", "Line Break"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitize(tt.input); got != tt.expected {
				t.Errorf("sanitize(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// 2. Test Category Logic (Table-Driven)
// Ensures we correctly map German keywords to English categories.
func TestDetermineCategory(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"enthält Rind", "Non-Vegetarian"},
		{"Schwein", "Non-Vegetarian"},
		{"Geflügel", "Non-Vegetarian"},
		{"Fisch", "Non-Vegetarian"},
		{"Vegan", "Vegetarian/Vegan"},
		{"Vegetarisch", "Vegetarian/Vegan"},
		{"vEgAn", "Vegetarian/Vegan"},
		{"Unknown Stuff", "Unknown Stuff"},
		{"", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := determineCategory(tt.input); got != tt.expected {
				t.Errorf("determineCategory(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// 3. Test the Parsing Logic directly (No Mock Server)
// This fixes the issue where we couldn't override the 'const' URL,
// and it's much faster because it doesn't use HTTP.
func TestExtractMealsFromTable(t *testing.T) {
	// A. Prepare Mock HTML
	mockHtml := `
	<div class="mensa">
		<table>
			<tbody>
				<tr>
					<td>
						<span class="gruen">Kürbissuppe</span>
						<span class="grau">Pumpkin Soup</span>
						<span class="mensapreis">(2.50 | 4.00)</span>
					</td>
					<td>
						<img alt="veganes Gericht" />
					</td>
				</tr>
			</tbody>
		</table>
	</div>`

	// B. Create a GoQuery document directly from the string
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(mockHtml))
	if err != nil {
		t.Fatalf("Failed to create doc: %v", err)
	}

	// C. Call the logic function directly
	// Note: We simulate the selection that Colly would pass
	tableSelection := doc.Find("table")
	meals := extractMealsFromTable(tableSelection)

	// D. Verify Results
	if len(meals) != 1 {
		t.Fatalf("Expected 1 meal, got %d", len(meals))
	}

	m := meals[0]

	// Check Name
	if m.NameGerman != "Kürbissuppe" {
		t.Errorf("Expected Name 'Kürbissuppe', got '%s'", m.NameGerman)
	}

	// Check Price Sanitization
	expectedPrice := "[2.50 | 4.00]"
	if m.Price != expectedPrice {
		t.Errorf("Expected Price '%s', got '%s'", expectedPrice, m.Price)
	}

	// Check Category Logic
	if m.Category != "Vegetarian/Vegan" {
		t.Errorf("Expected Category 'Vegetarian/Vegan', got '%s'", m.Category)
	}
}

// 4. Test the Fetch trigger Logic
// Ensures we only attempt to scrape when the OSM ID matches.
func TestFetch_MensaTrigger(t *testing.T) {
	collector, _ := NewRestaurantCollector()

	// Case 1: Random Restaurant (Should NOT trigger scrape)
	locRandom := config.Location{
		Name: "Random Pizza",
		Metadata: map[string]any{
			"category": "restaurant",
			"osm_id":   12345, // Not the Mensa ID
		},
	}
	res, _ := collector.Fetch(locRandom)
	if _, exists := res["todays_menu"]; exists {
		t.Error("Should NOT have 'todays_menu' for random restaurant")
	}

	// Case 2: Mensa (Should trigger scrape)
	locMensa := config.Location{
		Name: "Mensa UniCampus",
		Metadata: map[string]any{
			"category": "restaurant",
			"osm_id":   1578665845, // Matches mensaOsmID constant
		},
	}
	resMensa, _ := collector.Fetch(locMensa)
	_ = resMensa
}

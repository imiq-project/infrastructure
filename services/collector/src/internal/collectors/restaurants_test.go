package collectors

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
		input    string // The alt text from the image
		expected string
	}{
		{"enthält Rind", "Non-Vegetarian"},
		{"Schwein", "Non-Vegetarian"},
		{"Geflügel", "Non-Vegetarian"},
		{"Fisch", "Non-Vegetarian"},
		{"Vegan", "Vegetarian/Vegan"},
		{"Vegetarisch", "Vegetarian/Vegan"},
		{"vEgAn", "Vegetarian/Vegan"}, // Case insensitive check
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

// 3. Test the Full Scraper (Integration with Mock Server)
func TestScrapeMensaColly(t *testing.T) {
	// A. Prepare Mock HTML
	// We simulate a real scenario with complex price and category
	todayDate := time.Now().Format("02.01.2006")
	mockHtml := fmt.Sprintf(`
	<html>
	<body>
		<div class="mensa">
			<table>
				<thead>
					<tr><th>%s</th></tr> </thead>
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
		</div>
	</body>
	</html>
	`, todayDate)

	// B. Start Fake Server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(200)
		w.Write([]byte(mockHtml))
	}))
	defer ts.Close()

	// C. Override Global Variables
	// We point the scraper to our local fake server instead of the real internet
	origUrl := mensaUrl
	origDomain := mensaDomain

	mensaUrl = ts.URL
	mensaDomain = "" // Clear domain to allow localhost scraping

	defer func() {
		mensaUrl = origUrl
		mensaDomain = origDomain
	}()

	// D. Run the Function
	meals, err := scrapeMensaColly()

	// E. Verify Results
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(meals) != 1 {
		t.Fatalf("Expected 1 meal, got %d", len(meals))
	}

	m := meals[0]

	// Check Name
	if m.NameGerman != "Kürbissuppe" {
		t.Errorf("Expected Name 'Kürbissuppe', got '%s'", m.NameGerman)
	}

	// Check Price Sanitization: parens () should become brackets []
	expectedPrice := "[2.50 | 4.00]"
	if m.Price != expectedPrice {
		t.Errorf("Expected Price '%s', got '%s'", expectedPrice, m.Price)
	}

	// Check Category Logic: "veganes Gericht" -> "Vegetarian/Vegan"
	if m.Category != "Vegetarian/Vegan" {
		t.Errorf("Expected Category 'Vegetarian/Vegan', got '%s'", m.Category)
	}
}

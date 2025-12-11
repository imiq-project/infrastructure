package collectors

import (
	"fmt"
	"imiq/collector/internal/config"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

type RestaurantCollector struct{}

type MenuFileItem struct {
	Name     string `json:"name"`
	Schedule []struct {
		DateString string `json:"date"`
		Meals      []any  `json:"meals"`
	} `json:"schedule"`
}

type Meal struct {
	NameGerman  string `json:"name_german"`
	NameEnglish string `json:"name_english"`
	Price       string `json:"price"`
	Category    string `json:"category"`
}

func (collector RestaurantCollector) Fetch(loc config.Location) (map[string]any, error) {

	placeType := "Place"
	if cat, ok := loc.Metadata["category"].(string); ok && cat != "" {
		placeType = formatType(cat)
	}
	response := map[string]any{
		"type": placeType,
		"name": map[string]any{
			"value": sanitize(loc.Name),
			"type":  "String",
		},
		"location": map[string]any{
			"type":  "geo:point",
			"value": fmt.Sprintf("%f, %f", loc.Coord.Lat, loc.Coord.Lon),
		},
	}
	for key, value := range loc.Metadata {
		// Check if value is a string to sanitize
		safeValue := value
		if strVal, ok := value.(string); ok {
			safeValue = sanitize(strVal)
		}

		response[key] = map[string]any{
			"value": safeValue,
			"type":  "String",
		}
	}

	if placeType == "Mensa" || strings.Contains(strings.ToLower(loc.Name), "mensa") {
		if strings.Contains(strings.ToLower(loc.Name), "unicampus") {
			menu, err := scrapeMensaColly()
			if err == nil && len(menu) > 0 {
				response["todays_menu"] = map[string]any{
					"value": menu,
					"type":  "StructuredValue",
				}
			} else if err != nil {
				fmt.Printf("Error scraping mensa: %v\n", err)
			}
		}
	}

	return response, nil
}

var mensaUrl = "https://www.studentenwerk-magdeburg.de/mensen-cafeterien/mensa-unicampus-speiseplan-unten/"
var mensaDomain = "www.studentenwerk-magdeburg.de"

func scrapeMensaColly() ([]Meal, error) {
	var todaysMeals []Meal
	today := time.Now().Format("02.01.2006") // Format: DD.MM.YYYY

	// 1. Initialize the Collector
	c := colly.NewCollector(
	// Optional: Add a timeout or User-Agent here if needed
	)
	if mensaDomain != "" {
		c.AllowedDomains = []string{mensaDomain}
	}

	// 2. Define the Callback
	// "Whenever you find a table inside div.mensa, do this:"
	c.OnHTML("div.mensa table", func(e *colly.HTMLElement) {
		// e.DOM gives us the goquery Selection for this specific table
		dateHeader := strings.TrimSpace(e.DOM.Find("thead").Text())

		// Check if this table belongs to Today
		if strings.Contains(dateHeader, today) {

			// Iterate over the rows (tr) in this table
			e.DOM.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
				cols := row.Find("td")
				if cols.Length() < 2 {
					return
				}

				// Extract Data (using the same logic as before)
				col0 := cols.Eq(0)
				nameGer := strings.TrimSpace(col0.Find("span.gruen").Text())
				if nameGer == "" {
					nameGer = strings.TrimSpace(col0.Text())
				}

				nameEng := strings.TrimSpace(col0.Find("span.grau").Text())
				if nameEng == "" {
					nameEng = "No English name"
				}

				price := strings.TrimSpace(col0.Find("span.mensapreis").Text())

				col1 := cols.Eq(1)
				catRaw, _ := col1.Find("img").Attr("alt")
				category := determineCategory(catRaw)

				// Append to our result list
				todaysMeals = append(todaysMeals, Meal{
					NameGerman:  sanitize(nameGer),
					NameEnglish: sanitize(nameEng),
					Price:       sanitize(price),
					Category:    sanitize(category),
				})
			})
		}
	})

	// 3. Handle Errors
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	// 4. Visit the Page (This blocks until done)
	err := c.Visit(mensaUrl)
	if err != nil {
		return nil, err
	}

	return todaysMeals, nil
}

func determineCategory(catText string) string {
	lower := strings.ToLower(catText)
	if strings.Contains(lower, "vegetarisch") || strings.Contains(lower, "vegan") {
		return "Vegetarian/Vegan"
	}
	nonVegKeywords := []string{"rind", "schwein", "geflügel", "fisch", "hähnchen", "lamm", "suppe"}
	for _, kw := range nonVegKeywords {
		if strings.Contains(lower, kw) {
			return "Non-Vegetarian"
		}
	}
	if catText == "" {
		return "Unknown"
	}
	return catText
}

func formatType(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0]) // Capitalizes first letter
	return string(r)
}

func sanitize(s string) string {
	// Orion Forbidden chars: < > " ' = ; ( )
	r := strings.NewReplacer(
		";", ",",
		"(", "[",
		")", "]",
		"<", " ",
		">", " ",
		"=", ":",
		"\"", " ",
		"'", " ",
		"´", " ",
		"\n", " ",
		"\t", " ",
	)
	return strings.TrimSpace(r.Replace(s))
}

func (collector RestaurantCollector) Name() string {
	return "Restaurant"
}

func NewRestaurantCollector() (config.Collector, error) {
	return RestaurantCollector{}, nil
}

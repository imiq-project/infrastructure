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

const (
	mensaUrl = "https://www.studentenwerk-magdeburg.de/mensen-cafeterien/mensa-unicampus-speiseplan-unten/"
	//mensaDomain = "www.studentenwerk-magdeburg.de"
	mensaOsmID = 1578665845
)

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

	var currentOsmID int64
	if val, ok := loc.Metadata["osm_id"]; ok {
		switch v := val.(type) {
		case int:
			currentOsmID = int64(v)
		case int64:
			currentOsmID = v
		case float64:
			currentOsmID = int64(v)
		}
	}

	if currentOsmID == mensaOsmID {
		menu, err := scrapeMensaColly()
		if err == nil && len(menu) > 0 {
			//fmt.Println("---------------------------------------------------")
			//fmt.Printf("✅ SUCCESS! Fetched %d meals for Mensa:\n", len(menu))
			//for _, m := range menu {
			//	fmt.Printf(" - %s (%s)\n", m.NameGerman, m.Category)
			//}
			//fmt.Println("---------------------------------------------------")
			response["todays_menu"] = map[string]any{
				"value": menu,
				"type":  "StructuredValue",
			}
		} else if err != nil {
			fmt.Printf("Error scraping mensa: %v\n", err)
		}
	}

	return response, nil
}

func scrapeMensaColly() ([]Meal, error) {
	var todaysMeals []Meal
	today := time.Now().Format("02.01.2006")

	c := colly.NewCollector()

	c.OnHTML("div.mensa table", func(e *colly.HTMLElement) {
		dateHeader := strings.TrimSpace(e.DOM.Find("thead").Text())
		if strings.Contains(dateHeader, today) {
			// Pass the goquery selection (e.DOM) to our logic function
			todaysMeals = extractMealsFromTable(e.DOM)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	err := c.Visit(mensaUrl)
	if err != nil {
		return nil, err
	}

	return todaysMeals, nil
}

func extractMealsFromTable(table *goquery.Selection) []Meal {
	var meals []Meal

	table.Find("tbody tr").Each(func(_ int, row *goquery.Selection) {
		cols := row.Find("td")
		if cols.Length() < 2 {
			return
		}

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

		meals = append(meals, Meal{
			NameGerman:  sanitize(nameGer),
			NameEnglish: sanitize(nameEng),
			Price:       sanitize(price),
			Category:    sanitize(category),
		})
	})

	return meals
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

package collectors

import (
	"fmt"
	"imiq/collector/internal/config"
	"unicode"
)

type RestaurantCollector struct{}

func (collector RestaurantCollector) Fetch(loc config.Location) (map[string]any, error) {
	placeType := "Place"
	if cat, ok := loc.Metadata["category"].(string); ok && cat != "" {
		placeType = formatType(cat)
	}
	response := map[string]any{
		"type": placeType,
		"name": loc.Name, // Takes the name from YAML
		"location": map[string]any{
			"type":  "geo:point",
			"value": fmt.Sprintf("%f, %f", loc.Coord.Lat, loc.Coord.Lon),
		},
	}
	for key, value := range loc.Metadata {
		response[key] = value
	}
	return response, nil
}

func formatType(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0]) // Capitalizes first letter
	return string(r)
}

func (collector RestaurantCollector) Name() string {
	return "Restaurants"
}

func NewRestaurantCollector() (config.Collector, error) {
	return RestaurantCollector{}, nil
}

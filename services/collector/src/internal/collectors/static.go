package collectors

import (
	"fmt"
	"imiq/collector/internal/config"
)

type StaticCollector struct{}

func (collector StaticCollector) Fetch(loc config.Location) (map[string]any, error) {
	entity := map[string]any{
		"type": loc.Metadata["type"],
		"name": map[string]any{
			"type":  "string",
			"value": loc.Name,
		},
		"location": map[string]any{
			"type":  "geo:point",
			"value": fmt.Sprintf("%f, %f", loc.Coord.Lat, loc.Coord.Lon),
		},
	}
	for key, value := range loc.Metadata {
		entity[key] = value
	}
	return entity, nil
}

func (collector StaticCollector) Name() string {
	return "Static"
}

func NewStaticCollector() (config.Collector, error) {
	return StaticCollector{}, nil
}

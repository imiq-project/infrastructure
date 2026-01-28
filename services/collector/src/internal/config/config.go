package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Coord struct {
	Lat float64
	Lon float64
}

type Location struct {
	ID       string
	Name     string
	Coord    Coord
	Metadata map[string]any // map to hold additional fields
}

type Collector interface {
	Name() string
	Fetch(Location) (map[string]any, error)
}

type CollectorConfig struct {
	Collector Collector
	Intervall time.Duration
	Locations []Location
}

type Config struct {
	Collectors []CollectorConfig
}

func ReadConfig(path string, availableCollectors map[string]Collector) (*Config, error) {

	type RawLocation struct {
		ID       string         `yaml:"id"`
		Name     string         `yaml:"name"`
		Lat      float64        `yaml:"lat"`
		Lon      float64        `yaml:"lon"`
		Metadata map[string]any `yaml:",inline"` // captures all other fields
	}

	type RawCollectorConfig struct {
		Name      string        `yaml:"name"`
		Interval  time.Duration `yaml:"interval"`
		Locations []RawLocation `yaml:"locations"`
	}

	type RawConfig struct {
		Collectors []RawCollectorConfig `yaml:"collectors"`
	}

	data, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	decoder := yaml.NewDecoder(data)
	decoder.KnownFields(false) // not fail on unknown keys | changed from fail on unknown keys because of addition of metadata
	var rawCfg RawConfig
	if err := decoder.Decode(&rawCfg); err != nil {
		return nil, err
	}

	collectors := []CollectorConfig{}
	for _, rawCollectorCfg := range rawCfg.Collectors {
		collector := availableCollectors[rawCollectorCfg.Name]
		if collector == nil {
			return nil, fmt.Errorf("unknown collector %s", rawCollectorCfg.Name)
		}
		locations := []Location{}
		for _, rawLocation := range rawCollectorCfg.Locations {
			locations = append(locations, Location{rawLocation.ID, rawLocation.Name, Coord{rawLocation.Lat, rawLocation.Lon}, rawLocation.Metadata})
		}
		collectors = append(collectors, CollectorConfig{collector, rawCollectorCfg.Interval, locations})
	}

	return &Config{collectors}, nil
}

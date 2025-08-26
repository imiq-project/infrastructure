package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Coord struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

type Location struct {
	ID    string
	Name  string
	Coord Coord
}

type Collector interface {
	Name() string
	Fetch(Coord) (map[string]any, error)
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
		ID   string
		Name string
		Lat  float64
		Lon  float64
	}

	type RawCollectorConfig struct {
		Name         string        `yaml:"name"`
		IntervalSecs int           `yaml:"intervalSecs"`
		Locations    []RawLocation `yaml:"locations"`
	}

	type RawConfig struct {
		Collectors []RawCollectorConfig `yaml:"collectors"`
	}

	data, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	decoder := yaml.NewDecoder(data)
	decoder.KnownFields(true) // fail on unknown keys
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
			locations = append(locations, Location{rawLocation.ID, rawCollectorCfg.Name, Coord{rawLocation.Lat, rawLocation.Lon}})
		}
		interval := time.Duration(rawCollectorCfg.IntervalSecs) * time.Second
		collectors = append(collectors, CollectorConfig{collector, interval, locations})
	}

	return &Config{collectors}, nil
}

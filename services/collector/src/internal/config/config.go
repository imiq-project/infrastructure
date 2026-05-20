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
	/*
	 * Name returns the name of the collector, which is used in the configuration to reference it. It should be unique among all collectors.
	 */
	Name() string

	/*
	* Setup is called once at the start of the application with all locations that are assigned to this collector. It can be used to initialize the collector, e.g. by setting up database connections or initializing data structures.
	 */
	Setup([]Location)

	/*
	 * BeforeFetch is called before each fetch cycle. It can be used to perform any necessary steps before fetching data, e.g. refreshing authentication tokens.
	 */
	BeforeFetch() error

	/*
	 * Fetch is called for each location assigned to this collector in each fetch cycle. It should return a map representing the entity to be stored in the database. The map must contain a "type" field, which specifies the type of the entity (e.g. "Weather", "ParkingLot", etc.). The other fields can be defined as needed by the collector, but should follow the general structure of having a "type" and "value" subfield if they represent a property of the entity (e.g. "temperature": {"type": "Number", "value": 20}).
	 */
	Fetch(Location) (map[string]any, error)

	/*
	 * AfterFetch is called after each fetch cycle. It can be used to perform any necessary steps after fetching data, e.g. closing database connections or cleaning up resources.
	 */
	AfterFetch() error
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

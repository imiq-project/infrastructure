package main

import (
	"context"
	"flag"
	"imiq/collector/internal/collectors"
	"imiq/collector/internal/config"
	"imiq/collector/internal/fiware"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var collectorConstructors = [...]func() (config.Collector, error){
	collectors.NewOpenWeatherMapCollector,
	collectors.NewParkingCollector,
	collectors.NewTrafficCollector,
	collectors.NewRestaurantCollector,
	collectors.NewAirQualityCollector,
	collectors.NewEnergyCollector,
	collectors.NewStaticCollector,
	collectors.NewWaterLevelCollector,
}

func FetchAll(orionHost string, collectorConfig config.CollectorConfig) {
	log.Println("Running collector", collectorConfig.Collector.Name(), "on", len(collectorConfig.Locations), "locations")
	err := collectorConfig.Collector.BeforeFetch()
	if err != nil {
		log.Println("BeforeEach failed for:", collectorConfig.Collector.Name(), err)
		return
	}
	for _, location := range collectorConfig.Locations {
		result, err := collectorConfig.Collector.Fetch(location)
		if err != nil {
			log.Println("Failed to fetch:", location.ID, err)
		} else {
			err := fiware.UpdateEntity(orionHost, location.ID, result)
			if err != nil {
				log.Println("Failed to update:", location.ID, err)
			}
		}
	}
	err = collectorConfig.Collector.AfterFetch()
	if err != nil {
		log.Println("AfterEach failed for:", collectorConfig.Collector.Name(), err)
		return
	}

	log.Println("Finished collector", collectorConfig.Collector.Name())
}

func worker(ctx context.Context, orionHost string, collectorCfg config.CollectorConfig) {
	log.Println("Started worker for", collectorCfg.Collector.Name())
	ticker := time.NewTicker(collectorCfg.Intervall)
	defer ticker.Stop()
	FetchAll(orionHost, collectorCfg)
	for {
		select {
		case <-ticker.C:
			FetchAll(orionHost, collectorCfg)
		case <-ctx.Done():
			log.Println("Stopped worker for", collectorCfg.Collector.Name())
			return
		}
	}
}

func main() {
	log.Println("Starting")
	configFile := flag.String("config", "config.yml", "Config file")
	orionHost := flag.String("orion", "http://orion:1026", "Fiware Orion Host")
	flag.Parse()

	// Handle sigint / sigterm
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sigintReceived := make(chan bool, 1)
	go func() {
		sig := <-sigs
		log.Println("Received signal", sig)
		sigintReceived <- true
	}()

	// Create collectors
	var collectors = map[string]config.Collector{}
	for _, constructor := range collectorConstructors {
		collector, err := constructor()
		if err != nil {
			log.Println("Cannot create collector: ", err)
			os.Exit(1)
		}
		log.Println("Successfully created collector", collector.Name())
		collectors[collector.Name()] = collector
	}

	// Read config
	log.Println("Reading config from", *configFile)
	cfg, err := config.ReadConfig(*configFile, collectors)
	if err != nil {
		log.Println("Cannot read config file:", err)
		os.Exit(1)
	}

	// Setup all workers
	for _, collectorConfig := range cfg.Collectors {
		log.Println("Setting up collector", collectorConfig.Collector.Name())
		collectorConfig.Collector.Setup(collectorConfig.Locations)
	}

	// Start Workers
	log.Println("Read config file with", len(cfg.Collectors), "collectors")
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	for _, collectorConfig := range cfg.Collectors {
		wg.Go(func() {
			worker(ctx, *orionHost, collectorConfig)
		})
	}

	// Terminate
	log.Println("Press Ctrl+C to stop")
	<-sigintReceived
	cancel()
	wg.Wait()
	log.Println("Goodbye :)")
}

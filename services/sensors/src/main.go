package main

import (
	"log"
	"net/http"
	"os"
)

const (
	ID_KEY       = "wsid"
	PASSWORD_KEY = "wspw"
	FIWARE_URL   = "http://orion:1026"
	BRESSER_ID   = "Sensor:Weather:Walter"
	SENSECAP_ID  = "Sensor:Weather:Winfred"
)

func setupBresser() {
	expectedPassword, exists := os.LookupEnv("WEATHER_STATION_PASSWORD")
	if !exists {
		panic("Cannot lookup password")
	}

	entity := map[string]any{
		"type": "WeatherObserved",
		"humidity": map[string]any{
			"type":  "float",
			"value": 0,
		},
		"temperature": map[string]any{
			"type":  "float",
			"value": 0,
		},
		"location": map[string]any{
			"type":  "geo:point",
			"value": "52.141234471041685, 11.654583803189286",
		},
		"additionalData": map[string]any{
			"type":  "any",
			"value": map[string]any{},
		},
	}

	err := CreateEntity(FIWARE_URL, BRESSER_ID, entity)
	if err != nil {
		log.Println("Failed to create sensor", err)
		// we continue anyway, maybe the entity already exists
	}

	http.HandleFunc("/data/upload.php", func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		_, exists := params[ID_KEY]
		if !exists {
			http.Error(w, ID_KEY+" is missing", http.StatusBadRequest)
			return
		}
		delete(params, ID_KEY)
		password, exists := params[PASSWORD_KEY]
		if !exists {
			http.Error(w, PASSWORD_KEY+" is missing", http.StatusBadRequest)
			return
		}
		delete(params, PASSWORD_KEY)

		if password[0] != expectedPassword {
			http.Error(w, "invalid password", http.StatusBadRequest)
			return
		}

		entity := map[string]any{
			"humidity": map[string]any{
				"type":  "float",
				"value": params.Get("t1hum"),
			},
			"temperature": map[string]any{
				"type":  "float",
				"value": params.Get("t1tem"),
			},
			"additionalData": map[string]any{
				"type":  "any",
				"value": params,
			},
		}
		err := UpdateEntity(FIWARE_URL, BRESSER_ID, entity)
		if err != nil {
			log.Println("Failed to update weather data: ", err)
		}
	})
}

func setupSenseCap() {
	thingsNetUrl, exists := os.LookupEnv("THE_THINGS_NET_WEBHOOK_URL")
	if !exists {
		panic("Cannot lookup webhook url")
	}

	entity := map[string]any{
		"type": "WeatherObserved",
		"humidity": map[string]any{
			"type":  "float",
			"value": 0,
		},
		"temperature": map[string]any{
			"type":  "float",
			"value": 0,
		},
		"location": map[string]any{
			"type":  "geo:point",
			"value": "52.14614723277433, 11.661766246279447",
		},
	}

	err := CreateEntity(FIWARE_URL, SENSECAP_ID, entity)
	if err != nil {
		log.Println("Failed to create sensor", err)
		// we continue anyway, maybe the entity already exists
	}

	http.HandleFunc(thingsNetUrl, func(w http.ResponseWriter, r *http.Request) {
		resp, err := ProcessWebhook(r.Body)
		if err != nil {
			log.Println("Cannot decode webhook", err)
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		if resp.DeviceID != "winfred" {
			log.Println("Got invalid device id", resp.DeviceID)
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		data, err := decodeSenseCapPayload(resp.Payload)
		if err != nil {
			log.Println("Cannot decode SenseCap", err)
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		dataMapped := make(map[string]any)
		for key, value := range data {
			dataMapped[key] = map[string]any{
				"type":  "float",
				"value": value,
			}
		}
		err = UpdateEntity(FIWARE_URL, SENSECAP_ID, dataMapped)
		if err != nil {
			log.Println("Failed to update weather data: ", err)
		}
	})
}

func main() {
	setupBresser()
	setupSenseCap()
	log.Println("Listening on :80")
	log.Fatal(http.ListenAndServe(":80", nil))
}

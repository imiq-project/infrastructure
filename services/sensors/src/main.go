package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
)

const (
	ID_KEY              = "wsid"
	PASSWORD_KEY        = "wspw"
	FIWARE_URL          = "http://orion:1026"
	BRESSER_STATION_ID  = "Sensor:Weather:Walter"
	BRESSER_BUILDING_ID = "Building:UniversityG80"
	SENSECAP_ID         = "Sensor:Weather:Winfred"
)

func move(from *url.Values, fromKey string, to map[string]any, toKey string, typ string) {
	to[toKey] = map[string]any{
		"type":  typ,
		"value": from.Get(fromKey),
	}
	from.Del(fromKey)
}

func setupBresser() {
	expectedPassword, exists := os.LookupEnv("WEATHER_STATION_PASSWORD")
	if !exists {
		panic("Cannot lookup password")
	}

	stationEntity := map[string]any{
		"type": "Weather",
		"location": map[string]any{
			"type":  "geo:point",
			"value": "52.141234471041685, 11.654583803189286",
		},
	}

	err := CreateEntity(FIWARE_URL, BRESSER_STATION_ID, stationEntity)
	if err != nil {
		log.Println("Failed to create station", err)
		// we continue anyway, maybe the entity already exists
	}

	buildingEntity := map[string]any{
		"type": "Building",
		"location": map[string]any{
			"type":  "geo:point",
			"value": "52.14117022025373, 11.654859873233807",
		},
	}

	err = CreateEntity(FIWARE_URL, BRESSER_BUILDING_ID, buildingEntity)
	if err != nil {
		log.Println("Failed to create building", err)
		// we continue anyway, maybe the entity already exists
	}

	http.HandleFunc("/data/upload.php", func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()

		apiVersion, exists := params["apiver"]
		if !exists || apiVersion[0] != "8" {
			log.Println("Unexpected api version", apiVersion)
		}
		params.Del("apiver")

		_, exists = params[ID_KEY]
		if !exists {
			http.Error(w, ID_KEY+" is missing", http.StatusBadRequest)
			return
		}
		params.Del(ID_KEY)

		password, exists := params[PASSWORD_KEY]
		if !exists {
			http.Error(w, PASSWORD_KEY+" is missing", http.StatusBadRequest)
			return
		}
		params.Del(PASSWORD_KEY)

		if password[0] != expectedPassword {
			http.Error(w, "invalid password", http.StatusBadRequest)
			return
		}

		entity := make(map[string]any)
		move(&params, "intem", entity, "temperature", "float")
		move(&params, "inhum", entity, "humidity", "float")
		err = UpdateEntity(FIWARE_URL, BRESSER_BUILDING_ID, entity)
		if err != nil {
			log.Println("Failed to update weather data: ", err)
		}

		entity = make(map[string]any)
		move(&params, "t1hum", entity, "humidity", "float")
		move(&params, "t1tem", entity, "temperature", "float")
		move(&params, "rbar", entity, "airPressure", "float")
		move(&params, "abar", entity, "airPressureAbsolute", "float")
		move(&params, "t1wdir", entity, "windDirection", "float")
		move(&params, "t1ws", entity, "windSpeed", "float")
		move(&params, "t1uvi", entity, "uvIndex", "float")
		move(&params, "t1rainra", entity, "rain", "float")
		move(&params, "t1solrad", entity, "lightIntensity", "float")
		move(&params, "t1chill", entity, "windChillTemperature", "float")
		move(&params, "t1heat", entity, "heatIndexTemperature", "float")
		move(&params, "t1feels", entity, "feelsLikeTemperature", "float")
		move(&params, "t1dew", entity, "dewPointTemperature", "float")
		move(&params, "t1rainhr", entity, "rainHourly", "float")
		move(&params, "t1raindy", entity, "rainDaily", "float")
		move(&params, "t1rainwy", entity, "rainWeekly", "float")
		move(&params, "t1rainmth", entity, "rainMonthly", "float")
		move(&params, "t1rainyr", entity, "rainYearly", "float")
		move(&params, "t1ws10mav", entity, "windSpeed10Min", "float")
		move(&params, "t1wgust", entity, "windGust", "float")
		move(&params, "datetime", entity, "datetime", "string")
		entity["additionalData"] = map[string]any{
			"type":  "any",
			"value": params,
		}
		err := UpdateEntity(FIWARE_URL, BRESSER_STATION_ID, entity)
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
		"type": "Weather",
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

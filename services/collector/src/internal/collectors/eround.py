import requests
import yaml

response = requests.get(
    "https://map.eround.de/api/chargemap/rest/public/marker/search/24d89345-0ac5-40e4-a982-a501f85758cc",
    params={
        # "maximumMaxPowerInWatts": 999999,
        # "minimumMaxPowerInWatts": 1000,
        "lat": 52.13681446435348,
        "lng": 11.643464867382136,
        "distanceInMeters": 106250,
    },
)

data = response.json()
locations = []
for i in data["features"]:
    csname = i["properties"]["csName"]
    lon = i["geometry"]["coordinates"][0]
    lat = i["geometry"]["coordinates"][1]

    response = requests.get(
        "https://map.eround.de/api/chargemap/rest/public/marker/24d89345-0ac5-40e4-a982-a501f85758cc",
        params={"csName": csname},
    )
    station_data = response.json()["payload"]
    name = f"{station_data['address']['street']} {station_data['address']['number']}"
    id_ = name.replace(" ", "").replace("ä", "ae").replace("ö", "oe").replace("ü", "ue").replace("ß", "ss").replace("-", "")
    locations.append({"id": id_, "name": name, "lat": lat, "lon": lon, "csname": csname})

result = {"name": "Charging", "interval": "30m", "locations": locations}

print(yaml.safe_dump(result, sort_keys=False))

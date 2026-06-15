import requests
import time

orion_url = "http://orion:1026/v2/subscriptions"


def get_all_subscriptions():
    response = requests.get(orion_url)
    response.raise_for_status()
    return response.json()


def delete_subscription(sub_id):
    response = requests.delete(f"{orion_url}/{sub_id}")
    response.raise_for_status()


def delete_all_subscriptions():
    subscriptions = get_all_subscriptions()
    for sub in subscriptions:
        print(f"Delete subscription: {sub['description']}")
        delete_subscription(sub["id"])


def create_subscription(filter: dict, filter_description: str, attrs: list):
    print(f"New subscription: {filter_description}")
    response = requests.post(
        orion_url,
        json={
            "description": f"Feed {filter_description} into quantum leap",
            "subject": {
                "entities": [filter],
                "condition": {
                    "attrs": attrs,
                },
            },
            "notification": {
                "http": {"url": "http://quantumleap:8668/v2/notify"},
                "attrs": attrs,
            },
        },
    )
    response.raise_for_status()


def create_subscription_by_type(entity_type, attrs):
    filter = {"idPattern": ".*", "type": entity_type}
    desc = f"type is {entity_type}"
    create_subscription(filter, desc, attrs)


def create_subscription_for_id(entity_id, attrs):
    filter = {"id": entity_id}
    desc = f"id is {entity_id}"
    create_subscription(filter, desc, attrs)


def wait_fiware_ready():
    print("Waiting for Fiware to be ready...")
    while True:
        try:
            response = requests.get(orion_url)
            if response.status_code == 200:
                print("Fiware is ready.")
                break
        except requests.exceptions.ConnectionError:
            pass
        time.sleep(1)


def main():
    wait_fiware_ready()

    print("Setting up subscriptions...")
    delete_all_subscriptions()

    conf = {
        "AirQuality": [
            "no2",
            "o3",
            "pm10",
            "pm25",
            "co2",
            "temperature",
            "airPressure",
            "humidity",
        ],
        "Parking": ["freeSpots"],
        "Weather": [
            "airPressure",
            "airPressureAbsolute",
            "dewPointTemperature",
            "feelsLikeTemperature",
            "heatIndexTemperature",
            "humidity",
            "lightIntensity",
            "rain",
            "temperature",
            "uvIndex",
            "windChillTemperature",
            "windDirection",
            "windSpeed",
        ],
        "WaterLevel": ["level", "volume"],
        "Building": [
            "cold-water",
            "electricity",
            "heating",
            "humidity",
            "temperature",
        ],
        "EVChargingStation": [
            "availableCapacity",
        ],
    }

    for entity_type, attrs in conf.items():
        create_subscription_by_type(entity_type, attrs)

    large_streets = [
        "Traffic:AmKroekentor",
        "Traffic:AskanischerPlatz",
        "Traffic:ErnstLehmannStrasse",
        "Traffic:Erzbergerstrasse",
        "Traffic:Gareisstrasse",
        "Traffic:Gartenstrasse",
        "Traffic:GrosserWerder",
        "Traffic:GustavAdolfStrasse",
        "Traffic:Herrenkrugstrasse",
        "Traffic:Hohenstaufenring",
        "Traffic:Hohepfortestrasse",
        "Traffic:Hohepfortewall",
        "Traffic:Jakobstrasse",
        "Traffic:Jerusalembruecken",
        "Traffic:JohannGottlobNathusiusRing",
        "Traffic:Markgrafenstrasse",
        "Traffic:Muehlenstrasse",
        "Traffic:PfaelzerPlatz",
        "Traffic:RogaetzerStrasse",
        "Traffic:Sandtorstrasse",
        "Traffic:SarajevoUfer",
        "Traffic:Schifferstrasse",
        "Traffic:Schleinufer",
        "Traffic:TunnelAskanischerPlatz",
        "Traffic:Universitaetsplatz",
        "Traffic:WaltherRathenauStrasse",
        "Traffic:WendekreisBus",
        "Traffic:WittenbergerStrasse",
        "Traffic:Zschokkestrasse",
    ]

    for street in large_streets:
        create_subscription_for_id(street, ["avgSpeed"])

    print("All subscriptions set up.")

    # sleep until terminated
    while True:
        time.sleep(1)


if __name__ == "__main__":
    exit(main())

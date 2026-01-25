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
        delete_subscription(sub["id"])


def create_subscription(entity_type, attrs):
    response = requests.post(
        "http://orion:1026/v2/subscriptions",
        json={
            "description": f"Feed {entity_type} into quantum leap",
            "subject": {
                "entities": [{"idPattern": ".*", "type": entity_type}],
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


def main():
    print("Setting up subscriptions...")

    delete_all_subscriptions()

    conf = {
        "AirQuality": ["no2", "o3", "pm10", "pm25"],
        "Parking": ["freeSpaces"],
        "Weather": ["temperature", "humidity"],
    }

    for entity_type, attrs in conf.items():
        create_subscription(entity_type, attrs)

    print("All subscriptions set up.")

    # sleep until terminated
    while True:
        time.sleep(1)


if __name__ == "__main__":
    exit(main())

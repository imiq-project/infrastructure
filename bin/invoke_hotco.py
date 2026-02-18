#! /usr/bin/env python3
import requests
import json

response = requests.post(
    "http://localhost:8000/api/hotco",
    json={
        "needs": {
            "pro_env": 0.1,
            "physical": 0.7,
            "privacy": 0.2,
            "autonomy": 0.3,
            "hedonism": 0.7,
            "cost": 0.9,
            "speed": 0.3,
            "safety": 0.1,
            "comfort": 0,
        },
        "valences": {
            "car": 0,
            "bike": 0.1,
            "pt": -0.5,
            "walk": 0.3,
        },
        "stressors": {
            "rain": 0.1,
            "crowding": 0.2,
            "darkness": 0,
            "traffic": 0.1,
            "temperature": 0.7,
        },
        "tolerances": {
            "rain": 0.1,
            "crowding": 0.1,
            "darkness": 0.1,
            "traffic": 0.1,
            "temperature": 0.1,
        },
    },
)

response.raise_for_status()
print(json.dumps(response.json(), indent=4))

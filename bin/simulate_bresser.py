#! /usr/bin/env python3
import requests
import os


def simulate_data(host: str, password: str):
    response = requests.post(
        f"{host}/data/upload.php",
        params={
            "wsid": "unset",
            "wspw": password,
            "abar": "980.0",
            "apiver": "8",
            "datetime": "2026-02-11 08:46:02",
            "inbat": "1",
            "inhum": "39",
            "intem": "20.0",
            "rbar": "1013.0",
            "t10cn": "0",
            "t11cn": "0",
            "t1bat": "1",
            "t1chill": "4.4",
            "t1cn": "1",
            "t1dew": "2.8",
            "t1feels": "4.4",
            "t1heat": "4.4",
            "t1hum": "90",
            "t1raindy": "0.000",
            "t1rainhr": "0.000",
            "t1rainmth": "9.398",
            "t1rainra": "0.000",
            "t1rainwy": "0.762",
            "t1rainyr": "19.050",
            "t1solrad": "60.4",
            "t1tem": "4.4",
            "t1uvi": "0.0",
            "t1wdir": "186",
            "t1wgust": "0.5",
            "t1ws": "0.5",
            "t1ws10mav": "0.4",
            "t234c1cn": "0",
            "t234c2cn": "0",
            "t234c3cn": "0",
            "t234c4cn": "0",
            "t234c5cn": "0",
            "t234c6cn": "0",
            "t234c7cn": "0",
            "t5lscn": "0",
            "t6c1cn": "0",
            "t6c2cn": "0",
            "t6c3cn": "0",
            "t6c4cn": "0",
            "t6c5cn": "0",
            "t6c6cn": "0",
            "t6c7cn": "0",
            "t8cn": "0",
            "t9cn": "0",
        },
    )
    print(response.content)
    response.raise_for_status()


if __name__ == "__main__":
    password = os.environ["WEATHER_STATION_PASSWORD"]
    simulate_data("http://localhost:8000", password)

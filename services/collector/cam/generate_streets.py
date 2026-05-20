#! /usr/bin/env python3
import requests
import os
import json
from collections import defaultdict

with open("template.html", "r") as f:
    html_template = f.read()


def send_osm_query(query):
    overpass_url = "https://overpass-api.de/api/interpreter"
    headers = {"User-Agent": "overpass-python-client/1.0"}
    response = requests.post(overpass_url, data={"data": query}, headers=headers)
    response.raise_for_status()
    data = response.json()
    return data


def get_street_data():
    cache_file = ".cached.json"
    if os.path.exists(cache_file):
        print("Using cached data")
        with open(cache_file, "r") as f:
            data = json.load(f)
    else:
        query = """
        [out:json][timeout:60];
        area["name"="Magdeburg"]["boundary"="administrative"]->.a;
        way(area.a)["highway"~"residential|tertiary|secondary|primary|unclassified|service"]["name"];
        (._;>;);
        out geom;
        """
        data = send_osm_query(query)
        with open(cache_file, "w") as f:
            json.dump(data, f, indent=2)
    return data


def render_html(data, outfile):
    features = []
    for el in data.get("elements", []):
        if el["type"] != "way" or "geometry" not in el:
            continue
        coords = [[pt["lon"], pt["lat"]] for pt in el["geometry"]]
        features.append(
            {
                "type": "Feature",
                "properties": {"id": el["tags"]["name"]},
                "geometry": {"type": "LineString", "coordinates": coords},
            }
        )

    geojson = {"type": "FeatureCollection", "features": features}

    with open(outfile, "w") as f:
        html_content = html_template.replace(
            "GEOJSON_DATA_PLACEHOLDER", json.dumps(geojson)
        )
        f.write(html_content)
    print(f"Generated {outfile} with GeoJSON visualization.")


def render_yaml(data, outfile):
    all_streets = defaultdict(list)
    speed_limits = {}

    def to_coord(pt):
        # 5 decimals -> ~1m precision, should be enough for our purposes
        return f"{pt['lat']:.5f},{pt['lon']:.5f}"

    def name_to_id(name):
        id_ = (
            name.replace(" ", "")
            .replace("-", "")
            .replace(".", "")
            .replace("&", "")
            .replace("ä", "ae")
            .replace("Ä", "Ae")
            .replace("ö", "oe")
            .replace("Ö", "Oe")
            .replace("ü", "ue")
            .replace("Ü", "Ue")
            .replace("ß", "ss")
            .replace("é", "e") # Lennéstraße
        )
        return f"Traffic:{id_}"

    for el in data.get("elements", []):
        if el["type"] != "way" or "geometry" not in el:
            continue
        name = el["tags"]["name"]
        speed_limit = el["tags"].get("maxspeed")
        if speed_limit.endswith(" kmh"):
            speed_limit = int(speed_limit[:-4])
        if len(name) < 3:
            continue
        coords = ",".join([to_coord(pt) for pt in el["geometry"]])
        all_streets[name].append(coords)
        if speed_limit:
            speed_limits[name] = speed_limit

    # Sort streets by name for consistent output
    all_streets = sorted(all_streets.items(), key=lambda x: x[0])

    with open(outfile, "w") as f:
        f.write("  - name: Traffic\n")
        f.write("    interval: 15m\n")
        f.write("    locations:\n")
        for name, coords_list in all_streets:
            start = coords_list[0].split(",")
            f.write(f"      - id: {name_to_id(name)}\n")
            f.write(f"        name: {name}\n")
            f.write(f"        lat: {start[0]}\n")
            f.write(f"        lon: {start[1]}\n")
            f.write(f"        segments: {';'.join(coords_list)}\n")
            f.write(f"        speed_limit: {speed_limits.get(name, 30)}\n")
    print(f"Generated {outfile} with street coordinates in YAML format.")


if __name__ == "__main__":
    data = get_street_data()
    render_html(data, "streets.html")
    render_yaml(data, "streets.yaml")

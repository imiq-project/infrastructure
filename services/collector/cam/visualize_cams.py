#! /usr/bin/env python3
import sys
import csv
import json
from collections import defaultdict

with open("template.html", "r") as f:
    html_template = f.read()

def cams_to_geojson(input_file):

    with open(input_file, "r") as f:
        input_text = f.read()

    # Parse CSV-like input
    reader = csv.DictReader(input_text.strip().splitlines(), delimiter=";")

    # Group coordinates by id
    tracks = defaultdict(list)

    for row in reader:
        track_id = row["id"]
        lat = float(row["lat"])
        lon = float(row["lon"])

        # GeoJSON uses [lon, lat]
        tracks[track_id].append([lon, lat])

    # Build GeoJSON FeatureCollection
    geojson = {"type": "FeatureCollection", "features": []}

    for track_id, coords in tracks.items():
        feature = {
            "type": "Feature",
            "properties": {"id": track_id},
            "geometry": {"type": "LineString", "coordinates": coords},
        }
        geojson["features"].append(feature)

    return geojson


if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: visualizeCams.py <input_file>")
        sys.exit(1)
    input_file = sys.argv[1]
    geojson = cams_to_geojson(input_file)
    html_content = html_template.replace("GEOJSON_DATA_PLACEHOLDER", json.dumps(geojson))
    with open("cams.html", "w") as f:
        f.write(html_content)
    print("Generated cams.html with GeoJSON visualization.")

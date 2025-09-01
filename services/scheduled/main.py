#! /usr/bin/env python

from typing import List, Tuple
import json
import signal
import requests
import time
import datetime
from dataclasses import dataclass

SECS_PER_DAY = 24 * 60 * 60
ORION_HOST = "http://orion:1026"
TRACKS_FILE = "harbor.tracks"


@dataclass
class Position:
    lat: float
    lon: float

    @classmethod
    def from_dict(self, d):
        return Position(d[0], d[1])


@dataclass
class Track:
    id: int
    type: str
    start: int
    positions: List[Position]

    @classmethod
    def from_dict(self, d):
        return Track(
            d["id"],
            d["type"],
            d["start"],
            [Position.from_dict(i) for i in d["positions"]],
        )


@dataclass
class ActiveTrack:
    track: Track
    idx: int


def load(path) -> Tuple[int, List[Track]]:
    print(f"Reading {path}")
    with open(path) as f:
        data = json.load(f)
    tracks = [Track.from_dict(i) for i in data["tracks"]]
    period = int(data["period"])
    print(f"Tracks: {len(tracks)}")
    print(f"Period: {period}s")
    return period, tracks


def update_vehicle(id: int, pos: Position, category: str):
    response = requests.put(
        f"{ORION_HOST}/v2/entities/{id}/attrs/location/value",
        json={"coordinates": [pos.lat, pos.lon]},
    )
    if response.status_code == 204:
        return
    elif response.status_code == 404:
        print(f"Creating {id}")
        response = requests.post(
            f"{ORION_HOST}/v2/entities",
            json={
                "id": id,
                "type": "Vehicle",
                "category": {"type": "string", "value": category},
                "location": {
                    "type": "geo:json",
                    "value": {"type": "Point", "coordinates": [pos.lat, pos.lon]},
                },
            },
        )
        if response.status_code != 201:
            print(f"Failed to create: {response.status_code}, {response.content}")
    else:
        print(f"Got unexpected status {response.status_code}, {response.content}")


def delete_vehicle(id: int):
    response = requests.delete(f"{ORION_HOST}/v2/entities/{id}")
    if response.status_code != 204:
        print(f"Failed to delete {id}, {response.status_code}, {response.content}")


def delete_all_vehicles():
    response = requests.get(
        f"{ORION_HOST}/v2/entities", params={"type": "Vehicle", "limit": 200}
    )
    # TODO: pagination
    for vehicle in response.json():
        print(f"Deleting {vehicle['id']}")
        delete_vehicle(vehicle["id"])


def main():
    # Register sigint
    running = True

    def sigint_handler(signum, frame):
        nonlocal running
        running = False
        print("Received SIGINT, stopping...")

    signal.signal(signal.SIGINT, sigint_handler)

    period, tracks = load(TRACKS_FILE)

    delete_all_vehicles()

    now = datetime.datetime.now()
    step = int( (now.hour * 3600 + now.minute * 60 + now.second) / period)

    # Find active tracks for current time
    track_idx = 0
    active_tracks: List[ActiveTrack] = []
    while tracks[track_idx].start < step:
        track_idx += 1

    while running:
        # Start tracks
        while tracks[track_idx].start < step:
            track = tracks[track_idx]
            print(f"{step} Starting track {track.id}")
            active_tracks.append(ActiveTrack(track, 0))
            track_idx += 1

        # Advance active tracks
        new_active_tracks = []
        for active_track in active_tracks:
            pos = active_track.track.positions[active_track.idx]
            full_id = f"Vehicles:{active_track.track.id}"
            update_vehicle(full_id, pos, track.type)
            active_track.idx += 1
            if active_track.idx >= len(active_track.track.positions):
                print(f"{step} Stopping track {active_track.track.id}")
                delete_vehicle(full_id)
            else:
                new_active_tracks.append(active_track)
        active_tracks = new_active_tracks

        # Next sec
        time.sleep(period) # TODO: accurate timing
        step = (step + 1) % SECS_PER_DAY

    print("Goodbye :)")


if __name__ == "__main__":
    main()

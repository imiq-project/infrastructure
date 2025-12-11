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
                "source": {"type": "string", "value": "scheduled"},
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
        f"{ORION_HOST}/v2/entities", params={"type": "Vehicle", "limit": 1000, "q": "source==scheduled"}
    )
    # TODO: pagination
    for vehicle in response.json():
        print(f"Deleting {vehicle['id']}")
        delete_vehicle(vehicle["id"])
"""
#Mensa menu scraper
def check_veg_nonVeg(category_text): 
    veg_keywords = ['vegetarisch', 'vegan']
    non_veg_keywords = ['rind', 'schwein', 'geflügel', 'fisch', 'hähnchen', 'lamm', 'suppe']
    category_text_lower = category_text.lower()
    if any(keyword in category_text_lower for keyword in veg_keywords):
        return 'Vegetarian/Vegan'
    elif any(keyword in category_text_lower for keyword in non_veg_keywords):
        return 'Non-Vegetarian'
    else:
        return category_text

def scrape_mensa():
    print(f"[{datetime.datetime.now()}] Starting Mensa Scraper...")
    try:
        response = requests.get('https://www.studentenwerk-magdeburg.de/mensen-cafeterien/mensa-unicampus-speiseplan-unten/')
        response.raise_for_status()
        soup = BeautifulSoup(response.text, 'html.parser')

        speisekarte = []
        mensa_div = soup.find('div', class_='mensa')
        if not mensa_div:
            print("Error: Could not find div with class 'mensa'")
            return

        all_tables = mensa_div.find_all('table')
        print(f"Found {len(all_tables)} days.")

        for table in all_tables:
            date_header = table.find('thead').get_text().strip()
            day_menu = {'date': date_header, 'meals': []}
            
            rows = table.find('tbody').find_all('tr')
            for row in rows:
                columns = row.find_all('td')
                
                # Name Parsing
                name_node = columns[0].find('span', class_='gruen')
                name_ger = name_node.get_text().strip() if name_node else columns[0].find(string=True, recursive=False).strip()
                
                eng_node = columns[0].find('span', class_='grau')
                name_eng = eng_node.get_text().strip() if eng_node else "No English name"
                
                price = columns[0].find('span', class_='mensapreis').get_text().strip()
                
                # Category Parsing
                img = columns[1].find('img')
                cat_raw = img.get('alt', '').strip() if img else "Unknown"
                category = check_veg_nonVeg(cat_raw)

                day_menu['meals'].append({
                    'name_german': name_ger,
                    'name_english': name_eng,
                    'price': price,
                    'category': category
                })
            speisekarte.append(day_menu)

        # WRAPPER: This matches what Go expects
        final_output = [
            {
                "name": "Mensa UniCampus",
                "schedule": speisekarte
            }
        ]
        # SAVE TO SHARED VOLUME
        output_path = '/app/data/mensa_menus.json'
        with open(output_path, 'w', encoding='utf-8') as f:
            json.dump(final_output, f, ensure_ascii=False, indent=4)
            
        print(f"[{datetime.datetime.now()}] Success! Saved to {output_path}")

    except Exception as e:
        print(f"[{datetime.datetime.now()}] Error: {e}")
"""


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

    """
    # Schedule Mensa Scraper every Friday at 08:00 AM
    scrape_mensa()
    schedule.every().friday.at("08:00").do(scrape_mensa)
    """

    now = datetime.datetime.now()
    step = int((now.hour * 3600 + now.minute * 60 + now.second) / period)

    # Find active tracks for current time
    track_idx = 0
    active_tracks: List[ActiveTrack] = []
    while tracks[track_idx].start < step:
        track_idx += 1

    while running:
        
        
        """
        #Run Pending Schedule Jobs (Mensa)
        schedule.run_pending()
        """
        

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
        time.sleep(period)  # TODO: accurate timing
        step = (step + 1) % SECS_PER_DAY

    print("Goodbye :)")


if __name__ == "__main__":
    main()

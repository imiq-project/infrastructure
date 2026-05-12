import requests
import sys
import os
import time

REMOTE_ORION_URL = os.getenv("REMOTE_ORION_URL", "https://imiq-public.et.uni-magdeburg.de/api/orion/entities")
LOCAL_ORION_URL = os.getenv("LOCAL_ORION_URL", "http://orion:1026/v2/entities")

def sync_entities():
    print(f"Checking local environment at {LOCAL_ORION_URL}...")
    
    # 1. Check: Exit if data already exists
    max_retries = 5
    for attempt in range(max_retries):
        try:
            local_check = requests.get(f"{LOCAL_ORION_URL}?limit=1")
            if local_check.status_code == 200 and len(local_check.json()) > 0:
                print("Local environment already populated. Skipping sync.")
                sys.exit(0)
            else:
                break  # Local is empty, proceed with sync
        except requests.exceptions.ConnectionError:
            print(f"Attempt {attempt + 1}/{max_retries}: Local Orion is not ready yet. Waiting 3 seconds...")
            time.sleep(3)
    else:
        print(f"CRITICAL ERROR: Could not connect to local Orion at {LOCAL_ORION_URL} after {max_retries} attempts.")
        print("Run 'docker compose logs orion' to see if Orion is crashing.")
        sys.exit(1)

    # 2. Sync Logic
    limit = 1000
    offset = 0
    total_synced = 0

    print(f"Starting sync from {REMOTE_ORION_URL}...")

    while True:
        # Fetch from remote
        response = requests.get(f"{REMOTE_ORION_URL}?limit={limit}&offset={offset}")
        
        if response.status_code != 200:
            print(f"Failed to fetch data: HTTP {response.status_code} - {response.text}")
            sys.exit(1)

        entities = response.json()
        
        # Stop if pagination is exhausted
        if not entities:
            break 

        # Push to local
        for entity in entities:
            entity_id = entity.get('id', 'Unknown')
            
            post_resp = requests.post(LOCAL_ORION_URL, json=entity)
            
            if post_resp.status_code in (201, 204):
                total_synced += 1
            elif post_resp.status_code == 422:
                # 422 Unprocessable Entity means FIWARE already has this ID
                pass 
            else:
                print(f"Error pushing '{entity_id}': HTTP {post_resp.status_code} - {post_resp.text}")

        offset += limit

    print(f"Sync complete. {total_synced} new entities pushed to local environment.")

if __name__ == "__main__":
    sync_entities()
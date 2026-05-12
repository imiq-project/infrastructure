import requests
import sys
import argparse 

def sync_entities(remote_url, local_url, api_key):
    headers = {}
    if api_key:
        headers['x-api-key'] = api_key

    print(f"checking local environment at {local_url}...")

    try:
        local_check = requests.get(f"{local_url}/?limit=1", headers=headers, timeout=5)
        if local_check.status_code == 200 and len(local_check.json()) > 0:
            print("local environment is available, skipping sync")
            sys.exit(0)
    except requests.exceptions.ConnectionError:
        print(f"CRITICAL ERROR: Could not connect to {local_url}. Please check if the local Orion Context Broker is running and accessible.")
        sys.exit(1)
    
    limit = 1000
    offset = 0
    total_synced = 0

    while True:
        response = requests.get(f"{remote_url}?limit={limit}&offset={offset}", headers=headers)
        if response.status_code != 200:
            print(f"Error fetching entities from remote Orion: {response.status_code} - {response.text}")
            sys.exit(1)
        
        entities = response.json()
        if not entities:
            break
        for entity in entities:
            entity_id = entity.get('id', 'unknown')
            post_response = requests.post(local_url, json=entity, headers=headers)

            if post_response.status_code in [201, 204]:
                total_synced += 1
            elif post_response.status_code == 422:
                pass
            else:
                print(f"Error pushing '{entity_id}': HTTP {post_response.status_code} - {post_response.text}")
            
        offset += limit

    print(f"Sync completed. Total entities synced: {total_synced}")

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Sync entities from remote to local .')
    parser.add_argument('--remote', type=str, default="https://imiq-public.et.uni-magdeburg.de/api/orion/entities", help='URL of the remote Orion Context Broker')
    parser.add_argument('--local_url', type=str, default="http://localhost:8000/api/orion/entities", help='URL of the local Orion Context Broker')
    parser.add_argument('--api_key', required=True, type=str, default=None, help='API key for authentication (if required)')

    args = parser.parse_args()
    sync_entities(args.remote, args.local_url, args.api_key)
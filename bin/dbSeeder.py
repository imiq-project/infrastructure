import requests
import sys
import argparse 
from datetime import datetime, timedelta, timezone

def sync_entities(remote_orion, local_orion, remote_ql, local_ql, api_key, history_days):
    headers = {}
    if api_key:
        headers['x-api-key'] = api_key

    print(f"checking local environment at {local_orion}...")

    try:
        local_check = requests.get(f"{local_orion}/?limit=1", headers=headers, timeout=5)
        if local_check.status_code == 200 and len(local_check.json()) > 0:
            print("local environment is available, skipping sync")
            sys.exit(0)
    except requests.exceptions.ConnectionError:
        print(f"CRITICAL ERROR: Could not connect to {local_orion}. Please check if the local Orion Context Broker is running and accessible.")
        sys.exit(1)
    
    limit = 1000
    offset = 0
    total_orion_synced = 0
    synced_entities =[]

    while True:
        response = requests.get(f"{remote_orion}?limit={limit}&offset={offset}", headers=headers)
        if response.status_code != 200:
            print(f"Error fetching entities from remote Orion: {response.status_code} - {response.text}")
            sys.exit(1)
        
        entities = response.json()
        if not entities:
            break
        for entity in entities:
            entity_id = entity.get('id', 'unknown')
            entity_type = entity.get('type', 'unknown')

            synced_entities.append((entity_id, entity_type))

            post_response = requests.post(local_orion, json=entity, headers=headers)

            if post_response.status_code in [201, 204]:
                total_orion_synced += 1
            elif post_response.status_code == 422:
                pass
            else:
                print(f"Error pushing '{entity_id}': HTTP {post_response.status_code} - {post_response.text}")
            
        offset += limit

    print(f"Sync completed. Total entities synced: {total_orion_synced}")

    #Get historical data for each entity and sync to local QL
    print(f"Starting historical data sync for the last {history_days} days...")
    start_date = (datetime.now(timezone.utc) - timedelta(days=history_days)).strftime('%Y-%m-%dT%H:%M:%SZ')
    total_historical_records = 0
    total_entities = len(synced_entities)

    for index, (entity_id, entity_type) in enumerate(synced_entities):
        progress = (index + 1) / total_entities
        bar_length = 40
        filled_length = int(bar_length * progress)
        bar = '█' * filled_length + '-' * (bar_length - filled_length)
        print(f'\rSyncing History: |{bar}| {index+1}/{total_entities} Entities', end='', flush=True)
        
        history_url = f"{remote_ql}/entities/{entity_id}?fromDate={start_date}"
        response = requests.get(history_url)

        if response.status_code != 200:
            continue
        historical_data = response.json()
        if not historical_data.get('index'):
            continue
        
        timestamps = historical_data['index']
        attributes = historical_data.get('attributes', [])
       
        notifications = []
        for i, t in enumerate(timestamps):
            entity_state = {
                "id": entity_id,
                "type": entity_type,
                "timestamp": {"type":"DateTime", "value":t}
            }
            for attr in attributes:
                attr_name = attr['attrName']
                values = attr['values'][i]
                if values is not None:
                    t_value = "Number" if isinstance(values, (int, float)) else "Text"
                    entity_state[attr_name] = {"type": t_value, "value": values}

            notifications.append(entity_state)
        
        chunk_size = 50 #50 to avoid overwhelming the local QL
        
        for i in range(0, len(notifications), chunk_size):
            chunk = notifications[i:i+chunk_size]
            payload = {
                "data": chunk,
                "subscriptionId": "historical-seed"
            }
            q_response = requests.post(f"{local_ql}/notify", json=payload, headers=headers)

            if q_response.status_code in [200, 201, 204]:
                total_historical_records += len(chunk)
            else:                
                print(f"\nError pushing to local QL: HTTP {q_response.status_code} - {q_response.text}")
                break
            
    print()
    print(f"Historical data sync completed. Total records synced: {total_historical_records}")

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Sync entities and historical data from remote to local.')
    parser.add_argument('--remote-orion', type=str, default="https://imiq-public.et.uni-magdeburg.de/api/orion/entities", help='URL of the remote Orion Context Broker')
    parser.add_argument('--local-orion', type=str, default="http://localhost:8000/api/orion/entities", help='URL of the local Orion Context Broker')
    parser.add_argument('--remote-ql', type=str, default="https://imiq-public.et.uni-magdeburg.de/api/quantumleap", help='URL of the remote QL Broker')
    parser.add_argument('--local-ql', type=str, default="http://localhost:8000/api/quantumleap", help='URL of the local QL Broker')
    parser.add_argument('--api-key', required=True, type=str, default=None, help='API key for authentication (if required)')
    parser.add_argument('--history-days', type=int, default=7, help='Number of days of historical data to pull')

    args = parser.parse_args()
    sync_entities(args.remote_orion, args.local_orion, args.remote_ql, args.local_ql, args.api_key, args.history_days)
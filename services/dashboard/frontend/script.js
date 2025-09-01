// ------------------------------------
// Center map on Vision Hub / Science Hub (Magdeburg sample coords)
// ------------------------------------
const map = L.map('map').setView([52.140, 11.644], 16.5);

// -------------
// Base Maps tiles
// -------------
// 1. Satellite imagery (Esri)
// High-quality Esri Hybrid Layer (imagery + full labels)
const esriHybrid = L.tileLayer(
  'https://services.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}', {
    attribution: 'Imagery © Esri, i-cubed, USDA, USGS, AEX, GeoEye, and the GIS User Community'
  }
).addTo(map);

// Add this one last so it overlays cleanly with full labels
const esriReference = L.tileLayer(
  'https://services.arcgisonline.com/ArcGIS/rest/services/Reference/World_Transportation/MapServer/tile/{z}/{y}/{x}', {
    attribution: 'Labels © Esri & Contributors',
    opacity: 0.9
  }
).addTo(map);



// L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
//  attribution: '&copy; OpenStreetMap contributors'
// }).addTo(map);

let currentSensorMode = "";

// ---------------
// Weather sensor locations (initially no markers)
// ---------------
const weatherSensors = [
  { id: 'Sensor:Weather:FacultyCS', label: 'Faculty of Computer Science', coords: [52.13878, 11.64533], marker: null },
  { id: 'Sensor:Weather:ScienceHub', label: 'Science Harbor', coords: [52.14175, 11.65640], marker: null },
  { id: 'Sensor:Weather:UniMensa', label: 'University Mensa', coords: [52.13966, 11.64761], marker: null },
  { id: 'Sensor:Weather:Library', label: 'University Library', coords: [52.13888, 11.64707], marker: null },
  { id: 'Sensor:Weather:WelcomeCenter', label: 'Welcome Center', coords: [52.14031, 11.64039], marker: null },
  { id: 'Sensor:Weather:NorthPark', label: 'North Park', coords: [52.14276, 11.64513], marker: null },
  { id: 'Sensor:Weather:GeschwisterPark', label: 'Geschwister School Park', coords: [52.14020, 11.63655], marker: null }
];

// ---------------
// Parking simulation markers
// ---------------
const parkingSpots = [
  {
    id: "ParkingSpot:ScienceHarbor",
    label: "Parking A - Science Harbor",
    coords: [52.1412, 11.6558],
    marker: null
  },
  {
    id: "ParkingSpot:FacultyCS",
    label: "Parking B - Faculty CS",
    coords: [52.1383, 11.6448],
    marker: null
  },
  {
    id: "ParkingSpot:NorthPark",
    label: "Parking C - North Park",
    coords: [52.1431, 11.6457],
    marker: null
  }
];

// -----------------
// Get data from Orion Context Broker
// -----------------
async function getSensorData(sensorId) {
  try {
    const res = await fetch(`/entities/${sensorId}`);
    if (!res.ok) {
      console.error("Failed to fetch data for", sensorId);
      return null;
    }
    const json = await res.json();
    console.log("Fetched sensor:", sensorId, json);
    return json;
  } catch (error) {
    console.error("Error fetching data for", sensorId, error);
    return null;
  }
}

async function getAllVehicles() {
  try {
    const res = await fetch(`/entities?type=Vehicle`);
    if (!res.ok) {
      console.error("Error fetching vehicles", error);
      return null;
    }
    const json = await res.json();
    return json;
  } catch (error) {
    console.error("Error fetching vehicles", error);
    return null;
  }  
}

// -----------------------
// Attach popup with live weather data
// -----------------------
function attachLivePopup(marker, sensorId, label) {
  marker.on("click", async () => {
    const data = await getSensorData(sensorId);
    if (!data || (!data.temperature?.value && !data.humidity?.value)) {
      marker.setPopupContent(`${label}<br><i>No sensor data</i>`);
      return;
    }

    const popupContent = `
      <b>${label}</b><br>
      🌡️ Temp: ${data.temperature?.value ?? "n/a"} °C<br>
      💧 Humidity: ${data.humidity?.value ?? "n/a"} %
    `;

    marker.setPopupContent(popupContent);
    marker.openPopup();
  });
}

// --------------------------------------
// Update weather tooltips
// --------------------------------------
async function updateAllPopups(sensorType) {
  for (const entry of weatherSensors) {
    const data = await getSensorData(entry.id);
    if (!data || (!data.temperature?.value && !data.humidity?.value)) continue;

    const valueText = sensorType === "temperature"
      ? `🌡️ ${data.temperature.value} °C`
      : `💧 ${data.humidity.value} %`;

    if (!entry.marker) {
      entry.marker = L.marker(entry.coords).addTo(map);
      attachLivePopup(entry.marker, entry.id, entry.label);
    }

    const tooltip = `${entry.label}<br>${valueText}`;
    entry.marker.bindTooltip(tooltip, {
      permanent: true,
      direction: "top",
      offset: [0, -10]
    }).openTooltip();
  }
}

// --------------------------------------
// Update parking markers and tooltips
// --------------------------------------
async function updateParkingSpots() {
  for (const spot of parkingSpots) {
    const data = await getSensorData(spot.id);
    if (!data || data.freeSpaces?.value == null || data.totalSpaces?.value == null) continue;

    const free = data.freeSpaces.value;
    const total = data.totalSpaces.value;

    const label = `${spot.label}<br>🚗 ${free} of ${total} spaces free`;

    if (!spot.marker) {
      spot.marker = L.marker(spot.coords).addTo(map);
    }

    spot.marker.bindTooltip(label, {
      permanent: true,
      direction: "top",
      offset: [0, -10]
    }).openTooltip();
  }
}
// --------------------------------------
// Update all vehicle markers on map
// --------------------------------------
const createIcon = (html) => L.divIcon({
    html: html,
    className: 'vehicle-icon',
    iconSize: [64, 64],
    iconAnchor: [32, 32]
  })
icons = {
  'bus': createIcon('🚌'),
  'tram': createIcon('🚊'),
  'train': createIcon('🚆'),
}

let vehicleMarkers = []
async function updateVehicles() {
  const vehicles = await getAllVehicles();
  vehicleMarkers.forEach(m => m.remove())
  vehicleMarkers = []
  for(const vehicle of vehicles) {
    const icon = icons[vehicle.category.value]
    const marker = L.marker(vehicle.location.value.coordinates, {icon: icon}).addTo(map)
    vehicleMarkers.push(marker)
  }
}

// --------------------------------------
// Remove all markers from the map
// --------------------------------------
function clearAllMarkers() {
  weatherSensors.forEach(entry => {
    if (entry.marker) {
      entry.marker.remove();
      entry.marker = null;
    }
  });
  parkingSpots.forEach(entry => {
    if (entry.marker) {
      entry.marker.remove();
      entry.marker = null;
    }
  });
}

// --------------------------------------
// Handle dropdown change
// --------------------------------------
document.getElementById("sensorType").addEventListener("change", (e) => {
  const selected = e.target.value;
  currentSensorMode = selected;

  clearAllMarkers();

  if (selected === "temperature" || selected === "humidity") {
    updateAllPopups(selected);
  } else if (selected === "parking") {
    updateParkingSpots();
  }
});

// --------------------------------------
// Trigger first load (nothing selected)
// --------------------------------------
currentSensorMode = ""; // Start with clean map

// --------------------------------------
// Auto-refresh data every 10 seconds
// --------------------------------------
updateVehicles();
setInterval(updateVehicles, 1000);
setInterval(() => {
  if (currentSensorMode === "parking") {
    updateParkingSpots();
  } else if (currentSensorMode === "temperature" || currentSensorMode === "humidity") {
    updateAllPopups(currentSensorMode);
  }
}, 10000);

// ------------------------------------
// Center map on Vision Hub / Science Hub
// ------------------------------------
// OpenStreetMap standard tiles
const osmRoad = L.tileLayer(
  'https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png',
  { attribution: '&copy; OpenStreetMap contributors' }
);

// -------------
// Base Maps tiles
// -------------
// Satellite imagery (Esri)
// High-quality Esri Hybrid Layer (imagery + full labels)
const esriImagery = L.tileLayer(
  'https://services.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}',
  { attribution: 'Imagery © Esri' }
);

const esriLabels = L.tileLayer(
  'https://services.arcgisonline.com/ArcGIS/rest/services/Reference/World_Transportation/MapServer/tile/{z}/{y}/{x}',
  { attribution: 'Labels © Esri', opacity: 0.9 }
);

const openTopo = L.tileLayer(
  'https://{s}.tile.opentopomap.org/{z}/{x}/{y}.png',
  { attribution: 'Map data: &copy; OpenStreetMap contributors, SRTM | Map style: &copy; OpenTopoMap (CC-BY-SA)' }
);

const cartoLight = L.tileLayer(
  'https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}.png',
  {attribution: '© OpenStreetMap contributors © CARTO'}
);

const cartoDark = L.tileLayer(
  'https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}.png',
  {attribution: '© OpenStreetMap contributors © CARTO' }
);

const satellite = L.layerGroup([esriImagery, esriLabels]);

const map = L.map('map', {
  center: [52.140, 11.644],
  zoom: 16.5,
  layers: [satellite],
  fullscreenControl: true,
});

L.control.layers({
  "🛰️ Satellite": satellite,
  "🗺️ OpenStreetMap": osmRoad,
  "🏔️ Topographic": openTopo,
  "📄 Light": cartoLight,
  "🌙 Dark": cartoDark}, {}, { position: 'topleft', collapsed: true }).addTo(map);


let currentSensorMode = "";
let heatLayer = null;
let heatLegend = null;
let cesiumViewer = null;
let trafficMarkers = [];
let poiMarkers = [];


// --------------------------------------
// Weather Sensors
// --------------------------------------
const weatherSensors = [
  { id: 'Sensor:Weather:FacultyCS', label: 'Faculty of Computer Science', coords: [52.13878, 11.64533], marker: null },
  { id: 'Sensor:Weather:ScienceHub', label: 'Science Harbor', coords: [52.14175, 11.65640], marker: null },
  { id: 'Sensor:Weather:UniMensa', label: 'University Mensa', coords: [52.13966, 11.64761], marker: null },
  { id: 'Sensor:Weather:Library', label: 'University Library', coords: [52.13888, 11.64707], marker: null },
  { id: 'Sensor:Weather:WelcomeCenter', label: 'Welcome Center', coords: [52.14031, 11.64039], marker: null },
  { id: 'Sensor:Weather:NorthPark', label: 'North Park', coords: [52.14276, 11.64513], marker: null },
  { id: 'Sensor:Weather:GeschwisterPark', label: 'Geschwister School Park', coords: [52.14020, 11.63655], marker: null },
  { id: 'Sensor:Weather:Walter', label: 'Walter', coords: [52.14123, 11.654583], marker: null },
];

const parkingSpots = [
  { id: "ParkingSpot:ScienceHarbor", label: "Parking A - Science Harbor", coords: [52.1412, 11.6558], marker: null },
  { id: "ParkingSpot:FacultyCS", label: "Parking B - Faculty CS", coords: [52.1383, 11.6448], marker: null },
  { id: "ParkingSpot:NorthPark", label: "Parking C - North Park", coords: [52.1431, 11.6457], marker: null }
];



const trafficPoints = [
  { id: "Traffic:Junction:ScienceHub", label: "🚦 Traffic - Science Harbor", coords: [52.1417, 11.6564], marker: null},
  { id: "Traffic:Junction:FacultyCS", label: "🚦 Traffic - Faculty CS", coords: [52.1387, 11.6453], marker: null}
];

// --------------------------------------
// 2D/3D toggle buttons 
// --------------------------------------

const btn3D = document.getElementById('btn3D');
const btn2D = document.getElementById('btn2D');
const cesiumOverlay = document.getElementById('cesiumOverlay');
Cesium.Ion.defaultAccessToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJqdGkiOiIzY2IwMTEyNS0wYjMyLTQxNGYtOTU3OC1iMmY0YjE2ODlmNWEiLCJpZCI6MzM4MzczLCJpYXQiOjE3NTY5OTIyMjd9.8PKf7IaADHKaJTHTkiDz6mg25IHJa8C9ntk6RErFJoo"


// Initialize Cesium viewer once
async function initCesium() {
  cesiumViewer = new Cesium.Viewer('cesiumContainer', {
    animation: false,
    timeline: false,
    geocoder: false,
    baseLayerPicker: true,
    sceneModePicker: true,
    homeButton: true,
    navigationHelpButton: false,
    infoBox: false,
    selectionIndicator: false,
    shouldAnimate: false
  });

  cesiumViewer.scene.globe.show = true;
  cesiumViewer.scene.globe.enableLighting = false; // unified
  cesiumViewer.scene.highDynamicRange = true;
  cesiumViewer.scene.skyAtmosphere.show = true;

  const hasToken = !!Cesium.Ion.defaultAccessToken && Cesium.Ion.defaultAccessToken.trim() !== "";
  if (hasToken) {
    try {
      const terrain = await Cesium.createWorldTerrainAsync();
      cesiumViewer.terrainProvider = terrain;
      const buildings = await Cesium.createOsmBuildingsAsync();
      cesiumViewer.scene.primitives.add(buildings);
    } catch (e) {
      console.warn("Ion terrain/OSM buildings failed; using ellipsoid.", e);
      cesiumViewer.terrainProvider = new Cesium.EllipsoidTerrainProvider();
    }
  } else {
    cesiumViewer.terrainProvider = new Cesium.EllipsoidTerrainProvider();
    cesiumViewer.scene.globe.depthTestAgainstTerrain = false;
  }

  setTimeout(() => cesiumViewer.resize(), 50);
}
//  



// 2D button: hide 3D and re-render current layer on Leaflet
btn2D.addEventListener('click', () => {
  cesiumOverlay.style.display = 'none';
  cesiumOverlay.setAttribute('aria-hidden', 'true');

  if (currentSensorMode === "parking")       updateParkingSpots();
  else if (currentSensorMode === "traffic")  updateTrafficFlow();
  else if (currentSensorMode === "tempHeat") renderTemperatureHeatmap();
  else if (currentSensorMode === "temperature" || currentSensorMode === "humidity")
    updateAllPopups(currentSensorMode);

  setActiveButton();
});

// --------------------------------------
// 3D button handlers   
// --------------------------------------

btn3D.addEventListener('click', async () => {
  cesiumOverlay.style.display = 'block';
  cesiumOverlay.setAttribute('aria-hidden', 'false');
  if (!cesiumViewer) await initCesium();
  flyCesiumToLeafletCenter();
  setActiveButton();
})


// call setActiveButton() at the end of each btn2D/btn3D handler

function setActiveButton() {
  document.getElementById('btn2D').classList.toggle('mode-active', cesiumOverlay.style.display !== 'block');
  document.getElementById('btn3D').classList.toggle('mode-active', cesiumOverlay.style.display === 'block');
}

// --------------------------------------
// Fly Cesium camera to Leaflet map center
// --------------------------------------

const CESIUM_TARGET = { lat: 52.1385, lon: 11.6453, height: 400, heading: 15, pitch: -35 };

function flyCesiumToLeafletCenter() {
  if (!cesiumViewer) return;

  const center = map.getCenter();  
  const zoom = map.getZoom();

  // Lower height = closer. This mapping lands you near street/building level at high zooms.
  const height = Math.max(80, 100000 / Math.pow(2, (zoom - 8)));

  cesiumViewer.camera.flyTo({
    destination: Cesium.Cartesian3.fromDegrees(CESIUM_TARGET.lon, CESIUM_TARGET.lat, CESIUM_TARGET.height),
    orientation: {
      heading: Cesium.Math.toRadians(0),
      pitch: Cesium.Math.toRadians(-35),  // tilt to see facades
      roll: 0
    },
    duration: 1.2
  });
}


// --------------------------------------
// Fetch sensor data from Orion
// --------------------------------------
async function getSensorData(sensorId) {
  try {
    const res = await fetch(`/entities/${sensorId}`);
    if (!res.ok) {
      console.error("Failed to fetch", sensorId);
      return null;
    }
    return await res.json();
  } catch (err) {
    console.error("Fetch error:", err);
    return null;
  }
}

async function getAllVehicles() {
  try {
    const res = await fetch(`/entities?type=Vehicle&limit=1000`);
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

// --------------------------------------
// Weather tooltips
// --------------------------------------
function attachLivePopup(marker, sensorId, label) {
  marker.on("click", async () => {
    const data = await getSensorData(sensorId);
    if (!data || (!data.temperature?.value && !data.humidity?.value)) {
      marker.setPopupContent(`${label}<br><i>No sensor data</i>`);
      return;
    }

    const popup = `
      <b>${label}</b><br>
      🌡️ Temp: ${data.temperature?.value ?? "n/a"} °C<br>
      💧 Humidity: ${data.humidity?.value ?? "n/a"} %
    `;
    marker.setPopupContent(popup);
    marker.openPopup();
  });
}

// ---------------------------------------
// Update all weather popups
// ---------------------------------------
async function updateAllPopups(sensorType) {
  for (const entry of weatherSensors) {
    const data = await getSensorData(entry.id);
    if (!data || (!data.temperature && !data.humidity)) continue;

    const value = sensorType === "temperature"
      ? `🌡️ ${data.temperature.value} °C`
      : `💧 ${data.humidity.value} %`;

    if (!entry.marker) {
      entry.marker = L.marker(entry.coords).addTo(map);
      attachLivePopup(entry.marker, entry.id, entry.label);
    }

    entry.marker.bindTooltip(`${entry.label}<br>${value}`, {
      permanent: true,
      direction: "top",
      offset: [0, -10]
    }).openTooltip();
  }
}

// --------------------------------------
// Parking
// --------------------------------------
async function updateParkingSpots() {
  for (const spot of parkingSpots) {
    const data = await getSensorData(spot.id);
    if (!data || data.freeSpaces?.value == null || data.totalSpaces?.value == null) continue;

    const label = `${spot.label}<br>🚗 ${data.freeSpaces.value} of ${data.totalSpaces.value} spaces free`;

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
  'robot': createIcon('🤖'),
}

let vehicleMarkers = []
async function updateVehicles() {
  const vehicles = await getAllVehicles();
  vehicleMarkers.forEach(m => m.remove())
  vehicleMarkers = []
  for(const vehicle of vehicles) {
    const icon = icons[vehicle.category.value]
    const marker = L.marker(vehicle.location.value.coordinates, {icon: icon}).addTo(map)
    if(vehicle.category.value == "robot") {
      marker.bindTooltip("🤖 Delivery Robot", { direction: "top", offset: [0, -10] });
    }
    vehicleMarkers.push(marker)
  }
}

// --------------------------------------
// Points of Interest
// --------------------------------------

const poIPlaceMarker = (html)=>L.divIcon({
  html: html,
    className: 'poi-icon',
    iconSize: [64, 64],
    iconAnchor: [32, 32]
}) 

poiIcons = {
  'restaurant':poIPlaceMarker('🍽️'),
  'mensa':poIPlaceMarker('🍲'),
  'cafe':poIPlaceMarker('☕'),
  'supermarket':poIPlaceMarker('🛒'),
  'kiosk':poIPlaceMarker('🗞️'),
  'default':poIPlaceMarker('📍')
}

const poiConfig = ["Restaurant","Mensa","Cafe","Supermarket","Kiosk"] // Extend as needed

async function getPointsOfInterest() {
   try {
    const poIs = poiConfig.join(',') // e.g., "Restaurant,Mensa,Cafe,Supermarket,Kiosk"
    const res = await fetch(`/entities?type=${poIs}&limit=1000`);
    if (!res.ok) {
      console.error("Error fetching points of interest", error);
      return [];
    }
    const json = await res.json();
    return json;

  } catch (error) {
    console.error("Error fetching points of Interest", error);
    return [];
  }
}

function determineCategory(place) {

  const type = place.type ? place.type.toLowerCase() : ""
  const id = place.id ? place.id.toLowerCase() : ""

  if(type.includes("restaurant") || id.includes("restaurant")) return "restaurant"
  else if(type.includes("mensa") || id.includes("mensa")) return "mensa"
  else if(type.includes("cafe") || id.includes("cafe")) return "cafe"  
  else if(type.includes("supermarket") || id.includes("supermarket")) return "supermarket"
  else if(type.includes("kiosk") || id.includes("kiosk")) return "kiosk"
  else return "default";
}

async function updatePointsOfInterest() {
  try{
    const places = await getPointsOfInterest();
  
    poiMarkers.forEach(m => m.remove())
    poiMarkers = []
  
    for(const place of places) {
      const [lat, lon] = place.location.value.split(',').map(s=>parseFloat(s.trim()))
      let categoryKey = determineCategory(place) //for selecting icon
      
      const icon = poiIcons[categoryKey]

      let placeName = place.name?.value || "Unnamed Place"

      let content = `<b>${placeName}</b><br>`

      if(place.cuisine?.value) content += `🍕 Cuisine: ${place.cuisine.value}<br>`

      if(place.opening_hours?.value) content += `⏰ Hours: ${place.opening_hours.value}<br>`

      if(place.accessibility && place.accessibility.value != undefined) {
        const acc = place.accessibility.value.toLowerCase();
        // Handle boolean true/false or string "limited"
        if (acc === "true" || acc === true) content += `♿ Accessible: Yes<br>`;
        else if (acc === "limited") content += `♿ Accessible: ⚠️ Limited<br>`;
        else content += `♿ Accessible: ❌ No<br>`;
      }

      if(place.todays_menu && place.todays_menu.value) {
        content += `<hr style="margin:5px 0;"><b>📅 Today's Menu:</b>`;
            const menu = place.todays_menu.value;
            if (Array.isArray(menu) && menu.length > 0) {
              content += '<ul style="padding-left:15px; margin:5px 0; font-size:0.9em;">';
              menu.forEach(meal => {
                content += `<li>${meal.name_german} <b>(${meal.price})</b></li>`;
              });
              content += '</ul>';
            } else {
              content += '<br><i>No menu available.</i>';
              }
        }

      const marker = L.marker([lat, lon], {icon: icon}).addTo(map)
      marker.bindPopup(content, { direction: "top", offset: [0, -10] }); // Popup on click to read menu easily
      poiMarkers.push(marker)
    }
  } catch(error){
    console.error("Error updating restaurants", error);
  }
  

}
// --------------------------------------
// Cleanup
// --------------------------------------
function clearAllMarkers() {
  weatherSensors.forEach(e => { if (e.marker) { e.marker.remove(); e.marker = null; } });
  parkingSpots.forEach(e => { if (e.marker) { e.marker.remove(); e.marker = null; } });
  //removeHeatmap();
  trafficPoints.forEach(p => { if (p.marker) { map.removeLayer(p.marker); p.marker = null; } });
  trafficMarkers.forEach(marker => map.removeLayer(marker));
  trafficMarkers = [];
  poiMarkers.forEach(marker => map.removeLayer(marker));
  poiMarkers = [];
}

// --------------------------------------
// UI dropdown handler
// --------------------------------------


document.getElementById("sensorType").addEventListener("change", (e) => {
  const selected = e.target.value;
  currentSensorMode = selected;
  clearAllMarkers();

  if (selected === "temperature" || selected === "humidity") {
    updateAllPopups(selected);
  } else if (selected === "parking") {
    updateParkingSpots();
 // } else if (selected === "tempHeat") {
  //  renderTemperatureHeatmap();
  } else if (selected === "traffic") {
    updateTrafficFlow();
  }
  else if(selected === "all-amenities"){
    updatePointsOfInterest();
  }
});

// --------------------------------------
// Extend auto-refresh
// --------------------------------------

setInterval(() => {
  if (currentSensorMode === "parking") {
    updateParkingSpots();
  } else if (currentSensorMode === "temperature" || currentSensorMode === "humidity") {
    updateAllPopups(currentSensorMode);
  //} else if (currentSensorMode === "tempHeat") {
  //  renderTemperatureHeatmap();
  } else if (currentSensorMode === "traffic") {
    updateTrafficFlow();
  } else if (currentSensorMode === "all-amenities") {
    updatePointsOfInterest();
  }
}, 10000);
updateVehicles();
setInterval(updateVehicles, 1000);



// function closeGrafana() {
//   document.getElementById('grafanaSidebar').style.display = 'none';
//   document.getElementById('grafanaFrame').src = '';
// }

// --------------------------------------
// Traffic Flow
// --------------------------------------

async function updateTrafficFlow() {
  for (const point of trafficPoints) {
    const data = await getSensorData(point.id);
    if (!data) continue;

    const vIn  = Number(data.vehiclesIn?.value ?? 0);
    const vOut = Number(data.vehiclesOut?.value ?? 0);
    const cyc  = Number(data.cyclists?.value ?? 0);
    const ped  = Number(data.pedestrians?.value ?? 0);


    const label = `${point.label}<br>
    🚗 In: ${vIn}<br>
    🚙 Out: ${vOut}<br>
    🚴 Cyclists: ${cyc}<br>
    🚶 Pedestrians: ${ped}`;

    if (!point.marker) {
      point.marker = L.circleMarker(point.coords, {
        radius: 10, color: "#ff0000", fillColor: "#f03", fillOpacity: 0.5
      }).addTo(map);
      trafficMarkers.push(point.marker);

   //   // 🔥 Attach click AFTER marker is created
   //   point.marker.on('click', function () {
   //     const encodedId = encodeURIComponent(point.id);
   //     const grafanaURL = `http://localhost:3000/d/aad654c6-b987-4ef1-9466-76d129830a94/vehicle-count-sciencehabor?orgId=1&from=2025-08-26T08:17:40.887Z&to=2025-08-27T08:17:40.887Z&timezone=browser&viewPanel=panel-1&theme=light`;
//
   //     document.getElementById('grafanaSidebar').style.display = 'block';
   //     document.getElementById('grafanaFrame').src = grafanaURL;
   //   });
    }

    point.marker.bindTooltip(label, {
      permanent: true,
      direction: "top",
      offset: [0, -10]
    }).openTooltip();
  }
}

// Activate chatbot
if (initDashbot) {
  initDashbot({
    backendUrl: "/api/dashbot/chat",
    element: "#map",
  })
} else {
  console.error("Dashbot not present")
}

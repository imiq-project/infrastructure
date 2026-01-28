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
function orionUrl(entity) {
  return `<a href="/api/orion/entities/${entity.id}" target="_blank">${entity.id}</a>`
}

async function getSensorData(sensorId) {
  try {
    const res = await fetch(`/api/orion/entities/${sensorId}`);
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

async function getAllSensorsByType(type) {
  try {
    const res = await fetch(`/api/orion/entities?type=${type}&limit=1000`);
    if (!res.ok) {
      console.error("Error fetching sensors", error);
      return null;
    }
    const json = await res.json();
    return json;
  } catch (error) {
    console.error("Error fetching sensors", error);
    return null;
  }  
}

async function getEntityById(id) {
  try {
    const res = await fetch(`/api/orion/entities/${id}`);
    if (!res.ok) {
      console.error("Error fetching entity", error);
      return null;
    }
    const json = await res.json();
    return json;
  } catch (error) {
    console.error("Error fetching entity", error);
    return null;
  }  
}

async function fetchAllTypes() {
  try {
    const res = await fetch(`/api/orion/types`);
    if (!res.ok) {
      console.error("Error fetching types", error);
      return null;
    }
    const json = await res.json();
    return json;
  } catch (error) {
    console.error("Error fetching types", error);
    return null;
  }  
}

// --------------------------------------
// Air Quality
// --------------------------------------
// see https://umwelt.sachsen-anhalt.de/informationen-zum-lqi
function formatAirQualityAttr(value, description, thresholds) {
  colors = ["#3399FF", "#66CCFF", "#FFFF99", "#FF9933", "#FF3333"]
  let i = thresholds.length-1;
  while(value < thresholds[i]) {
    i -= 1;
    if (i == 0) {
      break;
    }
  }
  return `${description}: <span style="color: ${colors[i+1]}">${value} µg/m³</span>`
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

function createVehicleMarker(vehicle) {
    const icon = icons[vehicle.category.value]
    const marker = L.marker(vehicle.location.value.coordinates, {icon: icon}).addTo(map)
    if(vehicle.category.value == "robot") {
      marker.bindPopup("🤖 Delivery Robot")
    }
    return marker;
}

// --------------------------------------
// Traffic Flow
// --------------------------------------

function getTrafficPopupContent(data) {
    const vIn  = Number(data.vehiclesIn?.value ?? 0);
    const vOut = Number(data.vehiclesOut?.value ?? 0);
    const cyc  = Number(data.cyclists?.value ?? 0);
    const ped  = Number(data.pedestrians?.value ?? 0);

    const label = `${orionUrl(data)}<br>
    🚗 In: ${vIn}<br>
    🚙 Out: ${vOut}<br>
    🚴 Cyclists: ${cyc}<br>
    🚶 Pedestrians: ${ped}`;
    
    return label
}

function createTrafficFlowMarker(data) {
    return L.circleMarker(getEntityLocation(data), {
        radius: 10, color: "#ff0000", fillColor: "#f03", fillOpacity: 0.5
    })
}

// --------------------------------------
// Points of Interest
// --------------------------------------
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

function createPoiPopup(place) {
  let placeName = place.name?.value || "Unnamed Place"
  let content = `<b>${placeName}</b><br>`
  if(place.cuisine?.value) content += `🍕 Cuisine: ${place.cuisine.value}<br>`
  if(place.opening_hours?.value) content += `⏰ Hours: ${place.opening_hours.value}<br>`
  if(place.accessibility && place.accessibility.value) {
    const acc = place.accessibility.value.toLowerCase();
    // Handle "yes", "no" or "limited"
    if (acc === "yes" ) content += `♿ Accessible: Yes<br>`;
    else if (acc === "limited") content += `♿ Accessible: ⚠️ Limited<br>`;
    else content += `♿ Accessible: ❌ No<br>`;
  }

  if(place.todaysMenu && place.todaysMenu.value) {
    content += `<hr style="margin:5px 0;"><b>📅 Today's Menu:</b>`;
    const menu = place.todaysMenu.value;
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
  return content
}

// --------------------------------------
// Marker visualization
// --------------------------------------

function getEntityLocation(entity) {
  if (entity.location.type == "geo:point") {
    const loc = entity.location.value.split(",")
    return {lat: loc[0], lon: loc[1]}
  } else if (entity.location.type == "geo:json") {
    const loc = entity.location.value.coordinates
    return {lat: loc[0], lon: loc[1]}
  } else {
    console.error("Cannot extract location: ", entity.location.type)
    return {lat: 0, lon: 0}
  }
}

const chartDialog = document.getElementById('chartDialog')
const chartCanvas = document.getElementById("chartCanvas")
const chartText = document.getElementById("chartText")
let currentChart = null

async function showGraph(entityId) {
    chartDialog.showModal()
    const url = `/api/quantumleap/entities/${entityId}`
    chartText.innerHTML = `<a href="${url}" target="_blank">${entityId}</a>`
    const response = await fetch(url)
    if(!response.ok) {
      chartText.innerHTML += "<br><i>No historical data available.</i>"
      return
    }
    const data = await response.json()
    const datasets = data.attributes.map( e => { return {label: e.attrName, data: e.values} })
    currentChart = new Chart(chartCanvas, {
      type: 'line',
      data: {
        labels: data.index,
        datasets: datasets,
      },
      options: {
        responsive: true,
        scales: {
          x: {
            type: 'time',
            time: {
              tooltipFormat: 'yyyy-MM-dd HH:mm'
            }
          },
          y: {
            beginAtZero: false
          }
        }
      }
    });
}

function closeGraph() {
  if(currentChart) {
    currentChart.destroy();
  }
  currentChart = null;
  chartDialog.close()
}

function graphButton(entity) {
  return `<button onclick="showGraph('${entity.id}')">Graph</button>`
}

function popupFromAttributes(entity, config) {
  config = config || {attrs: {}}

  const attrs = Object.keys(entity).sort().filter(key => !["type", "id", "location"].includes(key))
  const lines = attrs.map(key => {
    const keyConfig = config.attrs ? config.attrs[key] : null
    if (keyConfig) {
      return keyConfig(entity[key].value)
    } else {
      return `${key}: ${entity[key].value}`
    }
  })

  return `${orionUrl(entity)}<br>${lines.join("<br>")}<br>${graphButton(entity)}`
}

function createDefaultMarker(entity) {
  const marker = L.marker(getEntityLocation(entity))
  return marker
}

function getConfigFor(type) {
  const config = {
    "AirQuality": {
      description: "🌬️ Air Quality",
      updateMinutes: 10,
      createMarker: createDefaultMarker,
      getPopupContent: entity => popupFromAttributes(entity, {
        attrs: {
          "pm10": value => formatAirQualityAttr(value, "Feinstaub (10µm)", [20, 35, 50, 100]),
          "pm25": value => formatAirQualityAttr(value, "Feinstaub (2.5µm)", [10, 20, 25, 50]),
          "no2": value => formatAirQualityAttr(value, "Stickstoffdioxid NO2", [20, 40, 100, 200]),
          "o3": value => formatAirQualityAttr(value, "Ozon O3", [60, 120, 180, 240]),
        }
      })
    },
    "Cafe": {
      description: "☕Cafe",
      updateMinutes: 'never',
      createMarker: createDefaultMarker,
      getPopupContent: createPoiPopup,
    },
    "Kiosk": {
      description: "🗞️ Kiosk",
      updateMinutes: 'never',
      createMarker: createDefaultMarker,
      getPopupContent: createPoiPopup,
    },
    "Mensa": {
      description: "🍲Mensa",
      updateMinutes: 60,
      createMarker: createDefaultMarker,
      getPopupContent: createPoiPopup,
    },
    "Parking": {
      description: "🅿️ Parking",
      updateMinutes: 30,
      createMarker: createDefaultMarker,
      getPopupContent: data => `${orionUrl(data)}<br>🚗 ${data.freeSpaces.value} of ${data.totalSpaces.value} spaces free<br>${graphButton(data)}`,
    },
    "Restaurant": {
      description: "🍽️ Restaurant",
      updateMinutes: 'never',
      createMarker: createDefaultMarker,
      getPopupContent: createPoiPopup,
    },
    "Supermarket": {
      description: "🛒Supermarket",
      updateMinutes: 'never',
      createMarker: createDefaultMarker,
      getPopupContent: createPoiPopup,
    },
    "Traffic": {
      description: "🚦Traffic",
      updateMinutes: 5,
      createMarker: createTrafficFlowMarker,
      getPopupContent: getTrafficPopupContent,
    },
    "Vehicle": {
      description: "🚗 Vehicle",
      updateMinutes: 'moving',
      createMarker: createVehicleMarker,
      getPopupContent: null,
    },
    "Weather": {
      description: "🌡️ Weather",
      updateMinutes: 10,
      createMarker: createDefaultMarker,
      getPopupContent: entity => popupFromAttributes(entity, {
        attrs: {
          "humidity": value => `💧 ${value} %`,
          "temperature": value => `🌡️ ${value} °C`,
          }
      }),
    },
  }

  return config[type] || {
     description: `📍 ${type}`,
     updateMinutes: 1,
     createMarker: createDefaultMarker,
     getPopupContent: popupFromAttributes,
  }
}

// --------------------------------------
// Marker handling
// --------------------------------------

const markersById = new Map()

function clearAllMarkers() {
  markersById.forEach(marker => {
    map.removeLayer(marker)
  });
  markersById.clear();
}

function updateAllPopups(data, typeConfig) {
  if (typeConfig.getPopupContent == null) {
    return
  }
  data.forEach( spot => {
      const marker = markersById.get(spot.id)
      const content = typeConfig.getPopupContent(spot)
      marker.setPopupContent(content)
    }
  )
}

function updateAllMarkers(data, typeConfig, fitMap) {
  clearAllMarkers();
  const bounds = L.latLngBounds();
  data.forEach(spot => {
    const marker = typeConfig.createMarker(spot)
    marker.addTo(map);
    bounds.extend(marker.getLatLng());
    if (typeConfig.getPopupContent != null) {
      const content = typeConfig.getPopupContent(spot)
      marker.bindPopup(content)
    }
    markersById.set(spot.id, marker);
    marker.on('popupopen', () => {
      const url = new URL(window.location.href);
      url.searchParams.set('id', spot.id);
      window.history.pushState({}, '', url);
    })
  })
  if (fitMap) {
    map.fitBounds(bounds);
  }
}

// --------------------------------------
// UI dropdown handler
// --------------------------------------
let currentEntityType = ""
let timerId = 0
async function onSensorTypeChanged(selected, targetId) {
  currentEntityType = selected;
  const typeConfig = getConfigFor(currentEntityType)
  data = await getAllSensorsByType(currentEntityType)
  updateAllMarkers(data, typeConfig, true)
  if(targetId == null) {
    targetId = data[0].id
  }
  markersById.get(targetId).openPopup()
  if (timerId) {
    clearTimeout(timerId)
  }

  async function update() {
    if (typeConfig.updateMinutes == 'never') {
      // no new timer
    } else if (typeConfig.updateMinutes == 'moving') {
      data = await getAllSensorsByType(currentEntityType)
      updateAllMarkers(data, typeConfig, false)
      timerId = setTimeout(update, 1_000)
    } else {
      data = await getAllSensorsByType(currentEntityType)
      updateAllPopups(data, typeConfig)
      timerId = setTimeout(update, typeConfig.updateMinutes * 60_000)
    }
  }
  await update()
}

// --------------------------------------
// Initialization
// --------------------------------------

const sensorTypeSelect = document.getElementById("sensorType");
sensorTypeSelect.addEventListener("change", (e) => onSensorTypeChanged(e.target.value, null));

async function init() {
  types = await fetchAllTypes()
  types.forEach(type => {
    const typeConfig = getConfigFor(type.type)
    sensorTypeSelect.insertAdjacentHTML("beforeend", `<option value="${type.type}">${typeConfig.description} (${type.count})</option>`)
  })
  const id = new URLSearchParams(window.location.search).get('id');
  if(id != null) {
    const entity = await getEntityById(id)
    if(entity != null) {
      sensorTypeSelect.value = entity.type
      onSensorTypeChanged(entity.type, id)
    }
  }
}
init().catch(console.log)

// Activate chatbot
if (initDashbot) {
  initDashbot({
    backendUrl: "/api/dashbot/chat",
    element: "#map",
  })
} else {
  console.error("Dashbot not present")
}

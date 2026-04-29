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

// if (map.zoomControl) map.zoomControl.remove();
const isMobileDevice = window.innerWidth <= 768;
const navPosition = isMobileDevice ? 'bottomright' : 'topright';

const map = L.map('map', {
  center: [52.140, 11.644],
  zoom: 14,
  layers: [cartoLight],
  zoomControl: false,
});


// if (map.zoomControl){
  //   map.zoomControl.remove();
  // }
L.control.zoom({ position: navPosition }).addTo(map);
  
if (L.control.fullscreen) {
  L.control.fullscreen({ position: navPosition }).addTo(map); 
}

L.control.layers({
  "📄 Light": cartoLight,
  "🌙 Dark": cartoDark, 
  "🛰️ Satellite": satellite,
  "🗺️ OpenStreetMap": osmRoad,
  "🏔️ Topographic": openTopo,
  }, {}, { position: 'topright', collapsed: true }).addTo(map);
  
const drawerTrigger = document.getElementById('mobileDrawerTrigger');
if (drawerTrigger) {
    drawerTrigger.addEventListener('click', () => {
        const container = document.getElementById('smartView');
        if (container) container.classList.toggle('drawer-open');
    });
}

function toggleSidebar() {

  const container = document.querySelector('.sidebar-container') || document.querySelector('.smart-view-container');
  if (container) {
      // Toggle the classes used in the new CSS
      container.classList.toggle('drawer-open'); 
      container.classList.toggle('open'); 
  }
}

const btnToggle = document.getElementById('btnModeToggle');
async function switchView() {
    const is3D = cesiumOverlay.style.display === 'block';

    if (is3D) {
        // Switch to 2D
        cesiumOverlay.style.display = 'none';
        cesiumOverlay.setAttribute('aria-hidden', 'true');
    } else {
        // Switch to 3D
        cesiumOverlay.style.display = 'block';
        cesiumOverlay.setAttribute('aria-hidden', 'false');
        if (!cesiumViewer) await initCesium();
        flyCesiumToLeafletCenter();
    }
    
    // Update the UI for all buttons
    setActiveButton();
}

['btnDesktopToggle', 'btnMobileToggle', 'btnModeToggle'].forEach(id => {
    const btn = document.getElementById(id);
    if (btn) btn.addEventListener('click', switchView);
});



function setActiveButton() {
    const is3D = cesiumOverlay.style.display === 'block';
    
    // Updates text and color for both mobile and desktop buttons
    ['btnDesktopToggle', 'btnMobileToggle', 'btnModeToggle'].forEach(id => {
        const btn = document.getElementById(id);
        if (btn) {
            btn.classList.toggle('mode-active', is3D);
            btn.innerHTML = is3D ? "🗺️ 2D" : "🧭 3D";
        }
    });
}

// document.addEventListener('DOMContentLoaded', () => {
//   const menuWrapper = document.querySelector('.menu-wrapper');
//   if (menuWrapper) {
//     menuWrapper.addEventListener('click', toggleDrawer);
//   }
// });


// --------------------------------------
// 2D/3D toggle buttons 
// --------------------------------------
let cesiumViewer = null;
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

if (btn2D) {
    btn2D.addEventListener('click', () => {
        cesiumOverlay.style.display = 'none';
        cesiumOverlay.setAttribute('aria-hidden', 'true');
        setActiveButton();
    });
}

if (btn3D) {
    btn3D.addEventListener('click', async () => {
        cesiumOverlay.style.display = 'block';
        cesiumOverlay.setAttribute('aria-hidden', 'false');
        if (!cesiumViewer) await initCesium();
        flyCesiumToLeafletCenter();
        setActiveButton();
    });
}

function setActiveButton() {
    const is3D = cesiumOverlay.style.display === 'block';
    if (btn2D) btn2D.classList.toggle('mode-active', !is3D);
    if (btn3D) btn3D.classList.toggle('mode-active', is3D);
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
    className: 'marker-icon',
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

const shuttleRoute = [
  [52.141099, 11.647549], // Official Starting Point (Mid-East Loop)
  [52.141012, 11.647562], [52.140963, 11.647613], [52.140928, 11.647705], [52.140925, 11.647820], [52.140933, 11.647871], [52.141030, 11.648013], 
  [52.141124, 11.648088], [52.141293, 11.648531], [52.142162, 11.651945], [52.142175, 11.652286], [52.142144, 11.652417], [52.142062, 11.652634],
  [52.141736, 11.653375], [52.141654, 11.653503], [52.141371, 11.653745], [52.141176, 11.654104], [52.141129, 11.654257], [52.141063, 11.654748], 
  [52.141017, 11.655065], [52.140972, 11.655156], [52.141002, 11.655298], [52.140997, 11.655456], [52.140964, 11.655687], [52.140854, 11.656172], 
  [52.141312, 11.656500], [52.142770, 11.657664], [52.147314, 11.660952], [52.146843, 11.662698], [52.142442, 11.658785], [52.142765, 11.657663], 
  [52.142197, 11.657250], [52.141316, 11.656502], [52.140854, 11.656177], [52.140961, 11.655703], [52.141053, 11.655333], [52.141117, 11.655311], 
  [52.141147, 11.655242], [52.141151, 11.655174], [52.141130, 11.655081], [52.141082, 11.655035], [52.141061, 11.654751], [52.141122, 11.654263], 
  [52.141170, 11.654123], [52.141365, 11.653750], [52.141653, 11.653509], [52.141732, 11.653375], [52.142062, 11.652642], [52.142192, 11.652597], 
  [52.142245, 11.652640], [52.142280, 11.652632], [52.142358, 11.652592], [52.142389, 11.652527], [52.142394, 11.652380], [52.142335, 11.652254], 
  [52.142266, 11.652219], [52.142162, 11.651945], [52.141293, 11.648539], [52.141214, 11.648110], [52.141218, 11.648051], [52.141257, 11.647957], 
  [52.141264, 11.647879], [52.141252, 11.647785], [52.141101, 11.647576], 
  //PART 2: TRANSITION TO WEST LOOP
  [52.141000, 11.647546],

  //PART 3: WEST LOOP
  [52.139891, 11.645580], [52.139687, 11.645274], [52.139712, 11.645191], [52.139819, 11.645036], [52.139871, 11.644529], [52.139244, 11.644405],
  [52.139177, 11.644373], [52.139154, 11.644287], [52.139397, 11.641251], [52.141012, 11.641578], [52.140995, 11.641820], [52.141359, 11.641892], 
  [52.141436, 11.641967], [52.141540, 11.642158], [52.141445, 11.642316], [52.141410, 11.642539], [52.141116, 11.647463], [52.141099, 11.647552]
]
function createMovingShuttleBus(path, color) {
    L.polyline(path, {
        color: color, weight: 3, opacity: 0.4, dashArray: '5, 10', interactive: false
    }).addTo(map);

    const marker = L.marker(path[0], {
        icon: L.divIcon({
            html: '🚌', className: 'moving-bus-icon', iconSize: [25, 25], iconAnchor: [12, 12]
        }),
        interactive: false
    }).addTo(map);

    let step = 0;
    const targetSpeed = 12; // meters per second (adjust as needed)

    function startSegment() {
        const startPt = path[step];
        const nextStep = (step + 1) % path.length;
        const endPt = path[nextStep];

        const startLatLng = L.latLng(startPt[0], startPt[1]);
        const endLatLng = L.latLng(endPt[0], endPt[1]);
        const distance = startLatLng.distanceTo(endLatLng);

        const durationInSeconds = distance / targetSpeed;
        const framesRequired = durationInSeconds * 60;
        const increment = 1 / framesRequired;

        let counter = 0.0;

        function frame() {
            if (map._animatingZoom) {
                requestAnimationFrame(frame);
                return;
            }
            counter += increment;

            if (counter <= 1.0) {
                // LERP Formula: Position = Start + (End - Start) * Counter
                const lat = startPt[0] + (endPt[0] - startPt[0]) * counter.toFixed(6);
                const lng = startPt[1] + (endPt[1] - startPt[1]) * counter.toFixed(6);
                marker.setLatLng([lat, lng]);
                requestAnimationFrame(frame);
            } else {
                step = nextStep;
                startSegment();
            }
        }

        frame();
    }

    startSegment();
}
 createMovingShuttleBus(shuttleRoute, "#009933");

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
// Digital Twins
// -------------------------------------
function getDigitalTwinPopup(entity) {
  return `
  <div>
    <img src="${entity.logo.value}" style="width: 100%">
    <br>
    <b>${entity.name.value}</b>
    <br>
    Website: <a href="${entity.website.value}" target="_blank">${entity.website.value}</a>
    <br>
    Data: <a href="${entity.data.value}" target="_blank">${entity.data.value}</a>
  </div>
  `
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
    let url = `/api/quantumleap/entities/${entityId}`
    chartText.innerHTML = `<a href="${url}" target="_blank">${entityId}</a>`

    const oneWeekAgo = new Date(new Date().getTime() - 7 * 24 * 60 * 60 * 1000);
    url += `?fromDate=${oneWeekAgo.toISOString()}`
    const response = await fetch(url)
    if(!response.ok) {
      chartText.innerHTML += "<br><i>No historical data available.</i>"
      return
    }
    const data = await response.json()
    const datasets = data.attributes.map( (e, idx) => { return {label: e.attrName, data: e.values, hidden: idx !== 0} })
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
          },
        },
        plugins: {
          // only show one dataset per time
          legend: {
            onClick: function (e, legendItem, legend) {
              const chart = legend.chart;
              const clickedIndex = legendItem.datasetIndex;
              chart.data.datasets.forEach((dataset, i) => {
                const meta = chart.getDatasetMeta(i);
                meta.hidden = (i !== clickedIndex)
              })
              chart.update()
            }
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

function createIconMarker(entity, icon) {
  const divIcon = L.divIcon({
    html: icon,
    className: 'marker-icon',
    iconSize: [64, 64],
    iconAnchor: [32, 0],
  })
  const marker = L.marker(getEntityLocation(entity), {icon: divIcon})
  return marker;
}

function getConfigFor(type) {
  const config = {
    "AirQuality": {
      description: "🌬️ Air Quality",
      updateMinutes: 10,
      createMarker: (entity) => createIconMarker(entity, '🌬️'),
      getPopupContent: entity => popupFromAttributes(entity, {
        attrs: {
          "pm10": value => formatAirQualityAttr(value, "Feinstaub (10µm)", [20, 35, 50, 100]),
          "pm25": value => formatAirQualityAttr(value, "Feinstaub (2.5µm)", [10, 20, 25, 50]),
          "no2": value => formatAirQualityAttr(value, "Stickstoffdioxid NO2", [20, 40, 100, 200]),
          "o3": value => formatAirQualityAttr(value, "Ozon O3", [60, 120, 180, 240]),
        }
      })
    },
    "Building":  {
      description: "🏠 Building",
      updateMinutes: 'never',
      createMarker: (entity) => createIconMarker(entity, '🏠'),
      getPopupContent: createPoiPopup,
    },
    "Cafe": {
      description: "☕Cafe",
      updateMinutes: 'never',
      createMarker: (entity) => createIconMarker(entity, '☕'),
      getPopupContent: createPoiPopup,
    },
    "Kiosk": {
      description: "🗞️ Kiosk",
      updateMinutes: 'never',
      createMarker: (entity) => createIconMarker(entity, '🗞️'),
      getPopupContent: createPoiPopup,
    },
    "Mensa": {
      description: "🍲Mensa",
      updateMinutes: 60,
      createMarker: (entity) => createIconMarker(entity, '🍲'),
      getPopupContent: createPoiPopup,
    },
    "Parking": {
      description: "🅿️ Parking",
      updateMinutes: 30,
      createMarker: (entity) => createIconMarker(entity, '🅿️'),
      getPopupContent: data => `${orionUrl(data)}<br>🚗 ${data.freeSpots.value} of ${data.totalSpots.value} spaces free<br>${graphButton(data)}`,
    },
    "Restaurant": {
      description: "🍽️ Restaurant",
      updateMinutes: 'never',
      createMarker: (entity) => createIconMarker(entity, '🍽️'),
      getPopupContent: createPoiPopup,
    },
    "Supermarket": {
      description: "🛒Supermarket",
      updateMinutes: 'never',
      createMarker: (entity) => createIconMarker(entity, '🛒'),
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
      createMarker: (entity) => createIconMarker(entity, '🌡️'),
      getPopupContent: entity => popupFromAttributes(entity, {
        attrs: {
          "humidity": value => `💧 ${value} %`,
          "temperature": value => `🌡️ ${value} °C`,
          }
      }),
    },
    "DigitalTwin": {
      description: "🧠 Digital Twin",
      updateMinutes: 'never',
      createMarker: (entity) => createIconMarker(entity, '🧠'),
      getPopupContent: getDigitalTwinPopup,
    },
    "WaterLevel": {
      description: "🌊 Water Level",
      updateMinutes: 60,
      createMarker: (entity) => createIconMarker(entity, '🌊'),
      getPopupContent: popupFromAttributes,
    }
  }

  return config[type] || {
     description: `📍 ${type}`,
     updateMinutes: 1,
     createMarker: (entity) => createIconMarker(entity, '📍'),
     getPopupContent: popupFromAttributes,
  }
}

// --------------------------------------
// Marker handling
// --------------------------------------

const markersByType = new Map()
const timersByType = new Map()

function clearAllMarkers(type, typeConfig) {
  const timerId = timersByType.get(type)
  if (timerId) {
    clearTimeout(timerId)
  }
  timersByType.delete(type)

  const markers = markersByType.get(type) || []
  markers.forEach(marker => {
    map.removeLayer(marker)
  });
  markersByType.delete(type)
}

async function addAllMarkers(type, typeConfig, data, fitMap, targetId) {
  const bounds = L.latLngBounds();
  const markers = []
  data.forEach(spot => {
    const marker = typeConfig.createMarker(spot)
    marker.addTo(map);
    bounds.extend(marker.getLatLng());
    if (typeConfig.getPopupContent != null) {
      const content = typeConfig.getPopupContent(spot)
      marker.bindPopup(content)
    }
    if (spot.id == targetId) {
      marker.openPopup()
    }
    markers.push(marker)
    marker.on('popupopen', () => {
      const url = new URL(window.location.href);
      url.searchParams.set('id', spot.id);
      window.history.pushState({}, '', url);
    })
  })
  markersByType.set(type, markers)
  if (fitMap) {
    map.fitBounds(bounds);
  }

  let interval = null
  if (typeConfig.updateMinutes == 'never') {
    // no new timer
  } else if (typeConfig.updateMinutes == 'moving') {
    interval = 1_000
  } else {
    interval = typeConfig.updateMinutes * 60_000
  }
  if (interval) {
    const timerId = setInterval(async () => {
      data = await getAllSensorsByType(type)
      clearAllMarkers(type, typeConfig)
      addAllMarkers(type, typeConfig, data, false, null)
    }, interval)
    timersByType.set(type, timerId)
  }
}

// --------------------------------------
// UI dropdown handler
// --------------------------------------
async function onSensorSelected(type, selected, targetId) {
  const typeConfig = getConfigFor(type)
  if (selected) {
    const data = await getAllSensorsByType(type)
    addAllMarkers(type, typeConfig, data, false, targetId)
  } else {
    clearAllMarkers(type, typeConfig)
  }
}

// --------------------------------------
// Initialization
// --------------------------------------


async function init() {
  const id = new URLSearchParams(window.location.search).get('id');
  types = await fetchAllTypes()

  const sensorTypesContainer = document.getElementById("sensorTypes");
  // //header label for drawer 
  // const header = document.createElement("h3")
  // header.className = "drawer-title"
  // header.innerText = "Mobile View"
  // header.style.cursor = "pointer"
  // sensorTypesContainer.parentElement.prepend(header);
  sensorTypesContainer.innerHTML = "" 
  
  types.forEach(async type => {
    const typeConfig = getConfigFor(type.type)
    const div = document.createElement("div")
    const input = document.createElement("input")
    input.checked = typeConfig.updateMinutes != "never"
    input.type = "checkbox"
    div.appendChild(input)
    const label = document.createElement("label")
    label.innerText = typeConfig.description
    div.appendChild(label)
    sensorTypesContainer.appendChild(div)
    input.onchange = async event => await onSensorSelected(type.type, event.target.checked, null)
    await onSensorSelected(type.type, input.checked, id)
  })
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

<script setup>
import { VMap, VMapOsmTileLayer, VMapGoogleTileLayer, VMapArcGisAeroTileLayer, VMapZoomControl, VMapMarker, VMapAttributionControl } from 'vue-map-ui';
import { onMounted, getCurrentInstance, ref } from 'vue'
const markers = ref([])
const bounds = ref(null)

async function updateMarkers() {
    const response = await fetch('/entities');
    const entities = await response.json()
    markers.value = entities.map( (i) => i['location']['value']['coordinates'])
    const group = L.featureGroup(markers.value.map(marker =>
        L.marker([marker[0], marker[1]])
    ));
    bounds.value = group.getBounds().pad(0.2);
}

onMounted(async (i) => {
    updateMarkers();
    setInterval(updateMarkers, 1_000);
})
</script>


<template>
    <VMap style="width: 100%; height: 100%;" :bounds="bounds">
        <VMapArcGisAeroTileLayer />
        <!-- <VMapGoogleTileLayer /> -->
        <!-- <VMapArcGisTileLayer /> -->
        <!-- <VMapOsmTileLayer /> -->
        <VMapZoomControl />
        <VMapMarker v-for="i in markers" :latlng="i" />
    </VMap>
</template>

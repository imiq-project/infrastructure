<script setup>
import { VMap, VMapOsmTileLayer, VMapGoogleTileLayer, VMapArcGisAeroTileLayer, VMapZoomControl, VMapMarker, VMapAttributionControl } from 'vue-map-ui';
import { onMounted, defineProps, ref, computed, watch } from 'vue'
const props = defineProps({
    'entities': Array,
})

const markers = computed(() =>
    props.entities.map((i) => i['location']['value']['coordinates'])
)

const bounds=ref(null);

// move map to the first set of valid markers
watch(markers, () => {
    if(bounds.value) return
    const group = L.featureGroup(markers.value.map(marker =>
        L.marker([marker[0], marker[1]])
    ));
    const groupBounds = group.getBounds()
    if (groupBounds.isValid()) {
        bounds.value = groupBounds.pad(.2)
    }
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

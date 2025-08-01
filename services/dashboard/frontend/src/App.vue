<script setup>
import Map from './components/Map.vue'
import EntityList from './components/EntityList.vue'
import Drawer from 'primevue/drawer';
import Button from 'primevue/button';
import { ListCollapse } from 'lucide-vue-next';

import { onMounted, ref } from 'vue'
const entities = ref([])

async function fetchEntities() {
  const response = await fetch('/entities');
  entities.value = await response.json()
}
onMounted(async () => {
  fetchEntities()
  setInterval(fetchEntities, 5_000);
})


const visible = ref(false);

</script>


<template>
  <Drawer v-model:visible="visible" header="Entities" :modal="false">
    <EntityList :entities="entities"></EntityList>
  </Drawer>
  <Button @click="visible = true" style="position: absolute; left: 0; top: 50%; z-index: 1000;" severity="secondary">
    <ListCollapse />
  </Button>
  <ArrowRightFromLine />
  <Map :entities="entities" style="width: 100%;"></Map>
</template>

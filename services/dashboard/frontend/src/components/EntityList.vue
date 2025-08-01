<script setup>
import {defineProps} from 'vue'
import DataTable from 'primevue/datatable';
import Column from 'primevue/column';
defineProps({
    'entities': Array
})
</script>

<template>
<DataTable :value="entities">
    <Column field="id" header="Entity"></Column>
    <Column field="type" header="Type"></Column>
</DataTable>
</template>
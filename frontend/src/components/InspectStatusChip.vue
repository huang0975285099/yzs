<template>
    <q-chip
        v-if="status === 'normal'"
        dense
        color="green-1"
        text-color="green-9"
        class="q-ma-none"
    >复查正常</q-chip>
    <q-chip
        v-else-if="status === 'abnormal'"
        dense
        color="red-1"
        text-color="red-9"
        class="q-ma-none"
    >复查异常</q-chip>
    <q-chip
        v-else
        dense
        color="grey-2"
        text-color="grey-7"
        class="q-ma-none"
    >未复查</q-chip>
</template>

<script setup>
defineProps({
    /** 'normal' | 'abnormal' | null/undefined = 未复查 */
    status: { type: String, default: null },
})
</script>

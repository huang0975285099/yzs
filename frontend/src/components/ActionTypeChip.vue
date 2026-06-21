<template>
    <q-chip
        dense
        class="q-ma-none"
        :color="type === 'pend' ? 'orange-1' : 'blue-1'"
        :text-color="type === 'pend' ? 'orange-9' : 'blue-9'"
    >
        {{ type === 'pend' ? '挂起' : '提交处理' }}
    </q-chip>
</template>

<script setup>
defineProps({
    /** 'submit' | 'pend' */
    type: { type: String, required: true },
})
</script>

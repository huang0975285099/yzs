<template>
    <q-card
        class="stat-card"
        :class="`stat-card--${color}`"
        v-bind="clickable ? { clickable: true } : {}"
        :style="clickable ? 'cursor:pointer' : ''"
    >
        <q-card-section class="text-center" :class="compact ? 'q-pa-sm' : 'q-pa-md'">
            <div class="stat-value">{{ value }}</div>
            <div class="stat-label">
                <slot>{{ label }}</slot>
            </div>
        </q-card-section>
    </q-card>
</template>

<script setup>
defineProps({
    /** Accent color: blue | purple | red | green | orange | teal */
    color: { type: String, default: 'blue' },
    /** The large number displayed */
    value: { type: [String, Number], default: 0 },
    /** Label text below the value (can also use default slot for rich content) */
    label: { type: String, default: '' },
    /** Compact padding (used in dense grids like StatsPage) */
    compact: { type: Boolean, default: false },
    /** Shows pointer cursor (for clickable cards) */
    clickable: { type: Boolean, default: false },
})
</script>

<template>
    <q-page class="q-pa-md">
        <!-- Stat cards -->
        <div class="row q-col-gutter-md q-mb-md">
            <div class="col-6 col-sm-3">
                <StatCard color="blue" :value="stats.total" label="异常订单总数" />
            </div>
            <div class="col-6 col-sm-3">
                <StatCard color="red" :value="stats.unhandledCount" label="总待处理" />
            </div>
            <div class="col-6 col-sm-3">
                <StatCard color="green" :value="stats.handledCount" label="总已处理" />
            </div>
            <div class="col-6 col-sm-3">
                <StatCard color="purple" :value="stats.todayCount" label="今日新增" />
            </div>
        </div>
        <!-- Charts row 1 -->
        <div class="row q-col-gutter-md q-mb-md">
            <div class="col-12 col-sm-8">
                <q-card>
                    <q-card-section class="q-pb-xs">
                        <div class="text-subtitle1 text-weight-medium">
                            近30天异常订单趋势
                        </div>
                    </q-card-section>
                    <q-card-section>
                        <div ref="lineChartRef" class="chart-box" />
                    </q-card-section>
                </q-card>
            </div>
            <div class="col-12 col-sm-4">
                <q-card>
                    <q-card-section class="q-pb-xs">
                        <div class="text-subtitle1 text-weight-medium">
                            异常类型分布
                        </div>
                    </q-card-section>
                    <q-card-section>
                        <div ref="pieChartRef" class="chart-box" />
                    </q-card-section>
                </q-card>
            </div>
        </div>

        <!-- Charts row 2 -->
        <div class="row q-col-gutter-md">
            <div class="col-12 col-sm-4">
                <q-card>
                    <q-card-section class="q-pb-xs">
                        <div class="text-subtitle1 text-weight-medium">
                            处理状态
                        </div>
                    </q-card-section>
                    <q-card-section>
                        <div ref="statusChartRef" class="chart-box" />
                    </q-card-section>
                </q-card>
            </div>
            <div
                v-if="authStore.user?.role !== 'operator'"
                class="col-12 col-sm-8"
            >
                <q-card>
                    <q-card-section class="q-pb-xs">
                        <div class="text-subtitle1 text-weight-medium">
                            操作员处理量排行
                        </div>
                    </q-card-section>
                    <q-card-section>
                        <div ref="opChartRef" class="chart-box" />
                    </q-card-section>
                </q-card>
            </div>
        </div>
    </q-page>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick } from "vue";
import * as echarts from "echarts";
import { statsApi } from "../api";
import { useAuthStore } from "../stores/auth";
import StatCard from "../components/StatCard.vue";

const authStore = useAuthStore();

const stats = ref({
    total: 0,
    todayCount: 0,
    handledCount: 0,
    unhandledCount: 0,
});

const lineChartRef = ref(null);
const pieChartRef = ref(null);
const statusChartRef = ref(null);
const opChartRef = ref(null);

let lineChart, pieChart, statusChart, opChart;

onMounted(async () => {
    try {
        const res = await statsApi.get();
        const data = res.data;
        stats.value = data;
        await nextTick();
        initCharts(data);
    } catch {}
});

onUnmounted(() => {
    [lineChart, pieChart, statusChart, opChart].forEach((c) => c?.dispose());
});

function initCharts(data) {
    lineChart = echarts.init(lineChartRef.value);
    lineChart.setOption({
        tooltip: { trigger: "axis" },
        grid: { left: 40, right: 20, top: 20, bottom: 30 },
        xAxis: {
            type: "category",
            data: (data.dailyCounts || []).map((d) => {
                const parts = (d.day || "").split("T")[0].split("-");
                return parts.length === 3 ? `${parts[1]}-${parts[2]}` : d.day;
            }),
            axisLabel: { fontSize: 12 },
        },
        yAxis: { type: "value", minInterval: 1 },
        series: [
            {
                name: "异常订单列表",
                type: "line",
                smooth: true,
                data: (data.dailyCounts || []).map((d) => d.count),
                areaStyle: { opacity: 0.15 },
                lineStyle: { width: 2 },
                itemStyle: { color: "#1890ff" },
            },
        ],
    });

    pieChart = echarts.init(pieChartRef.value);
    pieChart.setOption({
        tooltip: { trigger: "item", formatter: "{b}: {c} ({d}%)" },
        legend: { bottom: 0, type: "scroll" },
        series: [
            {
                type: "pie",
                radius: ["40%", "70%"],
                center: ["50%", "45%"],
                data: data.typeCounts || [],
                label: { show: false },
                emphasis: {
                    label: { show: true, fontSize: 14, fontWeight: "bold" },
                },
            },
        ],
    });

    statusChart = echarts.init(statusChartRef.value);
    statusChart.setOption({
        tooltip: { trigger: "item", formatter: "{b}: {c} ({d}%)" },
        legend: { bottom: 0 },
        series: [
            {
                type: "pie",
                radius: ["45%", "72%"],
                center: ["50%", "45%"],
                data: [
                    {
                        name: "待处理",
                        value: data.unhandledCount,
                        itemStyle: { color: "#ff4d4f" },
                    },
                    {
                        name: "已处理",
                        value: data.handledCount,
                        itemStyle: { color: "#52c41a" },
                    },
                ],
                label: { show: false },
                emphasis: {
                    label: { show: true, fontSize: 14, fontWeight: "bold" },
                },
            },
        ],
    });

    if (opChartRef.value && data.opStats?.length) {
        opChart = echarts.init(opChartRef.value);
        opChart.setOption({
            tooltip: { trigger: "axis" },
            grid: { left: 60, right: 20, top: 20, bottom: 30 },
            xAxis: {
                type: "category",
                data: (data.opStats || []).map((d) => d.name),
            },
            yAxis: { type: "value", minInterval: 1 },
            series: [
                {
                    name: "处理量",
                    type: "bar",
                    barMaxWidth: 50,
                    data: (data.opStats || []).map((d) => d.value),
                    itemStyle: { color: "#1890ff", borderRadius: [4, 4, 0, 0] },
                    label: { show: true, position: "top" },
                },
            ],
        });
    }
}
</script>

<template>
    <q-page class="q-pa-md">
        <q-card>
            <q-card-section class="q-pb-xs">
                <div class="text-subtitle1 text-weight-medium">
                    每小时订单分布情况
                </div>
            </q-card-section>
            <q-card-section>
                <div ref="lineChartRef" class="chart-box" />
            </q-card-section>
        </q-card>
        <q-card flat bordered>
            <q-card-section>
                <!-- Filter bar - Responsive Grid -->
                <div class="row q-col-gutter-sm q-mb-md">
                    <!-- 关键词搜索 -->
                    <div class="col-6 col-sm-6 col-md-4">
                        <q-input
                            v-model="filter.keyword"
                            outlined
                            dense
                            label="节点名/订单号/交易号"
                            clearable
                            class="full-width"
                            @keyup.enter="handleSearch"
                        >
                            <template #prepend>
                                <q-icon name="search" />
                            </template>
                        </q-input>
                    </div>

                    <!-- 处理状态 -->
                    <div class="col-6 col-sm-6 col-md-2">
                        <q-select
                            v-model="filter.isHandled"
                            outlined
                            dense
                            clearable
                            label="处理状态"
                            :options="statusOptions"
                            emit-value
                            map-options
                            class="full-width"
                        />
                    </div>

                    <!-- 日期筛选 -->
                    <div class="col-6 col-sm-3 col-md-2">
                        <q-input
                            v-model="filter.startDate"
                            outlined
                            dense
                            type="date"
                            label="开始日期"
                            class="full-width"
                        />
                    </div>
                    <div class="col-6 col-sm-3 col-md-2">
                        <q-input
                            v-model="filter.endDate"
                            outlined
                            dense
                            type="date"
                            label="结束日期"
                            class="full-width"
                        />
                    </div>

                    <!-- 操作按钮 -->
                    <div class="col-12 col-sm-6 col-md-2 flex flex-center">
                        <div class="row q-gutter-sm">
                            <q-btn
                                color="primary"
                                icon="search"
                                label="查询"
                                unelevated
                                @click="handleSearch"
                            />
                            <q-btn
                                flat
                                icon="refresh"
                                label="重置"
                                @click="handleReset"
                            />
                        </div>
                    </div>
                </div>

                <!-- Status chips -->
                <div class="row items-center q-gutter-sm q-mb-md">
                    <q-chip
                        v-if="lastSyncTime"
                        dense
                        color="grey-3"
                        text-color="grey-8"
                        icon="sync"
                    >
                        最近同步：{{ formatTime(lastSyncTime) }}
                    </q-chip>
                    <q-chip
                        dense
                        color="green-1"
                        text-color="green-9"
                        icon="schedule"
                        >每30分钟自动同步</q-chip
                    >
                </div>

                <!-- Table -->
                <div class="table-scroll">
                    <q-table
                        :rows="records"
                        :columns="columns"
                        :loading="loading"
                        :pagination="qPagination"
                        :rows-per-page-options="[10, 20, 30, 50]"
                        row-key="id"
                        flat
                        bordered
                        separator="cell"
                        @update:pagination="qPagination = $event"
                        @request="onRequest"
                    >
                        <template #body-cell-abnormalTypeDesc="props">
                            <q-td :props="props">
                                <q-chip
                                    dense
                                    color="orange-1"
                                    text-color="orange-9"
                                    class="q-ma-none"
                                >
                                    {{ props.row.abnormalTypeDesc }}
                                </q-chip>
                            </q-td>
                        </template>
                        <template #body-cell-isHandled="props">
                            <q-td :props="props" align="center">
                                <q-chip
                                    v-if="props.row.isHandled"
                                    dense
                                    color="green-1"
                                    text-color="green-9"
                                    icon="check_circle"
                                    class="q-ma-none"
                                    >已处理</q-chip
                                >
                                <q-chip
                                    v-else
                                    dense
                                    color="red-1"
                                    text-color="red-9"
                                    icon="cancel"
                                    class="q-ma-none"
                                    >未处理</q-chip
                                >
                            </q-td>
                        </template>
                        <template #body-cell-handledAt="props">
                            <q-td :props="props">{{
                                formatTime(props.row.handledAt)
                            }}</q-td>
                        </template>
                        <template #body-cell-handleDuration="props">
                            <q-td :props="props" align="center">
                                {{
                                    props.row.handleDuration
                                        ? formatDuration(
                                              props.row.handleDuration,
                                          )
                                        : "—"
                                }}
                            </q-td>
                        </template>
                        <template #body-cell-videoDuration="props">
                            <q-td :props="props" align="center">
                                {{ formatVideoDuration(props.row.videoDuration) }}
                            </q-td>
                        </template>
                        <template #body-cell-handleGoods="props">
                            <q-td :props="props" align="center">
                                <span
                                    v-if="
                                        goodsSummary(props.row.handleGoods)
                                            .detail
                                    "
                                >
                                    <q-tooltip>{{
                                        goodsSummary(props.row.handleGoods)
                                            .detail
                                    }}</q-tooltip>
                                    {{
                                        goodsSummary(props.row.handleGoods)
                                            .label
                                    }}
                                </span>
                                <span v-else>—</span>
                            </q-td>
                        </template>
                        <template #body-cell-syncedAt="props">
                            <q-td :props="props">{{
                                formatTime(props.row.syncedAt)
                            }}</q-td>
                        </template>
                    </q-table>
                </div>
            </q-card-section>
        </q-card>
    </q-page>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, nextTick, watch } from "vue";
import * as echarts from "echarts";
import { useQuasar } from "quasar";
import { tradeApi } from "../api";

const $q = useQuasar();
const loading = ref(false);
const records = ref([]);
const lastSyncTime = ref("");

const filter = reactive({
    keyword: "",
    isHandled: "",
    startDate: new Date(Date.now() - 7 * 24 * 3600 * 1000).toISOString().slice(0, 10),
    endDate: new Date().toISOString().slice(0, 10),
});

const qPagination = ref({ page: 1, rowsPerPage: 20, rowsNumber: 0 });
const statusOptions = [
    { label: "未处理", value: "false" },
    { label: "已处理", value: "true" },
];

const lineChartRef = ref(null);
let lineChart = null;

const columns = [
    {
        name: "id",
        label: "ID",
        field: "id",
        align: "center",
        style: "width:70px",
    },
    {
        name: "tradeId",
        label: "订单编号",
        field: "tradeId",
        align: "left",
        style: "width:90px",
    },
    {
        name: "createTime",
        label: "创建时间",
        field: "createTime",
        align: "left",
        style: "width:165px",
    },
    {
        name: "isHandled",
        label: "处理状态",
        field: "isHandled",
        align: "center",
        style: "width:100px",
    },
    {
        name: "handledByName",
        label: "处理人",
        field: "handledByName",
        align: "left",
        style: "width:100px",
    },
    {
        name: "handledAt",
        label: "处理时间",
        field: "handledAt",
        align: "left",
        style: "width:165px",
    },
    {
        name: "handleDuration",
        label: "作业时长",
        field: "handleDuration",
        align: "center",
        style: "width:90px",
    },
    {
        name: "videoDuration",
        label: "视频时长",
        field: "videoDuration",
        align: "center",
        style: "width:90px",
    },
    {
        name: "handleGoods",
        label: "处理商品",
        field: "handleGoods",
        align: "center",
        style: "width:120px",
    },
    {
        name: "nodeName",
        label: "节点名称",
        field: "nodeName",
        align: "left",
        style: "min-width:180px",
    },
    {
        name: "innerCode",
        label: "设备编号",
        field: "innerCode",
        align: "left",
        style: "width:110px",
    },
    {
        name: "abnormalTypeDesc",
        label: "异常类型",
        field: "abnormalTypeDesc",
        align: "center",
        style: "width:110px",
    },
    {
        name: "abnormalDesc",
        label: "异常描述",
        field: "abnormalDesc",
        align: "left",
        style: "min-width:160px",
    },
    {
        name: "tradeStatusDesc",
        label: "交易状态",
        field: "tradeStatusDesc",
        align: "left",
        style: "width:100px",
    },
    {
        name: "operatingModeDesc",
        label: "运营模式",
        field: "operatingModeDesc",
        align: "left",
        style: "min-width:140px",
    },
    {
        name: "outOrderNo",
        label: "商户单号",
        field: "outOrderNo",
        align: "left",
        style: "min-width:190px",
    },
    {
        name: "transactionId",
        label: "交易流水号",
        field: "transactionId",
        align: "left",
        style: "min-width:190px",
    },
    {
        name: "doorOpenTime",
        label: "开门时间",
        field: "doorOpenTime",
        align: "left",
        style: "width:165px",
    },
    {
        name: "doorCloseTime",
        label: "关门时间",
        field: "doorCloseTime",
        align: "left",
        style: "width:165px",
    },
    {
        name: "pendStatusDesc",
        label: "挂起状态",
        field: "pendStatusDesc",
        align: "center",
        style: "width:100px",
    },
    {
        name: "syncedAt",
        label: "同步时间",
        field: "syncedAt",
        align: "left",
        style: "width:165px",
    },
];

onMounted(() => {
    fetchData();
    fetchHourlyStats();
});

onUnmounted(() => {
    if (lineChart) {
        lineChart.dispose();
    }
});

async function fetchData() {
    loading.value = true;
    try {
        const params = {
            page: qPagination.value.page,
            size: qPagination.value.rowsPerPage,
            keyword: filter.keyword || undefined,
            isHandled: filter.isHandled || undefined,
            startDate: filter.startDate || undefined,
            endDate: filter.endDate || undefined,
        };
        const res = await tradeApi.list(params);
        records.value = res.data.records;
        qPagination.value = { ...qPagination.value, rowsNumber: res.data.total };
        if (res.data.lastSyncTime) lastSyncTime.value = res.data.lastSyncTime;
    } catch (err) {
        $q.notify({ type: "negative", message: err?.message || "加载失败" });
    } finally {
        loading.value = false;
    }
}

async function fetchHourlyStats() {
    try {
        const params = {
            keyword: filter.keyword || undefined,
            isHandled: filter.isHandled || undefined,
            startDate: filter.startDate || undefined,
            endDate: filter.endDate || undefined,
        };
        const res = await tradeApi.hourlyStats(params);
        await nextTick();
        renderChart(res.data || []);
    } catch (err) {
        console.error("Failed to fetch hourly stats:", err);
    }
}

function renderChart(data) {
    if (!lineChartRef.value) return;
    if (!lineChart) {
        lineChart = echarts.init(lineChartRef.value);
    }
    
    lineChart.setOption({
        tooltip: { trigger: "axis" },
        grid: { left: 40, right: 20, top: 20, bottom: 30 },
        xAxis: {
            type: "category",
            data: data.map(d => `${d.hour}时`),
            axisLabel: { fontSize: 12 },
        },
        yAxis: { type: "value", minInterval: 1 },
        series: [
            {
                name: "订单数量",
                type: "bar",
                data: data.map(d => d.count),
                itemStyle: { color: "#1890ff" },
            },
        ],
    });
}

function handleSearch() {
    qPagination.value = { ...qPagination.value, page: 1 };
    fetchData();
    fetchHourlyStats();
}

function handleReset() {
    filter.keyword = "";
    filter.isHandled = "";
    filter.startDate = "";
    filter.endDate = "";
    qPagination.value = { ...qPagination.value, page: 1 };
    fetchData();
    fetchHourlyStats();
}

async function onRequest(props) {
    const { page, rowsPerPage } = props.pagination;
    qPagination.value = { ...qPagination.value, page, rowsPerPage };
    await fetchData();
}

function formatTime(t) {
    if (!t) return "—";
    return new Date(t).toLocaleString("zh-CN");
}

function formatDuration(secs) {
    if (!secs) return "—";
    const m = Math.floor(secs / 60);
    const s = secs % 60;
    return m > 0 ? `${m}分${s}秒` : `${s}秒`;
}

function formatVideoDuration(v) {
    if (v === null || v === undefined) return "—";
    if (v === 0) return "无视频";
    if (v === 1) return "解析失败";
    const m = Math.floor(v / 60);
    const s = v % 60;
    return m > 0 ? `${m}分${s}秒` : `${s}秒`;
}

function goodsSummary(json) {
    if (!json) return { label: "", detail: "" };
    try {
        const goods = JSON.parse(json);
        if (!Array.isArray(goods) || goods.length === 0)
            return { label: "无消费", detail: "" };
        const label = `${goods.length}种商品`;
        const detail = goods
            .map(
                (g) =>
                    `${g.goodsName}×${g.goodsCount} ¥${((g.goodsPrice || 0) * (g.goodsCount || 0)).toFixed(2)}`,
            )
            .join("\n");
        return { label, detail };
    } catch {
        return { label: "", detail: "" };
    }
}
</script>

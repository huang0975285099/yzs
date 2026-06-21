<template>
    <q-page class="q-pa-md">
        <!-- Overview cards row 1: Operator stats -->
        <div class="row q-col-gutter-sm q-mb-sm">
            <div class="col-6 col-sm-4 col-md-2">
                <StatCard compact color="blue" :value="overview.operatorCount" label="操作员人数" />
            </div>
            <div class="col-6 col-sm-4 col-md-2">
                <StatCard compact clickable color="green" :value="overview.totalSubmit" @click="openDailyDialog('submit')">
                    总提交处理 <span class="text-caption">↗</span>
                </StatCard>
            </div>
            <div class="col-6 col-sm-4 col-md-2">
                <StatCard compact clickable color="orange" :value="overview.totalPend" @click="openDailyDialog('pend')">
                    总挂起 <span class="text-caption">↗</span>
                </StatCard>
            </div>
            <div class="col-6 col-sm-4 col-md-2">
                <StatCard compact color="purple" :value="overview.todaySubmit" label="今日提交处理" />
            </div>
            <div class="col-6 col-sm-4 col-md-2">
                <StatCard compact color="red" :value="overview.todayPend" label="今日挂起" />
            </div>
        </div>

        <!-- Overview cards row 2: Inspector stats -->
        <div class="row q-col-gutter-sm q-mb-md">
            <div class="col-6 col-sm-4 col-md-2">
                <StatCard compact color="teal" :value="inspectorOverview.inspectorCount" label="质检员人数" />
            </div>
            <div class="col-6 col-sm-4 col-md-2">
                <StatCard compact color="green" :value="inspectorOverview.todayNormal" label="今日复查正常" />
            </div>
            <div class="col-6 col-sm-4 col-md-2">
                <StatCard compact color="red" :value="inspectorOverview.todayAbnormal" label="今日复查异常" />
            </div>
            <div class="col-6 col-sm-4 col-md-2">
                <StatCard compact color="blue" :value="inspectorOverview.todayTotal" label="今日复查总计" />
            </div>
        </div>

        <!-- Operator stats table -->
        <q-card flat bordered class="q-mb-md">
            <q-card-section class="q-pb-xs row items-center justify-between">
                <div class="text-subtitle1 text-weight-medium">
                    操作员数据统计
                </div>
                <q-btn
                    flat
                    size="sm"
                    icon="refresh"
                    label="刷新"
                    :loading="statsLoading"
                    @click="fetchStats"
                />
            </q-card-section>
            <q-card-section class="q-pt-xs">
                <div class="table-scroll">
                    <q-table
                        :rows="operatorStats"
                        :columns="opColumns"
                        :loading="statsLoading"
                        row-key="userId"
                        flat
                        bordered
                        dense
                        :rows-per-page-options="[0]"
                        hide-pagination
                    >
                        <template #body-cell-todayStart="props">
                            <q-td :props="props" align="center">
                                <span class="text-primary text-weight-medium">{{
                                    props.row.todayStart
                                }}</span>
                            </q-td>
                        </template>
                        <template #body-cell-todaySkip="props">
                            <q-td :props="props" align="center">
                                <span class="text-grey-7">{{
                                    props.row.todaySkip
                                }}</span>
                            </q-td>
                        </template>
                        <template #body-cell-todaySubmit="props">
                            <q-td :props="props" align="center">
                                <span
                                    class="text-positive text-weight-medium"
                                    >{{ props.row.todaySubmit }}</span
                                >
                            </q-td>
                        </template>
                        <template #body-cell-todayPend="props">
                            <q-td :props="props" align="center">
                                <span class="text-warning text-weight-medium">{{
                                    props.row.todayPend
                                }}</span>
                            </q-td>
                        </template>
                        <template #body-cell-actions="props">
                            <q-td :props="props">
                                <q-btn
                                    flat
                                    dense
                                    size="sm"
                                    color="primary"
                                    label="每日统计"
                                    @click="openOperatorDailyDialog(props.row)"
                                />
                            </q-td>
                        </template>
                    </q-table>
                </div>
            </q-card-section>
        </q-card>

        <!-- Operator range stats -->
        <q-card flat bordered class="q-mb-md">
            <q-card-section class="q-pb-xs">
                <div class="text-subtitle1 text-weight-medium">
                    操作员时段统计
                </div>
            </q-card-section>
            <q-card-section class="q-pt-xs">
                <div class="row q-col-gutter-sm q-mb-sm items-end">
                    <!-- 操作员选择 -->
                    <div class="col-12 col-sm-12 col-md-4">
                        <q-select
                            v-model="range.userIds"
                            outlined
                            dense
                            multiple
                            use-chips
                            :options="operatorOptions"
                            option-value="userId"
                            option-label="label"
                            emit-value
                            map-options
                            label="选择操作员（可多选）"
                            class="full-width"
                        />
                    </div>

                    <!-- 时间选择 -->
                    <div class="col-6 col-md-3">
                        <q-input
                            v-model="range.startTime"
                            outlined
                            dense
                            type="datetime-local"
                            label="开始时间"
                            class="full-width"
                        />
                    </div>
                    <div class="col-6 col-md-3">
                        <q-input
                            v-model="range.endTime"
                            outlined
                            dense
                            type="datetime-local"
                            label="结束时间"
                            class="full-width"
                        />
                    </div>

                    <!-- 查询按钮 -->
                    <div class="col-12 col-sm-12 col-md-2 flex flex-center">
                        <div class="row q-gutter-sm">
                            <q-btn
                                color="primary"
                                label="查询"
                                unelevated
                                :loading="range.loading"
                                @click="fetchRangeStats"
                            />
                            <q-btn
                                flat
                                icon="download"
                                label="导出"
                                :disable="range.result.length === 0"
                                @click="exportOperatorRangeExcel"
                            />
                        </div>
                    </div>
                </div>
                <div class="table-scroll" v-if="range.result.length > 0">
                    <q-table
                        :rows="rangeTotalRows"
                        :columns="rangeColumns"
                        row-key="username"
                        flat
                        bordered
                        dense
                        :rows-per-page-options="[0]"
                        hide-pagination
                    >
                        <template #body-cell-startCount="props">
                            <q-td :props="props" align="center">
                                <span
                                    :class="
                                        props.row.isTotal
                                            ? 'text-weight-bold text-primary'
                                            : 'text-primary'
                                    "
                                    >{{ props.row.startCount }}</span
                                >
                            </q-td>
                        </template>
                        <template #body-cell-skipCount="props">
                            <q-td :props="props" align="center">
                                <span
                                    :class="
                                        props.row.isTotal
                                            ? 'text-weight-bold text-grey-7'
                                            : 'text-grey-7'
                                    "
                                    >{{ props.row.skipCount }}</span
                                >
                            </q-td>
                        </template>
                        <template #body-cell-submitCount="props">
                            <q-td :props="props" align="center">
                                <span
                                    :class="
                                        props.row.isTotal
                                            ? 'text-weight-bold text-positive'
                                            : 'text-positive'
                                    "
                                    >{{ props.row.submitCount }}</span
                                >
                            </q-td>
                        </template>
                        <template #body-cell-pendCount="props">
                            <q-td :props="props" align="center">
                                <span
                                    :class="
                                        props.row.isTotal
                                            ? 'text-weight-bold text-warning'
                                            : 'text-warning'
                                    "
                                    >{{ props.row.pendCount }}</span
                                >
                            </q-td>
                        </template>
                    </q-table>
                </div>
                <div
                    v-else-if="range.queried"
                    class="text-grey text-center q-pa-md"
                >
                    暂无数据
                </div>
            </q-card-section>
        </q-card>

        <!-- Inspector stats table -->
        <q-card flat bordered class="q-mb-md">
            <q-card-section class="q-pb-xs row items-center justify-between">
                <div class="text-subtitle1 text-weight-medium">
                    质检员数据统计
                </div>
                <q-btn
                    flat
                    size="sm"
                    icon="refresh"
                    label="刷新"
                    :loading="inspectorLoading"
                    @click="fetchInspectorStats"
                />
            </q-card-section>
            <q-card-section class="q-pt-xs">
                <div class="table-scroll">
                    <q-table
                        :rows="inspectorStats"
                        :columns="insColumns"
                        :loading="inspectorLoading"
                        row-key="userId"
                        flat
                        bordered
                        dense
                        :rows-per-page-options="[0]"
                        hide-pagination
                    >
                        <template #body-cell-todayNormal="props">
                            <q-td :props="props" align="center">
                                <span
                                    class="text-positive text-weight-medium"
                                    >{{ props.row.todayNormal }}</span
                                >
                            </q-td>
                        </template>
                        <template #body-cell-todayAbnormal="props">
                            <q-td :props="props" align="center">
                                <span
                                    class="text-negative text-weight-medium"
                                    >{{ props.row.todayAbnormal }}</span
                                >
                            </q-td>
                        </template>
                        <template #body-cell-todayTotal="props">
                            <q-td :props="props" align="center">
                                <span class="text-primary text-weight-medium">{{
                                    props.row.todayTotal
                                }}</span>
                            </q-td>
                        </template>
                    </q-table>
                </div>
            </q-card-section>
        </q-card>

        <!-- Inspector range stats -->
        <q-card flat bordered class="q-mb-md">
            <q-card-section class="q-pb-xs">
                <div class="text-subtitle1 text-weight-medium">
                    质检员时段统计
                </div>
            </q-card-section>
            <q-card-section class="q-pt-xs">
                <div class="row q-col-gutter-sm q-mb-sm items-end flex-wrap">
                    <div class="col-12 col-sm-12 col-md-4">
                        <q-select
                            v-model="inspectorRange.userIds"
                            outlined
                            dense
                            multiple
                            use-chips
                            :options="inspectorOptions"
                            option-value="userId"
                            option-label="label"
                            emit-value
                            map-options
                            label="选择质检员（可多选）"
                            class="full-width"
                        />
                    </div>

                    <!-- 时间选择 -->
                    <div class="col-6 col-md-3">
                        <q-input
                            v-model="inspectorRange.startTime"
                            outlined
                            dense
                            type="datetime-local"
                            label="开始时间"
                            class="full-width"
                        />
                    </div>
                    <div class="col-6 col-md-3">
                        <q-input
                            v-model="inspectorRange.endTime"
                            outlined
                            dense
                            type="datetime-local"
                            label="结束时间"
                            class="full-width"
                        />
                    </div>

                    <!-- 查询按钮 -->
                    <div class="col-12 col-sm-12 col-md-2 flex flex-center">
                        <div class="row q-gutter-sm">
                            <q-btn
                                color="primary"
                                label="查询"
                                unelevated
                                :loading="inspectorRange.loading"
                                @click="fetchInspectorRangeStats"
                            />
                            <q-btn
                                flat
                                icon="download"
                                label="导出"
                                :disable="inspectorRange.result.length === 0"
                                @click="exportInspectorRangeExcel"
                            />
                        </div>
                    </div>
                </div>
                <div
                    class="table-scroll"
                    v-if="inspectorRange.result.length > 0"
                >
                    <q-table
                        :rows="insRangeTotalRows"
                        :columns="insRangeColumns"
                        row-key="username"
                        flat
                        bordered
                        dense
                        :rows-per-page-options="[0]"
                        hide-pagination
                    >
                        <template #body-cell-normalCount="props">
                            <q-td :props="props" align="center">
                                <span
                                    :class="
                                        props.row.isTotal
                                            ? 'text-weight-bold text-positive'
                                            : 'text-positive'
                                    "
                                    >{{ props.row.normalCount }}</span
                                >
                            </q-td>
                        </template>
                        <template #body-cell-abnormalCount="props">
                            <q-td :props="props" align="center">
                                <span
                                    :class="
                                        props.row.isTotal
                                            ? 'text-weight-bold text-negative'
                                            : 'text-negative'
                                    "
                                    >{{ props.row.abnormalCount }}</span
                                >
                            </q-td>
                        </template>
                        <template #body-cell-total="props">
                            <q-td :props="props" align="center">
                                <span
                                    :class="
                                        props.row.isTotal
                                            ? 'text-weight-bold text-primary'
                                            : 'text-primary'
                                    "
                                    >{{ props.row.total }}</span
                                >
                            </q-td>
                        </template>
                    </q-table>
                </div>
                <div
                    v-else-if="inspectorRange.queried"
                    class="text-grey text-center q-pa-md"
                >
                    暂无数据
                </div>
            </q-card-section>
        </q-card>

        <!-- Daily stats dialog -->
        <q-dialog v-model="dailyDialog.visible" :maximized="$q.screen.lt.sm">
                <q-card class="full-width" style="max-width: 800px">
                <q-card-section class="row items-center q-pb-none">
                    <div class="text-h6">
                        {{
                            dailyDialog.operatorName
                                ? `${dailyDialog.operatorName} · 每日统计`
                                : "每日数据统计"
                        }}
                    </div>
                    <q-space />
                    <q-btn icon="close" flat round dense v-close-popup />
                </q-card-section>
                <q-card-section>
                    <q-table
                        :rows="dailyDialog.rows"
                        :columns="dailyColumns"
                        :loading="dailyDialog.loading"
                        row-key="date"
                        flat
                        bordered
                        dense
                        :rows-per-page-options="[0]"
                        hide-pagination
                        style="max-height: 400px"
                        virtual-scroll
                    >
                        <template #body-cell-date="props">
                            <q-td :props="props">{{
                                props.row.date?.slice(0, 10) || ""
                            }}</q-td>
                        </template>
                        <template #body-cell-submitCount="props">
                            <q-td :props="props" align="center">
                                <span
                                    :class="
                                        props.row.submitCount > 0
                                            ? 'text-positive text-weight-medium'
                                            : 'text-grey'
                                    "
                                >
                                    {{ props.row.submitCount }}
                                </span>
                            </q-td>
                        </template>
                        <template #body-cell-pendCount="props">
                            <q-td :props="props" align="center">
                                <span
                                    :class="
                                        props.row.pendCount > 0
                                            ? 'text-warning text-weight-medium'
                                            : 'text-grey'
                                    "
                                >
                                    {{ props.row.pendCount }}
                                </span>
                            </q-td>
                        </template>
                        <template #body-cell-total="props">
                            <q-td :props="props" align="center">
                                <span class="text-primary text-weight-medium">{{
                                    props.row.submitCount + props.row.pendCount
                                }}</span>
                            </q-td>
                        </template>
                    </q-table>
                    <div class="text-caption q-mt-sm text-grey-7">
                        提交处理
                        <span class="text-positive text-weight-medium">{{
                            dailyTotals.submit
                        }}</span>
                        条， 挂起
                        <span class="text-warning text-weight-medium">{{
                            dailyTotals.pend
                        }}</span>
                        条， 合计
                        <span class="text-primary text-weight-medium">{{
                            dailyTotals.submit + dailyTotals.pend
                        }}</span>
                        条
                    </div>
                </q-card-section>
            </q-card>
        </q-dialog>
    </q-page>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from "vue";
import { useQuasar } from "quasar";
import { statsApi } from "../api";
import StatCard from "../components/StatCard.vue";
import * as XLSX from "xlsx";

const $q = useQuasar();

// ─── Operator stats ───
const statsLoading = ref(false);
const operatorStats = ref([]);
const globalDailyRows = ref([]);

const overview = computed(() => {
    const ops = operatorStats.value;
    return {
        operatorCount: ops.length,
        totalSubmit: globalDailyRows.value.reduce(
            (s, r) => s + (r.submitCount || 0),
            0,
        ),
        totalPend: globalDailyRows.value.reduce(
            (s, r) => s + (r.pendCount || 0),
            0,
        ),
        todaySubmit: ops.reduce((s, o) => s + (o.todaySubmit || 0), 0),
        todayPend: ops.reduce((s, o) => s + (o.todayPend || 0), 0),
    };
});

const operatorOptions = computed(() =>
    operatorStats.value.map((o) => ({
        userId: o.userId,
        label: `${o.realname || o.username} (${o.username})`,
    })),
);

const opColumns = [
    { name: "username", label: "用户名", field: "username", align: "left" },
    { name: "realname", label: "姓名", field: "realname", align: "left" },
    {
        name: "todayStart",
        label: "今日开始",
        field: "todayStart",
        align: "center",
        style: "width:90px",
    },
    {
        name: "todaySkip",
        label: "今日跳过",
        field: "todaySkip",
        align: "center",
        style: "width:90px",
    },
    {
        name: "todaySubmit",
        label: "今日提交",
        field: "todaySubmit",
        align: "center",
        style: "width:90px",
    },
    {
        name: "todayPend",
        label: "今日挂起",
        field: "todayPend",
        align: "center",
        style: "width:90px",
    },
    {
        name: "actions",
        label: "操作",
        field: "actions",
        align: "center",
        style: "width:100px",
    },
];

async function fetchStats() {
    statsLoading.value = true;
    try {
        const res = await statsApi.operatorStats();
        operatorStats.value = res.data || [];
    } catch {
        $q.notify({ type: "negative", message: "加载操作员统计失败" });
    } finally {
        statsLoading.value = false;
    }
}

// ─── Daily dialog ───
const dailyDialog = reactive({
    visible: false,
    loading: false,
    type: "submit",
    operatorName: "",
    rows: [],
});

const dailyTotals = computed(() => ({
    submit: dailyDialog.rows.reduce((s, r) => s + (r.submitCount || 0), 0),
    pend: dailyDialog.rows.reduce((s, r) => s + (r.pendCount || 0), 0),
}));

const dailyColumns = [
    {
        name: "date",
        label: "日期",
        field: "date",
        align: "left",
        style: "width:120px",
    },
    {
        name: "submitCount",
        label: "提交处理",
        field: "submitCount",
        align: "center",
        style: "width:100px",
    },
    {
        name: "pendCount",
        label: "挂起",
        field: "pendCount",
        align: "center",
        style: "width:80px",
    },
    {
        name: "total",
        label: "合计",
        field: (row) => (row.submitCount || 0) + (row.pendCount || 0),
        align: "center",
        style: "width:80px",
    },
];

async function fetchGlobalDaily() {
    try {
        const res = await statsApi.daily();
        globalDailyRows.value = res.data || [];
    } catch {}
}

function openDailyDialog(type) {
    dailyDialog.type = type;
    dailyDialog.operatorName = "";
    dailyDialog.visible = true;
    dailyDialog.rows = globalDailyRows.value.slice().reverse();
}

async function openOperatorDailyDialog(row) {
    dailyDialog.operatorName = row.realname || row.username;
    dailyDialog.type = "submit";
    dailyDialog.rows = [];
    dailyDialog.visible = true;
    dailyDialog.loading = true;
    try {
        const res = await statsApi.daily({ userId: row.userId });
        dailyDialog.rows = (res.data || []).slice().reverse();
    } catch {
        $q.notify({ type: "negative", message: "加载每日数据失败" });
    } finally {
        dailyDialog.loading = false;
    }
}

// ─── Operator range stats ───
const range = reactive({
    userIds: [],
    startTime: "",
    endTime: "",
    loading: false,
    queried: false,
    result: [],
});

const rangeColumns = [
    { name: "username", label: "用户名", field: "username", align: "left" },
    { name: "realname", label: "姓名", field: "realname", align: "left" },
    {
        name: "startCount",
        label: "开始",
        field: "startCount",
        align: "center",
        style: "width:80px",
    },
    {
        name: "skipCount",
        label: "跳过",
        field: "skipCount",
        align: "center",
        style: "width:80px",
    },
    {
        name: "submitCount",
        label: "提交处理",
        field: "submitCount",
        align: "center",
        style: "width:90px",
    },
    {
        name: "pendCount",
        label: "挂起",
        field: "pendCount",
        align: "center",
        style: "width:80px",
    },
];

const rangeTotalRows = computed(() => {
    if (!range.result.length) return [];
    const total = {
        username: "合计",
        realname: "",
        startCount: range.result.reduce((s, r) => s + (r.startCount || 0), 0),
        skipCount: range.result.reduce((s, r) => s + (r.skipCount || 0), 0),
        submitCount: range.result.reduce((s, r) => s + (r.submitCount || 0), 0),
        pendCount: range.result.reduce((s, r) => s + (r.pendCount || 0), 0),
        isTotal: true,
    };
    return [...range.result, total];
});

async function fetchRangeStats() {
    if (range.userIds.length === 0) {
        $q.notify({ type: "warning", message: "请选择至少一个操作员" });
        return;
    }
    if (!range.startTime || !range.endTime) {
        $q.notify({ type: "warning", message: "请选择时间段" });
        return;
    }
    range.loading = true;
    try {
        const res = await statsApi.operatorRange({
            userIds: range.userIds.join(","),
            startTime: range.startTime.replace("T", " "),
            endTime: range.endTime.replace("T", " "),
        });
        range.result = res.data || [];
        range.queried = true;
    } catch {
        $q.notify({ type: "negative", message: "查询失败" });
    } finally {
        range.loading = false;
    }
}

function resetRange() {
    range.userIds = [];
    range.startTime = "";
    range.endTime = "";
    range.result = [];
    range.queried = false;
}

function exportOperatorRangeExcel() {
    if (range.result.length === 0) {
        $q.notify({ type: "warning", message: "暂无数据可导出" });
        return;
    }

    // 准备导出数据
    const exportData = rangeTotalRows.value.map((row) => ({
        用户名: row.username,
        姓名: row.realname || "",
        开始: row.startCount,
        跳过: row.skipCount,
        提交处理: row.submitCount,
        挂起: row.pendCount,
    }));

    // 创建工作簿和工作表
    const ws = XLSX.utils.json_to_sheet(exportData);
    const wb = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(wb, ws, "操作员时段统计");

    // 生成文件名（包含时间段）
    const start = range.startTime.replace("T", "_").replace(/:/g, "").slice(0, 15);
    const end = range.endTime.replace("T", "_").replace(/:/g, "").slice(0, 15);
    const filename = `操作员统计_${start}_${end}.xlsx`;

    // 导出文件
    try {
        XLSX.writeFile(wb, filename);
        $q.notify({ type: "positive", message: "导出成功" });
    } catch (e) {
        // 移动端备用方案：使用 Blob 下载
        const wbout = XLSX.write(wb, { bookType: "xlsx", type: "array" });
        const blob = new Blob([wbout], { type: "application/octet-stream" });
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
        $q.notify({ type: "positive", message: "导出成功" });
    }
}

// ─── Inspector stats ───
const inspectorLoading = ref(false);
const inspectorStats = ref([]);

const inspectorOverview = computed(() => {
    const ins = inspectorStats.value;
    return {
        inspectorCount: ins.length,
        todayNormal: ins.reduce((s, i) => s + (i.todayNormal || 0), 0),
        todayAbnormal: ins.reduce((s, i) => s + (i.todayAbnormal || 0), 0),
        todayTotal: ins.reduce((s, i) => s + (i.todayTotal || 0), 0),
    };
});

const inspectorOptions = computed(() =>
    inspectorStats.value.map((i) => ({
        userId: i.userId,
        label: `${i.realname || i.username} (${i.username})`,
    })),
);

const insColumns = [
    { name: "username", label: "用户名", field: "username", align: "left" },
    { name: "realname", label: "姓名", field: "realname", align: "left" },
    {
        name: "todayNormal",
        label: "今日复查正常",
        field: "todayNormal",
        align: "center",
        style: "width:110px",
    },
    {
        name: "todayAbnormal",
        label: "今日复查异常",
        field: "todayAbnormal",
        align: "center",
        style: "width:110px",
    },
    {
        name: "todayTotal",
        label: "今日总计",
        field: "todayTotal",
        align: "center",
        style: "width:90px",
    },
];

async function fetchInspectorStats() {
    inspectorLoading.value = true;
    try {
        const res = await statsApi.inspectorStats();
        inspectorStats.value = res.data || [];
    } catch {
        $q.notify({ type: "negative", message: "加载质检员统计失败" });
    } finally {
        inspectorLoading.value = false;
    }
}

// ─── Inspector range stats ───
const inspectorRange = reactive({
    userIds: [],
    startTime: "",
    endTime: "",
    loading: false,
    queried: false,
    result: [],
});

const insRangeColumns = [
    { name: "username", label: "用户名", field: "username", align: "left" },
    { name: "realname", label: "姓名", field: "realname", align: "left" },
    {
        name: "normalCount",
        label: "复查正常",
        field: "normalCount",
        align: "center",
        style: "width:90px",
    },
    {
        name: "abnormalCount",
        label: "复查异常",
        field: "abnormalCount",
        align: "center",
        style: "width:90px",
    },
    {
        name: "total",
        label: "总计",
        field: "total",
        align: "center",
        style: "width:80px",
    },
];

const insRangeTotalRows = computed(() => {
    if (!inspectorRange.result.length) return [];
    const total = {
        username: "合计",
        realname: "",
        normalCount: inspectorRange.result.reduce(
            (s, r) => s + (r.normalCount || 0),
            0,
        ),
        abnormalCount: inspectorRange.result.reduce(
            (s, r) => s + (r.abnormalCount || 0),
            0,
        ),
        total: inspectorRange.result.reduce((s, r) => s + (r.total || 0), 0),
        isTotal: true,
    };
    return [...inspectorRange.result, total];
});

async function fetchInspectorRangeStats() {
    if (inspectorRange.userIds.length === 0) {
        $q.notify({ type: "warning", message: "请选择至少一个质检员" });
        return;
    }
    if (!inspectorRange.startTime || !inspectorRange.endTime) {
        $q.notify({ type: "warning", message: "请选择时间段" });
        return;
    }
    inspectorRange.loading = true;
    try {
        const res = await statsApi.inspectorRange({
            userIds: inspectorRange.userIds.join(","),
            startTime: inspectorRange.startTime.replace("T", " "),
            endTime: inspectorRange.endTime.replace("T", " "),
        });
        inspectorRange.result = res.data || [];
        inspectorRange.queried = true;
    } catch {
        $q.notify({ type: "negative", message: "查询失败" });
    } finally {
        inspectorRange.loading = false;
    }
}

function resetInspectorRange() {
    inspectorRange.userIds = [];
    inspectorRange.startTime = "";
    inspectorRange.endTime = "";
    inspectorRange.result = [];
    inspectorRange.queried = false;
}

function exportInspectorRangeExcel() {
    if (inspectorRange.result.length === 0) {
        $q.notify({ type: "warning", message: "暂无数据可导出" });
        return;
    }

    // 准备导出数据
    const exportData = insRangeTotalRows.value.map((row) => ({
        用户名: row.username,
        姓名: row.realname || "",
        复查正常: row.normalCount,
        复查异常: row.abnormalCount,
        总计: row.total,
    }));

    // 创建工作簿和工作表
    const ws = XLSX.utils.json_to_sheet(exportData);
    const wb = XLSX.utils.book_new();
    XLSX.utils.book_append_sheet(wb, ws, "质检员时段统计");

    // 生成文件名
    const start = inspectorRange.startTime.replace("T", "_").replace(/:/g, "").slice(0, 15);
    const end = inspectorRange.endTime.replace("T", "_").replace(/:/g, "").slice(0, 15);
    const filename = `质检员统计_${start}_${end}.xlsx`;

    // 导出文件
    try {
        XLSX.writeFile(wb, filename);
        $q.notify({ type: "positive", message: "导出成功" });
    } catch (e) {
        // 移动端备用方案：使用 Blob 下载
        const wbout = XLSX.write(wb, { bookType: "xlsx", type: "array" });
        const blob = new Blob([wbout], { type: "application/octet-stream" });
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
        $q.notify({ type: "positive", message: "导出成功" });
    }
}

onMounted(() => {
    fetchStats();
    fetchGlobalDaily();
    fetchInspectorStats();
});
</script>

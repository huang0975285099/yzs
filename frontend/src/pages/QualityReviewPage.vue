<template>
    <q-page class="q-pa-md">
        <!-- Personal inspect stats card -->
        <q-card flat bordered class="q-mb-lg">
            <q-card-section class="q-pb-none">
                <div class="row items-center justify-between">
                    <div class="text-subtitle1 text-weight-medium">
                        {{ selectedDate }} 复查统计
                    </div>
                    <q-input
                        v-model="selectedDate"
                        type="date"
                        dense
                        outlined
                        label="选择日期"
                        class="col-12 col-sm-auto"
                        style="max-width: 180px"
                        @update:model-value="fetchInspectStats"
                    />
                </div>
            </q-card-section>
            <q-card-section>
                <div class="row q-col-gutter-md text-center">
                    <div class="col-4 col-sm-4">
                        <div class="text-h4 text-weight-bold text-positive">
                            {{ todayStats.normalCount }}
                        </div>
                        <div class="text-caption text-grey">复查正常</div>
                    </div>
                    <div class="col-4 col-sm-4">
                        <div class="text-h4 text-weight-bold text-negative">
                            {{ todayStats.abnormalCount }}
                        </div>
                        <div class="text-caption text-grey">复查异常</div>
                    </div>
                    <div class="col-4 col-sm-4">
                        <div class="text-h4 text-weight-bold text-primary">
                            {{ todayStats.total }}
                        </div>
                        <div class="text-caption text-grey">总计</div>
                    </div>
                </div>
            </q-card-section>
        </q-card>

        <!-- History stats -->
        <div class="q-mb-lg">
            <div class="text-subtitle2 text-grey-7 q-mb-sm">
                历史统计（最近30天）
            </div>
            <div
                class="table-scroll"
                style="max-height: 200px; overflow-y: auto"
            >
                <q-table
                    :rows="historyStats"
                    :columns="histColumns"
                    row-key="date"
                    flat
                    bordered
                    dense
                    :rows-per-page-options="[0]"
                    hide-pagination
                    virtual-scroll
                    :virtual-scroll-item-size="48"
                >
                    <template #body-cell-date="props">
                        <q-td :props="props">{{
                            props.row.date?.slice(0, 10) || ""
                        }}</q-td>
                    </template>
                    <template #body-cell-normalCount="props">
                        <q-td :props="props" align="center">
                            <span
                                :class="
                                    props.row.normalCount > 0
                                        ? 'text-positive text-weight-medium'
                                        : ''
                                "
                            >
                                {{ props.row.normalCount }}
                            </span>
                        </q-td>
                    </template>
                    <template #body-cell-abnormalCount="props">
                        <q-td :props="props" align="center">
                            <span
                                :class="
                                    props.row.abnormalCount > 0
                                        ? 'text-negative text-weight-medium'
                                        : ''
                                "
                            >
                                {{ props.row.abnormalCount }}
                            </span>
                        </q-td>
                    </template>
                </q-table>
            </div>
        </div>

        <!-- Export section -->
        <div class="q-mb-lg">
            <div class="text-subtitle2 text-grey-7 q-mb-sm">质检记录导出</div>
            <q-card flat bordered>
                <q-card-section>
                    <div class="row q-col-gutter-sm">
                        <div class="col-6 col-md-4">
                            <q-input
                                v-model="exportStartTime"
                                outlined
                                dense
                                type="datetime-local"
                                label="导出开始日期时间"
                                class="full-width"
                            />
                        </div>
                        <div class="col-6 col-md-4">
                            <q-input
                                v-model="exportEndTime"
                                outlined
                                dense
                                type="datetime-local"
                                label="导出结束日期时间"
                                class="full-width"
                            />
                        </div>
                        <div class="col-12 col-sm-12 col-md-4">
                            <q-btn
                                color="positive"
                                label="导出 Excel"
                                unelevated
                                :loading="exporting"
                                :disable="!exportStartTime || !exportEndTime"
                                @click="exportExcel"
                                class="full-width"
                                style="height: 40px"
                            />
                        </div></div
                ></q-card-section>
            </q-card>
        </div>

        <!-- Records section -->
        <q-card flat bordered>
            <q-card-section>
                <div class="text-subtitle1 text-weight-medium q-mb-md">
                    已处理订单
                </div>

                <!-- Filters - Responsive Grid -->
                <div class="row q-col-gutter-sm q-mb-md">
                    <div class="col-12 col-sm-auto">
                        <q-btn
                            color="orange"
                            label="随机复查"
                            unelevated
                            :loading="randomInspecting"
                            @click="randomInspect"
                            style="height: 40px"
                        />
                    </div>
                    <!-- 操作员筛选 -->
                    <div class="col-12 col-sm-6 col-md-3 col-lg-2">
                        <q-select
                            v-model="filter.userId"
                            outlined
                            dense
                            clearable
                            label="操作员"
                            :options="operatorOptions"
                            option-value="userId"
                            option-label="label"
                            emit-value
                            map-options
                            class="full-width"
                            @update:model-value="onFilterChange"
                        />
                    </div>

                    <!-- 类型筛选 -->
                    <div class="col-6 col-sm-3 col-md-2 col-lg-2">
                        <q-select
                            v-model="filter.actionType"
                            outlined
                            dense
                            clearable
                            label="操作类型"
                            :options="[
                                { label: '提交处理', value: 'submit' },
                                { label: '挂起', value: 'pend' },
                            ]"
                            emit-value
                            map-options
                            class="full-width"
                            @update:model-value="onFilterChange"
                        />
                    </div>

                    <!-- 复查状态筛选 -->
                    <div class="col-6 col-sm-3 col-md-2 col-lg-2">
                        <q-select
                            v-model="filter.inspectStatus"
                            outlined
                            dense
                            clearable
                            label="复查状态"
                            :options="[
                                { label: '未复查', value: 'none' },
                                { label: '复查正常', value: 'normal' },
                                { label: '复查异常', value: 'abnormal' },
                            ]"
                            emit-value
                            map-options
                            class="full-width"
                            @update:model-value="onFilterChange"
                        />
                    </div>

                    <!-- 日期筛选 -->
                    <div class="col-6 col-sm-3 col-md-2 col-lg-2">
                        <q-input
                            v-model="filter.startDate"
                            outlined
                            dense
                            type="date"
                            label="开始日期"
                            class="full-width"
                            @update:model-value="onFilterChange"
                        />
                    </div>
                    <div class="col-6 col-sm-3 col-md-2 col-lg-2">
                        <q-input
                            v-model="filter.endDate"
                            outlined
                            dense
                            type="date"
                            label="结束日期"
                            class="full-width"
                            @update:model-value="onFilterChange"
                        />
                    </div>

                    <!-- 重置按钮 -->
                    <div
                        class="col-12 col-sm-12 col-md-1 col-lg-1 flex flex-center"
                    >
                        <q-btn flat label="重置" @click="resetFilter" />
                    </div>
                </div>

                <!-- Table with pagination -->
                <q-table
                    :rows="records"
                    :columns="columns"
                    :loading="recordsLoading"
                    row-key="id"
                    flat
                    bordered
                    separator="cell"
                    dense
                    :pagination="qPagination"
                    :rows-per-page-options="[10, 20, 30, 50]"
                    @update:pagination="qPagination = $event"
                    @request="onRequest"
                >
                    <template #body-cell-id="props">
                        <q-td :props="props">
                            <q-btn
                                flat
                                dense
                                color="primary"
                                :label="String(props.row.id)"
                                @click="viewHistory(props.row.id)"
                            />
                        </q-td>
                    </template>
                    <template #body-cell-actionType="props">
                        <q-td :props="props" align="center">
                            <ActionTypeChip :type="props.row.actionType" />
                        </q-td>
                    </template>
                    <template #body-cell-handleGoods="props">
                        <q-td :props="props">
                            <span v-if="props.row.actionType !== 'pend'">{{
                                goodsSummary(props.row.handleGoods)
                            }}</span>
                            <span v-else class="text-caption text-grey">—</span>
                        </q-td>
                    </template>
                    <template #body-cell-handledAt="props">
                        <q-td :props="props">{{
                            formatTime(props.row.handledAt)
                        }}</q-td>
                    </template>
                    <template #body-cell-inspectedAt="props">
                        <q-td :props="props">
                            <span v-if="props.row.inspectStatus">{{
                                formatTime(props.row.inspectedAt)
                            }}</span>
                            <span v-else class="text-caption text-grey">—</span>
                        </q-td>
                    </template>
                    <template #body-cell-inspectStatus="props">
                        <q-td :props="props" align="center">
                            <InspectStatusChip
                                :status="props.row.inspectStatus"
                            />
                        </q-td>
                    </template>
                    <template #body-cell-inspectRemark="props">
                        <q-td :props="props">
                            <span v-if="props.row.inspectRemark">{{
                                props.row.inspectRemark
                            }}</span>
                            <span v-else class="text-caption text-grey">—</span>
                        </q-td>
                    </template>
                    <template #body-cell-videoDuration="props">
                        <q-td :props="props" align="center">
                            {{ formatVideoDuration(props.row.videoDuration) }}
                        </q-td>
                    </template>
                </q-table>
            </q-card-section>
        </q-card>
    </q-page>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useQuasar } from "quasar";
import { statsApi, authApi, tradeApi } from "../api";
import * as XLSX from "xlsx";
import ActionTypeChip from "../components/ActionTypeChip.vue";
import InspectStatusChip from "../components/InspectStatusChip.vue";
import { formatTime, goodsSummary, formatVideoDuration } from "../utils/format.js";

const $q = useQuasar();
const router = useRouter();

// ─── Personal inspect stats ───
const selectedDate = ref(new Date().toISOString().slice(0, 10));
const todayStats = reactive({ normalCount: 0, abnormalCount: 0, total: 0 });
const historyStats = ref([]);

const histColumns = [
    {
        name: "date",
        label: "日期",
        field: "date",
        align: "left",
        style: "width:120px",
    },
    {
        name: "normalCount",
        label: "复查正常",
        field: "normalCount",
        align: "center",
        style: "width:100px",
    },
    {
        name: "abnormalCount",
        label: "复查异常",
        field: "abnormalCount",
        align: "center",
        style: "width:100px",
    },
    {
        name: "total",
        label: "总计",
        field: "total",
        align: "center",
        style: "width:80px",
    },
];

async function fetchInspectStats() {
    try {
        const res = await authApi.inspectStats({ date: selectedDate.value });
        const t = res.data.today;
        todayStats.normalCount = t.normalCount || 0;
        todayStats.abnormalCount = t.abnormalCount || 0;
        todayStats.total = t.total || 0;
        historyStats.value = res.data.history || [];
    } catch {}
}

// ─── Operator filter ───
const operatorStats = ref([]);
const operatorOptions = computed(() =>
    operatorStats.value.map((o) => ({
        userId: o.userId,
        label: o.realname || o.username,
    })),
);

async function fetchOperatorStats() {
    try {
        const res = await statsApi.operatorStats();
        operatorStats.value = res.data || [];
    } catch {}
}

// ─── Records table ───
const recordsLoading = ref(false);
const records = ref([]);
const qPagination = ref({ page: 1, rowsPerPage: 20, rowsNumber: 0 });

const FILTER_KEY = "quality-review-filter";

function loadStoredFilter() {
    try {
        const s = localStorage.getItem(FILTER_KEY);
        return s ? JSON.parse(s) : {};
    } catch {
        return {};
    }
}

const _stored = loadStoredFilter();
const filter = reactive({
    userId: _stored.userId ?? null,
    startDate: _stored.startDate ?? "",
    endDate: _stored.endDate ?? "",
    actionType: _stored.actionType ?? null,
    inspectStatus: _stored.inspectStatus ?? null,
});

const columns = [
    {
        name: "id",
        label: "ID",
        field: "id",
        align: "center",
        style: "width:80px",
    },
    {
        name: "inspectStatus",
        label: "复查状态",
        field: "inspectStatus",
        align: "center",
        style: "width:100px",
    },
    {
        name: "inspectRemark",
        label: "复查反馈",
        field: "inspectRemark",
        align: "left",
        style: "min-width:150px",
    },
    {
        name: "inspectedAt",
        label: "复查时间",
        field: "inspectedAt",
        align: "left",
        style: "width:155px",
    },
    {
        name: "handledByName",
        label: "处理人",
        field: "handledByName",
        align: "left",
        style: "width:100px",
    },
    {
        name: "handleGoods",
        label: "处理商品",
        field: "handleGoods",
        align: "left",
        style: "min-width:180px",
    },
    {
        name: "handledAt",
        label: "处理时间",
        field: "handledAt",
        align: "left",
        style: "width:155px",
    },
    {
        name: "tradeId",
        label: "订单ID",
        field: "tradeId",
        align: "left",
        style: "width:85px",
    },
    {
        name: "outOrderNo",
        label: "商户单号",
        field: "outOrderNo",
        align: "left",
        style: "min-width:190px",
    },
    {
        name: "createTime",
        label: "订单时间",
        field: "createTime",
        align: "left",
        style: "width:155px",
    },
    {
        name: "nodeName",
        label: "节点名称",
        field: "nodeName",
        align: "left",
        style: "min-width:160px",
    },
    {
        name: "actionType",
        label: "操作类型",
        field: "actionType",
        align: "center",
        style: "width:120px",
    },
    {
        name: "videoDuration",
        label: "视频时长",
        field: "videoDuration",
        align: "center",
        style: "width:90px",
    },
];

function onFilterChange() {
    qPagination.value = { ...qPagination.value, page: 1 };
    localStorage.setItem(
        FILTER_KEY,
        JSON.stringify({
            userId: filter.userId,
            startDate: filter.startDate,
            endDate: filter.endDate,
            actionType: filter.actionType,
            inspectStatus: filter.inspectStatus,
        }),
    );
    fetchRecords();
}

function resetFilter() {
    filter.userId = null;
    filter.startDate = "";
    filter.endDate = "";
    filter.actionType = null;
    filter.inspectStatus = null;
    localStorage.removeItem(FILTER_KEY);
    qPagination.value = { ...qPagination.value, page: 1 };
    fetchRecords();
}

async function fetchRecords() {
    recordsLoading.value = true;
    try {
        const params = {
            page: qPagination.value.page,
            size: qPagination.value.rowsPerPage,
        };
        if (filter.userId) params.userId = filter.userId;
        if (filter.actionType) params.actionType = filter.actionType;
        if (filter.startDate) params.startDate = filter.startDate;
        if (filter.endDate) params.endDate = filter.endDate;
        if (filter.inspectStatus) params.inspectStatus = filter.inspectStatus;
        const res = await statsApi.operatorRecords(params);
        records.value = res.data.records || [];
        qPagination.value = {
            ...qPagination.value,
            rowsNumber: res.data.total || 0,
        };
    } catch {
        $q.notify({ type: "negative", message: "加载记录失败" });
    } finally {
        recordsLoading.value = false;
    }
}

async function onRequest(props) {
    const { page, rowsPerPage } = props.pagination;
    qPagination.value = { ...qPagination.value, page, rowsPerPage };
    await fetchRecords();
}

// ─── Export Excel ───
const exportStartTime = ref("");
const exportEndTime = ref("");
const exporting = ref(false);

async function exportExcel() {
    if (!exportStartTime.value || !exportEndTime.value) return;
    exporting.value = true;
    try {
        const res = await statsApi.inspectExport({
            startTime: exportStartTime.value.replace("T", " "),
            endTime: exportEndTime.value.replace("T", " "),
        });
        const rows = res.data || [];
        const statusLabel = (s) =>
            s === "normal" ? "复查正常" : s === "abnormal" ? "复查异常" : "";
        const data = [
            [
                "复查时间",
                "复查人",
                "商户单号",
                "复查结果",
                "复查备注",
                "操作人",
            ],
            ...rows.map((r) => [
                r.inspectedAt
                    ? new Date(r.inspectedAt).toLocaleString("zh-CN")
                    : "",
                r.inspectedByName || "",
                r.outOrderNo || "",
                statusLabel(r.inspectStatus),
                r.inspectRemark || "",
                r.handledByName || "",
            ]),
        ];
        const ws = XLSX.utils.aoa_to_sheet(data);
        ws["!cols"] = [22, 12, 32, 12, 30, 12].map((w) => ({ wch: w }));
        const wb = XLSX.utils.book_new();
        XLSX.utils.book_append_sheet(wb, ws, "复查记录");
        const start = exportStartTime.value.slice(0, 10);
        const end = exportEndTime.value.slice(0, 10);
        XLSX.writeFile(wb, `复查记录_${start}_${end}.xlsx`);
        $q.notify({ type: "positive", message: `导出 ${rows.length} 条` });
    } catch {
        $q.notify({ type: "negative", message: "导出失败" });
    } finally {
        exporting.value = false;
    }
}

function viewHistory(id) {
    router.push({ path: "/app/operations", query: { id, history: "true" } });
}

// ─── Random inspect ───
const randomInspecting = ref(false);
async function randomInspect() {
    randomInspecting.value = true;
    try {
        const res = await tradeApi.randomUninspected();
        if (res.code === 200) {
            router.push({ path: "/app/operations", query: { id: res.data.id, history: "true" } });
        } else {
            $q.notify({ type: "warning", message: "没有待复查的订单" });
        }
    } catch {
        $q.notify({ type: "negative", message: "获取随机订单失败" });
    } finally {
        randomInspecting.value = false;
    }
}

onMounted(() => {
    fetchInspectStats();
    fetchOperatorStats();
    fetchRecords();
});
</script>

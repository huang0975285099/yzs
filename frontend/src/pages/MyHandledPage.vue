<template>
    <q-page class="q-pa-md">
        <div class="page-container">
            <!-- Header -->
            <div class="page-header">
                <div class="row items-center q-gutter-sm flex-wrap">
                    <q-chip
                        v-if="myPendingCount > 0"
                        dense
                        color="orange-2"
                        text-color="orange-9"
                    >
                        待审核 {{ myPendingCount }} 条
                    </q-chip>
                    <q-chip dense color="blue-1" text-color="blue-9">
                        累计提交 {{ cumulativeSubmit }} 条
                    </q-chip>
                    <q-chip dense color="green-1" text-color="green-9">
                        累计金额 ¥{{ cumulativeAmount.toFixed(2) }}
                    </q-chip>
                    <q-chip dense color="grey-2" text-color="grey-8">
                        累计挂起 {{ cumulativePend }} 条
                    </q-chip>
                </div>
                <q-btn
                    color="primary"
                    label="开始处理"
                    unelevated
                    @click="startHandling"
                />
            </div>

            <!-- Daily stats card -->
            <q-card flat bordered class="q-mb-md">
                <q-card-section class="q-pb-none">
                    <div class="row items-center justify-between">
                        <div class="text-subtitle1 text-weight-medium">
                            {{ selectedDate }} 统计
                        </div>
                        <q-input
                            v-model="selectedDate"
                            type="date"
                            dense
                            outlined
                            style="width: 150px"
                            @update:model-value="onDateChange"
                        />
                    </div>
                </q-card-section>
                <q-card-section>
                    <div class="row q-col-gutter-md text-center">
                        <div class="col-6 col-sm-3">
                            <div class="text-h4 text-weight-bold">
                                {{ dayStats.startCount }}
                            </div>
                            <div class="text-caption text-grey">
                                点击开始处理
                            </div>
                        </div>
                        <div class="col-6 col-sm-3">
                            <div class="text-h4 text-weight-bold">
                                {{ dayStats.skipCount }}
                            </div>
                            <div class="text-caption text-grey">点击跳过</div>
                        </div>
                        <div class="col-6 col-sm-3">
                            <div class="text-h4 text-weight-bold text-positive">
                                {{ dayStats.submitCount }}
                            </div>
                            <div class="text-caption text-grey">提交处理</div>
                        </div>
                        <div class="col-6 col-sm-3">
                            <div class="text-h4 text-weight-bold text-warning">
                                {{ dayStats.pendCount }}
                            </div>
                            <div class="text-caption text-grey">挂起</div>
                        </div>
                    </div>
                </q-card-section>
            </q-card>

            <!-- History stats (last 30 days) -->
            <div class="q-mb-md">
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
                        <template #body-cell-submitCount="props">
                            <q-td :props="props" align="center">
                                <span
                                    :class="
                                        props.row.submitCount > 0
                                            ? 'text-positive text-weight-medium'
                                            : ''
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
                                            : ''
                                    "
                                >
                                    {{ props.row.pendCount }}
                                </span>
                            </q-td>
                        </template>
                    </q-table>
                </div>
            </div>

            <!-- Records table -->
            <div class="table-scroll">
                <div class="row items-center q-col-gutter-sm q-mb-sm">
                    <div class="col-12 col-sm-auto text-subtitle2 text-grey-7 self-center">
                        我的处理记录（{{ selectedDate }}）
                    </div>
                    <div class="col-12 col-sm-auto">
                        <q-select
                            v-model="inspectStatusFilter"
                            :options="inspectStatusOptions"
                            option-value="value"
                            option-label="label"
                            emit-value
                            map-options
                            dense
                            outlined
                            label="复查状态"
                            style="min-width: 120px"
                            @update:model-value="onRecordFilterChange"
                        />
                    </div>
                </div>
                <q-table
                    :rows="records"
                    :columns="columns"
                    :loading="loading"
                    row-key="id"
                    flat
                    bordered
                    separator="cell"
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
                    <template #body-cell-inspectStatus="props">
                        <q-td :props="props" align="center">
                            <InspectStatusChip
                                :status="props.row.inspectStatus"
                            />
                        </q-td>
                    </template>
                    <template #body-cell-actionType="props">
                        <q-td :props="props" align="center">
                            <ActionTypeChip :type="props.row.actionType" />
                            <div
                                v-if="
                                    props.row.actionType !== 'pend' &&
                                    goodsSummary(props.row.handleGoods)
                                "
                                class="text-caption text-grey-7 q-mt-xs"
                            >
                                {{ goodsSummary(props.row.handleGoods) }}
                            </div>
                        </q-td>
                    </template>
                    <template #body-cell-handledAt="props">
                        <q-td :props="props">{{
                            formatTimeLocale(props.row.handledAt)
                        }}</q-td>
                    </template>
                    <template #body-cell-videoDuration="props">
                        <q-td :props="props" align="center">
                            {{ formatVideoDuration(props.row.videoDuration) }}
                        </q-td>
                    </template>
                </q-table>
            </div>
        </div>
    </q-page>
</template>

<script setup>
import { ref, reactive, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useQuasar } from "quasar";
import { tradeApi, authApi, reviewApi } from "../api";
import ActionTypeChip from "../components/ActionTypeChip.vue";
import InspectStatusChip from "../components/InspectStatusChip.vue";
import { formatTimeLocale, goodsSummary, formatVideoDuration } from "../utils/format.js";

const $q = useQuasar();
const router = useRouter();

const loading = ref(false);
const records = ref([]);
const qPagination = ref({ page: 1, rowsPerPage: 20, rowsNumber: 0 });
const myPendingCount = ref(0);
const cumulativeSubmit = ref(0);
const cumulativePend = ref(0);
const cumulativeAmount = ref(0);

const selectedDate = ref(new Date().toISOString().slice(0, 10));
const inspectStatusFilter = ref("");
const inspectStatusOptions = [
    { label: "全部", value: "" },
    { label: "未复查", value: "uninspected" },
    { label: "复查正常", value: "normal" },
    { label: "复查异常", value: "abnormal" },
];

const dayStats = reactive({
    startCount: 0,
    skipCount: 0,
    submitCount: 0,
    pendCount: 0,
});
const historyStats = ref([]);

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
        label: "复查",
        field: "inspectStatus",
        align: "center",
        style: "width:90px",
    },
    {
        name: "tradeId",
        label: "订单编号",
        field: "tradeId",
        align: "left",
        style: "width:90px",
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
        style: "width:165px",
    },
    {
        name: "nodeName",
        label: "节点名称",
        field: "nodeName",
        align: "left",
        style: "min-width:180px",
    },
    {
        name: "actionType",
        label: "操作类型",
        field: "actionType",
        align: "center",
        style: "min-width:150px",
    },
    {
        name: "handledByName",
        label: "处理人",
        field: "handledByName",
        align: "left",
        style: "width:100px",
    },
    {
        name: "handledById",
        label: "处理人ID",
        field: "handledById",
        align: "center",
        style: "width:90px",
    },
    {
        name: "handledAt",
        label: "处理时间",
        field: "handledAt",
        align: "left",
        style: "width:165px",
    },
    {
        name: "videoDuration",
        label: "视频时长",
        field: "videoDuration",
        align: "center",
        style: "width:90px",
    },
];

const histColumns = [
    {
        name: "date",
        label: "日期",
        field: "date",
        align: "left",
        style: "width:120px",
    },
    {
        name: "startCount",
        label: "开始处理",
        field: "startCount",
        align: "center",
        style: "width:100px",
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
        style: "width:100px",
    },
    {
        name: "pendCount",
        label: "挂起",
        field: "pendCount",
        align: "center",
        style: "width:80px",
    },
];

onMounted(() => {
    fetchData();
    fetchDailyStats();
    fetchMyPendingCount();
});

async function startHandling() {
    try {
        const res = await tradeApi.randomUnhandled();
        authApi.recordStart();
        router.push({ path: "/app/operations", query: { id: res.data.id } });
    } catch (err) {
        $q.notify({
            type: "warning",
            message: err?.message || "暂无待处理订单",
        });
    }
}

async function fetchMyPendingCount() {
    try {
        const res = await reviewApi.list({
            status: "pending",
            mine: true,
            page: 1,
            size: 1,
        });
        myPendingCount.value = res.data.total || 0;
    } catch {}
}

async function fetchData() {
    loading.value = true;
    try {
        const res = await tradeApi.myHandled({
            page: qPagination.value.page,
            size: qPagination.value.rowsPerPage,
            date: selectedDate.value,
            inspectStatus: inspectStatusFilter.value || undefined,
        });
        records.value = res.data.records;
        qPagination.value = {
            ...qPagination.value,
            rowsNumber: res.data.total,
        };
        cumulativeSubmit.value = res.data.cumulativeSubmit || 0;
        cumulativePend.value = res.data.cumulativePend || 0;
        cumulativeAmount.value = res.data.cumulativeAmount || 0;
    } catch {
        $q.notify({ type: "negative", message: "加载数据失败" });
    } finally {
        loading.value = false;
    }
}

async function onRequest(props) {
    const { page, rowsPerPage } = props.pagination;
    qPagination.value = { ...qPagination.value, page, rowsPerPage };
    await fetchData();
}

async function fetchDailyStats() {
    try {
        const res = await authApi.dailyStats({ date: selectedDate.value });
        dayStats.startCount = res.data.today.startCount || 0;
        dayStats.skipCount = res.data.today.skipCount || 0;
        dayStats.submitCount = res.data.today.submitCount || 0;
        dayStats.pendCount = res.data.today.pendCount || 0;
        historyStats.value = res.data.history || [];
    } catch {}
}

function onDateChange() {
    qPagination.value = { ...qPagination.value, page: 1 };
    fetchData();
    fetchDailyStats();
}

function onRecordFilterChange() {
    qPagination.value = { ...qPagination.value, page: 1 };
    fetchData();
}

function viewHistory(id) {
    router.push({ path: "/app/operations", query: { id, history: "true" } });
}
</script>

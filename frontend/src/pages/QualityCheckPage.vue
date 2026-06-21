<template>
    <q-page class="q-pa-md">
        <!-- Filter bar -->
        <div class="row items-center q-gutter-md q-mb-md">
            <q-btn-toggle
                v-model="statusFilter"
                unelevated
                :options="[
                    { label: '待审核', value: 'pending' },
                    { label: '已通过', value: 'approved' },
                    { label: '全部', value: 'all' },
                ]"
                color="grey-4"
                text-color="dark"
                toggle-color="primary"
                toggle-text-color="white"
                @update:model-value="loadData"
            />
            <q-chip
                v-if="pendingCount > 0"
                color="orange-2"
                text-color="orange-9"
                icon="pending_actions"
                dense
            >
                待审核 {{ pendingCount }} 条
            </q-chip>
        </div>

        <!-- Table -->
        <q-card flat bordered>
            <div class="table-scroll">
                <q-table
                    :rows="records"
                    :columns="columns"
                    :loading="loading"
                    row-key="id"
                    flat
                    bordered
                    separator="cell"
                    :pagination="{ rowsPerPage: 0 }"
                    hide-pagination
                >
                    <!-- 机器/商户单号 -->
                    <template #body-cell-trade="props">
                        <q-td :props="props">
                            <div class="text-weight-medium text-body2">
                                {{ props.row.trade?.nodeName || "—" }}
                            </div>
                            <div class="text-caption text-grey">
                                {{ props.row.trade?.outOrderNo || "—" }}
                            </div>
                        </q-td>
                    </template>

                    <!-- 操作类型 -->
                    <template #body-cell-actionType="props">
                        <q-td :props="props" align="center">
                            <ActionTypeChip :type="props.row.actionType" />
                        </q-td>
                    </template>

                    <!-- 商品明细 -->
                    <template #body-cell-goodsJson="props">
                        <q-td :props="props">
                            <template v-if="props.row.actionType === 'submit'">
                                <div
                                    v-if="
                                        parseGoods(props.row.goodsJson)
                                            .length === 0
                                    "
                                    class="text-caption text-grey"
                                >
                                    无消费订单
                                </div>
                                <div
                                    v-for="g in parseGoods(props.row.goodsJson)"
                                    :key="g.goodsId"
                                    class="text-caption"
                                >
                                    {{ g.goodsName }} × {{ g.goodsCount }}
                                    <span class="text-negative"
                                        >¥{{
                                            (
                                                (g.goodsPrice || 0) *
                                                (g.goodsCount || 0)
                                            ).toFixed(2)
                                        }}</span
                                    >
                                </div>
                            </template>
                            <span v-else class="text-caption text-grey">—</span>
                        </q-td>
                    </template>

                    <!-- 作业时长 -->
                    <template #body-cell-duration="props">
                        <q-td :props="props" align="center">{{
                            formatDuration(props.row.duration)
                        }}</q-td>
                    </template>

                    <!-- 操作员 -->
                    <template #body-cell-submittedBy="props">
                        <q-td :props="props">
                            <div class="text-body2">
                                {{ props.row.submittedByName }}
                            </div>
                            <div class="text-caption text-grey">
                                {{ formatTime(props.row.submittedAt) }}
                            </div>
                        </q-td>
                    </template>

                    <!-- 状态 -->
                    <template #body-cell-reviewStatus="props">
                        <q-td :props="props" align="center">
                            <q-chip
                                dense
                                class="q-ma-none"
                                :color="
                                    props.row.reviewStatus === 'approved'
                                        ? 'green-1'
                                        : 'orange-1'
                                "
                                :text-color="
                                    props.row.reviewStatus === 'approved'
                                        ? 'green-9'
                                        : 'orange-9'
                                "
                            >
                                {{
                                    props.row.reviewStatus === "approved"
                                        ? "已通过"
                                        : "待审核"
                                }}
                            </q-chip>
                        </q-td>
                    </template>

                    <!-- 审核人 -->
                    <template #body-cell-reviewedBy="props">
                        <q-td :props="props">
                            <template
                                v-if="props.row.reviewStatus === 'approved'"
                            >
                                <div class="text-body2">
                                    {{ props.row.reviewedByName }}
                                </div>
                                <div class="text-caption text-grey">
                                    {{ formatTime(props.row.reviewedAt) }}
                                </div>
                            </template>
                            <span v-else class="text-caption text-grey">—</span>
                        </q-td>
                    </template>

                    <!-- 备注 -->
                    <template #body-cell-reviewRemark="props">
                        <q-td :props="props">
                            <q-input
                                v-model="props.row.reviewRemark"
                                dense
                                outlined
                                type="textarea"
                                :rows="2"
                                placeholder="写备注..."
                                :disable="props.row.reviewStatus === 'approved'"
                                @blur="saveRemark(props.row)"
                                style="min-width: 150px"
                            />
                        </q-td>
                    </template>

                    <!-- 操作 -->
                    <template #body-cell-actions="props">
                        <q-td :props="props" align="center">
                            <q-btn
                                v-if="props.row.reviewStatus === 'pending'"
                                color="primary"
                                size="sm"
                                label="通过"
                                unelevated
                                :loading="approvingId === props.row.id"
                                @click="handleApprove(props.row)"
                            />
                            <span v-else class="text-caption text-grey">—</span>
                        </q-td>
                    </template>
                </q-table>
            </div>

            <!-- Pagination -->
            <q-card-section class="row justify-end items-center q-gutter-sm">
                <div class="text-caption text-grey-7">共 {{ total }} 条</div>
                <q-pagination
                    v-model="page"
                    :max="Math.ceil(total / pageSize) || 1"
                    :max-pages="7"
                    direction-links
                    boundary-links
                    @update:model-value="loadData"
                />
            </q-card-section>
        </q-card>
    </q-page>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useQuasar } from "quasar";
import { reviewApi } from "../api/index.js";
import ActionTypeChip from "../components/ActionTypeChip.vue";
import { formatTime, formatDuration, parseGoods } from "../utils/format.js";

const $q = useQuasar();
const loading = ref(false);
const records = ref([]);
const total = ref(0);
const pendingCount = ref(0);
const page = ref(1);
const pageSize = ref(20);
const statusFilter = ref("pending");
const approvingId = ref(null);

const columns = [
    {
        name: "trade",
        label: "机器/商户单号",
        field: "trade",
        align: "left",
        style: "min-width:160px",
    },
    {
        name: "abnormalTypeDesc",
        label: "异常类型",
        field: (row) => row.trade?.abnormalTypeDesc,
        align: "left",
        style: "width:120px",
    },
    {
        name: "createTime",
        label: "下单时间",
        field: (row) => row.trade?.createTime,
        align: "left",
        style: "width:150px",
    },
    {
        name: "actionType",
        label: "操作类型",
        field: "actionType",
        align: "center",
        style: "width:90px",
    },
    {
        name: "goodsJson",
        label: "商品明细",
        field: "goodsJson",
        align: "left",
        style: "min-width:180px",
    },
    {
        name: "duration",
        label: "作业时长",
        field: "duration",
        align: "center",
        style: "width:90px",
    },
    {
        name: "submittedBy",
        label: "操作员",
        field: "submittedByName",
        align: "left",
        style: "width:110px",
    },
    {
        name: "reviewStatus",
        label: "状态",
        field: "reviewStatus",
        align: "center",
        style: "width:90px",
    },
    {
        name: "reviewedBy",
        label: "审核人",
        field: "reviewedByName",
        align: "left",
        style: "width:120px",
    },
    {
        name: "reviewRemark",
        label: "备注",
        field: "reviewRemark",
        align: "left",
        style: "min-width:160px",
    },
    {
        name: "actions",
        label: "操作",
        field: "actions",
        align: "center",
        style: "width:90px",
    },
];

async function loadPendingCount() {
    try {
        const res = await reviewApi.list({
            status: "pending",
            page: 1,
            size: 1,
        });
        pendingCount.value = res.data.total || 0;
    } catch {}
}

async function loadData() {
    loading.value = true;
    try {
        const [res] = await Promise.all([
            reviewApi.list({
                status: statusFilter.value,
                page: page.value,
                size: pageSize.value,
            }),
            loadPendingCount(),
        ]);
        records.value = res.data.records || [];
        total.value = res.data.total || 0;
    } catch (err) {
        $q.notify({ type: "negative", message: err?.message || "加载失败" });
    } finally {
        loading.value = false;
    }
}

function handleApprove(row) {
    $q.dialog({
        title: "审核确认",
        message: `确认审核通过该条${row.actionType === "submit" ? "提交处理" : "挂起"}记录？通过后将提交到外部系统。`,
        cancel: { label: "取消", flat: true },
        ok: { label: "确认通过", color: "primary" },
        persistent: true,
    }).onOk(async () => {
        approvingId.value = row.id;
        try {
            const res = await reviewApi.approve(row.id);
            $q.notify({ type: "positive", message: res.message || "审核通过" });
            loadData();
        } catch (err) {
            $q.notify({
                type: "negative",
                message: err?.message || "操作失败，请重试",
            });
        } finally {
            approvingId.value = null;
        }
    });
}

async function saveRemark(row) {
    if (row.reviewStatus === "approved") return;
    try {
        await reviewApi.remark(row.id, { remark: row.reviewRemark || "" });
    } catch {}
}


onMounted(loadData);
</script>

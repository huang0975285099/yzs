<template>
    <q-page class="ops-page">
        <div class="ops-container" :class="{ 'ops-fullscreen': isFullscreen }">
            <!-- Loading state -->
            <div v-if="loading" class="center-state">
                <q-spinner-dots color="primary" size="48px" />
                <div class="text-grey q-mt-sm">加载订单中...</div>
            </div>

            <!-- Empty state -->
            <div v-else-if="!detail" class="center-state">
                <div class="text-center">
                    <q-icon name="inbox" size="64px" color="grey-4" />
                    <div class="text-grey-6 q-mt-sm q-mb-md">
                        暂无待处理订单
                    </div>
                    <q-btn
                        color="primary"
                        label="重新获取"
                        unelevated
                        @click="fetchRandom"
                    />
                </div>
            </div>

            <template v-else>
                <!-- Order Info Bar -->
                <div class="info-bar">
                    <q-btn
                        flat
                        dense
                        size="sm"
                        :icon="isFullscreen ? 'fullscreen_exit' : 'fullscreen'"
                        :label="isFullscreen ? '退出全屏' : '全屏作业'"
                        @click="toggleFullscreen"
                    />
                    <div class="info-item" v-if="!$q.screen.lt.lg">
                        <span class="label">货柜名称</span
                        ><span>{{ detail.nodeName }}</span>
                    </div>
                    <div class="info-item" v-if="!$q.screen.lt.lg">
                        <span class="label">机器编号</span
                        ><span>{{ detail.innerCode }}</span>
                    </div>
                    <div class="info-item" v-if="!$q.screen.lt.lg">
                        <span class="label">商户单号</span
                        ><span>{{ detail.outOrderNo }}</span>
                    </div>
                    <div class="info-item" v-if="!$q.screen.lt.lg">
                        <span class="label">异常描述</span
                        ><span>{{ detail.unrecognizedDesc || "—" }}</span>
                    </div>
                    <div class="info-item" v-if="!$q.screen.lt.lg">
                        <span class="label">创建时间</span
                        ><span>{{ detail.createTime || "—" }}</span>
                    </div>
                    <div class="info-item" v-if="!$q.screen.lt.lg">
                        <span class="label">作业时间</span>
                        <span class="text-weight-medium text-primary">{{
                            elapsedFormatted
                        }}</span>
                    </div>
                    <!-- <q-btn
                        v-if="!isHistory"
                        flat
                        dense
                        size="sm"
                        icon="edit_note"
                        label="填写备注"
                        @click="remarkDialogVisible = true"
                    /> -->
                    <q-btn
                        v-if="!isHistory"
                        flat
                        dense
                        size="sm"
                        icon="history"
                        label="我处理过的商品"
                        @click="openMyGoodsDrawer"
                    />
                </div>

                <!-- Main content: responsive layout -->
                <div class="main-content">
                    <!-- Video section (left on PC, top on Mobile) -->
                    <div class="video-panel">
                        <div
                            v-if="
                                detail.doorOpenFileUrlList &&
                                detail.doorOpenFileUrlList.length
                            "
                        >
                            <div
                                v-for="(url, idx) in detail.doorOpenFileUrlList"
                                :key="'open-' + idx"
                                class="video-block"
                            >
                                <video
                                    :ref="
                                        (el) => {
                                            if (el) openVideos[idx] = el;
                                        }
                                    "
                                    :src="url"
                                    controls
                                    preload="auto"
                                    class="video-el"
                                    :style="{
                                        transform: `rotate(${getRotation('open', idx)}deg)`,
                                    }"
                                />
                                <div class="speed-bar">
                                    <q-btn
                                        v-for="s in speeds"
                                        :key="s"
                                        dense
                                        size="sm"
                                        unelevated
                                        :color="
                                            (openRates[idx] || 1) === s
                                                ? 'positive'
                                                : 'primary'
                                        "
                                        :label="`${s}x`"
                                        @click="setRate('open', idx, s)"
                                    />
                                </div>
                            </div>
                        </div>

                        <div
                            v-if="
                                detail.doorCloseFileUrlList &&
                                detail.doorCloseFileUrlList.length
                            "
                        >
                            <div
                                v-for="(
                                    url, idx
                                ) in detail.doorCloseFileUrlList"
                                :key="'close-' + idx"
                                class="video-block"
                            >
                                <video
                                    :ref="
                                        (el) => {
                                            if (el) closeVideos[idx] = el;
                                        }
                                    "
                                    :src="url"
                                    controls
                                    preload="auto"
                                    class="video-el"
                                    :style="{
                                        transform: `rotate(${getRotation('close', idx)}deg)`,
                                    }"
                                />
                                <div class="speed-bar">
                                    <q-btn
                                        v-for="s in speeds"
                                        :key="s"
                                        unelevated
                                        :color="
                                            (closeRates[idx] || 1) === s
                                                ? 'positive'
                                                : 'primary'
                                        "
                                        :label="`${s}x`"
                                        @click="setRate('close', idx, s)"
                                    />
                                </div>
                            </div>
                        </div>

                        <div
                            v-if="
                                !detail.doorOpenFileUrlList?.length &&
                                !detail.doorCloseFileUrlList?.length
                            "
                            class="center-state"
                        >
                            <div class="text-center">
                                <q-icon
                                    name="videocam_off"
                                    size="48px"
                                    color="grey-4"
                                />
                                <div class="text-grey-6 q-mt-sm">暂无视频</div>
                            </div>
                        </div>
                    </div>

                    <!-- Product section (right on PC, bottom on Mobile with horizontal scroll) -->
                    <div class="product-panel">
                        <div class="product-header">
                            <div class="panel-title">
                                <strong v-if="!$q.screen.lt.lg"
                                    >机器商品</strong
                                >
                            </div>
                            <div class="row q-gutter-sm" v-if="!isHistory">
                                <q-input
                                    v-model="searchKeyword"
                                    dense
                                    outlined
                                    placeholder="搜索商品名称"
                                    clearable
                                    class="mobile-search-input"
                                >
                                    <template #prepend>
                                        <q-icon name="search" />
                                    </template>
                                </q-input>
                                <q-btn
                                    color="warning"
                                    label="添加分公司商品"
                                    unelevated
                                    @click="openBranchDialog"
                                />
                            </div>
                        </div>

                        <!-- Goods grid - responsive layout -->
                        <div class="goods-grid">
                            <q-card
                                v-for="(p, idx) in filteredProducts"
                                :key="idx"
                                flat
                                bordered
                                :class="[
                                    'goods-card',
                                    quantities[p.goodsId] > 0 ? 'selected' : '',
                                ]"
                            >
                                <q-card-section class="q-pa-xs">
                                    <div class="card-name">
                                        <span
                                            @click="copyGoodsName(p.goodsName)"
                                            style="cursor: copy"
                                            >{{
                                                p.goodsName || "未知商品"
                                            }}</span
                                        >
                                    </div>
                                </q-card-section>
                                <div
                                    class="goods-img-wrap"
                                    style="cursor: pointer"
                                    @click="openGoodsCatalog(p.goodsName)"
                                >
                                    <img
                                        v-if="p.goodsImage"
                                        :src="p.goodsImage"
                                        alt="商品图片"
                                        @error="
                                            (e) =>
                                                (e.target.style.display =
                                                    'none')
                                        "
                                    />
                                    <div v-else class="no-img">无图片</div>
                                </div>
                                <q-card-section class="q-pa-xs card-footer">
                                    <q-chip
                                        dense
                                        color="green-8"
                                        text-color="white"
                                        class="footer-price"
                                    >
                                        ¥{{ p.goodsPrice }}
                                    </q-chip>
                                    <div
                                        v-if="!isHistory"
                                        class="goods-count-control"
                                    >
                                        <q-btn
                                            dense
                                            push
                                            icon="remove"
                                            @click="decreaseCount(p.goodsId)"
                                        />
                                        <span class="count-val">{{
                                            quantities[p.goodsId] || 0
                                        }}</span>
                                        <q-btn
                                            dense
                                            push
                                            icon="add"
                                            @click="increaseCount(p.goodsId)"
                                        />
                                    </div>
                                </q-card-section>
                            </q-card>
                        </div>

                        <div
                            v-if="filteredProducts.length === 0"
                            class="center-state"
                            style="flex: 1"
                        >
                            <div class="text-center">
                                <q-icon
                                    name="inventory_2"
                                    size="40px"
                                    color="grey-4"
                                />
                                <div class="text-grey-6 q-mt-xs">暂无商品</div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Bottom action bar -->
                <div class="action-bar">
                    <!-- History mode -->
                    <template v-if="isHistory">
                        <div class="action-btns">
                            <q-btn
                                flat
                                label="返回"
                                icon="arrow_back"
                                @click="router.back()"
                            />
                            <q-btn
                                v-if="
                                    !authStore.reviewEnabled &&
                                    (authStore.user?.role === 'inspector' ||
                                        authStore.user?.role === 'admin')
                                "
                                unelevated
                                :color="
                                    inspectInfo.status === 'abnormal'
                                        ? 'negative'
                                        : inspectInfo.status === 'normal'
                                          ? 'positive'
                                          : 'warning'
                                "
                                :label="
                                    inspectInfo.status === 'normal'
                                        ? '已复查：正常'
                                        : inspectInfo.status === 'abnormal'
                                          ? '已复查：异常'
                                          : '复查'
                                "
                                @click="openInspectDialog"
                            />
                        </div>
                        <div class="history-notice">
                            <q-chip dense color="grey-3" text-color="grey-8"
                                >历史查看模式（只读）</q-chip
                            >
                            <template
                                v-if="
                                    inspectInfo.status &&
                                    !authStore.reviewEnabled
                                "
                            >
                                <q-chip
                                    dense
                                    :color="
                                        inspectInfo.status === 'normal'
                                            ? 'green-1'
                                            : 'red-1'
                                    "
                                    :text-color="
                                        inspectInfo.status === 'normal'
                                            ? 'green-9'
                                            : 'red-9'
                                    "
                                >
                                    复查{{
                                        inspectInfo.status === "normal"
                                            ? "正常"
                                            : "异常"
                                    }}
                                </q-chip>
                                <span
                                    v-if="inspectInfo.remark"
                                    style="font-size: 12px; color: #606266"
                                    >{{ inspectInfo.remark }}</span
                                >
                                <span style="font-size: 12px; color: #909399"
                                    >{{ inspectInfo.byName }}
                                    {{ inspectInfo.at }}</span
                                >
                            </template>
                            <template
                                v-else-if="historyPendStatus === 'PENDING'"
                            >
                                <q-chip
                                    dense
                                    color="orange-1"
                                    text-color="orange-9"
                                    >已挂起</q-chip
                                >
                            </template>
                            <template v-else-if="historyHandleGoods">
                                <div class="history-goods">
                                    <span class="history-goods-label"
                                        >处理商品：</span
                                    >
                                    <span class="history-goods-content">{{
                                        formatHistoryGoods(historyHandleGoods)
                                    }}</span>
                                </div>
                            </template>
                            <template v-else>
                                <q-chip
                                    dense
                                    color="green-1"
                                    text-color="green-9"
                                    >无消费</q-chip
                                >
                            </template>
                        </div>
                    </template>

                    <!-- Normal mode -->
                    <template v-else>
                        <div class="action-btns">
                            <q-btn
                                outline
                                :label="
                                    $q.screen.lt.lg
                                        ? `跳过${remainingCount !== null ? '，剩余' + remainingCount : ''}`
                                        : `跳过，处理下一单（Esc）${remainingCount !== null ? '，剩余' + remainingCount : ''}`
                                "
                                :loading="skipping"
                                @click="goNext"
                            />
                            <q-btn
                                color="warning"
                                unelevated
                                :label="$q.screen.lt.lg ? '挂起' : '挂起（G）'"
                                :loading="pending"
                                @click="handlePend"
                            />
                        </div>
                        <div class="selected-summary">
                            <div
                                style="padding-right: 10px"
                                v-if="!$q.screen.lt.md"
                            >
                                <div
                                    v-if="
                                        selectedGoods &&
                                        selectedGoods.length > 0
                                    "
                                    class="selected-goods-list"
                                >
                                    <div
                                        v-for="(item, index) in selectedGoods"
                                        :key="index"
                                        class="selected-goods-item"
                                    >
                                        <div>
                                            <b class="text-positive">{{
                                                item.goodsName
                                            }}</b>
                                            <span class="text-grey-6 q-ml-xs"
                                                >#{{ item.goodsId }}</span
                                            >
                                        </div>
                                        <div
                                            class="row items-center q-gutter-xs"
                                        >
                                            <div class="goods-count-control">
                                                <q-btn
                                                    dense
                                                    push
                                                    round
                                                    size="xs"
                                                    icon="remove"
                                                    @click="
                                                        decreaseCount(
                                                            item.goodsId,
                                                        )
                                                    "
                                                />
                                                <span class="count-val">{{
                                                    quantities[item.goodsId]
                                                }}</span>
                                                <q-btn
                                                    dense
                                                    push
                                                    round
                                                    size="xs"
                                                    icon="add"
                                                    @click="
                                                        increaseCount(
                                                            item.goodsId,
                                                        )
                                                    "
                                                />
                                            </div>
                                            <span
                                                >× {{ item.goodsPrice }} 元 =
                                                {{
                                                    (
                                                        item.goodsPrice *
                                                        quantities[item.goodsId]
                                                    ).toFixed(2)
                                                }}元</span
                                            >
                                        </div>
                                    </div>
                                </div>
                            </div>
                            <div style="text-align: center">
                                <q-btn
                                    color="positive"
                                    unelevated
                                    :loading="submitting"
                                    @click="handleSubmit"
                                    :label="
                                        $q.screen.lt.lg
                                            ? '提交处理'
                                            : '提交处理（Ctrl+Enter）'
                                    "
                                />
                                <div class="q-mt-xs">
                                    <span
                                        v-if="selectedCount > 0"
                                        class="text-caption"
                                    >
                                        合计：¥{{ totalPrice }} 总计：{{
                                            selectedCount
                                        }}件
                                    </span>
                                    <span
                                        v-else
                                        class="text-caption text-grey-5"
                                        >未选择商品（将提交为本次无消费）</span
                                    >
                                </div>
                            </div>
                        </div>
                    </template>
                </div>
            </template>

            <!-- Branch drawer backdrop -->
            <div
                v-if="branchDialog.visible"
                class="drawer-backdrop"
                @click="branchDialog.visible = false"
            />

            <!-- Branch product search drawer -->
            <q-drawer
                v-model="branchDialog.visible"
                side="right"
                :width="$q.screen.lt.sm ? $q.screen.width : 720"
                overlay
                elevated
                class="branch-drawer"
            >
                <q-toolbar class="bg-grey-2">
                    <q-toolbar-title>搜索分公司商品</q-toolbar-title>
                    <q-btn
                        flat
                        round
                        icon="close"
                        @click="branchDialog.visible = false"
                    />
                </q-toolbar>
                <div class="q-pa-md">
                    <div class="row q-gutter-sm q-mb-md">
                        <q-input
                            v-model="branchDialog.keyword"
                            outlined
                            dense
                            placeholder="输入商品名称关键词"
                            clearable
                            style="flex: 1"
                            @keyup.enter="searchBranchProducts"
                        />
                        <q-btn
                            color="warning"
                            unelevated
                            :loading="branchDialog.loading"
                            label="搜索"
                            @click="searchBranchProducts"
                        />
                        <q-btn
                            v-if="!isHistory"
                            flat
                            label="我处理过的商品"
                            @click="openMyGoodsDrawer"
                        />
                    </div>
                    <q-inner-loading :showing="branchDialog.loading" />
                    <div
                        v-if="branchDialog.results.length"
                        class="branch-results"
                    >
                        <div
                            v-for="p in branchDialog.results"
                            :key="p.productId"
                            class="branch-item"
                            @click="addBranchProduct(p)"
                        >
                            <div class="branch-img-wrap">
                                <img
                                    v-if="p.productUrl"
                                    :src="p.productUrl"
                                    class="branch-img"
                                    alt=""
                                />
                                <div v-else class="branch-img no-img">无图</div>
                                <div class="branch-add-mask">
                                    <q-icon
                                        name="add_circle"
                                        size="28px"
                                        color="white"
                                    />
                                </div>
                            </div>
                            <div class="branch-name">{{ p.productName }}</div>
                            <div class="branch-id">#{{ p.productId }}</div>
                        </div>
                    </div>
                    <div
                        v-else-if="!branchDialog.loading"
                        class="text-center q-mt-xl"
                    >
                        <q-icon name="search" size="48px" color="grey-4" />
                        <div class="text-grey-6 q-mt-sm">搜索后显示结果</div>
                    </div>
                </div>
            </q-drawer>

            <!-- Remark dialog -->
            <q-dialog v-model="remarkDialogVisible" persistent>
                <q-card style="min-width: 480px">
                    <q-card-section class="row items-center">
                        <div class="text-h6">填写备注</div>
                        <q-space />
                        <q-btn flat round icon="close" v-close-popup />
                    </q-card-section>
                    <q-card-section>
                        <q-input
                            v-model="handleRemark"
                            type="textarea"
                            :rows="4"
                            outlined
                            placeholder="请输入处理备注（可选）"
                            maxlength="500"
                            counter
                        />
                    </q-card-section>
                    <q-card-actions align="right">
                        <q-btn flat label="取消" v-close-popup />
                        <q-btn
                            color="primary"
                            label="确定"
                            unelevated
                            v-close-popup
                        />
                    </q-card-actions>
                </q-card>
            </q-dialog>

            <!-- My handled goods dialog -->
            <q-dialog
                v-model="myGoodsDrawer.visible"
                :maximized="$q.screen.lt.md"
            >
                <q-card class="my-goods-dialog" :style="dialogStyle">
                    <q-card-section class="row items-center q-pa-sm-md">
                        <div class="text-subtitle1 text-weight-medium">
                            我处理过的商品
                        </div>
                        <q-space />
                        <q-btn flat round dense icon="close" v-close-popup />
                    </q-card-section>
                    <q-card-section class="my-goods-content">
                        <q-inner-loading :showing="myGoodsDrawer.loading" />
                        <div
                            v-if="myGoodsDrawer.list.length"
                            class="my-goods-list"
                        >
                            <div
                                v-for="item in myGoodsDrawer.list"
                                :key="item.goodsId + '-' + item.type"
                                class="my-goods-item"
                            >
                                <div class="my-goods-img-wrap">
                                    <img
                                        v-if="item.goodsImage"
                                        :src="item.goodsImage"
                                        class="my-goods-img"
                                        alt=""
                                    />
                                    <div v-else class="my-goods-img no-img">
                                        无图
                                    </div>
                                    <q-chip
                                        dense
                                        :color="
                                            item.type === 1
                                                ? 'green-1'
                                                : 'orange-1'
                                        "
                                        :text-color="
                                            item.type === 1
                                                ? 'green-9'
                                                : 'orange-9'
                                        "
                                        class="my-goods-type-tag"
                                        size="sm"
                                    >
                                        {{
                                            item.type === 1 ? "机器" : "分公司"
                                        }}
                                    </q-chip>
                                </div>
                                <div class="my-goods-info">
                                    <div class="my-goods-name">
                                        {{ item.goodsName }}
                                    </div>
                                    <div class="my-goods-meta">
                                        <span class="my-goods-total"
                                            >处理次数：</span
                                        >
                                        <span class="my-goods-price">{{
                                            item.goodsCount
                                        }}</span>
                                        <q-btn
                                            flat
                                            dense
                                            size="xs"
                                            color="primary"
                                            label="搜索它"
                                            @click.stop="
                                                openBranchDialogWithName(
                                                    item.goodsName,
                                                )
                                            "
                                        />
                                    </div>
                                    <div class="my-goods-time">
                                        最近处理：{{
                                            formatTime(item.createdAt)
                                        }}
                                    </div>
                                </div>
                            </div>
                        </div>
                        <div
                            v-else-if="!myGoodsDrawer.loading"
                            class="text-center q-pa-xl"
                        >
                            <q-icon name="inbox" size="48px" color="grey-4" />
                            <div class="text-grey-6 q-mt-sm">暂无处理记录</div>
                        </div>
                    </q-card-section>
                    <q-card-actions
                        v-if="myGoodsDrawer.hasMore"
                        align="center"
                        class="q-pa-sm-md"
                    >
                        <q-btn
                            flat
                            :loading="myGoodsDrawer.loading"
                            label="加载更多"
                            @click="loadMoreMyGoods"
                        />
                    </q-card-actions>
                </q-card>
            </q-dialog>

            <!-- Inspect dialog -->
            <q-dialog v-model="inspectDialog.visible" persistent>
                <q-card style="min-width: 420px">
                    <q-card-section class="row items-center">
                        <div class="text-h6">订单复查</div>
                        <q-space />
                        <q-btn
                            flat
                            round
                            icon="close"
                            @click="inspectDialog.visible = false"
                        />
                    </q-card-section>
                    <q-card-section>
                        <div class="q-mb-md">
                            <div class="text-caption text-grey q-mb-xs">
                                复查结果
                            </div>
                            <div class="row q-gutter-md">
                                <q-radio
                                    v-model="inspectDialog.status"
                                    val="normal"
                                    label="正常"
                                    color="positive"
                                />
                                <q-radio
                                    v-model="inspectDialog.status"
                                    val="abnormal"
                                    label="异常"
                                    color="negative"
                                />
                            </div>
                        </div>
                        <div v-if="inspectDialog.status === 'abnormal'">
                            <div class="text-caption text-grey q-mb-xs">
                                异常备注 <span class="text-negative">*</span>
                            </div>
                            <q-input
                                v-model="inspectDialog.remark"
                                type="textarea"
                                :rows="3"
                                outlined
                                placeholder="请输入异常原因（必填）"
                                maxlength="500"
                                counter
                            />
                        </div>
                    </q-card-section>
                    <q-card-actions align="right">
                        <q-btn
                            flat
                            label="取消"
                            @click="inspectDialog.visible = false"
                        />
                        <q-btn
                            color="primary"
                            unelevated
                            label="提交复查"
                            :loading="inspectDialog.submitting"
                            @click="submitInspect"
                        />
                    </q-card-actions>
                </q-card>
            </q-dialog>
        </div>
    </q-page>
</template>

<script setup>
import { ref, reactive, computed, inject, onMounted, onUnmounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useQuasar } from "quasar";
import { tradeApi, authApi } from "../api";
import { useAuthStore } from "../stores/auth";

const $q = useQuasar();
const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();

const isHistory = computed(() => route.query.history === "true");

// Responsive dialog style
const dialogStyle = computed(() => {
    if ($q.screen.lt.md) {
        return {};
    }
    return {
        width: $q.screen.lt.lg ? "90vw" : "1000px",
        maxWidth: "1000px",
    };
});

const isFullscreen = ref(false);
const remainingCount = ref(null);

async function fetchRemainingCount() {
    try {
        const res = await tradeApi.list({ isHandled: false, page: 1, size: 1 });
        remainingCount.value = res.data?.total ?? null;
    } catch {
        // 静默失败，不影响主流程
    }
}

function toggleFullscreen() {
    isFullscreen.value = !isFullscreen.value;
    document.body.classList.toggle("ops-fullscreen", isFullscreen.value);
}
onUnmounted(() => {
    document.body.classList.remove("ops-fullscreen");
    window.removeEventListener("keydown", onKeyDown, true);
});

const loading = ref(false);
const skipping = ref(false);
const pending = ref(false);
const submitting = ref(false);

// Timer
const elapsedSeconds = ref(0);
let timerHandle = null;
function startTimer() {
    stopTimer();
    elapsedSeconds.value = 0;
    timerHandle = setInterval(() => {
        elapsedSeconds.value++;
    }, 1000);
}
function stopTimer() {
    if (timerHandle) {
        clearInterval(timerHandle);
        timerHandle = null;
    }
}
const elapsedFormatted = computed(() => {
    const m = Math.floor(elapsedSeconds.value / 60)
        .toString()
        .padStart(2, "0");
    const s = (elapsedSeconds.value % 60).toString().padStart(2, "0");
    return `${m}:${s}`;
});
onUnmounted(stopTimer);

const detail = ref(null);
const products = ref([]);
const quantities = reactive({});
const searchKeyword = ref("");
const currentId = ref(null);
const handleRemark = ref("");
const remarkDialogVisible = ref(false);

const historyHandleGoods = ref("");
const historyPendStatus = ref("");

// Video refs and state
const openVideos = ref([]);
const closeVideos = ref([]);
const openRates = reactive({});
const closeRates = reactive({});
const openRotations = reactive({});
const closeRotations = reactive({});
const speeds = [0.5, 1, 2, 3, 5];

function getRotation(type, idx) {
    return type === "open" ? openRotations[idx] || 0 : closeRotations[idx] || 0;
}

function setRate(type, idx, speed) {
    if (type === "open") {
        openRates[idx] = speed;
        if (openVideos.value[idx]) openVideos.value[idx].playbackRate = speed;
    } else {
        closeRates[idx] = speed;
        if (closeVideos.value[idx]) closeVideos.value[idx].playbackRate = speed;
    }
}

const filteredProducts = computed(() => {
    const kw = searchKeyword.value?.trim()?.toLowerCase() || "";
    if (!kw) return products.value;
    return products.value.filter((p) =>
        (p.goodsName || "").toLowerCase().includes(kw),
    );
});

const selectedCount = computed(() =>
    Object.values(quantities).reduce((sum, v) => sum + (v || 0), 0),
);

const selectedGoods = computed(() =>
    products.value.filter((p) => (quantities[p.goodsId] || 0) > 0),
);

function copyGoodsName(name) {
    const text = name || "未知商品";
    navigator.clipboard.writeText(text).then(() => {
        $q.notify({
            message: `已复制：${text}`,
            color: "positive",
            timeout: 1200,
        });
    });
}

function increaseCount(goodsId) {
    quantities[goodsId] = Math.min((quantities[goodsId] || 0) + 1, 50);
}

function decreaseCount(goodsId) {
    quantities[goodsId] = Math.max((quantities[goodsId] || 0) - 1, 0);
}

const totalPrice = computed(() => {
    let total = 0;
    for (const p of products.value) {
        const qty = quantities[p.goodsId] || 0;
        if (qty > 0) total += p.goodsPrice * qty;
    }
    return total.toFixed(2);
});

function formatHistoryGoods(json) {
    if (!json) return "";
    try {
        const goods = JSON.parse(json);
        if (!Array.isArray(goods) || goods.length === 0) return "无消费";
        const list = goods
            .map((g) => `${g.goodsName}×${g.goodsCount}`)
            .join("、");
        const total = goods.reduce(
            (s, g) => s + (g.goodsPrice || 0) * (g.goodsCount || 0),
            0,
        );
        return `${list}（¥${total.toFixed(2)}）`;
    } catch {
        return "";
    }
}

// Track whether any dialog is open to suppress keyboard shortcuts
const anyDialogOpen = computed(
    () =>
        remarkDialogVisible.value ||
        myGoodsDrawer.visible ||
        inspectDialog.visible ||
        branchDialog.visible,
);

function onKeyDown(e) {
    if (
        !isHistory.value &&
        !skipping.value &&
        !submitting.value &&
        !pending.value &&
        !anyDialogOpen.value
    ) {
        if (e.key === "Escape") {
            e.stopPropagation();
            goNext();
        } else if (e.ctrlKey && e.key === "Enter") {
            e.stopPropagation();
            handleSubmit();
        } else if (e.key === "g") {
            e.stopPropagation();
            handlePend();
        }
    }
}

onMounted(async () => {
    window.addEventListener("keydown", onKeyDown, true);
    fetchRemainingCount();
    const id = route.query.id;
    if (id) {
        if (!isHistory.value) {
            try {
                await tradeApi.lock(id);
            } catch {}
        }
        await loadOrder(id);
    } else {
        await fetchRandom();
    }
});

async function loadOrder(id, redirectsLeft = 5) {
    loading.value = true;
    currentId.value = id;
    detail.value = null;
    products.value = [];
    Object.keys(quantities).forEach((k) => delete quantities[k]);
    searchKeyword.value = "";
    handleRemark.value = "";
    openVideos.value = [];
    closeVideos.value = [];

    let redirected = false;
    try {
        if (!isHistory.value) {
            try {
                const checkRes = await tradeApi.check(id);
                if (checkRes.data?.alreadyHandled) {
                    if (redirectsLeft <= 0) {
                        $q.notify({
                            type: "warning",
                            message: "多个订单已被外部处理，暂无待处理订单",
                        });
                        detail.value = null;
                        return;
                    }
                    $q.notify({
                        type: "info",
                        message: "当前订单已被处理，自动跳下一单",
                    });
                    tradeApi.unlock(id);
                    redirected = true;
                    await fetchRandom(redirectsLeft - 1);
                    return;
                }
            } catch {}
        }

        const res = await tradeApi.detail(id);
        detail.value = res.data.detail;
        products.value = (res.data.products || []).map((p) => ({
            ...p,
            type: 1,
        }));
        for (const p of products.value) {
            quantities[p.goodsId] = 0;
        }
        historyHandleGoods.value = res.data.handleGoods || "";
        historyPendStatus.value = res.data.pendStatus || "";
        inspectInfo.status = res.data.inspectStatus || "";
        inspectInfo.remark = res.data.inspectRemark || "";
        inspectInfo.byName = res.data.inspectedByName || "";
        inspectInfo.at = res.data.inspectedAt
            ? new Date(res.data.inspectedAt).toLocaleString("zh-CN")
            : "";
        if (!isHistory.value) startTimer();
    } catch (err) {
        $q.notify({
            type: "negative",
            message: err?.message || "加载订单详情失败",
        });
        detail.value = null;
    } finally {
        if (!redirected) loading.value = false;
    }
}

async function fetchRandom(redirectsLeft = 5) {
    skipping.value = true;
    try {
        for (let lockRetry = 0; lockRetry < 5; lockRetry++) {
            const res = await tradeApi.randomUnhandled();
            const id = res.data.id;
            try {
                await tradeApi.lock(id);
            } catch (lockErr) {
                if (lockErr?.code === 409) continue;
                throw lockErr;
            }
            router.replace({ path: "/app/operations", query: { id } });
            await loadOrder(id, redirectsLeft);
            return;
        }
        $q.notify({ type: "warning", message: "暂无可处理订单" });
        detail.value = null;
    } catch (err) {
        $q.notify({
            type: "warning",
            message: err?.message || "暂无待处理订单",
        });
        detail.value = null;
    } finally {
        skipping.value = false;
        fetchRemainingCount();
    }
}

async function goNext() {
    if (skipping.value) return;
    skipping.value = true;
    $q.dialog({
        title: "跳过确认",
        message: "确定要跳过该订单吗？",
        cancel: { label: "取消", flat: true },
        ok: { label: "确认跳过", color: "warning" },
        persistent: true,
    })
        .onOk(async () => {
            authApi.recordSkip();
            if (currentId.value) tradeApi.unlock(currentId.value);
            await fetchRandom();
        })
        .onCancel(() => {
            skipping.value = false;
        })
        .onDismiss(() => {
            skipping.value = false;
        });
}

async function handlePend() {
    if (pending.value) return;
    pending.value = true;
    $q.dialog({
        title: "挂起确认",
        message: "确定要挂起该订单吗？挂起后订单将标记为已完结。",
        cancel: { label: "取消", flat: true },
        ok: { label: "确认挂起", color: "warning" },
        persistent: true,
    })
        .onOk(async () => {
            stopTimer();
            try {
                await tradeApi.pend(currentId.value, {
                    duration: elapsedSeconds.value,
                    remark: handleRemark.value,
                });
                tradeApi.unlock(currentId.value);
                $q.notify({
                    type: "positive",
                    message: "挂起成功，订单已完结",
                });
                await fetchRandom();
            } catch (err) {
                startTimer();
                $q.notify({
                    type: "negative",
                    message: err?.message || "挂起失败",
                });
            } finally {
                pending.value = false;
            }
        })
        .onCancel(() => {
            pending.value = false;
        })
        .onDismiss(() => {
            pending.value = false;
        });
}

async function handleSubmit() {
    submitting.value = true;
    try {
        try {
            const checkRes = await tradeApi.check(currentId.value);
            if (checkRes.data?.alreadyHandled) {
                $q.notify({
                    type: "warning",
                    message:
                        checkRes.data.message || "该订单已被处理，请处理下一单",
                });
                if (currentId.value) tradeApi.unlock(currentId.value);
                await fetchRandom();
                return;
            }
        } catch (err) {
            $q.notify({
                type: "negative",
                message: err?.message || "本订单异常，请处理其他订单",
            });
            return;
        }

        const msg =
            selectedCount.value === 0
                ? "是否确认此笔为无消费订单？"
                : `确定提交处理（${selectedCount.value}件）吗？`;
        const okLabel = selectedCount.value === 0 ? "确认" : "确认提交";

        await new Promise((resolve, reject) => {
            $q.dialog({
                title: "提交确认",
                message: msg,
                cancel: { label: "取消", flat: true },
                ok: { label: okLabel, color: "positive" },
                persistent: true,
            })
                .onOk(resolve)
                .onCancel(reject)
                .onDismiss(reject);
        });

        const orderGoodsDetailList = products.value
            .filter((p) => (quantities[p.goodsId] || 0) > 0)
            .map((p) => ({
                goodsId: p.goodsId,
                goodsName: p.goodsName,
                goodsPrice: p.goodsPrice,
                goodsImage: p.goodsImage || "",
                type: p.type || 1,
                goodsCount: quantities[p.goodsId],
            }));

        stopTimer();
        const res = await tradeApi.submit(currentId.value, {
            orderGoodsDetailList,
            duration: elapsedSeconds.value,
            remark: handleRemark.value,
        });
        tradeApi.unlock(currentId.value);
        if (res.message?.includes("完结订单状态不合法")) {
            $q.notify({ type: "warning", message: res.message });
        } else {
            $q.notify({
                type: "positive",
                message: res.message || "本订单处理成功",
            });
            if (orderGoodsDetailList.length > 0) {
                try {
                    await authApi.saveHandledGoods({
                        tradeId: currentId.value,
                        outOrderNo: detail.value?.outOrderNo || "",
                        goodsList: orderGoodsDetailList,
                        duration: elapsedSeconds.value,
                        remark: handleRemark.value,
                    });
                } catch {}
            }
        }
        await fetchRandom();
    } catch {
        // dialog cancelled — restore timer if submit was in progress
        startTimer();
    } finally {
        submitting.value = false;
    }
}

// Branch product search
const branchDialog = reactive({
    visible: false,
    keyword: "",
    loading: false,
    results: [],
});

// Open goods catalog dialog provided by MainLayout
const openGoodsCatalog = inject("openGoodsDialog", () => {});

function openBranchDialog() {
    branchDialog.visible = true;
    branchDialog.keyword = searchKeyword.value;
    branchDialog.results = [];
    searchBranchProducts();
}

function openBranchDialogWithName(goodsName) {
    branchDialog.visible = true;
    branchDialog.keyword = goodsName;
    branchDialog.results = [];
    searchBranchProducts();
}

async function searchBranchProducts() {
    if (!branchDialog.keyword.trim()) {
        $q.notify({ type: "warning", message: "请输入搜索关键词" });
        return;
    }
    branchDialog.loading = true;
    try {
        const res = await tradeApi.branchProducts(
            currentId.value,
            branchDialog.keyword.trim(),
        );
        branchDialog.results = res.data || [];
        if (!branchDialog.results.length)
            $q.notify({ type: "info", message: "未找到匹配商品" });
    } catch (err) {
        $q.notify({ type: "negative", message: err?.message || "搜索失败" });
    } finally {
        branchDialog.loading = false;
    }
}

async function addBranchProduct(p) {
    const goodsId = p.productId;
    const existing = products.value.find((item) => item.goodsId === goodsId);
    if (existing) {
        quantities[goodsId] = Math.min((quantities[goodsId] || 0) + 1, 50);
        $q.notify({ type: "positive", message: `已添加：${p.productName}` });
        branchDialog.visible = false;
        branchDialog.keyword = "";
        branchDialog.results = [];
        return;
    }
    let price = 0;
    try {
        const res = await tradeApi.productPrice(currentId.value, goodsId);
        price = res.data.productPrice ?? 0;
    } catch {
        $q.notify({ type: "warning", message: "价格查询失败，已按 0 元添加" });
    }
    products.value.unshift({
        goodsId: p.productId,
        goodsName: p.productName,
        goodsPrice: price,
        goodsImage: p.productUrl,
        type: 2,
    });
    quantities[goodsId] = 1;
    $q.notify({
        type: "positive",
        message: `已添加：${p.productName}（¥${price}）`,
    });
    branchDialog.visible = false;
    branchDialog.keyword = "";
    branchDialog.results = [];
}

// My handled goods dialog
const myGoodsDrawer = reactive({
    visible: false,
    loading: false,
    list: [],
    page: 1,
    size: 20,
    hasMore: false,
});

function openMyGoodsDrawer() {
    myGoodsDrawer.visible = true;
    myGoodsDrawer.list = [];
    myGoodsDrawer.page = 1;
    myGoodsDrawer.hasMore = false;
    loadMyGoods();
}

async function loadMyGoods() {
    myGoodsDrawer.loading = true;
    try {
        const res = await authApi.listHandledGoods({
            page: myGoodsDrawer.page,
            size: myGoodsDrawer.size,
        });
        myGoodsDrawer.list = [...myGoodsDrawer.list, ...(res.data?.records || [])];
        myGoodsDrawer.hasMore = myGoodsDrawer.list.length < (res.data?.total || 0);
    } catch (err) {
        $q.notify({ type: "negative", message: err?.message || "加载失败" });
    } finally {
        myGoodsDrawer.loading = false;
    }
}

async function loadMoreMyGoods() {
    myGoodsDrawer.page++;
    await loadMyGoods();
}

function formatTime(timeStr) {
    if (!timeStr) return "";
    const d = new Date(timeStr);
    const pad = (n) => n.toString().padStart(2, "0");
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

// Inspect dialog
const inspectInfo = reactive({ status: "", remark: "", byName: "", at: "" });
const inspectDialog = reactive({
    visible: false,
    status: "normal",
    remark: "",
    submitting: false,
});

function openInspectDialog() {
    inspectDialog.status = inspectInfo.status || "normal";
    inspectDialog.remark = inspectInfo.remark || "";
    inspectDialog.visible = true;
}

async function submitInspect() {
    if (inspectDialog.status === "abnormal" && !inspectDialog.remark.trim()) {
        $q.notify({ type: "warning", message: "标记异常时备注不能为空" });
        return;
    }
    inspectDialog.submitting = true;
    try {
        await tradeApi.inspect(currentId.value, {
            status: inspectDialog.status,
            remark: inspectDialog.remark,
        });
        inspectInfo.status = inspectDialog.status;
        inspectInfo.remark = inspectDialog.remark;
        inspectInfo.byName =
            authStore.user?.realname || authStore.user?.username || "";
        inspectInfo.at = new Date().toLocaleString("zh-CN");
        inspectDialog.visible = false;
        $q.notify({
            type: "positive",
            message:
                inspectDialog.status === "normal" ? "已标记正常" : "已标记异常",
        });
    } catch (err) {
        $q.notify({ type: "negative", message: err?.message || "提交失败" });
    } finally {
        inspectDialog.submitting = false;
    }
}
</script>

<style scoped>
.ops-page {
    padding: 0;
    height: calc(100vh - 64px); /* 减去 header 高度 64px */
}

.ops-container {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: #fff;
    border-radius: 8px;
    overflow: hidden;
}

/* 全屏模式时覆盖整个视口（包括 header） */
.ops-container.ops-fullscreen {
    height: 100vh;
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 2100; /* 高于 q-header (2000)，低于 q-dialog (6000) / q-drawer (3000) */
    border-radius: 0;
}

.center-state {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 300px;
    overflow: auto;
}

.info-bar {
    display: flex;
    flex-wrap: wrap;
    align-content: center;
    gap: 6px 20px;
    padding: 10px 14px;
    background: #f5f7fa;
    border-bottom: 1px solid #ebeef5;
    font-size: 13px;
    flex-shrink: 0; /* 不压缩，保持固定高度 */
    min-height: 44px;
}

.info-item {
    display: flex;
    align-items: center;
    gap: 4px;
}

.label {
    color: #909399;
    white-space: nowrap;
}

/* 主内容区域 - 填充剩余空间 */
.main-content {
    display: flex;
    min-height: 0; /* 重要：允许 flex 子项收缩 */
}

/* PC端 (≥1024px) - 左右结构 */
.video-panel {
    width: 55%;
    background: #1a1a1a;
    overflow-y: auto;
}
.video-block {
    margin-bottom: 30px;
}

.video-el {
    width: 100%;
    display: block;
}

.speed-bar {
    display: flex;
    gap: 10px;
    justify-content: center;
}

.product-panel {
    flex: 1;
    display: flex;
    flex-direction: column;
    background: #f6f6f6;
    overflow: hidden;
}

.product-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 5px;
}

/* 商品网格 - 自适应列数 */
.goods-grid {
    overflow-y: auto;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 10px;
    padding: 5px;
}

.goods-card {
    display: flex;
    flex-direction: column;
    min-width: 0;
}

.goods-card.selected {
    outline: 2px solid #21ba45;
}

.card-name {
    font-size: 12px;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

/* 商品图片固定比例容器 */
.goods-img-wrap {
    width: 100%;
    aspect-ratio: 1 / 1;
    overflow: hidden;
    background: #f0f0f0;
}

.goods-img-wrap img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
}

.no-img {
    width: 100%;
    aspect-ratio: 1 / 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #c0c4cc;
    font-size: 12px;
    background: #f0f0f0;
}

.card-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 4px;
    position: relative;
}

.goods-count-control {
    display: flex;
    align-items: center;
    gap: 2px;
}

.count-val {
    min-width: 20px;
    text-align: center;
}

/* 平板 (<1024px) - 上下结构，最小列宽 110px */
@media (max-width: 1023px) {
    .main-content {
        flex-direction: column;
    }

    .video-panel {
        width: 100%;
        max-height: 40vh;
        min-height: 200px;
    }

    .goods-grid {
        grid-template-columns: repeat(auto-fill, minmax(130px, 1fr));
        gap: 8px;
    }
}

/* 手机 (<600px) - 最小列宽 100px，减去底部导航栏 48px */
@media (max-width: 599px) {
    .ops-page {
        height: calc(
            100vh - 50px - 52px
        ); /* header 44px(mobile) + 底部导航栏 48px */
    }

    .video-panel {
        max-height: 35vh;
        min-height: 180px;
    }

    .goods-grid {
        grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
    }
    .card-footer {
        display: block;
    }
    .footer-price {
        position: absolute;
        bottom: 40px;
        left: 0;
    }
    .goods-count-control {
        justify-content: flex-end;
    }
}

/* 底部操作栏 - 固定在底部 */
.action-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 5px 20px;
    border-top: 1px solid rgba(0, 0, 0, 0.12);
}

.action-btns {
    display: flex;
    gap: 16px;
    flex-shrink: 0;
    flex-wrap: wrap;
}

.history-notice {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 16px;
    flex-wrap: wrap;
}

.history-goods {
    display: flex;
    align-items: center;
    font-size: 14px;
}

.history-goods-label {
    color: #909399;
}
.history-goods-content {
    color: #606266;
}

.selected-summary {
    font-size: 13px;
    color: #606266;
    display: flex;
    align-items: flex-end;
}

.selected-goods-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
    max-height: 150px;
    overflow-y: scroll;
}

.selected-goods-item {
    display: flex;
    align-items: center;
    gap: 12px;
    background: #f0f9eb;
    border: 1px solid #b3e19d;
    border-radius: 4px;
    padding: 2px 8px;
    font-size: 12px;
    white-space: nowrap;
    justify-content: space-between;
}

/* Branch drawer */
.drawer-backdrop {
    position: fixed;
    inset: 0;
    z-index: 1999;
    background: transparent;
}

.branch-drawer {
    z-index: 3000 !important;
}

.branch-results {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: 10px;
}

@media (max-width: 599px) {
    .branch-results {
        grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
        gap: 8px;
    }
}

.branch-item {
    cursor: pointer;
    border-radius: 6px;
    overflow: hidden;
    border: 1px solid #e0e0e0;
    transition: box-shadow 0.2s;
}

.branch-item:hover {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.branch-img-wrap {
    position: relative;
    width: 100%;
    padding-top: 100%;
    overflow: hidden;
}

.branch-img {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.branch-add-mask {
    position: absolute;
    inset: 0;
    background: rgba(0, 0, 0, 0.4);
    display: flex;
    align-items: center;
    justify-content: center;
    opacity: 0;
    transition: opacity 0.2s;
}

.branch-item:hover .branch-add-mask {
    opacity: 1;
}

.branch-name {
    font-size: 12px;
    padding: 4px 6px 2px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.branch-id {
    font-size: 11px;
    color: #999;
    padding: 0 6px 4px;
}

/* My goods dialog */
.my-goods-dialog {
    border-radius: 12px;
}

@media (max-width: 1023px) {
    .my-goods-dialog {
        border-radius: 8px;
    }
}

/* Mobile: no border-radius for maximized dialog */
.body--mobile .my-goods-dialog {
    border-radius: 0;
}

.my-goods-content {
    max-height: 70vh;
    overflow-y: auto;
}

@media (max-width: 599px) {
    .my-goods-content {
        max-height: calc(100vh - 120px);
    }
}

.my-goods-list {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(170px, 1fr));
    gap: 12px;
}

@media (max-width: 599px) {
    .my-goods-list {
        grid-template-columns: repeat(2, 1fr);
        gap: 8px;
    }
}

.my-goods-item {
    border: 1px solid #e0e0e0;
    border-radius: 6px;
    overflow: hidden;
}

.my-goods-img-wrap {
    position: relative;
    width: 100%;
    padding-top: 75%;
}

.my-goods-img {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.my-goods-type-tag {
    position: absolute;
    top: 4px;
    right: 4px;
}

.my-goods-info {
    padding: 6px 8px;
}

@media (max-width: 599px) {
    .my-goods-info {
        padding: 4px 6px;
    }
}

.my-goods-name {
    font-size: 13px;
    font-weight: 600;
    margin-bottom: 4px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

@media (max-width: 599px) {
    .my-goods-name {
        font-size: 12px;
    }
}

.my-goods-meta {
    font-size: 12px;
    color: #606266;
    display: flex;
    align-items: center;
    gap: 4px;
    flex-wrap: wrap;
}

@media (max-width: 599px) {
    .my-goods-meta {
        font-size: 11px;
    }
}

.my-goods-time {
    font-size: 11px;
    color: #999;
    margin-top: 2px;
}

@media (max-width: 599px) {
    .my-goods-time {
        font-size: 10px;
    }
}

/* Responsive padding utility */
.q-pa-sm-md {
    padding: 8px 12px;
}

@media (min-width: 600px) {
    .q-pa-sm-md {
        padding: 16px;
    }
}
</style>

<!-- 全屏模式下 drawer 不受 layout header 偏移影响 -->
<style>
body.ops-fullscreen .q-drawer {
    top: 0 !important;
}
</style>

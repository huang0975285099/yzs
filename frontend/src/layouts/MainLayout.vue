<template>
    <!--
    view="lHh LpR lFf"
      l = left drawer fixed
      H = header above drawer
      h = header in scroll area (lowercase = scrolls)
      L = left drawer above content
      p = page container
      R = right space
      F = footer fixed
      f = footer in scroll area
  -->
    <q-layout view="lHh LpR lFf">
        <!-- ─────────── HEADER ─────────── -->
        <q-header elevated class="bg-white text-dark">
            <q-toolbar
                :style="
                    $q.screen.lt.sm
                        ? 'height: 44px; padding: 0 10px'
                        : 'height: 64px; padding: 0 16px'
                "
            >
                <!-- Mobile: hamburger to open overlay drawer -->
                <q-btn
                    v-if="$q.screen.lt.sm"
                    flat
                    round
                    dense
                    icon="menu"
                    color="grey-7"
                    @click="desktopDrawerOpen = !desktopDrawerOpen"
                />

                <!-- Current page title -->
                <span
                    class="text-subtitle1 text-weight-medium text-dark"
                    >{{ currentTitle }}</span
                >

                <q-space />

                <q-btn
                    flat
                    dense
                    :round="$q.screen.lt.sm"
                    color="primary"
                    icon="inventory_2"
                    :label="$q.screen.lt.sm ? '' : '商品大全'"
                    class="goods-btn"
                    @click="goodsDialog.visible = true"
                />

                <!-- User dropdown -->
                <q-btn flat dense round class="user-btn">
                    <q-avatar
                        size="32px"
                        color="primary"
                        text-color="white"
                        style="font-size: 14px"
                    >
                        {{
                            authStore.user?.realname?.[0] ||
                            authStore.user?.username?.[0] ||
                            "U"
                        }}
                    </q-avatar>
                    <span
                        v-if="!$q.screen.lt.sm"
                        class="q-ml-sm text-body2 text-dark"
                    >
                        {{
                            authStore.user?.realname || authStore.user?.username
                        }}
                    </span>
                    <q-icon
                        v-if="!$q.screen.lt.sm"
                        name="arrow_drop_down"
                        color="grey-7"
                    />

                    <q-menu anchor="bottom right" self="top right">
                        <q-list style="min-width: 160px">
                            <q-item dense>
                                <q-item-section>
                                    <div class="text-caption text-grey">
                                        {{ roleLabel }}
                                    </div>
                                </q-item-section>
                            </q-item>
                            <q-separator />
                            <q-item
                                clickable
                                v-close-popup
                                @click="$router.push('/')"
                            >
                                <q-item-section avatar
                                    ><q-icon name="home"
                                /></q-item-section>
                                <q-item-section>返回首页</q-item-section>
                            </q-item>
                            <q-item
                                clickable
                                v-close-popup
                                @click="showPasswordDialog = true"
                            >
                                <q-item-section avatar
                                    ><q-icon name="lock"
                                /></q-item-section>
                                <q-item-section>修改密码</q-item-section>
                            </q-item>
                            <q-item
                                clickable
                                v-close-popup
                                @click="handleLogout"
                            >
                                <q-item-section avatar
                                    ><q-icon name="logout"
                                /></q-item-section>
                                <q-item-section>退出登录</q-item-section>
                            </q-item>
                        </q-list>
                    </q-menu>
                </q-btn>
            </q-toolbar>
        </q-header>

        <!-- ─────────── SIDEBAR DRAWER ─────────── -->
        <!--
      Desktop (≥1024px): always visible, full width 220px, mini=false
      Tablet  (600-1023px): always visible, mini=true (60px icon mode)
      Mobile  (<600px): overlay drawer, opened by hamburger button
    -->
        <q-drawer
            v-model="desktopDrawerOpen"
            :mini="miniMode"
            :width="220"
            :mini-width="60"
            :breakpoint="600"
            show-if-above
            class="sidebar-dark"
        >
            <div
                class="sidebar-logo"
                style="
                    height: 64px;
                    display: flex;
                    align-items: center;
                    justify-content: space-between;
                    padding: 0 20px;
                    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
                "
            >
                <div style="display: flex; align-items: center">
                    <q-icon
                        name="cloud"
                        size="26px"
                        color="white"
                        @click="miniMode = !miniMode"
                    />
                    <span
                        v-if="!miniMode"
                        style="
                            color: #fff;
                            font-size: 16px;
                            margin-left: 8px;
                            font-weight: bold;
                        "
                    >
                        云值守系统
                    </span>
                </div>
            </div>
            <q-scroll-area
                style="width: 100%; height: calc(100% - 64px - 48px)"
            >
                <q-list padding>
                    <template v-if="authStore.user?.role === 'operator'">
                        <q-item
                            clickable
                            v-ripple
                            to="/app/my-handled"
                            active-class="active-menu-item"
                        >
                            <q-item-section avatar>
                                <q-icon name="task_alt" />
                            </q-item-section>
                            <q-item-section>我的处理记录</q-item-section>
                            <q-tooltip
                                v-if="miniMode"
                                anchor="center right"
                                self="center left"
                            >
                                我的处理记录
                            </q-tooltip>
                        </q-item>
                    </template>

                    <template
                        v-else-if="authStore.user?.role === 'statistician'"
                    >
                        <q-item
                            clickable
                            v-ripple
                            to="/app/stats"
                            active-class="active-menu-item"
                        >
                            <q-item-section avatar
                                ><q-icon name="bar_chart"
                            /></q-item-section>
                            <q-item-section>操作员统计</q-item-section>
                            <q-tooltip
                                v-if="miniMode"
                                anchor="center right"
                                self="center left"
                            >
                                操作员统计
                            </q-tooltip>
                        </q-item>
                    </template>

                    <template v-else-if="authStore.user?.role === 'inspector'">
                        <q-item
                            v-if="authStore.reviewEnabled"
                            clickable
                            v-ripple
                            to="/app/quality-check"
                            active-class="active-menu-item"
                        >
                            <q-item-section avatar
                                ><q-icon name="fact_check"
                            /></q-item-section>
                            <q-item-section>质检审核</q-item-section>
                            <q-tooltip
                                v-if="miniMode"
                                anchor="center right"
                                self="center left"
                                >质检审核</q-tooltip
                            >
                        </q-item>
                        <q-item
                            v-else
                            clickable
                            v-ripple
                            to="/app/quality-review"
                            active-class="active-menu-item"
                        >
                            <q-item-section avatar
                                ><q-icon name="rate_review"
                            /></q-item-section>
                            <q-item-section>质检复查</q-item-section>
                            <q-tooltip
                                v-if="miniMode"
                                anchor="center right"
                                self="center left"
                                >质检复查</q-tooltip
                            >
                        </q-item>
                    </template>

                    <template v-else>
                        <!-- Admin: full menu -->
                        <q-item
                            clickable
                            v-ripple
                            to="/app/dashboard"
                            active-class="active-menu-item"
                        >
                            <q-item-section avatar
                                ><q-icon name="dashboard"
                            /></q-item-section>
                            <q-item-section>数据看板</q-item-section>
                            <q-tooltip
                                v-if="miniMode"
                                anchor="center right"
                                self="center left"
                                >数据看板</q-tooltip
                            >
                        </q-item>
                        <q-item
                            clickable
                            v-ripple
                            to="/app/orders"
                            active-class="active-menu-item"
                        >
                            <q-item-section avatar
                                ><q-icon name="list_alt"
                            /></q-item-section>
                            <q-item-section>异常订单列表</q-item-section>
                            <q-tooltip
                                v-if="miniMode"
                                anchor="center right"
                                self="center left"
                                >异常订单列表</q-tooltip
                            >
                        </q-item>
                        <q-item
                            clickable
                            v-ripple
                            to="/app/operations"
                            active-class="active-menu-item"
                        >
                            <q-item-section avatar
                                ><q-icon name="play_circle"
                            /></q-item-section>
                            <q-item-section>处理订单</q-item-section>
                            <q-tooltip
                                v-if="miniMode"
                                anchor="center right"
                                self="center left"
                                >处理订单</q-tooltip
                            >
                        </q-item>
                        <q-item
                            clickable
                            v-ripple
                            to="/app/my-handled"
                            active-class="active-menu-item"
                        >
                            <q-item-section avatar
                                ><q-icon name="task_alt"
                            /></q-item-section>
                            <q-item-section>我的处理记录</q-item-section>
                            <q-tooltip
                                v-if="miniMode"
                                anchor="center right"
                                self="center left"
                                >我的处理记录</q-tooltip
                            >
                        </q-item>
                        <q-item
                            clickable
                            v-ripple
                            to="/app/stats"
                            active-class="active-menu-item"
                        >
                            <q-item-section avatar
                                ><q-icon name="bar_chart"
                            /></q-item-section>
                            <q-item-section>操作员统计</q-item-section>
                            <q-tooltip
                                v-if="miniMode"
                                anchor="center right"
                                self="center left"
                                >操作员统计</q-tooltip
                            >
                        </q-item>
                        <q-item
                            v-if="authStore.reviewEnabled"
                            clickable
                            v-ripple
                            to="/app/quality-check"
                            active-class="active-menu-item"
                        >
                            <q-item-section avatar
                                ><q-icon name="fact_check"
                            /></q-item-section>
                            <q-item-section>质检审核</q-item-section>
                            <q-tooltip
                                v-if="miniMode"
                                anchor="center right"
                                self="center left"
                                >质检审核</q-tooltip
                            >
                        </q-item>
                        <q-item
                            v-else
                            clickable
                            v-ripple
                            to="/app/quality-review"
                            active-class="active-menu-item"
                        >
                            <q-item-section avatar
                                ><q-icon name="rate_review"
                            /></q-item-section>
                            <q-item-section>质检复查</q-item-section>
                            <q-tooltip
                                v-if="miniMode"
                                anchor="center right"
                                self="center left"
                                >质检复查</q-tooltip
                            >
                        </q-item>
                        <q-item
                            v-if="authStore.isAdmin"
                            clickable
                            v-ripple
                            to="/app/users"
                            active-class="active-menu-item"
                        >
                            <q-item-section avatar
                                ><q-icon name="manage_accounts"
                            /></q-item-section>
                            <q-item-section>用户管理</q-item-section>
                            <q-tooltip
                                v-if="miniMode"
                                anchor="center right"
                                self="center left"
                                >用户管理</q-tooltip
                            >
                        </q-item>
                        <q-item
                            v-if="authStore.isAdmin"
                            clickable
                            v-ripple
                            to="/app/teams"
                            active-class="active-menu-item"
                        >
                            <q-item-section avatar
                                ><q-icon name="groups"
                            /></q-item-section>
                            <q-item-section>团队管理</q-item-section>
                            <q-tooltip
                                v-if="miniMode"
                                anchor="center right"
                                self="center left"
                                >团队管理</q-tooltip
                            >
                        </q-item>
                    </template>
                </q-list>
            </q-scroll-area>
            <div
                style="
                    height: 48px;
                    display: flex;
                    align-items: center;
                    justify-content: flex-end;
                    padding: 0 12px;
                    border-top: 1px solid rgba(255, 255, 255, 0.1);
                "
            >
                <q-btn
                    flat
                    round
                    dense
                    :icon="miniMode ? 'chevron_right' : 'chevron_left'"
                    color="grey-5"
                    @click="miniMode = !miniMode"
                />
            </div>
        </q-drawer>

        <!-- ─────────── PAGE CONTENT ─────────── -->
        <q-page-container>
            <router-view />
        </q-page-container>

        <!-- ─────────── MOBILE BOTTOM NAV ─────────── -->
        <q-footer v-if="$q.screen.lt.sm" elevated class="bg-white">
            <q-tabs
                dense
                active-color="primary"
                indicator-color="primary"
                align="justify"
                class="text-grey-7"
                no-caps
            >
                <!-- Operator -->
                <template v-if="authStore.user?.role === 'operator'">
                    <q-route-tab
                        name="my-handled"
                        icon="task_alt"
                        label="我的记录"
                        to="/app/my-handled"
                        exact
                    />
                </template>

                <!-- Statistician -->
                <template v-else-if="authStore.user?.role === 'statistician'">
                    <q-route-tab
                        name="stats"
                        icon="bar_chart"
                        label="统计"
                        to="/app/stats"
                        exact
                    />
                </template>

                <!-- Inspector -->
                <template v-else-if="authStore.user?.role === 'inspector'">
                    <q-route-tab
                        v-if="authStore.reviewEnabled"
                        name="quality-check"
                        icon="fact_check"
                        label="质检审核"
                        to="/app/quality-check"
                        exact
                    />
                    <q-route-tab
                        v-else
                        name="quality-review"
                        icon="rate_review"
                        label="质检复查"
                        to="/app/quality-review"
                        exact
                    />
                </template>

                <!-- Admin -->
                <template v-else>
                    <q-route-tab
                        name="dashboard"
                        icon="dashboard"
                        label="看板"
                        to="/app/dashboard"
                        exact
                    />
                    <q-route-tab
                        name="operations"
                        icon="play_circle"
                        label="处理"
                        to="/app/operations"
                        exact
                    />
                    <q-route-tab
                        name="my-handled"
                        icon="task_alt"
                        label="记录"
                        to="/app/my-handled"
                        exact
                    />
                    <q-route-tab
                        name="orders"
                        icon="list_alt"
                        label="订单"
                        to="/app/orders"
                        exact
                    />
                    <q-tab
                        name="more"
                        icon="more_horiz"
                        label="更多"
                        @click="desktopDrawerOpen = true"
                    />
                </template>
            </q-tabs>
        </q-footer>

        <!-- 修改密码对话框 -->
        <q-dialog v-model="showPasswordDialog" persistent>
            <q-card style="min-width: 350px">
                <q-card-section>
                    <div class="text-h6">修改密码</div>
                </q-card-section>

                <q-card-section>
                    <q-form @submit="handleChangePassword" class="q-gutter-md">
                        <q-input
                            v-model="passwordForm.oldPassword"
                            type="password"
                            label="原密码"
                            :rules="[(v) => !!v || '请输入原密码']"
                            outlined
                            dense
                        />
                        <q-input
                            v-model="passwordForm.newPassword"
                            type="password"
                            label="新密码"
                            :rules="[
                                (v) => !!v || '请输入新密码',
                                (v) => v.length >= 6 || '密码至少6位',
                            ]"
                            outlined
                            dense
                        />
                        <q-input
                            v-model="passwordForm.confirmPassword"
                            type="password"
                            label="确认新密码"
                            :rules="[
                                (v) => !!v || '请确认新密码',
                                (v) =>
                                    v === passwordForm.newPassword ||
                                    '两次密码不一致',
                            ]"
                            outlined
                            dense
                        />
                    </q-form>
                </q-card-section>

                <q-card-actions align="right">
                    <q-btn
                        flat
                        label="取消"
                        color="grey"
                        @click="resetPasswordForm"
                    />
                    <q-btn
                        unelevated
                        label="确定"
                        color="primary"
                        @click="handleChangePassword"
                        :loading="passwordLoading"
                    />
                </q-card-actions>
            </q-card>
        </q-dialog>

        <!-- Goods catalog dialog -->
        <q-dialog v-model="goodsDialog.visible" :maximized="$q.screen.lt.md">
            <q-card class="goods-dialog" :style="goodsDialogStyle">
                <q-card-section class="row items-center q-pa-sm-md">
                    <div class="text-subtitle1 text-weight-medium">
                        商品大全
                    </div>
                    <q-space />
                    <q-btn flat round dense icon="close" v-close-popup />
                </q-card-section>
                <div class="row q-gutter-sm q-pa-sm-md" style="padding-top: 0;padding-bottom: 0;">
                    <q-input
                        v-model="goodsDialog.keyword"
                        dense
                        outlined
                        placeholder="搜索商品名称"
                        clearable
                        @clear="loadGoods"
                        @keyup.enter="loadGoods"
                        class="col"
                    >
                        <template #append>
                            <q-btn
                                flat
                                dense
                                round
                                icon="search"
                                @click="loadGoods"
                            />
                        </template>
                    </q-input>
                    <q-btn
                        :outline="!goodsDialog.showFavorites"
                        :color="
                            goodsDialog.showFavorites ? 'primary' : 'grey-7'
                        "
                        icon="favorite"
                        :label="$q.screen.lt.sm ? '' : '我的收藏'"
                        @click="toggleFavorites"
                    />
                </div>

                <q-card-section class="goods-content">
                    <q-inner-loading :showing="goodsDialog.loading" />
                    <div v-if="goodsDialog.list.length" class="goods-grid">
                        <q-card
                            v-for="item in goodsDialog.list"
                            :key="item.id"
                            flat
                            bordered
                            class="goods-item"
                        >
                            <div class="goods-img-wrap">
                                <img
                                    v-if="item.frontImg"
                                    :src="formatImageUrl(item.frontImg)"
                                    class="goods-img"
                                    alt=""
                                />
                                <div v-else class="goods-img no-img">无图</div>
                                <q-btn
                                    flat
                                    dense
                                    round
                                    size="sm"
                                    icon="image"
                                    color="white"
                                    class="goods-zoom-btn"
                                    @click="openImageViewer(item)"
                                />
                                <q-btn
                                    flat
                                    dense
                                    round
                                    size="sm"
                                    :icon="
                                        goodsDialog.favorites[item.id]
                                            ? 'favorite'
                                            : 'favorite_border'
                                    "
                                    :color="
                                        goodsDialog.favorites[item.id]
                                            ? 'red'
                                            : 'white'
                                    "
                                    class="goods-favorite-btn"
                                    @click="toggleFavorite(item)"
                                />
                            </div>
                            <div class="goods-info">
                                <div class="goods-title">{{ item.title }}</div>
                                <div class="goods-sn">编号: {{ item.sn }}</div>
                            </div>
                        </q-card>
                    </div>
                    <div
                        v-else-if="!goodsDialog.loading"
                        class="text-center q-pa-xl"
                    >
                        <q-icon name="inventory_2" size="48px" color="grey-4" />
                        <div class="text-grey-6 q-mt-sm">暂无商品</div>
                    </div>
                </q-card-section>

                <q-card-actions
                    v-if="goodsDialog.hasMore"
                    align="center"
                >
                    <q-btn
                        flat
                        :loading="goodsDialog.loading"
                        label="加载更多"
                        @click="loadMoreGoods"
                    />
                </q-card-actions>
            </q-card>
        </q-dialog>

        <!-- Image viewer dialog -->
        <q-dialog v-model="imageViewer.visible" :maximized="$q.screen.lt.md">
            <q-card class="image-viewer-dialog" :style="imageViewerStyle">
                <q-card-section class="row items-center q-pa-sm-md">
                    <div class="text-subtitle1 text-weight-medium">
                        {{ imageViewer.goods?.title }}
                    </div>
                    <q-space />
                    <q-btn flat round dense icon="close" v-close-popup />
                </q-card-section>

                <q-card-section class="image-viewer-content">
                    <div class="image-viewer-wrapper">
                        <q-btn
                            flat
                            round
                            dense
                            icon="chevron_left"
                            class="image-nav-btn image-nav-prev"
                            @click="prevImage"
                            :disable="imageViewer.currentIndex === 0"
                        />
                        <img
                            :src="currentViewImage"
                            class="viewer-image"
                            alt=""
                        />
                        <q-btn
                            flat
                            round
                            dense
                            icon="chevron_right"
                            class="image-nav-btn image-nav-next"
                            @click="nextImage"
                            :disable="imageViewer.currentIndex === 5"
                        />
                    </div>
                    <div class="image-viewer-labels">
                        <q-chip
                            v-for="(label, idx) in imageLabels"
                            :key="idx"
                            :color="
                                idx === imageViewer.currentIndex
                                    ? 'primary'
                                    : 'grey-3'
                            "
                            :text-color="
                                idx === imageViewer.currentIndex
                                    ? 'white'
                                    : 'grey-7'
                            "
                            dense
                            clickable
                            @click="imageViewer.currentIndex = idx"
                            class="image-label-chip"
                        >
                            {{ label }}
                        </q-chip>
                    </div>
                </q-card-section>
            </q-card>
        </q-dialog>
    </q-layout>
</template>

<script setup>
import { ref, computed, reactive, watch, provide } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useQuasar } from "quasar";
import { useAuthStore } from "../stores/auth";
import { authApi, goodsApi, favoriteApi } from "../api";

const $q = useQuasar();
const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();

// Tablet mini mode (600–1023px): icon-only sidebar
const miniMode = ref(true);

// Desktop/tablet persistent drawer state
const desktopDrawerOpen = ref(false);

const currentTitle = computed(() => route.meta?.title || "云值守系统");

const roleLabel = computed(() => {
    const map = {
        admin: "管理员",
        statistician: "统计员",
        operator: "操作员",
        inspector: "质检员",
    };
    return map[authStore.user?.role] || authStore.user?.role || "";
});

// 修改密码相关
const showPasswordDialog = ref(false);
const passwordLoading = ref(false);
const passwordForm = ref({
    oldPassword: "",
    newPassword: "",
    confirmPassword: "",
});

function resetPasswordForm() {
    showPasswordDialog.value = false;
    passwordForm.value = {
        oldPassword: "",
        newPassword: "",
        confirmPassword: "",
    };
}

async function handleChangePassword() {
    if (passwordForm.value.newPassword !== passwordForm.value.confirmPassword) {
        $q.notify({ type: "warning", message: "两次密码不一致" });
        return;
    }

    if (passwordForm.value.newPassword.length < 6) {
        $q.notify({ type: "warning", message: "密码至少6位" });
        return;
    }

    passwordLoading.value = true;
    try {
        await authApi.changePassword({
            oldPassword: passwordForm.value.oldPassword,
            newPassword: passwordForm.value.newPassword,
        });
        $q.notify({ type: "positive", message: "密码修改成功" });
        resetPasswordForm();
    } catch (err) {
        $q.notify({
            type: "negative",
            message: err?.message || "密码修改失败",
        });
    } finally {
        passwordLoading.value = false;
    }
}

async function handleLogout() {
    $q.dialog({
        title: "提示",
        message: "确定要退出登录吗？",
        cancel: { label: "取消", flat: true },
        ok: { label: "退出", color: "negative" },
        persistent: true,
    }).onOk(async () => {
        await authStore.logout();
        $q.notify({ type: "positive", message: "已退出登录" });
        router.push("/");
    });
}

// Goods catalog dialog
const goodsDialog = reactive({
    visible: false,
    loading: false,
    keyword: "",
    list: [],
    page: 1,
    pageSize: 20,
    total: 0,
    hasMore: false,
    favorites: {}, // { goodsId: true }
    favoriteList: [], // 用户收藏列表
    showFavorites: false, // 是否只显示收藏
});

provide("openGoodsDialog", (keyword) => {
    goodsDialog.keyword = keyword ?? "";
    goodsDialog.list = [];
    goodsDialog.visible = true;
});

const goodsDialogStyle = computed(() => {
    if ($q.screen.lt.md) {
        return {};
    }
    return {
        width: $q.screen.lt.lg ? "90vw" : "1000px",
        maxWidth: "1000px",
    };
});

async function loadGoods() {
    goodsDialog.loading = true;
    goodsDialog.page = 1;
    goodsDialog.showFavorites = false;
    try {
        const res = await goodsApi.list({
            page: goodsDialog.page,
            size: goodsDialog.pageSize,
            keyword: goodsDialog.keyword,
        });
        goodsDialog.list = res.data?.records || [];
        goodsDialog.total = res.data?.total || 0;
        goodsDialog.hasMore = goodsDialog.list.length < goodsDialog.total;
        // 加载收藏状态
        await loadFavoritesStatus();
    } catch (err) {
        $q.notify({ type: "negative", message: err?.message || "加载失败" });
    } finally {
        goodsDialog.loading = false;
    }
}

async function loadMoreGoods() {
    goodsDialog.loading = true;
    goodsDialog.page++;
    try {
        const res = await goodsApi.list({
            page: goodsDialog.page,
            size: goodsDialog.pageSize,
            keyword: goodsDialog.keyword,
        });
        goodsDialog.list = [...goodsDialog.list, ...(res.data?.records || [])];
        goodsDialog.hasMore = goodsDialog.list.length < goodsDialog.total;
    } catch (err) {
        goodsDialog.page--;
        $q.notify({ type: "negative", message: err?.message || "加载失败" });
    } finally {
        goodsDialog.loading = false;
    }
}

// Watch dialog visibility to load data
watch(
    () => goodsDialog.visible,
    (val) => {
        if (val && goodsDialog.list.length === 0) {
            loadGoods();
        }
    },
);

// Load favorites status
async function loadFavoritesStatus() {
    try {
        const goodsIds = goodsDialog.list.map((item) => item.id).join(",");
        if (!goodsIds) return;
        const res = await favoriteApi.check(goodsIds);
        goodsDialog.favorites = res.data?.favorites || {};
    } catch (err) {
        console.error("Failed to load favorites status:", err);
    }
}

// Toggle favorite
async function toggleFavorite(item) {
    try {
        if (goodsDialog.favorites[item.id]) {
            await favoriteApi.remove(item.id);
            delete goodsDialog.favorites[item.id];
            $q.notify({ type: "positive", message: "已取消收藏" });
        } else {
            await favoriteApi.add({
                goodsId: item.id,
                title: item.title,
                sn: item.sn,
                frontImg: item.frontImg,
            });
            goodsDialog.favorites[item.id] = true;
            $q.notify({ type: "positive", message: "已收藏" });
        }
    } catch (err) {
        $q.notify({ type: "negative", message: err?.message || "操作失败" });
    }
}

// Toggle favorites filter
async function toggleFavorites() {
    if (goodsDialog.showFavorites) {
        // 关闭收藏过滤，显示全部商品
        goodsDialog.showFavorites = false;
        loadGoods();
    } else {
        // 显示收藏列表
        goodsDialog.showFavorites = true;
        goodsDialog.loading = true;
        try {
            const res = await favoriteApi.list();
            goodsDialog.favoriteList = res.data?.records || [];
            goodsDialog.list = goodsDialog.favoriteList.map((f) => ({
                id: f.goodsId,
                title: f.title,
                sn: f.sn,
                frontImg: f.frontImg,
            }));
            goodsDialog.total = goodsDialog.list.length;
            goodsDialog.hasMore = false;
            // 所有收藏的商品都标记为已收藏
            goodsDialog.favorites = {};
            goodsDialog.list.forEach((item) => {
                goodsDialog.favorites[item.id] = true;
            });
        } catch (err) {
            $q.notify({
                type: "negative",
                message: err?.message || "加载失败",
            });
        } finally {
            goodsDialog.loading = false;
        }
    }
}

// Image viewer
const imageLabels = ["正面", "背面", "左侧", "右侧", "顶部", "底部"];

const imageViewer = reactive({
    visible: false,
    goods: null,
    currentIndex: 0,
});

const imageViewerStyle = computed(() => {
    if ($q.screen.lt.md) {
        return {};
    }
    return {
        width: "800px",
        maxWidth: "90vw",
    };
});

const currentViewImage = computed(() => {
    if (!imageViewer.goods) return "";
    const images = [
        imageViewer.goods.frontImg,
        imageViewer.goods.backImg,
        imageViewer.goods.leftImg,
        imageViewer.goods.rightImg,
        imageViewer.goods.topImg,
        imageViewer.goods.bottomImg,
    ];
    const url = images[imageViewer.currentIndex] || "";
    return formatImageUrl(url);
});

function formatImageUrl(url) {
    if (!url) return "";
    const prefix =
        "https://ubox-goods-image.oss-cn-beijing.aliyuncs.com/segment-image/";
    if (url.startsWith("http://") || url.startsWith("https://")) {
        return url;
    }
    return prefix + url;
}

function openImageViewer(goods) {
    imageViewer.goods = goods;
    imageViewer.currentIndex = 0;
    imageViewer.visible = true;
}

function prevImage() {
    if (imageViewer.currentIndex > 0) {
        imageViewer.currentIndex--;
    }
}

function nextImage() {
    if (imageViewer.currentIndex < 5) {
        imageViewer.currentIndex++;
    }
}
</script>

<style lang="scss" scoped>
// Active menu item — dark sidebar (PC/tablet)
.sidebar-dark :deep(.active-menu-item) {
    color: #fff !important;
    background: rgba(24, 144, 255, 0.25) !important;
    border-right: 3px solid #1890ff;
}

// Active menu item — light sidebar (mobile)
:deep(.q-drawer:not(.sidebar-dark) .active-menu-item) {
    color: #1890ff !important;
    background: #e6f4ff !important;
    border-right: 3px solid #1890ff;
}

:deep(.q-item__section--avatar) {
    min-width: unset;
    padding-right: 12px;
}

:deep(.q-item__section--avatar .q-icon) {
    font-size: 20px;
}

.goods-btn {
    border-radius: 6px;
    padding: 4px 12px;
    font-weight: 500;

    &:hover {
        background: rgba(25, 118, 210, 0.08);
    }
}

.user-btn {
    border-radius: 6px;
    padding: 4px 8px;

    &:hover {
        background: #f5f5f5;
    }
}

:deep(.q-tab__label) {
    line-height: 16px;
}

/* Goods catalog dialog */
.goods-dialog {
    border-radius: 12px;
}

@media (max-width: 1023px) {
    .goods-dialog {
        border-radius: 8px;
    }
}

.goods-content {
    max-height: 70vh;
    overflow-y: auto;
}

@media (max-width: 599px) {
    .goods-content {
        max-height: calc(100vh - 145px);
    }
}

.goods-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 12px;
}

@media (max-width: 599px) {
    .goods-grid {
        grid-template-columns: repeat(2, 1fr);
        gap: 8px;
    }
}

.goods-item {
    border-radius: 6px;
    overflow: hidden;
}

.goods-img-wrap {
    position: relative;
    width: 100%;
    padding-top: 75%;
    background: #f5f5f5;
}

.goods-img {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.goods-img.no-img {
    display: flex;
    align-items: center;
    justify-content: center;
    color: #999;
    font-size: 12px;
}

.goods-zoom-btn {
    position: absolute;
    top: 4px;
    right: 4px;
    background: rgba(0, 0, 0, 0.4);
    border-radius: 4px;
    min-width: 28px;
    min-height: 28px;
}

.goods-zoom-btn:hover {
    background: rgba(0, 0, 0, 0.6);
}

.goods-favorite-btn {
    position: absolute;
    top: 4px;
    left: 4px;
    background: rgba(0, 0, 0, 0.4);
    border-radius: 4px;
    min-width: 28px;
    min-height: 28px;
}

.goods-favorite-btn:hover {
    background: rgba(0, 0, 0, 0.6);
}

.goods-info {
    padding: 8px;
}

@media (max-width: 599px) {
    .goods-info {
        padding: 6px;
    }
}

.goods-title {
    font-size: 13px;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

@media (max-width: 599px) {
    .goods-title {
        font-size: 12px;
    }
}

.goods-sn {
    font-size: 11px;
    color: #999;
    margin-top: 2px;
}

/* Image viewer dialog */
.image-viewer-dialog {
    border-radius: 12px;
}

@media (max-width: 1023px) {
    .image-viewer-dialog {
        border-radius: 8px;
    }
}

.image-viewer-content {
    padding: 16px;
}

@media (max-width: 599px) {
    .image-viewer-content {
        padding: 8px;
    }
}

.image-viewer-wrapper {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 16px;
    min-height: 400px;
}

@media (max-width: 599px) {
    .image-viewer-wrapper {
        min-height: 280px;
        gap: 8px;
    }
}

.viewer-image {
    max-width: 100%;
    max-height: 500px;
    border-radius: 8px;
    object-fit: contain;
}

@media (max-width: 599px) {
    .viewer-image {
        max-height: calc(100vh - 200px);
    }
}

.image-nav-btn {
    background: rgba(0, 0, 0, 0.1);
    border-radius: 50%;
}

.image-nav-btn:hover {
    background: rgba(0, 0, 0, 0.2);
}

@media (max-width: 599px) {
    .image-nav-btn {
        padding: 8px;
    }
}

.image-viewer-labels {
    display: flex;
    justify-content: center;
    gap: 8px;
    margin-top: 16px;
    flex-wrap: wrap;
}

@media (max-width: 599px) {
    .image-viewer-labels {
        margin-top: 8px;
        gap: 4px;
    }
}

.image-label-chip {
    cursor: pointer;
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

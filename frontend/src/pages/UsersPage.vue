<template>
    <q-page class="q-pa-md">
        <!-- Records section -->
        <q-card flat bordered>
            <q-card-section>
                <div class="page-header">
                    <div class="row q-gutter-md">
                        <q-input
                            v-model="filter"
                            dense
                            outlined
                            label="搜索用户..."
                            style="min-width: 200px"
                        >
                            <template #prepend>
                                <q-icon name="search" />
                            </template>
                        </q-input>
                        <q-select
                            v-model="roleFilter"
                            dense
                            outlined
                            label="角色筛选"
                            :options="roleFilterOptions"
                            emit-value
                            map-options
                            clearable
                            style="min-width: 150px"
                        >
                            <template #prepend>
                                <q-icon name="filter_list" />
                            </template>
                        </q-select>
                        <q-select
                            v-model="teamFilter"
                            dense
                            outlined
                            label="团队筛选"
                            :options="teamFilterOptions"
                            emit-value
                            map-options
                            clearable
                            style="min-width: 150px"
                        >
                            <template #prepend>
                                <q-icon name="groups" />
                            </template>
                        </q-select>
                    </div>
                    <q-btn
                        color="primary"
                        icon="add"
                        label="新增用户"
                        unelevated
                        @click="openCreateDialog"
                    />
                    <q-btn
                        v-if="authStore.user?.role === 'admin'"
                        color="secondary"
                        icon="history"
                        label="登录日志"
                        outline
                        unelevated
                        @click="openAllLoginLogs"
                    />
                </div>

                <div class="table-scroll">
                    <q-table
                        :rows="filteredUsers"
                        :columns="columns"
                        :loading="loading"
                        :pagination="{ rowsPerPage: 20 }"
                        :rows-per-page-options="[10, 20, 30, 50]"
                        row-key="id"
                        flat
                        bordered
                        separator="cell"
                    >
                        <template #body-cell-role="props">
                            <q-td :props="props">
                                <q-chip
                                    dense
                                    :class="roleChipClass(props.row.role)"
                                    class="q-ma-none"
                                >
                                    {{ roleLabel(props.row.role) }}
                                </q-chip>
                            </q-td>
                        </template>
                        <template #body-cell-createdAt="props">
                            <q-td :props="props">{{
                                formatTime(props.row.createdAt)
                            }}</q-td>
                        </template>
                        <template #body-cell-actions="props">
                            <q-td :props="props" class="q-gutter-sm">
                                <q-btn
                                    unelevated
                                    size="sm"
                                    color="primary"
                                    label="编辑"
                                    @click="openEditDialog(props.row)"
                                />
                                <q-btn
                                    v-if="authStore.user?.role === 'admin'"
                                    unelevated
                                    size="sm"
                                    color="info"
                                    label="登录历史"
                                    @click="openUserLoginLogs(props.row)"
                                />
                                <q-btn
                                    unelevated
                                    size="sm"
                                    color="negative"
                                    label="删除"
                                    :disable="
                                        props.row.id === authStore.user?.id
                                    "
                                    @click="handleDelete(props.row)"
                                />
                            </q-td>
                        </template>
                    </q-table>
                </div>
            </q-card-section>
        </q-card>

        <!-- Create / Edit Dialog -->
        <q-dialog v-model="dialogVisible" @hide="resetForm" persistent>
            <q-card style="min-width: 380px; max-width: 480px; width: 100%">
                <q-card-section class="row items-center q-pb-none">
                    <div class="text-h6">
                        {{ editingUser ? "编辑用户" : "新增用户" }}
                    </div>
                    <q-space />
                    <q-btn icon="close" flat round dense v-close-popup />
                </q-card-section>

                <q-card-section>
                    <q-form ref="formRef" class="q-gutter-md">
                        <q-input
                            v-model="form.username"
                            outlined
                            label="用户名"
                            :disable="!!editingUser"
                            :rules="[
                                (v) => !!v || '请输入用户名',
                                (v) => v.length >= 3 || '至少3个字符',
                            ]"
                        />
                        <q-input
                            v-model="form.realname"
                            outlined
                            label="姓名"
                        />
                        <q-input
                            v-model="form.password"
                            outlined
                            :type="showPwd ? 'text' : 'password'"
                            :label="
                                editingUser ? '新密码（不填则不修改）' : '密码'
                            "
                            :rules="
                                editingUser
                                    ? [
                                          (v) =>
                                              !v ||
                                              v.length >= 6 ||
                                              '密码至少6位',
                                      ]
                                    : [
                                          (v) => !!v || '请输入密码',
                                          (v) => v.length >= 6 || '密码至少6位',
                                      ]
                            "
                        >
                            <template #append>
                                <q-icon
                                    :name="
                                        showPwd
                                            ? 'visibility_off'
                                            : 'visibility'
                                    "
                                    class="cursor-pointer"
                                    @click="showPwd = !showPwd"
                                />
                            </template>
                        </q-input>
                        <q-select
                            v-model="form.role"
                            outlined
                            label="角色"
                            :options="roleOptions"
                            emit-value
                            map-options
                            :rules="[(v) => !!v || '请选择角色']"
                        />
                        <q-select
                            v-model="form.teamId"
                            outlined
                            label="所属团队（可选）"
                            :options="teamFilterOptions"
                            emit-value
                            map-options
                            clearable
                        />
                    </q-form>
                </q-card-section>

                <q-card-actions align="right" class="q-pa-md">
                    <q-btn flat label="取消" v-close-popup />
                    <q-btn
                        color="primary"
                        label="确定"
                        :loading="saving"
                        unelevated
                        @click="handleSave"
                    />
                </q-card-actions>
            </q-card>
        </q-dialog>

        <!-- Login Logs Dialog -->
        <q-dialog v-model="logDialogVisible" persistent>
            <q-card style="min-width: 800px; max-width: 95vw; width: 100%">
                <q-card-section class="row items-center q-pb-none">
                    <div class="text-h6">
                        登录日志{{ logFilterTitle }}
                    </div>
                    <q-space />
                    <q-btn icon="close" flat round dense v-close-popup />
                </q-card-section>

                <q-card-section class="row q-gutter-md">
                    <q-input
                        v-model="logFilter.username"
                        dense
                        outlined
                        label="用户名"
                        style="min-width: 160px"
                        :disable="!!logFilter.userId"
                        @keyup.enter="reloadLoginLogs"
                    />
                    <q-input
                        v-model="logFilter.ip"
                        dense
                        outlined
                        label="IP地址"
                        style="min-width: 160px"
                        @keyup.enter="reloadLoginLogs"
                    />
                    <q-select
                        v-model="logFilter.status"
                        dense
                        outlined
                        label="状态"
                        :options="logStatusOptions"
                        emit-value
                        map-options
                        clearable
                        style="min-width: 140px"
                    />
                    <q-btn
                        color="primary"
                        label="查询"
                        unelevated
                        @click="reloadLoginLogs"
                    />
                    <q-btn
                        flat
                        label="重置"
                        @click="resetLogFilter"
                    />
                </q-card-section>

                <q-card-section>
                    <q-table
                        :rows="loginLogs"
                        :columns="logColumns"
                        :loading="logLoading"
                        :pagination="logPagination"
                        :rows-per-page-options="[10, 20, 50]"
                        row-key="id"
                        flat
                        bordered
                        separator="cell"
                        @request="onLogPageChange"
                    >
                        <template #body-cell-status="props">
                            <q-td :props="props">
                                <q-chip
                                    dense
                                    :color="
                                        props.row.status === 'success'
                                            ? 'positive'
                                            : 'negative'
                                    "
                                    text-color="white"
                                    class="q-ma-none"
                                >
                                    {{
                                        props.row.status === "success"
                                            ? "成功"
                                            : "失败"
                                    }}
                                </q-chip>
                            </q-td>
                        </template>
                        <template #body-cell-loginAt="props">
                            <q-td :props="props">{{
                                formatTime(props.row.loginAt)
                            }}</q-td>
                        </template>
                        <template #body-cell-ua="props">
                            <q-td :props="props" class="ua-cell">
                                <q-tooltip>{{ props.row.ua }}</q-tooltip>
                                {{ truncateUA(props.row.ua) }}
                            </q-td>
                        </template>
                    </q-table>
                </q-card-section>

                <q-card-actions align="right" class="q-pa-md">
                    <q-btn flat label="关闭" v-close-popup />
                </q-card-actions>
            </q-card>
        </q-dialog>
    </q-page>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { useQuasar } from "quasar";
import { userApi, teamApi, loginLogApi } from "../api";
import { useAuthStore } from "../stores/auth";

const $q = useQuasar();
const authStore = useAuthStore();

const users = ref([]);
const teams = ref([]);
const loading = ref(false);
const dialogVisible = ref(false);
const saving = ref(false);
const editingUser = ref(null);
const formRef = ref(null);
const showPwd = ref(false);
const filter = ref("");
const roleFilter = ref(null);
const teamFilter = ref(null);

const form = ref({
    username: "",
    realname: "",
    password: "",
    role: "operator",
    teamId: null,
});

const columns = [
    {
        name: "id",
        label: "ID",
        field: "id",
        align: "center",
        style: "width: 70px",
    },
    { name: "username", label: "用户名", field: "username", align: "left" },
    { name: "realname", label: "姓名", field: "realname", align: "left" },
    {
        name: "team",
        label: "团队",
        field: (row) => row.team?.name || "-",
        align: "left",
        style: "width: 140px",
    },
    {
        name: "role",
        label: "角色",
        field: "role",
        align: "center",
        style: "width: 110px",
    },
    {
        name: "createdAt",
        label: "创建时间",
        field: "createdAt",
        align: "left",
        style: "width: 180px",
    },
    {
        name: "actions",
        label: "操作",
        field: "actions",
        align: "center",
        style: "width: 240px",
    },
];

const roleOptions = [
    { label: "管理员", value: "admin" },
    { label: "统计员", value: "statistician" },
    { label: "操作员", value: "operator" },
    { label: "质检员", value: "inspector" },
];

const roleFilterOptions = roleOptions;

const teamFilterOptions = computed(() =>
    teams.value.map((t) => ({ label: t.name, value: t.id })),
);

const filteredUsers = computed(() => {
    let result = users.value;

    if (roleFilter.value) {
        result = result.filter((user) => user.role === roleFilter.value);
    }

    if (teamFilter.value) {
        result = result.filter((user) => user.teamId === teamFilter.value);
    }

    if (filter.value) {
        const searchTerm = filter.value.toLowerCase();
        result = result.filter(
            (user) =>
                user.username.toLowerCase().includes(searchTerm) ||
                user.realname.toLowerCase().includes(searchTerm) ||
                roleLabel(user.role).toLowerCase().includes(searchTerm) ||
                (user.team?.name || "").toLowerCase().includes(searchTerm),
        );
    }

    return result;
});

onMounted(() => {
    loadUsers();
    loadTeams();
});

async function loadTeams() {
    try {
        const res = await teamApi.list();
        teams.value = Array.isArray(res.data) ? res.data : [];
    } catch {
        // 不影响主功能
    }
}

async function loadUsers() {
    loading.value = true;
    try {
        const res = await userApi.list();
        users.value = Array.isArray(res.data) ? res.data : [];
    } catch (error) {
        console.error("Error loading users:", error);
        $q.notify({ type: "negative", message: "获取用户列表失败" });
        users.value = [];
    } finally {
        loading.value = false;
    }
}

function openCreateDialog() {
    editingUser.value = null;
    resetForm();
    showPwd.value = false;
    dialogVisible.value = true;
}

function openEditDialog(user) {
    editingUser.value = user;
    form.value = {
        username: user.username,
        realname: user.realname,
        password: "",
        role: user.role,
        teamId: user.teamId || null,
    };
    showPwd.value = false;
    dialogVisible.value = true;
}

function resetForm() {
    form.value = { username: "", realname: "", password: "", role: "operator", teamId: null };
    formRef.value?.resetValidation();
}

async function handleSave() {
    const valid = await formRef.value?.validate();
    if (!valid) return;
    saving.value = true;
    try {
        if (editingUser.value) {
            const payload = {
                realname: form.value.realname,
                role: form.value.role,
                teamId: form.value.teamId ?? 0,
            };
            if (form.value.password) payload.password = form.value.password;
            await userApi.update(editingUser.value.id, payload);
            $q.notify({ type: "positive", message: "更新成功" });
        } else {
            await userApi.create(form.value);
            $q.notify({ type: "positive", message: "创建成功" });
        }
        dialogVisible.value = false;
        loadUsers();
    } catch (err) {
        $q.notify({ type: "negative", message: err?.message || "操作失败" });
    } finally {
        saving.value = false;
    }
}

function handleDelete(user) {
    $q.dialog({
        title: "警告",
        message: `确定要删除用户「${user.realname || user.username}」吗？`,
        cancel: { label: "取消", flat: true },
        ok: { label: "删除", color: "negative" },
        persistent: true,
    }).onOk(async () => {
        try {
            await userApi.delete(user.id);
            $q.notify({ type: "positive", message: "删除成功" });
            loadUsers();
        } catch (err) {
            $q.notify({
                type: "negative",
                message: err?.message || "删除失败",
            });
        }
    });
}

function roleLabel(role) {
    const map = {
        admin: "管理员",
        statistician: "统计员",
        operator: "操作员",
        inspector: "质检员",
    };
    return map[role] || role;
}

function roleChipClass(role) {
    const map = {
        admin: "role-admin",
        statistician: "role-statistician",
        operator: "role-operator",
        inspector: "role-inspector",
    };
    return map[role] || "";
}

function formatTime(t) {
    if (!t) return "-";
    return new Date(t).toLocaleString("zh-CN");
}

// ============ 登录日志 ============
const logDialogVisible = ref(false);
const logLoading = ref(false);
const loginLogs = ref([]);
const logFilter = ref({
    userId: null,
    username: "",
    ip: "",
    status: null,
});
const logPagination = ref({
    page: 1,
    rowsPerPage: 20,
    rowsNumber: 0,
});

const logStatusOptions = [
    { label: "成功", value: "success" },
    { label: "失败", value: "failed" },
];

const logColumns = [
    { name: "id", label: "ID", field: "id", align: "center", style: "width: 70px" },
    { name: "username", label: "用户名", field: "username", align: "left" },
    { name: "ip", label: "IP地址", field: "ip", align: "left", style: "width: 160px" },
    { name: "status", label: "状态", field: "status", align: "center", style: "width: 90px" },
    { name: "message", label: "说明", field: "message", align: "left", style: "width: 150px" },
    { name: "loginAt", label: "登录时间", field: "loginAt", align: "left", style: "width: 180px" },
    { name: "ua", label: "User-Agent", field: "ua", align: "left" },
];

const logFilterTitle = computed(() => {
    if (!logFilter.value.userId) return "（全部）";
    const u = users.value.find((x) => x.id === logFilter.value.userId);
    return ` - ${u?.realname || u?.username || ""}`;
});

function openUserLoginLogs(user) {
    // 仅按 userId 过滤，避免与 username 模糊匹配产生冲突
    // （用户改名后旧日志的 username 与新 username 不一致会被 LIKE 过滤掉）
    logFilter.value = {
        userId: user.id,
        username: "",
        ip: "",
        status: null,
    };
    logPagination.value = { page: 1, rowsPerPage: 20, rowsNumber: 0 };
    logDialogVisible.value = true;
    loadLoginLogs();
}

function openAllLoginLogs() {
    logFilter.value = {
        userId: null,
        username: "",
        ip: "",
        status: null,
    };
    logPagination.value = { page: 1, rowsPerPage: 20, rowsNumber: 0 };
    logDialogVisible.value = true;
    loadLoginLogs();
}

function reloadLoginLogs() {
    logPagination.value.page = 1;
    loadLoginLogs();
}

function resetLogFilter() {
    const keepUserId = logFilter.value.userId;
    logFilter.value = {
        userId: keepUserId,
        username: keepUserId ? logFilter.value.username : "",
        ip: "",
        status: null,
    };
    logPagination.value.page = 1;
    loadLoginLogs();
}

async function loadLoginLogs() {
    logLoading.value = true;
    try {
        const params = {
            page: logPagination.value.page,
            pageSize: logPagination.value.rowsPerPage,
        };
        if (logFilter.value.userId) params.userId = logFilter.value.userId;
        if (logFilter.value.username) params.username = logFilter.value.username;
        if (logFilter.value.ip) params.ip = logFilter.value.ip;
        if (logFilter.value.status) params.status = logFilter.value.status;
        const res = await loginLogApi.list(params);
        const data = res.data || {};
        loginLogs.value = data.list || [];
        logPagination.value.rowsNumber = data.total || 0;
    } catch (err) {
        $q.notify({
            type: "negative",
            message: err?.message || "获取登录日志失败",
        });
        loginLogs.value = [];
        logPagination.value.rowsNumber = 0;
    } finally {
        logLoading.value = false;
    }
}

function onLogPageChange(props) {
    logPagination.value.page = props.pagination.page;
    logPagination.value.rowsPerPage = props.pagination.rowsPerPage;
    loadLoginLogs();
}

function truncateUA(ua) {
    if (!ua) return "-";
    return ua.length > 60 ? ua.slice(0, 60) + "…" : ua;
}
</script>

<style scoped>
.ua-cell {
    max-width: 280px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}
</style>

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
    </q-page>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { useQuasar } from "quasar";
import { userApi, teamApi } from "../api";
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
        style: "width: 140px",
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
</script>

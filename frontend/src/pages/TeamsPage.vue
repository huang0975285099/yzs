<template>
    <q-page class="q-pa-md">
        <q-card flat bordered>
            <q-card-section>
                <div class="page-header">
                    <q-input
                        v-model="filter"
                        dense
                        outlined
                        label="搜索团队..."
                        style="min-width: 200px"
                    >
                        <template #prepend>
                            <q-icon name="search" />
                        </template>
                    </q-input>
                    <q-btn
                        color="primary"
                        icon="add"
                        label="新增团队"
                        unelevated
                        @click="openCreateDialog"
                    />
                </div>

                <q-table
                    :rows="filteredTeams"
                    :columns="columns"
                    :loading="loading"
                    :pagination="{ rowsPerPage: 20 }"
                    :rows-per-page-options="[10, 20, 50]"
                    row-key="id"
                    flat
                    bordered
                    separator="cell"
                >
                    <template #body-cell-createdAt="props">
                        <q-td :props="props">{{
                            formatTime(props.row.createdAt)
                        }}</q-td>
                    </template>
                    <template #body-cell-userCount="props">
                        <q-td :props="props">
                            {{ teamUserCount(props.row.id) }}
                        </q-td>
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
                                @click="handleDelete(props.row)"
                            />
                        </q-td>
                    </template>
                </q-table>
            </q-card-section>
        </q-card>

        <!-- Create / Edit Dialog -->
        <q-dialog v-model="dialogVisible" @hide="resetForm" persistent>
            <q-card style="min-width: 340px; max-width: 420px; width: 100%">
                <q-card-section class="row items-center q-pb-none">
                    <div class="text-h6">
                        {{ editingTeam ? "编辑团队" : "新增团队" }}
                    </div>
                    <q-space />
                    <q-btn icon="close" flat round dense v-close-popup />
                </q-card-section>

                <q-card-section>
                    <q-form ref="formRef" class="q-gutter-md">
                        <q-input
                            v-model="form.name"
                            outlined
                            label="团队名称"
                            :rules="[
                                (v) => !!v || '请输入团队名称',
                                (v) => v.length <= 100 || '最多100个字符',
                            ]"
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
import { teamApi, userApi } from "../api";

const $q = useQuasar();

const teams = ref([]);
const users = ref([]);
const loading = ref(false);
const dialogVisible = ref(false);
const saving = ref(false);
const editingTeam = ref(null);
const formRef = ref(null);
const filter = ref("");

const form = ref({ name: "" });

const columns = [
    { name: "id", label: "ID", field: "id", align: "center", style: "width: 70px" },
    { name: "name", label: "团队名称", field: "name", align: "left" },
    { name: "userCount", label: "成员数", field: "id", align: "center", style: "width: 100px" },
    { name: "createdAt", label: "创建时间", field: "createdAt", align: "left", style: "width: 180px" },
    { name: "actions", label: "操作", field: "actions", align: "center", style: "width: 140px" },
];

const filteredTeams = computed(() => {
    if (!filter.value) return teams.value;
    const s = filter.value.toLowerCase();
    return teams.value.filter((t) => t.name.toLowerCase().includes(s));
});

function teamUserCount(teamId) {
    return users.value.filter((u) => u.teamId === teamId).length;
}

onMounted(async () => {
    await Promise.all([loadTeams(), loadUsers()]);
});

async function loadTeams() {
    loading.value = true;
    try {
        const res = await teamApi.list();
        teams.value = Array.isArray(res.data) ? res.data : [];
    } catch {
        $q.notify({ type: "negative", message: "获取团队列表失败" });
    } finally {
        loading.value = false;
    }
}

async function loadUsers() {
    try {
        const res = await userApi.list();
        users.value = Array.isArray(res.data) ? res.data : [];
    } catch {
        // 不影响主功能
    }
}

function openCreateDialog() {
    editingTeam.value = null;
    resetForm();
    dialogVisible.value = true;
}

function openEditDialog(team) {
    editingTeam.value = team;
    form.value = { name: team.name };
    dialogVisible.value = true;
}

function resetForm() {
    form.value = { name: "" };
    formRef.value?.resetValidation();
}

async function handleSave() {
    const valid = await formRef.value?.validate();
    if (!valid) return;
    saving.value = true;
    try {
        if (editingTeam.value) {
            await teamApi.update(editingTeam.value.id, form.value);
            $q.notify({ type: "positive", message: "更新成功" });
        } else {
            await teamApi.create(form.value);
            $q.notify({ type: "positive", message: "创建成功" });
        }
        dialogVisible.value = false;
        await loadTeams();
    } catch (err) {
        $q.notify({ type: "negative", message: err?.message || "操作失败" });
    } finally {
        saving.value = false;
    }
}

function handleDelete(team) {
    const count = teamUserCount(team.id);
    const hint = count > 0 ? `，该团队下有 ${count} 名成员，删除后成员团队将被清空` : "";
    $q.dialog({
        title: "警告",
        message: `确定要删除团队「${team.name}」吗${hint}？`,
        cancel: { label: "取消", flat: true },
        ok: { label: "删除", color: "negative" },
        persistent: true,
    }).onOk(async () => {
        try {
            await teamApi.delete(team.id);
            $q.notify({ type: "positive", message: "删除成功" });
            await Promise.all([loadTeams(), loadUsers()]);
        } catch (err) {
            $q.notify({ type: "negative", message: err?.message || "删除失败" });
        }
    });
}

function formatTime(t) {
    if (!t) return "-";
    return new Date(t).toLocaleString("zh-CN");
}
</script>

<style scoped>
.page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
    gap: 12px;
    flex-wrap: wrap;
}
</style>

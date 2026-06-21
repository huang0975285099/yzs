<template>
    <div class="login-page flex flex-center column">
        <!-- Browser warning dialog -->
        <q-dialog v-model="showBrowserWarning" persistent>
            <q-card style="width: min(420px, calc(100vw - 32px))">
                <q-card-section class="text-center q-pt-lg">
                    <q-icon name="warning" color="warning" size="48px" />
                    <div class="text-h6 q-mt-sm">浏览器兼容性提示</div>
                </q-card-section>
                <q-card-section class="text-center">
                    <p>
                        系统需要使用
                        <strong>Chrome 浏览器</strong> 的视频播放能力，请使用
                        Chrome 浏览器访问。
                    </p>
                    <p class="text-caption text-grey q-mt-sm">
                        当前浏览器：{{ currentBrowser }}
                    </p>
                </q-card-section>
                <q-card-actions align="center" class="q-pb-lg">
                    <q-btn
                        color="primary"
                        label="下载 Chrome 浏览器"
                        @click="downloadChrome"
                    />
                    <q-btn
                        flat
                        label="忽略继续使用"
                        @click="showBrowserWarning = false"
                    />
                </q-card-actions>
            </q-card>
        </q-dialog>

        <!-- Login box -->
        <div class="login-box">
            <div class="login-header text-center q-mb-lg">
                <div class="logo">☁</div>
                <div class="text-h5 text-weight-bold" style="color: #1a237e">
                    AI云值守 · 管理系统
                </div>
                <div
                    class="text-caption text-grey q-mt-xs"
                    style="letter-spacing: 1px"
                >
                    AI Cloud Monitoring · Management System
                </div>
            </div>

            <q-form @submit.prevent="handleLogin" class="login-form">
                <q-input
                    v-model="form.username"
                    outlined
                    label="请输入用户名"
                    :rules="[(v) => !!v || '请输入用户名']"
                    class="login-input"
                >
                    <template #prepend>
                        <q-icon name="person" color="grey-6" />
                    </template>
                </q-input>

                <q-input
                    v-model="form.password"
                    outlined
                    :type="showPwd ? 'text' : 'password'"
                    label="请输入密码"
                    :rules="[(v) => !!v || '请输入密码']"
                    class="login-input q-mt-md"
                >
                    <template #prepend>
                        <q-icon name="lock" color="grey-6" />
                    </template>
                    <template #append>
                        <q-icon
                            :name="showPwd ? 'visibility_off' : 'visibility'"
                            class="cursor-pointer"
                            color="grey-6"
                            @click="showPwd = !showPwd"
                        />
                    </template>
                </q-input>

                <q-btn
                    type="submit"
                    color="primary"
                    label="登 录"
                    :loading="loading"
                    class="login-btn q-mt-lg"
                    size="lg"
                    style="letter-spacing: 4px"
                    unelevated
                />
            </q-form>
        </div>

        <div
            class="login-footer text-caption"
            style="color: rgba(255, 255, 255, 0.6)"
        >
            © 2026 云值守管理系统
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useQuasar } from "quasar";
import { useAuthStore } from "../stores/auth";

const $q = useQuasar();
const router = useRouter();
const authStore = useAuthStore();

const loading = ref(false);
const showPwd = ref(false);
const showBrowserWarning = ref(false);
const currentBrowser = ref("");

const form = ref({ username: "", password: "" });

onMounted(() => {
    checkBrowser();
});

function checkBrowser() {
    const ua = navigator.userAgent;
    const isChrome = /Chrome\//.test(ua) && !/Chromium|Edg|OPR|Brave/.test(ua);
    if (!isChrome) {
        currentBrowser.value = detectBrowserName(ua);
        showBrowserWarning.value = true;
    }
}

function detectBrowserName(ua) {
    if (/Edg\//.test(ua)) return "Microsoft Edge";
    if (/OPR\//.test(ua)) return "Opera";
    if (/Firefox\//.test(ua)) return "Firefox";
    if (/Safari\//.test(ua) && !/Chrome/.test(ua)) return "Safari";
    if (/Chromium\//.test(ua)) return "Chromium";
    return "未知浏览器";
}

function downloadChrome() {
    window.open("https://www.google.cn/chrome/", "_blank");
}

async function handleLogin() {
    loading.value = true;
    try {
        await authStore.login(form.value.username, form.value.password);
        $q.notify({ type: "positive", message: "登录成功" });
        const user = JSON.parse(localStorage.getItem("user") || "{}");
        if (user.role === "operator" || user.role === "inspector") {
            router.push("/app/my-handled");
        } else if (user.role === "statistician") {
            router.push("/app/stats");
        } else {
            router.push("/app/dashboard");
        }
    } catch (err) {
        $q.notify({
            type: "negative",
            message: err?.message || "用户名或密码错误",
        });
    } finally {
        loading.value = false;
    }
}
</script>

<style scoped>
.login-page {
    min-height: 100vh;
    background: linear-gradient(135deg, #1a237e 0%, #0d47a1 50%, #01579b 100%);
    position: relative;
}

.login-box {
    background: #fff;
    border-radius: 12px;
    padding: 40px;
    width: min(420px, calc(100vw - 32px));
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

@media (max-width: 480px) {
    .login-box {
        padding: 28px 20px;
        border-radius: 10px;
    }
}

.logo {
    font-size: 56px;
    line-height: 1;
    margin-bottom: 10px;
}

.login-footer {
    position: absolute;
    bottom: 24px;
}

/* 登录表单样式 - 确保按钮和输入框对齐 */
.login-form {
    width: 100%;
}

.login-input {
    width: 100%;
}

.login-btn {
    width: 100%;
    display: block;
}

/* 确保移动端对齐 */
@media (max-width: 480px) {
    .login-form :deep(.q-field__control) {
        padding-left: 12px;
        padding-right: 12px;
    }

    .login-btn {
        border-radius: 4px;
    }
}
</style>

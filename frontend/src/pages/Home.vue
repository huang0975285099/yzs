<template>
    <div class="home-page" :class="{ 'theme-light': isLight }">
        <!-- 顶部导航 -->
        <header class="home-header">
            <div class="header-content">
                <div class="logo">
                    <div class="logo-icon-wrap">
                        <span class="logo-icon">◈</span>
                        <span class="logo-pulse"></span>
                    </div>
                    <div class="logo-text-wrap">
                        <span class="logo-text">AI云值守</span>
                        <span class="logo-sub">AI Cloud Monitoring</span>
                    </div>
                </div>
                <div class="header-actions">
                    <button class="theme-toggle" @click="toggleTheme" :title="isLight ? '切换深色' : '切换白色'">
                        <span class="theme-toggle-track">
                            <span class="theme-toggle-thumb"></span>
                        </span>
                        <span class="theme-toggle-label">{{ isLight ? '◑ 深色' : '☀ 白色' }}</span>
                    </button>
                    <q-btn class="login-btn" flat @click="goLogin">
                        <span class="btn-text">进入系统</span>
                        <span class="btn-arrow">→</span>
                    </q-btn>
                </div>
            </div>
            <div class="header-line"></div>
        </header>

        <!-- 主要内容 -->
        <main class="home-main">
            <!-- Hero 区域 -->
            <section class="hero-section">
                <div class="grid-bg"></div>
                <div class="particles">
                    <span v-for="i in 20" :key="i" class="particle" :style="getParticleStyle(i)"></span>
                </div>
                <div class="hero-content">
                    <div class="hero-badge">
                        <span class="badge-dot"></span>
                        <span>AI 智能驱动 · 全天候无人值守</span>
                    </div>
                    <h1 class="hero-title">
                        <span class="title-line">AI 云值守</span>
                        <span class="title-line title-gradient">智能运营平台</span>
                    </h1>
                    <p class="hero-subtitle">覆盖多行业场景，AI驱动异常识别与处理，让每一次运营决策更高效、更精准</p>
                    <div class="hero-stats">
                        <div class="stat-item">
                            <span class="stat-num">9</span>
                            <span class="stat-label">业务场景</span>
                        </div>
                        <div class="stat-divider"></div>
                        <div class="stat-item">
                            <span class="stat-num">24/7</span>
                            <span class="stat-label">全天候运行</span>
                        </div>
                        <div class="stat-divider"></div>
                        <div class="stat-item">
                            <span class="stat-num">AI</span>
                            <span class="stat-label">智能决策</span>
                        </div>
                    </div>
                    <div class="hero-actions">
                        <q-btn class="primary-btn" flat @click="goLogin">立即进入系统</q-btn>
                    </div>
                </div>
                <div class="hero-visual">
                    <div class="orbit-ring ring-1"></div>
                    <div class="orbit-ring ring-2"></div>
                    <div class="orbit-ring ring-3"></div>
                    <div class="center-core">
                        <span class="core-icon">◈</span>
                    </div>
                </div>
            </section>

            <!-- 业务模块 -->
            <section class="modules-section">
                <div class="section-header">
                    <div class="section-tag">BUSINESS MODULES</div>
                    <h2 class="section-title">业务场景覆盖</h2>
                    <p class="section-desc">多行业场景深度融合，AI赋能每一个运营环节</p>
                </div>
                <div class="modules-grid">
                    <!-- 货柜结算值守 - 运行中 -->
                    <div class="module-card active" @click="goLogin">
                        <div class="card-glow"></div>
                        <div class="card-header">
                            <div class="module-icon-wrap">
                                <q-icon name="shopping_cart" size="32px" />
                            </div>
                            <div class="module-status running">
                                <span class="status-dot"></span>
                                <span>运行中</span>
                            </div>
                        </div>
                        <div class="card-body">
                            <h3 class="module-name">货柜结算值守</h3>
                            <p class="module-desc">智能识别货柜异常订单，AI辅助视频审核，自动化结算处理，大幅提升人工效率</p>
                        </div>
                        <div class="module-tags">
                            <span class="tag">异常识别</span>
                            <span class="tag">视频审核</span>
                            <span class="tag">自动结算</span>
                        </div>
                        <div class="card-footer">
                            <span class="enter-text">进入系统 →</span>
                        </div>
                        <div class="card-border-animate"></div>
                    </div>

                    <!-- 商家值守 -->
                    <div class="module-card coming">
                        <div class="card-header">
                            <div class="module-icon-wrap icon-purple">
                                <q-icon name="storefront" size="32px" />
                            </div>
                            <div class="module-status coming-soon">
                                <span class="status-dot dot-purple"></span>
                                <span>即将上线</span>
                            </div>
                        </div>
                        <div class="card-body">
                            <h3 class="module-name">商家零售值守</h3>
                            <p class="module-desc">AI监控门店偷损与异常行为，顾客求助实时响应，云端坐席远程语音引导，一名坐席可同时看管数十家门店</p>
                        </div>
                        <div class="module-tags">
                            <span class="tag tag-purple">行为识别</span>
                            <span class="tag tag-purple">远程引导</span>
                            <span class="tag tag-purple">夜间值守</span>
                        </div>
                        <div class="card-footer">
                            <span class="coming-text">开发中...</span>
                        </div>
                    </div>

                    <!-- 医疗与养老值守 -->
                    <div class="module-card coming">
                        <div class="card-header">
                            <div class="module-icon-wrap icon-blue">
                                <q-icon name="person" size="32px" />
                            </div>
                            <div class="module-status coming-soon">
                                <span class="status-dot dot-blue"></span>
                                <span>即将上线</span>
                            </div>
                        </div>
                        <div class="card-body">
                            <h3 class="module-name">医疗与养老值守</h3>
                            <p class="module-desc">AI实时监测老人健康状态，异常行为智能预警，远程视频关怀，守护夕阳安康</p>
                        </div>
                        <div class="module-tags">
                            <span class="tag tag-blue">健康监测</span>
                            <span class="tag tag-blue">异常预警</span>
                            <span class="tag tag-blue">远程关怀</span>
                        </div>
                        <div class="card-footer">
                            <span class="coming-text">开发中...</span>
                        </div>
                    </div>

                    <!-- 车辆违章值守 -->
                    <div class="module-card coming">
                        <div class="card-header">
                            <div class="module-icon-wrap icon-orange">
                                <q-icon name="local_shipping" size="32px" />
                            </div>
                            <div class="module-status coming-soon">
                                <span class="status-dot dot-orange"></span>
                                <span>即将上线</span>
                            </div>
                        </div>
                        <div class="card-body">
                            <h3 class="module-name">车辆违章值守</h3>
                            <p class="module-desc">智能摄像头联动，AI识别违规行为，自动推送违章记录，精准高效取证执法</p>
                        </div>
                        <div class="module-tags">
                            <span class="tag tag-orange">违规识别</span>
                            <span class="tag tag-orange">自动取证</span>
                            <span class="tag tag-orange">实时推送</span>
                        </div>
                        <div class="card-footer">
                            <span class="coming-text">开发中...</span>
                        </div>
                    </div>

                    <!-- 新能源值守 -->
                    <div class="module-card coming">
                        <div class="card-header">
                            <div class="module-icon-wrap icon-green">
                                <q-icon name="bolt" size="32px" />
                            </div>
                            <div class="module-status coming-soon">
                                <span class="status-dot dot-green"></span>
                                <span>即将上线</span>
                            </div>
                        </div>
                        <div class="card-body">
                            <h3 class="module-name">新能源值守</h3>
                            <p class="module-desc">新能源设备全天候智能监控，故障预测与告警，远程运维调度，保障稳定供能</p>
                        </div>
                        <div class="module-tags">
                            <span class="tag tag-green">设备监控</span>
                            <span class="tag tag-green">故障预测</span>
                            <span class="tag tag-green">远程运维</span>
                        </div>
                        <div class="card-footer">
                            <span class="coming-text">开发中...</span>
                        </div>
                    </div>

                    <!-- 物业安防值守 -->
                    <div class="module-card coming">
                        <div class="card-header">
                            <div class="module-icon-wrap icon-rose">
                                <q-icon name="gps_fixed" size="32px" />
                            </div>
                            <div class="module-status coming-soon">
                                <span class="status-dot dot-rose"></span>
                                <span>即将上线</span>
                            </div>
                        </div>
                        <div class="card-body">
                            <h3 class="module-name">物业安防值守</h3>
                            <p class="module-desc">AI周界防范取代保安巡逻，翻墙闯禁区自动弹窗提醒并远程喊话；消控室24小时在岗监测，实时分析烟雾火情</p>
                        </div>
                        <div class="module-tags">
                            <span class="tag tag-rose">周界防范</span>
                            <span class="tag tag-rose">消控值守</span>
                            <span class="tag tag-rose">远程喊话</span>
                        </div>
                        <div class="card-footer">
                            <span class="coming-text">开发中...</span>
                        </div>
                    </div>

                    <!-- 智慧工地巡检值守 -->
                    <div class="module-card coming">
                        <div class="card-header">
                            <div class="module-icon-wrap icon-amber">
                                <q-icon name="build" size="32px" />
                            </div>
                            <div class="module-status coming-soon">
                                <span class="status-dot dot-amber"></span>
                                <span>即将上线</span>
                            </div>
                        </div>
                        <div class="card-body">
                            <h3 class="module-name">智慧工地巡检值守</h3>
                            <p class="module-desc">识别工人是否佩戴安全帽与违规操作，发现立即云端纠偏；机房监测设备指示灯异常、漏水及非法闯入</p>
                        </div>
                        <div class="module-tags">
                            <span class="tag tag-amber">安全帽识别</span>
                            <span class="tag tag-amber">违规纠偏</span>
                            <span class="tag tag-amber">机房监控</span>
                        </div>
                        <div class="card-footer">
                            <span class="coming-text">开发中...</span>
                        </div>
                    </div>

                    <!-- 智慧停车场值守 -->
                    <div class="module-card coming">
                        <div class="card-header">
                            <div class="module-icon-wrap icon-teal">
                                <q-icon name="vpn_key" size="32px" />
                            </div>
                            <div class="module-status coming-soon">
                                <span class="status-dot dot-teal"></span>
                                <span>即将上线</span>
                            </div>
                        </div>
                        <div class="card-body">
                            <h3 class="module-name">智慧停车场值守</h3>
                            <p class="module-desc">取代岗亭保安，车主按求助键后云端客服秒级介入，远程识别车牌并一键抬杆；1名坐席可管理上百个停车场</p>
                        </div>
                        <div class="module-tags">
                            <span class="tag tag-teal">秒级介入</span>
                            <span class="tag tag-teal">远程抬杆</span>
                            <span class="tag tag-teal">多场联管</span>
                        </div>
                        <div class="card-footer">
                            <span class="coming-text">开发中...</span>
                        </div>
                    </div>

                    <!-- 智慧校园值守 -->
                    <div class="module-card coming">
                        <div class="card-header">
                            <div class="module-icon-wrap icon-indigo">
                                <q-icon name="business" size="32px" />
                            </div>
                            <div class="module-status coming-soon">
                                <span class="status-dot dot-indigo"></span>
                                <span>即将上线</span>
                            </div>
                        </div>
                        <div class="card-body">
                            <h3 class="module-name">智慧校园值守</h3>
                            <p class="module-desc">AI视觉自动统计宿舍查寝人数；实时监测走廊打架、倒地、翻越围栏等突发行为，云端值班员按告警级别决策出警</p>
                        </div>
                        <div class="module-tags">
                            <span class="tag tag-indigo">宿舍查寝</span>
                            <span class="tag tag-indigo">突发监测</span>
                            <span class="tag tag-indigo">分级告警</span>
                        </div>
                        <div class="card-footer">
                            <span class="coming-text">开发中...</span>
                        </div>
                    </div>
                </div>
            </section>

            <!-- 优势分析 -->
            <section class="compare-section">
                <div class="section-header">
                    <div class="section-tag">WHY AI CLOUD MONITORING</div>
                    <h2 class="section-title">为什么选择 AI 云值守？</h2>
                    <p class="section-desc">一名云端坐席同时管理 10-50 个场所，全天候 AI 驱动，成本降低 80%</p>
                </div>
                <div class="compare-wrap">
                    <!-- 表头 -->
                    <div class="compare-header">
                        <div class="compare-dim-label">对比维度</div>
                        <div class="compare-col-label traditional">
                            <span class="col-icon">✕</span>
                            <span>传统人工值守</span>
                        </div>
                        <div class="compare-col-label ai-guard">
                            <span class="col-icon ai-dot">◈</span>
                            <span>AI 云值守</span>
                        </div>
                    </div>
                    <!-- 成本 -->
                    <div class="compare-row">
                        <div class="compare-dim">
                            <div class="dim-icon">💰</div>
                            <span>成本</span>
                        </div>
                        <div class="compare-cell cell-bad">
                            <p>需支付多班倒工资、社保，长期人力成本极高，且难以扩张</p>
                        </div>
                        <div class="compare-cell cell-good">
                            <p>一名坐席可同时看管 <em>10–50</em> 个场所，平摊后单店成本极低</p>
                        </div>
                    </div>
                    <!-- 效率 -->
                    <div class="compare-row">
                        <div class="compare-dim">
                            <div class="dim-icon">⚡</div>
                            <span>效率</span>
                        </div>
                        <div class="compare-cell cell-bad">
                            <p>容易疲劳、漏看、脱岗，夜班尤其薄弱，异常响应慢</p>
                        </div>
                        <div class="compare-cell cell-good">
                            <p>AI <em>24/7</em> 不间断秒级响应，发现异常才由人工精准介入</p>
                        </div>
                    </div>
                    <!-- 覆盖 -->
                    <div class="compare-row">
                        <div class="compare-dim">
                            <div class="dim-icon">🌐</div>
                            <span>覆盖</span>
                        </div>
                        <div class="compare-cell cell-bad">
                            <p>物理距离受限，跨城市、跨门店管理需大量驻场人员</p>
                        </div>
                        <div class="compare-cell cell-good">
                            <p>只要有网络，即可<em>跨城市、跨区域</em>远程统一管理所有场所</p>
                        </div>
                    </div>
                    <!-- 质量 -->
                    <div class="compare-row">
                        <div class="compare-dim">
                            <div class="dim-icon">🎯</div>
                            <span>质量</span>
                        </div>
                        <div class="compare-cell cell-bad">
                            <p>主观判断差异大，缺乏数据留存，质检难以追溯</p>
                        </div>
                        <div class="compare-cell cell-good">
                            <p>AI 客观识别，全程录像留存，处理数据<em>可追溯、可审计</em></p>
                        </div>
                    </div>
                </div>
            </section>

            <!-- 核心能力 -->
            <section class="features-section">
                <div class="section-header">
                    <div class="section-tag">CORE CAPABILITIES</div>
                    <h2 class="section-title">AI 核心能力</h2>
                </div>
                <div class="features-grid">
                    <div class="feature-item">
                        <div class="feature-icon-wrap">
                            <q-icon name="videocam" size="28px" />
                        </div>
                        <div class="feature-info">
                            <h4>视频智能分析</h4>
                            <p>AI实时解析视频流，精准识别异常行为与事件</p>
                        </div>
                    </div>
                    <div class="feature-item">
                        <div class="feature-icon-wrap">
                            <q-icon name="bar_chart" size="28px" />
                        </div>
                        <div class="feature-info">
                            <h4>数据驾驶舱</h4>
                            <p>多维度数据聚合，可视化洞察运营全貌</p>
                        </div>
                    </div>
                    <div class="feature-item">
                        <div class="feature-icon-wrap">
                            <q-icon name="notifications" size="28px" />
                        </div>
                        <div class="feature-info">
                            <h4>智能预警推送</h4>
                            <p>毫秒级异常检测，多通道实时告警通知</p>
                        </div>
                    </div>
                    <div class="feature-item">
                        <div class="feature-icon-wrap">
                            <q-icon name="check_circle" size="28px" />
                        </div>
                        <div class="feature-info">
                            <h4>质检审核闭环</h4>
                            <p>完整的审核流程管理，保障处理质量达标</p>
                        </div>
                    </div>
                    <div class="feature-item">
                        <div class="feature-icon-wrap">
                            <q-icon name="lock" size="28px" />
                        </div>
                        <div class="feature-info">
                            <h4>权限安全管控</h4>
                            <p>多角色精细化权限体系，数据安全有保障</p>
                        </div>
                    </div>
                    <div class="feature-item">
                        <div class="feature-icon-wrap">
                            <q-icon name="hub" size="28px" />
                        </div>
                        <div class="feature-info">
                            <h4>多系统集成</h4>
                            <p>开放API接口，无缝对接现有业务系统</p>
                        </div>
                    </div>
                </div>
            </section>
        </main>

        <!-- 底部 -->
        <footer class="home-footer">
            <div class="footer-content">
                <div class="footer-logo">
                    <span class="logo-icon">◈</span>
                    <span>AI云值守 · 智能运营解决方案</span>
                </div>
                <div class="footer-info">
                    <p>Copyright 2009 - 2026 四川佳华互联信息科技有限公司 All Rights Reserved</p>
                    <p class="icp">蜀ICP备2024048409号</p>
                </div>
            </div>
        </footer>
    </div>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";

const router = useRouter();
const isLight = ref(false);

function toggleTheme() {
    isLight.value = !isLight.value;
}

function goLogin() {
    router.push("/login");
}

function getParticleStyle(i) {
    const angle = (i / 20) * 360;
    const radius = 30 + (i % 5) * 10;
    const size = 2 + (i % 3);
    const duration = 3 + (i % 4);
    const delay = (i % 5) * 0.5;
    return {
        width: size + "px",
        height: size + "px",
        left: (50 + radius * Math.cos((angle * Math.PI) / 180)) + "%",
        top: (50 + radius * Math.sin((angle * Math.PI) / 180)) + "%",
        animationDuration: duration + "s",
        animationDelay: delay + "s",
        opacity: 0.3 + (i % 3) * 0.2,
    };
}
</script>

<style scoped>
/* ===== 全局 ===== */
.home-page {
    min-height: 100vh;
    background: #060b18;
    color: #e2e8f0;
    font-family: -apple-system, "PingFang SC", "Microsoft YaHei", sans-serif;
    overflow-x: hidden;
}

/* ===== 顶部导航 ===== */
.home-header {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    height: 68px;
    background: rgba(6, 11, 24, 0.85);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    z-index: 100;
}

.header-line {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 1px;
    background: linear-gradient(90deg, transparent, #00d4ff44, #7c3aed44, transparent);
}

.header-content {
    max-width: 1300px;
    margin: 0 auto;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 32px;
}

.logo {
    display: flex;
    align-items: center;
    gap: 14px;
}

.logo-icon-wrap {
    position: relative;
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
}

.logo-icon {
    font-size: 28px;
    color: #00d4ff;
    filter: drop-shadow(0 0 8px #00d4ff);
    position: relative;
    z-index: 1;
}

.logo-pulse {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    border: 1px solid #00d4ff44;
    animation: pulse 2s ease-in-out infinite;
}

.logo-text-wrap {
    display: flex;
    flex-direction: column;
}

.logo-text {
    font-size: 18px;
    font-weight: 700;
    color: #fff;
    letter-spacing: 1px;
}

.logo-sub {
    font-size: 10px;
    color: #00d4ff99;
    letter-spacing: 2px;
    font-weight: 400;
}

.login-btn {
    background: linear-gradient(135deg, #00d4ff22, #7c3aed22) !important;
    border: 1px solid #00d4ff55 !important;
    color: #00d4ff !important;
    padding: 8px 24px !important;
    border-radius: 6px !important;
    font-size: 14px !important;
    cursor: pointer;
    transition: all 0.3s;
    display: flex;
    align-items: center;
    gap: 8px;
}

.login-btn:hover {
    background: linear-gradient(135deg, #00d4ff33, #7c3aed33) !important;
    border-color: #00d4ff !important;
    box-shadow: 0 0 16px #00d4ff44;
}

.btn-arrow {
    transition: transform 0.3s;
}

.login-btn:hover .btn-arrow {
    transform: translateX(4px);
}

/* ===== Hero ===== */
.hero-section {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 160px 80px 0px;
    max-width: 1300px;
    margin: 0 auto;
    gap: 60px;
    overflow: hidden;
}

.grid-bg {
    position: fixed;
    inset: 0;
    background-image:
        linear-gradient(rgba(0, 212, 255, 0.04) 1px, transparent 1px),
        linear-gradient(90deg, rgba(0, 212, 255, 0.04) 1px, transparent 1px);
    background-size: 60px 60px;
    pointer-events: none;
    z-index: 0;
}

.particles {
    position: absolute;
    inset: 0;
    pointer-events: none;
}

.particle {
    position: absolute;
    background: #00d4ff;
    border-radius: 50%;
    animation: float linear infinite;
}

.hero-content {
    flex: 1;
    position: relative;
    z-index: 1;
    max-width: 600px;
}

.hero-badge {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    background: rgba(0, 212, 255, 0.1);
    border: 1px solid rgba(0, 212, 255, 0.3);
    border-radius: 20px;
    padding: 6px 16px;
    font-size: 13px;
    color: #00d4ff;
    margin-bottom: 28px;
}

.badge-dot {
    width: 6px;
    height: 6px;
    background: #00d4ff;
    border-radius: 50%;
    animation: blink 1.5s ease-in-out infinite;
}

.hero-title {
    display: flex;
    flex-direction: column;
    font-size: 56px;
    font-weight: 800;
    line-height: 1.2;
    margin-bottom: 20px;
    letter-spacing: -1px;
}

.title-line {
    color: #fff;
}

.title-gradient {
    background: linear-gradient(135deg, #00d4ff, #7c3aed);
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
}

.hero-subtitle {
    font-size: 17px;
    color: #94a3b8;
    line-height: 1.7;
    margin-bottom: 36px;
}

.hero-stats {
    display: flex;
    align-items: center;
    gap: 32px;
    margin-bottom: 40px;
}

.stat-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
}

.stat-num {
    font-size: 28px;
    font-weight: 800;
    color: #00d4ff;
    line-height: 1;
}

.stat-label {
    font-size: 12px;
    color: #64748b;
    letter-spacing: 1px;
}

.stat-divider {
    width: 1px;
    height: 40px;
    background: linear-gradient(180deg, transparent, #334155, transparent);
}

.primary-btn {
    background: linear-gradient(135deg, #00d4ff, #7c3aed) !important;
    border: none !important;
    color: #fff !important;
    padding: 14px 36px !important;
    border-radius: 8px !important;
    font-size: 16px !important;
    font-weight: 600 !important;
    cursor: pointer;
    transition: all 0.3s;
    box-shadow: 0 4px 24px rgba(0, 212, 255, 0.3);
    height: auto !important;
}

.primary-btn:hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 32px rgba(0, 212, 255, 0.5);
}

/* ===== Hero Visual ===== */
.hero-visual {
    flex-shrink: 0;
    width: 380px;
    height: 380px;
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
}

.orbit-ring {
    position: absolute;
    border-radius: 50%;
    border: 1px solid transparent;
}

.ring-1 {
    width: 200px;
    height: 200px;
    border-color: rgba(0, 212, 255, 0.3);
    animation: spin 8s linear infinite;
    box-shadow: 0 0 20px rgba(0, 212, 255, 0.1);
}

.ring-2 {
    width: 300px;
    height: 300px;
    border-color: rgba(124, 58, 237, 0.25);
    animation: spin 15s linear infinite reverse;
}

.ring-3 {
    width: 380px;
    height: 380px;
    border-color: rgba(0, 212, 255, 0.1);
    animation: spin 25s linear infinite;
    border-style: dashed;
}

.center-core {
    width: 80px;
    height: 80px;
    background: linear-gradient(135deg, rgba(0, 212, 255, 0.15), rgba(124, 58, 237, 0.15));
    border: 1px solid rgba(0, 212, 255, 0.5);
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 0 40px rgba(0, 212, 255, 0.2), inset 0 0 20px rgba(0, 212, 255, 0.05);
}

.core-icon {
    font-size: 36px;
    color: #00d4ff;
    filter: drop-shadow(0 0 12px #00d4ff);
}

/* ===== 业务模块 ===== */
.modules-section {
    padding: 100px 80px;
    max-width: 1300px;
    margin: 0 auto;
    position: relative;
    z-index: 1;
}

.features-section {
    padding: 80px 80px 100px;
    max-width: 1300px;
    margin: 0 auto;
    position: relative;
    z-index: 1;
}

.section-header {
    text-align: center;
    margin-bottom: 60px;
}

.section-tag {
    font-size: 11px;
    letter-spacing: 4px;
    color: #00d4ff;
    font-weight: 600;
    margin-bottom: 12px;
}

.section-title {
    font-size: 36px;
    font-weight: 800;
    color: #fff;
    margin-bottom: 14px;
}

.section-desc {
    font-size: 15px;
    color: #64748b;
}

.modules-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 20px;
}

/* 模块卡片 */
.module-card {
    position: relative;
    background: rgba(13, 20, 40, 0.8);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 16px;
    padding: 28px;
    transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
    overflow: hidden;
}

.module-card.active {
    border-color: rgba(0, 212, 255, 0.35);
    background: rgba(0, 212, 255, 0.04);
    cursor: pointer;
}

.module-card.active:hover {
    transform: translateY(-6px);
    border-color: rgba(0, 212, 255, 0.7);
    box-shadow: 0 20px 60px rgba(0, 212, 255, 0.15);
}

.module-card.coming {
    opacity: 0.9;
}

.card-glow {
    position: absolute;
    top: -40px;
    left: -40px;
    width: 180px;
    height: 180px;
    background: radial-gradient(circle, rgba(0, 212, 255, 0.12), transparent 70%);
    pointer-events: none;
}

.card-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    margin-bottom: 20px;
}

.module-icon-wrap {
    width: 60px;
    height: 60px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: linear-gradient(135deg, rgba(0, 212, 255, 0.2), rgba(124, 58, 237, 0.2));
    border: 1px solid rgba(0, 212, 255, 0.3);
    border-radius: 14px;
    color: #00d4ff;
}

.module-icon-wrap.icon-blue {
    background: linear-gradient(135deg, rgba(59, 130, 246, 0.2), rgba(99, 102, 241, 0.2));
    border-color: rgba(59, 130, 246, 0.3);
    color: #60a5fa;
}

.module-icon-wrap.icon-orange {
    background: linear-gradient(135deg, rgba(251, 146, 60, 0.2), rgba(239, 68, 68, 0.2));
    border-color: rgba(251, 146, 60, 0.3);
    color: #fb923c;
}

.module-icon-wrap.icon-green {
    background: linear-gradient(135deg, rgba(16, 185, 129, 0.2), rgba(5, 150, 105, 0.2));
    border-color: rgba(16, 185, 129, 0.3);
    color: #34d399;
}

.module-icon-wrap.icon-purple {
    background: linear-gradient(135deg, rgba(167, 139, 250, 0.2), rgba(124, 58, 237, 0.2));
    border-color: rgba(167, 139, 250, 0.3);
    color: #a78bfa;
}

.module-icon-wrap.icon-rose {
    background: linear-gradient(135deg, rgba(251, 113, 133, 0.2), rgba(225, 29, 72, 0.2));
    border-color: rgba(251, 113, 133, 0.3);
    color: #fb7185;
}

.module-icon-wrap.icon-amber {
    background: linear-gradient(135deg, rgba(251, 191, 36, 0.2), rgba(217, 119, 6, 0.2));
    border-color: rgba(251, 191, 36, 0.3);
    color: #fbbf24;
}

.module-icon-wrap.icon-teal {
    background: linear-gradient(135deg, rgba(45, 212, 191, 0.2), rgba(13, 148, 136, 0.2));
    border-color: rgba(45, 212, 191, 0.3);
    color: #2dd4bf;
}

.module-icon-wrap.icon-indigo {
    background: linear-gradient(135deg, rgba(129, 140, 248, 0.2), rgba(79, 70, 229, 0.2));
    border-color: rgba(129, 140, 248, 0.3);
    color: #818cf8;
}

.module-status {
    display: flex;
    align-items: center;
    gap: 5px;
    font-size: 12px;
    padding: 4px 10px;
    border-radius: 20px;
}

.module-status.running {
    background: rgba(16, 185, 129, 0.15);
    border: 1px solid rgba(16, 185, 129, 0.4);
    color: #34d399;
}

.module-status.coming-soon {
    background: rgba(100, 116, 139, 0.15);
    border: 1px solid rgba(100, 116, 139, 0.3);
    color: #94a3b8;
}

.status-dot {
    width: 6px;
    height: 6px;
    background: #34d399;
    border-radius: 50%;
    animation: blink 1.5s ease-in-out infinite;
}

.status-dot.dot-blue { background: #60a5fa; animation: none; }
.status-dot.dot-orange { background: #fb923c; animation: none; }
.status-dot.dot-green { background: #34d399; animation: none; }
.status-dot.dot-purple { background: #a78bfa; animation: none; }
.status-dot.dot-rose { background: #fb7185; animation: none; }
.status-dot.dot-amber { background: #fbbf24; animation: none; }
.status-dot.dot-teal { background: #2dd4bf; animation: none; }
.status-dot.dot-indigo { background: #818cf8; animation: none; }

.module-name {
    font-size: 20px;
    font-weight: 700;
    color: #fff;
    margin-bottom: 10px;
}

.module-desc {
    font-size: 13px;
    color: #64748b;
    line-height: 1.7;
    margin-bottom: 18px;
}

.module-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin-bottom: 20px;
}

.tag {
    font-size: 11px;
    padding: 3px 10px;
    border-radius: 4px;
    background: rgba(0, 212, 255, 0.1);
    border: 1px solid rgba(0, 212, 255, 0.25);
    color: #00d4ff;
}

.tag-blue {
    background: rgba(59, 130, 246, 0.1);
    border-color: rgba(59, 130, 246, 0.25);
    color: #60a5fa;
}

.tag-orange {
    background: rgba(251, 146, 60, 0.1);
    border-color: rgba(251, 146, 60, 0.25);
    color: #fb923c;
}

.tag-green {
    background: rgba(16, 185, 129, 0.1);
    border-color: rgba(16, 185, 129, 0.25);
    color: #34d399;
}

.tag-purple {
    background: rgba(167, 139, 250, 0.1);
    border-color: rgba(167, 139, 250, 0.25);
    color: #a78bfa;
}

.tag-rose {
    background: rgba(251, 113, 133, 0.1);
    border-color: rgba(251, 113, 133, 0.25);
    color: #fb7185;
}

.tag-amber {
    background: rgba(251, 191, 36, 0.1);
    border-color: rgba(251, 191, 36, 0.25);
    color: #fbbf24;
}

.tag-teal {
    background: rgba(45, 212, 191, 0.1);
    border-color: rgba(45, 212, 191, 0.25);
    color: #2dd4bf;
}

.tag-indigo {
    background: rgba(129, 140, 248, 0.1);
    border-color: rgba(129, 140, 248, 0.25);
    color: #818cf8;
}

.card-footer {
    border-top: 1px solid rgba(255, 255, 255, 0.06);
    padding-top: 16px;
}

.enter-text {
    font-size: 14px;
    color: #00d4ff;
    font-weight: 600;
    letter-spacing: 0.5px;
    transition: letter-spacing 0.3s;
}

.module-card.active:hover .enter-text {
    letter-spacing: 2px;
}

.coming-text {
    font-size: 13px;
    color: #475569;
}

/* 卡片边框动画 */
.card-border-animate {
    position: absolute;
    inset: 0;
    border-radius: 16px;
    pointer-events: none;
    background: linear-gradient(90deg, transparent, rgba(0, 212, 255, 0.4), transparent);
    background-size: 200% 100%;
    opacity: 0;
    transition: opacity 0.3s;
}

.module-card.active:hover .card-border-animate {
    opacity: 1;
    animation: shimmer 2s infinite;
}

/* ===== 核心能力 ===== */
.features-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 20px;
}

.feature-item {
    display: flex;
    align-items: flex-start;
    gap: 16px;
    background: rgba(13, 20, 40, 0.6);
    border: 1px solid rgba(255, 255, 255, 0.06);
    border-radius: 12px;
    padding: 24px;
    transition: all 0.3s;
}

.feature-item:hover {
    border-color: rgba(0, 212, 255, 0.25);
    background: rgba(0, 212, 255, 0.04);
    transform: translateY(-3px);
}

.feature-icon-wrap {
    width: 52px;
    height: 52px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(0, 212, 255, 0.1);
    border: 1px solid rgba(0, 212, 255, 0.2);
    border-radius: 10px;
    color: #00d4ff;
}

.feature-info h4 {
    font-size: 16px;
    font-weight: 600;
    color: #fff;
    margin-bottom: 6px;
}

.feature-info p {
    font-size: 13px;
    color: #64748b;
    line-height: 1.6;
}

/* ===== 优势分析 ===== */
.compare-section {
    padding: 80px 80px 100px;
    position: relative;
    z-index: 1;
}

.compare-wrap {
    max-width: 1100px;
    margin: 0 auto;
    border: 1px solid rgba(0, 212, 255, 0.15);
    border-radius: 20px;
    overflow: hidden;
    background: rgba(13, 20, 40, 0.7);
}

.compare-header {
    display: grid;
    grid-template-columns: 140px 1fr 1fr;
    background: rgba(0, 212, 255, 0.06);
    border-bottom: 1px solid rgba(0, 212, 255, 0.15);
}

.compare-dim-label {
    padding: 20px 24px;
    font-size: 12px;
    letter-spacing: 2px;
    color: #475569;
    text-transform: uppercase;
    border-right: 1px solid rgba(255, 255, 255, 0.06);
    text-align: center;
}

.compare-col-label {
    padding: 18px 28px;
    font-size: 15px;
    font-weight: 700;
    display: flex;
    align-items: center;
    gap: 10px;
}

.compare-col-label.traditional {
    color: #64748b;
    border-right: 1px solid rgba(255, 255, 255, 0.06);
}

.compare-col-label.ai-guard {
    color: #00d4ff;
}

.col-icon {
    font-size: 16px;
    color: #475569;
}

.col-icon.ai-dot {
    color: #00d4ff;
    filter: drop-shadow(0 0 6px #00d4ff);
}

.compare-row {
    display: grid;
    grid-template-columns: 140px 1fr 1fr;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    transition: background 0.3s;
}

.compare-row:last-child {
    border-bottom: none;
}

.compare-row:hover {
    background: rgba(0, 212, 255, 0.03);
}

.compare-dim {
    padding: 24px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    border-right: 1px solid rgba(255, 255, 255, 0.06);
    background: rgba(0, 0, 0, 0.15);
}

.dim-icon {
    font-size: 22px;
    line-height: 1;
}

.compare-dim span {
    font-size: 14px;
    font-weight: 700;
    color: #fff;
    letter-spacing: 1px;
}

.compare-cell {
    padding: 24px 28px;
    display: flex;
    align-items: center;
}

.compare-cell p {
    font-size: 14px;
    line-height: 1.7;
    margin: 0;
}

.cell-bad {
    border-right: 1px solid rgba(255, 255, 255, 0.06);
}

.cell-bad p {
    color: #94a3b8;
    position: relative;
    padding-left: 14px;
}

.cell-bad p::before {
    content: '';
    position: absolute;
    left: 0;
    top: 10px;
    width: 4px;
    height: 4px;
    border-radius: 50%;
    background: #475569;
}

.cell-good p {
    color: #fff;
    position: relative;
    padding-left: 14px;
}

.cell-good p::before {
    content: '';
    position: absolute;
    left: 0;
    top: 10px;
    width: 4px;
    height: 4px;
    border-radius: 50%;
    background: #00d4ff;
    box-shadow: 0 0 6px #00d4ff;
}

.cell-good p em {
    font-style: normal;
    color: #00d4ff;
    font-weight: 700;
}

/* ===== 底部 ===== */
.home-footer {
    border-top: 1px solid rgba(255, 255, 255, 0.06);
    padding: 32px 80px;
    background: rgba(6, 11, 24, 0.9);
    position: relative;
    z-index: 1;
}

.footer-content {
    max-width: 1300px;
    margin: 0 auto;
    display: flex;
    align-items: center;
    justify-content: space-between;
}

.footer-logo {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 15px;
    color: #475569;
}

.footer-logo .logo-icon {
    color: #00d4ff66;
    font-size: 20px;
    filter: none;
}

.footer-info {
    text-align: right;
    font-size: 13px;
    color: #334155;
    line-height: 1.8;
}

.icp {
    color: #1e293b;
}

/* ===== 主题切换按钮 ===== */
.header-actions {
    display: flex;
    align-items: center;
    gap: 12px;
}

.theme-toggle {
    display: flex;
    align-items: center;
    gap: 8px;
    background: rgba(255,255,255,0.07);
    border: 1px solid rgba(255,255,255,0.15);
    border-radius: 20px;
    padding: 6px 14px 6px 10px;
    cursor: pointer;
    color: #94a3b8;
    font-size: 13px;
    transition: all 0.3s;
    white-space: nowrap;
}

.theme-toggle:hover {
    border-color: rgba(0,212,255,0.5);
    color: #00d4ff;
}

.theme-toggle-track {
    display: none;
}

/* ===== 白色主题（遗忘战境风格）===== */
.theme-light {
    background: #f0f4f8;
    color: #1e293b;
}

.theme-light .grid-bg {
    background-image:
        linear-gradient(rgba(14,165,233,0.06) 1px, transparent 1px),
        linear-gradient(90deg, rgba(14,165,233,0.06) 1px, transparent 1px);
}

.theme-light .home-header {
    background: rgba(255,255,255,0.82);
    box-shadow: 0 1px 24px rgba(0,0,0,0.06);
}

.theme-light .header-line {
    background: linear-gradient(90deg, transparent, #38bdf844, transparent);
}

.theme-light .logo-icon {
    color: #0ea5e9;
    filter: drop-shadow(0 0 6px #38bdf8aa);
}

.theme-light .logo-pulse {
    border-color: #38bdf866;
}

.theme-light .logo-text {
    color: #0f172a;
}

.theme-light .logo-sub {
    color: #0ea5e9aa;
}

.theme-light .theme-toggle {
    background: rgba(14,165,233,0.08);
    border-color: rgba(14,165,233,0.25);
    color: #0ea5e9;
}

.theme-light .theme-toggle:hover {
    background: rgba(14,165,233,0.15);
    border-color: #0ea5e9;
}

.theme-light .login-btn {
    background: linear-gradient(135deg, rgba(14,165,233,0.12), rgba(56,189,248,0.08)) !important;
    border-color: rgba(14,165,233,0.4) !important;
    color: #0ea5e9 !important;
}

.theme-light .login-btn:hover {
    background: linear-gradient(135deg, rgba(14,165,233,0.2), rgba(56,189,248,0.15)) !important;
    border-color: #0ea5e9 !important;
    box-shadow: 0 0 16px rgba(14,165,233,0.25);
}

.theme-light .hero-badge {
    background: rgba(14,165,233,0.1);
    border-color: rgba(14,165,233,0.3);
    color: #0ea5e9;
}

.theme-light .badge-dot {
    background: #0ea5e9;
}

.theme-light .title-line {
    color: #0f172a;
}

.theme-light .title-gradient {
    background: linear-gradient(135deg, #0ea5e9, #6366f1);
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
}

.theme-light .hero-subtitle {
    color: #475569;
}

.theme-light .stat-num {
    color: #0ea5e9;
}

.theme-light .stat-label {
    color: #94a3b8;
}

.theme-light .stat-divider {
    background: linear-gradient(180deg, transparent, #cbd5e1, transparent);
}

.theme-light .primary-btn {
    background: linear-gradient(135deg, #0ea5e9, #6366f1) !important;
    box-shadow: 0 4px 24px rgba(14,165,233,0.35);
}

.theme-light .primary-btn:hover {
    box-shadow: 0 8px 32px rgba(14,165,233,0.5);
}

.theme-light .particle {
    background: #0ea5e9;
}

.theme-light .ring-1 {
    border-color: rgba(14,165,233,0.3);
    box-shadow: 0 0 20px rgba(14,165,233,0.08);
}

.theme-light .ring-2 {
    border-color: rgba(99,102,241,0.2);
}

.theme-light .ring-3 {
    border-color: rgba(14,165,233,0.12);
}

.theme-light .center-core {
    background: linear-gradient(135deg, rgba(14,165,233,0.15), rgba(99,102,241,0.15));
    border-color: rgba(14,165,233,0.5);
    box-shadow: 0 0 40px rgba(14,165,233,0.15), inset 0 0 20px rgba(14,165,233,0.05);
}

.theme-light .core-icon {
    color: #0ea5e9;
    filter: drop-shadow(0 0 10px #38bdf8);
}

.theme-light .section-tag {
    color: #0ea5e9;
}

.theme-light .section-title {
    color: #0f172a;
}

.theme-light .section-desc {
    color: #64748b;
}

.theme-light .module-card {
    background: rgba(255,255,255,0.9);
    border-color: rgba(0,0,0,0.07);
    box-shadow: 0 2px 16px rgba(0,0,0,0.06);
}

.theme-light .module-card.active {
    border-color: rgba(14,165,233,0.35);
    background: rgba(255,255,255,0.95);
}

.theme-light .module-card.active:hover {
    border-color: rgba(14,165,233,0.7);
    box-shadow: 0 20px 60px rgba(14,165,233,0.12);
}

.theme-light .card-glow {
    background: radial-gradient(circle, rgba(14,165,233,0.1), transparent 70%);
}

.theme-light .module-icon-wrap {
    background: linear-gradient(135deg, rgba(14,165,233,0.15), rgba(99,102,241,0.1));
    border-color: rgba(14,165,233,0.25);
    color: #0ea5e9;
}

.theme-light .module-name {
    color: #0f172a;
}

.theme-light .module-desc {
    color: #64748b;
}

.theme-light .module-status.running {
    background: rgba(16,185,129,0.1);
    border-color: rgba(16,185,129,0.35);
}

.theme-light .module-status.coming-soon {
    background: rgba(148,163,184,0.12);
    border-color: rgba(148,163,184,0.3);
}

.theme-light .tag {
    background: rgba(14,165,233,0.08);
    border-color: rgba(14,165,233,0.2);
    color: #0ea5e9;
}

.theme-light .card-footer {
    border-top-color: rgba(0,0,0,0.06);
}

.theme-light .enter-text {
    color: #0ea5e9;
}

.theme-light .coming-text {
    color: #94a3b8;
}

.theme-light .card-border-animate {
    background: linear-gradient(90deg, transparent, rgba(14,165,233,0.35), transparent);
}

.theme-light .feature-item {
    background: rgba(255,255,255,0.9);
    border-color: rgba(0,0,0,0.07);
    box-shadow: 0 2px 12px rgba(0,0,0,0.05);
}

.theme-light .feature-item:hover {
    border-color: rgba(14,165,233,0.3);
    background: #fff;
    box-shadow: 0 8px 28px rgba(14,165,233,0.1);
}

.theme-light .feature-icon-wrap {
    background: rgba(14,165,233,0.1);
    border-color: rgba(14,165,233,0.2);
    color: #0ea5e9;
}

.theme-light .feature-info h4 {
    color: #0f172a;
}

.theme-light .feature-info p {
    color: #64748b;
}

.theme-light .compare-wrap {
    background: rgba(255,255,255,0.9);
    border-color: rgba(14,165,233,0.15);
    box-shadow: 0 4px 32px rgba(0,0,0,0.06);
}

.theme-light .compare-header {
    background: rgba(14,165,233,0.05);
    border-bottom-color: rgba(14,165,233,0.12);
}

.theme-light .compare-dim-label {
    color: #94a3b8;
    border-right-color: rgba(0,0,0,0.06);
}

.theme-light .compare-col-label.traditional {
    color: #94a3b8;
    border-right-color: rgba(0,0,0,0.06);
}

.theme-light .compare-col-label.ai-guard {
    color: #0ea5e9;
}

.theme-light .col-icon { color: #94a3b8; }

.theme-light .col-icon.ai-dot {
    color: #0ea5e9;
    filter: drop-shadow(0 0 5px #38bdf8);
}

.theme-light .compare-row {
    border-bottom-color: rgba(0,0,0,0.05);
}

.theme-light .compare-row:hover {
    background: rgba(14,165,233,0.03);
}

.theme-light .compare-dim {
    background: rgba(14,165,233,0.03);
    border-right-color: rgba(0,0,0,0.06);
}

.theme-light .compare-dim span {
    color: #1e293b;
}

.theme-light .cell-bad {
    border-right-color: rgba(0,0,0,0.05);
}

.theme-light .cell-bad p {
    color: #94a3b8;
}

.theme-light .cell-bad p::before {
    background: #cbd5e1;
}

.theme-light .cell-good p {
    color: #1e293b;
}

.theme-light .cell-good p::before {
    background: #0ea5e9;
    box-shadow: 0 0 6px #38bdf8;
}

.theme-light .cell-good p em {
    color: #0ea5e9;
}

.theme-light .home-footer {
    background: rgba(255,255,255,0.85);
    border-top-color: rgba(0,0,0,0.07);
}

.theme-light .footer-logo {
    color: #64748b;
}

.theme-light .footer-logo .logo-icon {
    color: #0ea5e966;
}

.theme-light .footer-info {
    color: #94a3b8;
}

.theme-light .icp {
    color: #cbd5e1;
}

/* ===== 动画 ===== */
@keyframes pulse {
    0%, 100% { transform: scale(1); opacity: 0.5; }
    50% { transform: scale(1.5); opacity: 0; }
}

@keyframes blink {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.3; }
}

@keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
}

@keyframes float {
    0% { transform: translateY(0) translateX(0); opacity: 0; }
    10% { opacity: 1; }
    90% { opacity: 1; }
    100% { transform: translateY(-100px) translateX(20px); opacity: 0; }
}

@keyframes shimmer {
    0% { background-position: -200% 0; }
    100% { background-position: 200% 0; }
}

/* ===== 响应式 ===== */
@media (max-width: 1200px) {
    .hero-section {
        padding: 100px 40px 60px;
    }
    .hero-visual {
        width: 280px;
        height: 280px;
    }
    .ring-1 { width: 160px; height: 160px; }
    .ring-2 { width: 220px; height: 220px; }
    .ring-3 { width: 280px; height: 280px; }
    .modules-section, .features-section, .compare-section { padding-left: 40px; padding-right: 40px; }
    .modules-grid { grid-template-columns: repeat(2, 1fr); }
}

@media (max-width: 900px) {
    .hero-section { padding-top: 120px; text-align: center; }
    .hero-visual { display: none; }
    .hero-title { font-size: 40px; }
    .hero-stats { justify-content: center; }
    .modules-section, .features-section, .compare-section { padding-left: 20px; padding-right: 20px; }
    .compare-header, .compare-row { grid-template-columns: 80px 1fr 1fr; }
    .compare-dim { padding: 16px 12px; }
    .compare-cell { padding: 16px 16px; }
    .home-footer { padding: 24px 20px; }
    .footer-content { flex-direction: column; gap: 16px; text-align: center; }
    .footer-info { text-align: center; }
    .features-grid { grid-template-columns: 1fr; }
}

@media (max-width: 600px) {
    .modules-grid { grid-template-columns: 1fr; }
    .compare-header { display: none; }
    .compare-row { grid-template-columns: 1fr; border-bottom: 1px solid rgba(255,255,255,0.06); }
    .compare-dim { flex-direction: row; justify-content: flex-start; padding: 14px 16px; border-right: none; border-bottom: 1px solid rgba(255,255,255,0.04); background: rgba(0,212,255,0.04); }
    .compare-dim span { font-size: 13px; }
    .cell-bad { border-right: none; border-bottom: 1px solid rgba(255,255,255,0.04); }
    .cell-bad p::before { content: '✕  '; position: static; background: none; box-shadow: none; width: auto; height: auto; border-radius: 0; color: #475569; }
    .cell-good p::before { content: '✓  '; position: static; background: none; box-shadow: none; width: auto; height: auto; border-radius: 0; color: #00d4ff; }
    .hero-title { font-size: 32px; }
    .header-content { padding: 0 16px; }
    .logo-sub { display: none; }
}
</style>

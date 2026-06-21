/**
 * Playwright E2E 测试配置
 * 文档: https://playwright.dev/docs/test-configuration
 */
const { defineConfig, devices } = require('@playwright/test')

module.exports = defineConfig({
  // 测试目录
  testDir: './tests/e2e',

  // 完全并行执行测试
  fullyParallel: true,

  // CI 上禁止 test.only
  forbidOnly: !!process.env.CI,

  // CI 上重试失败用例
  retries: process.env.CI ? 2 : 0,

  // 并发 worker 数量
  workers: process.env.CI ? 1 : undefined,

  // Reporter 配置
  reporter: [
    ['list'],
    ['html', { open: 'never', outputFolder: 'playwright-report' }],
  ],

  // 全局测试配置
  use: {
    // 基础 URL (Quasar dev 默认端口)
    baseURL: 'http://localhost:9000',

    // 收集失败用例的 trace
    trace: 'on-first-retry',

    // 截图策略
    screenshot: 'only-on-failure',

    // 视频录制（CI 环境）
    video: process.env.CI ? 'retain-on-failure' : 'off',

    // 操作超时
    actionTimeout: 10000,

    // 导航超时
    navigationTimeout: 30000,

    // 浏览器上下文配置
    contextOptions: {
      // 忽略 HTTPS 错误（本地开发环境）
      ignoreHTTPSErrors: true,
    },
  },

  // 项目配置 - 多平台支持（使用 Chromium 模拟各设备视口）
  projects: [
    // PC 端 - Desktop Chrome
    {
      name: 'desktop-chrome',
      use: { ...devices['Desktop Chrome'] },
    },
    // 平板横屏 - iPad Pro 横屏 (使用 Chromium)
    {
      name: 'tablet-landscape',
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 1024, height: 768 },
      },
    },
    // 平板竖屏 - iPad Pro 竖屏 (使用 Chromium)
    {
      name: 'tablet-portrait',
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 768, height: 1024 },
      },
    },
    // 手机 - iPhone 14 (使用 Chromium)
    {
      name: 'mobile-iphone',
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 390, height: 844 },
      },
    },
    // 手机 - Android (使用 Chromium)
    {
      name: 'mobile-android',
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 393, height: 851 },
      },
    },
  ],

  // Web Server - 自动启动开发服务器
  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:9000',
    reuseExistingServer: !process.env.CI,
    timeout: 120000,
    stdout: 'ignore',
    stderr: 'pipe',
  },
})
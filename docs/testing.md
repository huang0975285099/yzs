# 云值守系统 - Playwright 自动化测试文档

> 状态：规划中 | 最后更新：2026-04-14

---

## 1. 概述

本文档描述云值守系统前端的 Playwright 自动化测试方案。测试范围覆盖 Web SPA 应用，使用 Chromium 浏览器引擎进行端到端（E2E）测试。

### 1.1 测试目标

| 目标 | 说明 |
|------|------|
| 功能验证 | 验证核心业务流程（登录、数据查询、表单提交等） |
| 回归测试 | 代码变更后自动验证已有功能不受影响 |
| CI/CD 集成 | 为持续集成流水线提供自动化测试支持 |
| 视觉回归 | 可选：页面 UI 变化检测（截图比对） |

### 1.2 技术选型

| 技术 | 版本 | 说明 |
|------|------|------|
| Playwright | ^1.40 | 微软开源 E2E 测试框架 |
| Chromium | 内置 | 唯一目标浏览器 |
| Node.js | 20.x | 运行环境 |

### 1.3 为什么选择 Playwright

| 特性 | Playwright | Cypress | Selenium |
|------|------------|---------|----------|
| 自动等待 | ✅ 内置 | ✅ 内置 | ❌ 需手动 |
| 多标签页/iframe | ✅ 原生支持 | ❌ 有限 | ✅ 支持 |
| 网络拦截 | ✅ 强大 | ✅ 支持 | ✅ 支持 |
| 文件上传/下载 | ✅ 原生 | ⚠️ 有限 | ✅ 支持 |
| 执行速度 | ⚡ 快 | ⚡ 快 | 🐢 较慢 |
| 调试体验 | ✅ Trace Viewer | ✅ Time Travel | ⚠️ 一般 |
| 学习曲线 | 📈 中等 | 📉 简单 | 📈 较高 |

---

## 2. 目录结构

```
frontend/
├── tests/                          ← Playwright 测试目录
│   ├── e2e/                        ← 端到端测试
│   │   ├── auth.spec.ts            ← 认证相关测试
│   │   ├── dashboard.spec.ts       ← 仪表盘测试
│   │   ├── operations.spec.ts     ← 运维管理测试
│   │   ├── orders.spec.ts          ← 订单管理测试
│   │   └── chat.spec.ts            ← 聊天功能测试
│   ├── fixtures/                   ← 测试固件
│   │   ├── auth.fixture.ts         ← 认证固件（登录状态复用）
│   │   └── test-data.ts            ← 测试数据
│   ├── mocks/                      ← API Mock 数据
│   │   ├── users.mock.ts
│   │   └── orders.mock.ts
│   └── utils/                      ← 测试工具函数
│       └── helpers.ts
├── playwright.config.ts            ← Playwright 配置
└── package.json
```

---

## 3. 配置说明

### 3.1 playwright.config.ts

```typescript
import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  // 测试目录
  testDir: './tests/e2e',
  
  // 完全并行执行测试
  fullyParallel: true,
  
  // CI 上失败时禁止 test.only
  forbidOnly: !!process.env.CI,
  
  // CI 上重试失败用例
  retries: process.env.CI ? 2 : 0,
  
  // 并发 worker 数量
  workers: process.env.CI ? 1 : undefined,
  
  // Reporter 配置
  reporter: [
    ['html', { open: 'never' }],        // HTML 报告
    ['json', { outputFile: 'test-results.json' }]  // JSON 报告
  ],
  
  // 全局测试配置
  use: {
    // 基础 URL
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
  },
  
  // 项目配置 - 仅 Chromium
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  
  // Web Server - 自动启动开发服务器
  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:9000',
    reuseExistingServer: !process.env.CI,
    timeout: 120000,
  },
})
```

### 3.2 配置项详解

| 配置项 | 值 | 说明 |
|--------|-----|------|
| `testDir` | `./tests/e2e` | 测试文件目录 |
| `fullyParallel` | `true` | 测试用例并行执行 |
| `retries` | CI: 2, 本地: 0 | CI 环境重试失败用例 |
| `workers` | CI: 1, 本地: 自动 | 并发进程数 |
| `baseURL` | `http://localhost:9000` | Quasar dev 默认端口 |
| `trace` | `on-first-retry` | 首次重试时记录 trace |
| `screenshot` | `only-on-failure` | 仅失败时截图 |
| `video` | CI: `retain-on-failure` | CI 失败时保留视频 |

---

## 4. 测试用例设计

### 4.1 认证测试 (auth.spec.ts)

```typescript
import { test, expect } from '@playwright/test'

test.describe('认证模块', () => {
  test.describe('登录页面', () => {
    test('应显示登录表单', async ({ page }) => {
      await page.goto('/login')
      await expect(page.locator('input[type="text"]')).toBeVisible()
      await expect(page.locator('input[type="password"]')).toBeVisible()
      await expect(page.locator('button[type="submit"]')).toBeVisible()
    })

    test('空表单提交应显示验证错误', async ({ page }) => {
      await page.goto('/login')
      await page.click('button[type="submit"]')
      await expect(page.locator('.q-field--error')).toBeVisible()
    })

    test('错误凭据应显示错误提示', async ({ page }) => {
      await page.goto('/login')
      await page.fill('input[type="text"]', 'wronguser')
      await page.fill('input[type="password"]', 'wrongpass')
      await page.click('button[type="submit"]')
      await expect(page.locator('.q-notification')).toContainText('用户名或密码错误')
    })

    test('正确凭据应跳转到首页', async ({ page }) => {
      await page.goto('/login')
      await page.fill('input[type="text"]', 'admin')
      await page.fill('input[type="password"]', 'admin123')
      await page.click('button[type="submit"]')
      await expect(page).toHaveURL('/dashboard')
    })
  })

  test.describe('登录状态保持', () => {
    test('刷新页面应保持登录状态', async ({ page }) => {
      // 先登录
      await page.goto('/login')
      await page.fill('input[type="text"]', 'admin')
      await page.fill('input[type="password"]', 'admin123')
      await page.click('button[type="submit"]')
      await expect(page).toHaveURL('/dashboard')
      
      // 刷新页面
      await page.reload()
      await expect(page).toHaveURL('/dashboard')
    })

    test('Token 过期应跳转登录页', async ({ page }) => {
      // 模拟 Token 过期
      await page.goto('/dashboard')
      await page.evaluate(() => {
        localStorage.removeItem('token')
      })
      await page.reload()
      await expect(page).toHaveURL('/login')
    })
  })

  test.describe('退出登录', () => {
    test('退出后应跳转登录页', async ({ page }) => {
      // 登录
      await page.goto('/login')
      await page.fill('input[type="text"]', 'admin')
      await page.fill('input[type="password"]', 'admin123')
      await page.click('button[type="submit"]')
      await expect(page).toHaveURL('/dashboard')
      
      // 点击退出
      await page.click('[data-testid="logout-button"]')
      await expect(page).toHaveURL('/login')
    })
  })
})
```

### 4.2 仪表盘测试 (dashboard.spec.ts)

```typescript
import { test, expect } from '@playwright/test'

test.describe('仪表盘页面', () => {
  test.beforeEach(async ({ page }) => {
    // 登录前置条件
    await page.goto('/login')
    await page.fill('input[type="text"]', 'admin')
    await page.fill('input[type="password"]', 'admin123')
    await page.click('button[type="submit"]')
    await expect(page).toHaveURL('/dashboard')
  })

  test('应显示统计卡片', async ({ page }) => {
    await expect(page.locator('[data-testid="stat-cards"]')).toBeVisible()
  })

  test('应显示图表', async ({ page }) => {
    // 等待 ECharts 渲染完成
    await page.waitForSelector('[data-testid="chart-container"] canvas')
    await expect(page.locator('[data-testid="chart-container"]')).toBeVisible()
  })

  test('响应式布局 - 移动端', async ({ page }) => {
    // 设置移动端视口
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/dashboard')
    
    // 验证移动端布局
    await expect(page.locator('[data-testid="mobile-nav"]')).toBeVisible()
    await expect(page.locator('[data-testid="desktop-sidebar"]')).not.toBeVisible()
  })
})
```

### 4.3 运维管理测试 (operations.spec.ts)

```typescript
import { test, expect } from '@playwright/test'

test.describe('运维管理页面', () => {
  test.beforeEach(async ({ page }) => {
    // 登录
    await page.goto('/login')
    await page.fill('input[type="text"]', 'admin')
    await page.fill('input[type="password"]', 'admin123')
    await page.click('button[type="submit"]')
    await expect(page).toHaveURL('/dashboard')
    
    // 导航到运维管理
    await page.click('[data-testid="nav-operations"]')
    await expect(page).toHaveURL('/operations')
  })

  test('应显示视频监控区域', async ({ page }) => {
    await expect(page.locator('[data-testid="video-container"]')).toBeVisible()
  })

  test('应显示数据表格', async ({ page }) => {
    await expect(page.locator('.q-table')).toBeVisible()
  })

  test('表格应支持排序', async ({ page }) => {
    // 点击表头排序
    const header = page.locator('.q-table th').first()
    await header.click()
    await expect(page.locator('.q-table')).toBeVisible()
  })

  test('表格应支持分页', async ({ page }) => {
    const pagination = page.locator('.q-table__bottom')
    await expect(pagination).toBeVisible()
  })

  test('移动端 - 视频应在表格上方', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/operations')
    
    const video = page.locator('[data-testid="video-container"]')
    const table = page.locator('.q-table')
    
    // 验证视频在表格上方
    const videoBox = await video.boundingBox()
    const tableBox = await table.boundingBox()
    expect(videoBox?.y).toBeLessThan(tableBox?.y || 0)
  })
})
```

---

## 5. 测试固件 (Fixtures)

### 5.1 认证固件

```typescript
// tests/fixtures/auth.fixture.ts
import { test as base, Page } from '@playwright/test'

// 扩展 Page 对象
type AuthFixtures = {
  authenticatedPage: Page
}

export const test = base.extend<AuthFixtures>({
  authenticatedPage: async ({ page }, use) => {
    // 登录
    await page.goto('/login')
    await page.fill('input[type="text"]', 'admin')
    await page.fill('input[type="password"]', 'admin123')
    await page.click('button[type="submit"]')
    await page.waitForURL('/dashboard')
    
    // 保存存储状态
    await page.context().storageState({ path: 'tests/.auth/user.json' })
    
    await use(page)
  },
})

export { expect } from '@playwright/test'
```

### 5.2 使用固件

```typescript
// tests/e2e/profile.spec.ts
import { test, expect } from '../fixtures/auth.fixture'

test.describe('用户资料页面', () => {
  test('应显示用户信息', async ({ authenticatedPage }) => {
    await authenticatedPage.goto('/profile')
    await expect(authenticatedPage.locator('[data-testid="user-info"]')).toBeVisible()
  })
})
```

---

## 6. API Mock

### 6.1 拦截网络请求

```typescript
import { test, expect } from '@playwright/test'

test.describe('订单管理 - Mock API', () => {
  test('应显示空状态', async ({ page }) => {
    // Mock 空数据
    await page.route('**/api/orders*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: [], total: 0 })
      })
    })
    
    await page.goto('/orders')
    await expect(page.locator('.q-table__bottom--nodata')).toBeVisible()
  })

  test('应处理 API 错误', async ({ page }) => {
    // Mock 500 错误
    await page.route('**/api/orders*', async (route) => {
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ message: '服务器错误' })
      })
    })
    
    await page.goto('/orders')
    await expect(page.locator('.q-notification')).toContainText('错误')
  })

  test('应显示模拟订单数据', async ({ page }) => {
    // Mock 订单数据
    await page.route('**/api/orders*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: [
            { id: 1, orderNo: 'ORD001', status: 'pending' },
            { id: 2, orderNo: 'ORD002', status: 'completed' },
          ],
          total: 2
        })
      })
    })
    
    await page.goto('/orders')
    await expect(page.locator('.q-table tbody tr')).toHaveCount(2)
  })
})
```

### 6.2 Mock 数据管理

```typescript
// tests/mocks/orders.mock.ts
export const mockOrders = {
  empty: { data: [], total: 0 },
  
  single: {
    data: [{ id: 1, orderNo: 'ORD001', status: 'pending' }],
    total: 1
  },
  
  list: {
    data: [
      { id: 1, orderNo: 'ORD001', status: 'pending' },
      { id: 2, orderNo: 'ORD002', status: 'completed' },
      { id: 3, orderNo: 'ORD003', status: 'cancelled' },
    ],
    total: 3
  }
}
```

---

## 7. 工具函数

### 7.1 测试辅助函数

```typescript
// tests/utils/helpers.ts
import { Page } from '@playwright/test'

/**
 * 登录辅助函数
 */
export async function login(page: Page, username: string, password: string) {
  await page.goto('/login')
  await page.fill('input[type="text"]', username)
  await page.fill('input[type="password"]', password)
  await page.click('button[type="submit"]')
  await page.waitForURL('/dashboard')
}

/**
 * 等待表格加载
 */
export async function waitForTable(page: Page) {
  await page.waitForSelector('.q-table tbody tr', { state: 'visible' })
}

/**
 * 获取表格行数
 */
export async function getTableRowCount(page: Page): Promise<number> {
  return await page.locator('.q-table tbody tr').count()
}

/**
 * 响应式视口预设
 */
export const viewports = {
  mobile: { width: 375, height: 667 },
  tablet: { width: 768, height: 1024 },
  desktop: { width: 1440, height: 900 },
}
```

---

## 8. 运行测试

### 8.1 命令说明

```bash
# 安装依赖
cd frontend
npm install -D @playwright/test
npx playwright install chromium

# 运行所有测试
npx playwright test

# 运行指定文件
npx playwright test tests/e2e/auth.spec.ts

# 运行指定测试用例
npx playwright test -g "登录页面"

# UI 模式（交互式调试）
npx playwright test --ui

# 调试模式
npx playwright test --debug

# 生成测试报告
npx playwright show-report

# 仅运行失败的测试
npx playwright test --last-failed
```

### 8.2 package.json 脚本

```json
{
  "scripts": {
    "test": "playwright test",
    "test:ui": "playwright test --ui",
    "test:debug": "playwright test --debug",
    "test:report": "playwright show-report",
    "test:e2e": "playwright test tests/e2e"
  }
}
```

---

## 9. CI/CD 集成

### 9.1 GitHub Actions

```yaml
# .github/workflows/e2e.yml
name: E2E Tests

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json
      
      - name: Install dependencies
        working-directory: frontend
        run: npm ci
      
      - name: Install Playwright
        working-directory: frontend
        run: npx playwright install --with-deps chromium
      
      - name: Run E2E tests
        working-directory: frontend
        run: npx playwright test
        
      - name: Upload test results
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: playwright-report
          path: frontend/playwright-report/
          retention-days: 30
```

### 9.2 本地预提交钩子

```json
// package.json
{
  "husky": {
    "hooks": {
      "pre-commit": "lint-staged",
      "pre-push": "npm run test"
    }
  }
}
```

---

## 10. 最佳实践

### 10.1 测试编写原则

| 原则 | 说明 |
|------|------|
| 独立性 | 每个测试用例独立运行，不依赖其他用例 |
| 幂等性 | 多次执行结果一致 |
| 快速 | 避免不必要的等待，使用智能断言 |
| 可读性 | 使用语义化的选择器，清晰的描述 |
| 稳定性 | 避免依赖动态数据，使用 Mock 或测试数据 |

### 10.2 选择器策略

```typescript
// ✅ 推荐：使用 data-testid
await page.locator('[data-testid="login-button"]').click()

// ✅ 推荐：使用语义化选择器
await page.getByRole('button', { name: '登录' }).click()
await page.getByLabel('用户名').fill('admin')

// ⚠️ 谨慎：使用文本内容（可能变化）
await page.getByText('登录').click()

// ❌ 避免：使用 CSS 类名（样式可能变化）
await page.locator('.btn-primary').click()

// ❌ 避免：使用 XPath（脆弱）
await page.locator('//div[@class="form"]/button[1]').click()
```

### 10.3 等待策略

```typescript
// ✅ 推荐：自动等待
await expect(page.locator('.q-table')).toBeVisible()

// ✅ 推荐：等待网络请求
await page.waitForResponse('**/api/orders')

// ✅ 推荐：等待导航完成
await page.waitForURL('/dashboard')

// ❌ 避免：硬编码等待
await page.waitForTimeout(2000)  // 不要这样做
```

### 10.4 data-testid 命名规范

| 元素类型 | 命名格式 | 示例 |
|----------|----------|------|
| 页面区域 | `{page}-{section}` | `dashboard-charts` |
| 按钮 | `{action}-button` | `login-button` |
| 表单输入 | `{field}-input` | `username-input` |
| 导航项 | `nav-{page}` | `nav-orders` |
| 表格 | `{entity}-table` | `orders-table` |
| 弹窗 | `{action}-dialog` | `confirm-delete-dialog` |

---

## 11. 调试技巧

### 11.1 Trace Viewer

```bash
# 运行测试并生成 trace
npx playwright test --trace on

# 查看 trace
npx playwright show-trace trace.zip
```

Trace Viewer 提供：
- 完整的操作时间线
- 每个 action 的 DOM 快照
- 网络请求详情
- 控制台日志

### 11.2 截图与视频

```typescript
// 手动截图
await page.screenshot({ path: 'screenshot.png' })

// 元素截图
await page.locator('.q-table').screenshot({ path: 'table.png' })

// 全页面截图
await page.screenshot({ path: 'full.png', fullPage: true })
```

### 11.3 调试模式

```bash
# 启动调试模式
npx playwright test --debug

# 或在代码中设置断点
await page.pause()
```

---

## 12. 测试覆盖率

### 12.1 功能模块覆盖率目标

| 模块 | 优先级 | 目标覆盖率 | 说明 |
|------|--------|------------|------|
| 认证（登录/退出） | P0 | 100% | 核心安全功能 |
| 用户管理 | P0 | 90% | 权限相关 |
| 订单管理 | P1 | 80% | 核心业务 |
| 运维管理 | P1 | 80% | 核心业务 |
| 仪表盘 | P2 | 60% | 数据展示 |
| 聊天功能 | P1 | 80% | 即时通讯 |
| 系统设置 | P2 | 50% | 配置功能 |

### 12.2 测试场景清单

- [ ] 登录成功/失败
- [ ] Token 过期处理
- [ ] 退出登录
- [ ] 路由权限守卫
- [ ] 表格分页/排序/筛选
- [ ] 表单验证
- [ ] 文件上传/下载
- [ ] WebSocket 连接
- [ ] 响应式布局（移动端/桌面端）
- [ ] 错误边界处理

---

## 13. 常见问题

### Q1: 测试运行很慢怎么办？

**A:** 优化策略：
1. 使用 `fullyParallel: true` 并行执行
2. 合理使用 Fixtures 复用登录状态
3. 避免不必要的 `waitForTimeout`
4. Mock 外部 API 减少网络等待

### Q2: 测试偶发性失败？

**A:** 排查方向：
1. 检查是否有竞态条件
2. 增加重试机制 `retries: 2`
3. 使用更稳健的断言
4. 检查网络请求是否稳定

### Q3: 如何测试 WebSocket？

**A:** 示例：
```typescript
test('WebSocket 消息测试', async ({ page }) => {
  // 监听 WebSocket
  const wsPromise = page.waitForEvent('websocket')
  await page.goto('/chat')
  const ws = await wsPromise
  
  // 监听消息
  const msgPromise = new Promise(resolve => {
    ws.on('framereceived', frame => resolve(frame.payload))
  })
  
  // 触发消息
  await page.click('[data-testid="send-button"]')
  const msg = await msgPromise
  expect(msg).toContain('hello')
})
```

### Q4: 如何测试文件下载？

```typescript
test('文件下载测试', async ({ page }) => {
  const downloadPromise = page.waitForEvent('download')
  await page.click('[data-testid="export-button"]')
  const download = await downloadPromise
  
  expect(download.suggestedFilename()).toContain('.xlsx')
})
```

---

## 14. 参考资料

- [Playwright 官方文档](https://playwright.dev/)
- [Quasar 测试指南](https://quasar.dev/quasar-cli/testing/testing)
- [Testing Library](https://testing-library.com/)
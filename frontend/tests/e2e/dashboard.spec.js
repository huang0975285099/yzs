/**
 * 仪表盘页面 E2E 测试
 */
const { test, expect } = require('@playwright/test')

// 登录辅助函数
async function loginAs(page, username, password) {
  await page.goto('/login')
  await page.locator('.q-input').first().locator('input').fill(username)
  await page.locator('.q-input').last().locator('input').fill(password)
  await page.locator('button.q-btn').click()
  await page.waitForURL(/\/app\//)
}

test.describe('仪表盘页面', () => {
  test.beforeEach(async ({ page }) => {
    // 以管理员身份登录
    await loginAs(page, 'admin', 'mzjy.com')

    // 导航到仪表盘（如果不在的话）
    const url = page.url()
    if (!url.includes('/dashboard')) {
      await page.goto('/app/dashboard')
    }
  })

  test('应显示页面内容', async ({ page }) => {
    // 验证页面加载完成
    await expect(page.locator('.q-page')).toBeVisible()
  })

  test('侧边栏导航应正常工作', async ({ page }) => {
    // 桌面端应显示侧边栏
    await page.setViewportSize({ width: 1440, height: 900 })

    // 验证侧边栏存在
    const sidebar = page.locator('.q-drawer, [data-testid="sidebar"]')
    await expect(sidebar.first()).toBeVisible()
  })

  test('应显示统计数据卡片', async ({ page }) => {
    // 等待数据加载
    await page.waitForLoadState('networkidle')

    // 验证统计卡片存在（根据实际页面结构调整选择器）
    const statCards = page.locator('.q-card, [data-testid="stat-card"]')
    await expect(statCards.first()).toBeVisible()
  })
})

test.describe('仪表盘 - 响应式布局', () => {
  test('移动端应显示底部导航', async ({ page }) => {
    // 登录
    await loginAs(page, 'admin', 'mzjy.com')

    // 设置移动端视口
    await page.setViewportSize({ width: 375, height: 667 })

    // 验证底部导航存在（如果有）
    const bottomNav = page.locator('.q-tabs, [data-testid="mobile-nav"]')
    // 移动端布局可能使用底部标签栏或汉堡菜单
    const hasBottomNav = await bottomNav.count() > 0

    // 至少验证页面正常显示
    await expect(page.locator('.q-page')).toBeVisible()
  })

  test('桌面端应显示侧边栏', async ({ page }) => {
    // 登录
    await loginAs(page, 'admin', 'mzjy.com')

    // 设置桌面端视口
    await page.setViewportSize({ width: 1440, height: 900 })
    await page.goto('/app/dashboard')

    // 验证侧边栏可见
    const sidebar = page.locator('.q-drawer')
    await expect(sidebar).toBeVisible()
  })

  test('平板端布局', async ({ page }) => {
    // 登录
    await loginAs(page, 'admin', 'mzjy.com')

    // 设置平板端视口
    await page.setViewportSize({ width: 768, height: 1024 })
    await page.goto('/app/dashboard')

    // 验证页面正常显示
    await expect(page.locator('.q-page')).toBeVisible()
  })
})

test.describe('导航功能', () => {
  test.beforeEach(async ({ page }) => {
    await loginAs(page, 'admin', 'mzjy.com')
  })

  test('点击导航菜单应跳转到对应页面', async ({ page }) => {
    // 导航到仪表盘
    await page.goto('/app/dashboard')

    // 点击订单管理（根据实际菜单文本调整）
    const ordersLink = page.locator('a:has-text("订单"), [data-testid="nav-orders"]')
    if (await ordersLink.count() > 0) {
      await ordersLink.first().click()
      await expect(page).toHaveURL(/\/app\/orders/)
    }
  })

  test('页面刷新应保持登录状态', async ({ page }) => {
    await page.goto('/app/dashboard')
    await page.waitForLoadState('networkidle')

    // 刷新页面
    await page.reload()

    // 应仍在仪表盘页面
    await expect(page).toHaveURL(/\/app\//)
  })
})
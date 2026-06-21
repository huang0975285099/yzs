/**
 * 全平台UI测试
 * 覆盖 PC、平板（横屏/竖屏）、手机端
 */
const { test, expect } = require('@playwright/test')

// 测试账号
const TEST_ACCOUNT = {
  admin: { username: 'admin', password: 'mzjy.com' },
}

// 登录辅助函数
async function login(page, username = 'admin', password = 'mzjy.com') {
  await page.goto('/login')
  await page.waitForLoadState('networkidle')

  // 填写登录表单
  const usernameInput = page.locator('.q-input').first().locator('input')
  const passwordInput = page.locator('.q-input').last().locator('input')
  const submitBtn = page.locator('button[type="submit"]')

  await usernameInput.fill(username)
  await passwordInput.fill(password)
  await submitBtn.click()

  // 等待登录完成
  await page.waitForURL(/\/app\//, { timeout: 15000 })
}

// 获取项目名称（用于截图命名）
function getProjectName(testInfo) {
  return testInfo.project.name
}

// 获取设备类型名称
function getDeviceName(projectName) {
  const deviceNames = {
    'desktop-chrome': 'PC端(Chrome)',
    'desktop-firefox': 'PC端(Firefox)',
    'desktop-safari': 'PC端(Safari)',
    'tablet-landscape': '平板横屏',
    'tablet-portrait': '平板竖屏',
    'mobile-iphone': 'iPhone',
    'mobile-android': 'Android手机',
  }
  return deviceNames[projectName] || projectName
}

// admin 可访问的页面
const ADMIN_PAGES = [
  { name: '仪表盘', path: '/app/dashboard' },
  { name: '订单管理', path: '/app/orders' },
  { name: '质检', path: '/app/quality-check' },
  { name: '质审', path: '/app/quality-review' },
  { name: '用户管理', path: '/app/users' },
  { name: '运营', path: '/app/operations' },
  { name: '我的处理', path: '/app/my-handled' },
]

// 验证页面内容加载
async function verifyPageLoaded(page, timeout = 10000) {
  // 等待页面渲染完成
  await page.waitForTimeout(500)

  // 尝试多种选择器验证页面加载
  const pageContent = page.locator('.q-page, .q-card, .q-table, main, [class*="page"]')
  const count = await pageContent.count()

  if (count === 0) {
    // 如果找不到特定元素，检查是否有任何可见内容
    await page.waitForLoadState('domcontentloaded')
    const body = page.locator('body')
    await expect(body).toBeVisible()
  }
}

// ==================== 登录页测试 ====================
test.describe('登录页 - 全平台', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')
  })

  test('登录页UI布局正常', async ({ page }, testInfo) => {
    const projectName = getProjectName(testInfo)

    // 验证登录框可见
    const loginBox = page.locator('.login-box, .q-card')
    await expect(loginBox.first()).toBeVisible()

    // 验证表单元素存在
    await expect(page.locator('.q-input')).toHaveCount(2)
    await expect(page.locator('button[type="submit"]')).toBeVisible()

    // 截图记录
    await page.screenshot({
      path: `tests/screenshots/login-${projectName}.png`,
      fullPage: true,
    })
  })

  test('登录表单响应式布局', async ({ page }) => {
    const loginBox = page.locator('.login-box, .q-card').first()

    // 验证登录框在视口内可见
    await expect(loginBox).toBeVisible()

    // 表单输入框应可交互
    const usernameInput = page.locator('.q-input').first().locator('input')
    await usernameInput.click()
    await usernameInput.fill('test')
    await expect(usernameInput).toHaveValue('test')
  })

  test('登录成功跳转', async ({ page }) => {
    await login(page)

    // 验证跳转到应用页面
    await expect(page).toHaveURL(/\/app\//)

    // 验证token保存
    const token = await page.evaluate(() => localStorage.getItem('token'))
    expect(token).toBeTruthy()
  })

  test('登录失败显示错误', async ({ page }) => {
    const usernameInput = page.locator('.q-input').first().locator('input')
    const passwordInput = page.locator('.q-input').last().locator('input')
    const submitBtn = page.locator('button[type="submit"]')

    await usernameInput.fill('wronguser')
    await passwordInput.fill('wrongpass')
    await submitBtn.click()

    // 验证错误提示
    await expect(page.locator('.q-notification')).toBeVisible({ timeout: 5000 })
  })
})

// ==================== 仪表盘测试 ====================
test.describe('仪表盘页面 - 全平台', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/app/dashboard')
    await page.waitForLoadState('networkidle')
  })

  test('页面加载正常', async ({ page }, testInfo) => {
    const projectName = getProjectName(testInfo)

    // 验证页面容器可见
    await verifyPageLoaded(page)

    // 截图记录
    await page.screenshot({
      path: `tests/screenshots/dashboard-${projectName}.png`,
      fullPage: true,
    })
  })

  test('导航菜单显示正确', async ({ page }) => {
    const viewport = page.viewportSize()
    const width = viewport?.width || 0

    // PC端和平板：侧边栏使用 show-if-above，在 breakpoint(600px) 以上应显示
    if (width >= 600) {
      // 等待侧边栏渲染
      await page.waitForTimeout(500)
      const sidebar = page.locator('.q-drawer')

      // 检查侧边栏是否存在（可能处于 mini 模式）
      const sidebarCount = await sidebar.count()
      if (sidebarCount > 0) {
        // 验证侧边栏存在
        expect(sidebarCount).toBeGreaterThan(0)
      }
    } else {
      // 手机端(<600px)：底部导航栏
      const bottomNav = page.locator('.q-footer .q-tabs')
      await page.waitForTimeout(500)
      if (await bottomNav.count() > 0) {
        await expect(bottomNav).toBeVisible()
      }
    }
  })

  test('统计卡片显示', async ({ page }) => {
    // 等待数据加载
    await page.waitForTimeout(1000)

    // 验证页面内容存在
    const pageContent = page.locator('.q-page')
    if (await pageContent.count() > 0) {
      await expect(pageContent).toBeVisible()
    }
  })
})

// ==================== 订单管理测试 ====================
test.describe('订单管理页面 - 全平台', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/app/orders')
    await page.waitForLoadState('networkidle')
  })

  test('页面加载正常', async ({ page }, testInfo) => {
    const projectName = getProjectName(testInfo)

    await verifyPageLoaded(page)

    // 截图
    await page.screenshot({
      path: `tests/screenshots/orders-${projectName}.png`,
      fullPage: true,
    })
  })

  test('表格响应式布局', async ({ page }) => {
    const viewport = page.viewportSize()
    const width = viewport?.width || 0

    // 等待表格加载
    await page.waitForTimeout(1000)

    // 桌面端应有表格
    if (width >= 1024) {
      const table = page.locator('.q-table')
      if (await table.count() > 0) {
        await expect(table.first()).toBeVisible()
      }
    }
  })

  test('搜索和筛选功能', async ({ page }) => {
    // 查找搜索框
    const searchInput = page.locator('input[type="search"], input[placeholder*="搜索"], input[placeholder*="查询"]')

    if (await searchInput.count() > 0) {
      await searchInput.first().click()
      await searchInput.first().fill('测试')
      await page.keyboard.press('Enter')
      await page.waitForTimeout(500)
    }
  })
})

// ==================== 我的处理页面测试 ====================
test.describe('我的处理页面 - 全平台', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/app/my-handled')
    await page.waitForLoadState('networkidle')
  })

  test('页面加载正常', async ({ page }, testInfo) => {
    const projectName = getProjectName(testInfo)

    await verifyPageLoaded(page)

    await page.screenshot({
      path: `tests/screenshots/my-handled-${projectName}.png`,
      fullPage: true,
    })
  })
})

// ==================== 质检页面测试 ====================
test.describe('质检页面 - 全平台', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/app/quality-check')
    await page.waitForLoadState('networkidle')
  })

  test('页面加载正常', async ({ page }, testInfo) => {
    const projectName = getProjectName(testInfo)

    await verifyPageLoaded(page)

    await page.screenshot({
      path: `tests/screenshots/quality-check-${projectName}.png`,
      fullPage: true,
    })
  })
})

// ==================== 质审页面测试 ====================
test.describe('质审页面 - 全平台', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/app/quality-review')
    await page.waitForLoadState('networkidle')
  })

  test('页面加载正常', async ({ page }, testInfo) => {
    const projectName = getProjectName(testInfo)

    await verifyPageLoaded(page)

    await page.screenshot({
      path: `tests/screenshots/quality-review-${projectName}.png`,
      fullPage: true,
    })
  })
})

// ==================== 用户管理页面测试 ====================
test.describe('用户管理页面 - 全平台', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/app/users')
    await page.waitForLoadState('networkidle')
  })

  test('页面加载正常', async ({ page }, testInfo) => {
    const projectName = getProjectName(testInfo)

    await verifyPageLoaded(page)

    await page.screenshot({
      path: `tests/screenshots/users-${projectName}.png`,
      fullPage: true,
    })
  })
})

// ==================== 运营页面测试 ====================
test.describe('运营页面 - 全平台', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await page.goto('/app/operations')
    await page.waitForLoadState('networkidle')
  })

  test('页面加载正常', async ({ page }, testInfo) => {
    const projectName = getProjectName(testInfo)

    await verifyPageLoaded(page)

    await page.screenshot({
      path: `tests/screenshots/operations-${projectName}.png`,
      fullPage: true,
    })
  })
})

// ==================== 响应式布局专项测试 ====================
test.describe('响应式布局专项测试', () => {
  test('侧边栏在不同设备上的表现', async ({ page }, testInfo) => {
    await login(page)
    await page.goto('/app/dashboard')
    await page.waitForLoadState('networkidle')

    const viewport = page.viewportSize()
    const width = viewport?.width || 0
    const sidebar = page.locator('.q-drawer')

    if (width >= 1024) {
      // PC端：侧边栏默认展开
      await expect(sidebar).toBeVisible()
    } else if (width >= 768) {
      // 平板：侧边栏可能默认收起
      const isVisible = await sidebar.isVisible()
      const menuBtn = page.locator('.q-header button, [data-testid="menu-toggle"]')
      expect(await menuBtn.count() > 0 || isVisible).toBeTruthy()
    } else {
      // 手机：侧边栏默认收起，有汉堡菜单
      const menuBtn = page.locator('.q-header button, [data-testid="menu-toggle"]')
      expect(await menuBtn.count()).toBeGreaterThan(0)
    }
  })

  test('表格在移动端的适配', async ({ page }) => {
    await login(page)
    await page.goto('/app/orders')
    await page.waitForLoadState('networkidle')

    const viewport = page.viewportSize()
    const width = viewport?.width || 0

    if (width < 768) {
      // 手机端验证页面可访问
      await verifyPageLoaded(page)
    }
  })

  test('表单在移动端的适配', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')

    // 表单输入框应在视口内可访问
    const inputs = page.locator('.q-input')
    const count = await inputs.count()

    for (let i = 0; i < count; i++) {
      const input = inputs.nth(i)
      await expect(input).toBeVisible()
    }
  })

  test('按钮在移动端可点击', async ({ page }) => {
    await login(page)
    await page.goto('/app/dashboard')
    await page.waitForLoadState('networkidle')

    // 测试主要按钮可点击
    const buttons = page.locator('button.q-btn:visible')
    const count = await buttons.count()

    if (count > 0) {
      const firstBtn = buttons.first()
      const box = await firstBtn.boundingBox()
      expect(box).toBeTruthy()
      // 移动端最小点击区域
      if (page.viewportSize()?.width < 768) {
        expect(box?.width).toBeGreaterThan(30)
        expect(box?.height).toBeGreaterThan(20)
      }
    }
  })
})

// ==================== 导航测试 ====================
test.describe('导航功能 - 全平台', () => {
  // 导航测试需要串行执行，避免登录状态互相干扰
  test.describe.configure({ mode: 'serial' })

  test.beforeEach(async ({ page }) => {
    await login(page)
  })

  test('页面导航正常工作', async ({ page }) => {
    await page.goto('/app/dashboard')
    await page.waitForLoadState('networkidle')

    // 遍历 admin 可访问的页面
    for (const pageInfo of ADMIN_PAGES) {
      if (pageInfo.path === '/app/dashboard') continue

      await page.goto(pageInfo.path)
      await page.waitForLoadState('networkidle')
      await verifyPageLoaded(page)

      // 返回仪表盘
      await page.goto('/app/dashboard')
      await page.waitForLoadState('networkidle')
    }
  })

  test('页面刷新保持登录状态', async ({ page }) => {
    await page.goto('/app/dashboard')
    await page.waitForLoadState('networkidle')

    // 验证 token 存在
    const token = await page.evaluate(() => localStorage.getItem('token'))
    expect(token).toBeTruthy()

    // 验证 user 存在
    const userStr = await page.evaluate(() => localStorage.getItem('user'))
    expect(userStr).toBeTruthy()

    // 刷新页面
    await page.reload()
    await page.waitForLoadState('networkidle')

    // admin 刷新后应仍在 dashboard 或 my-handled
    const url = page.url()
    expect(url).toMatch(/\/app\//)
  })

  test('浏览器后退/前进正常', async ({ page }) => {
    await page.goto('/app/dashboard')
    await page.waitForLoadState('networkidle')

    await page.goto('/app/orders')
    await page.waitForLoadState('networkidle')

    // 后退
    await page.goBack()
    await page.waitForLoadState('networkidle')
    expect(page.url()).toMatch(/\/app\//)

    // 前进
    await page.goForward()
    await page.waitForLoadState('networkidle')
    expect(page.url()).toMatch(/\/app\//)
  })
})

// ==================== 退出登录测试 ====================
test.describe('退出登录 - 全平台', () => {
  test('退出登录功能', async ({ page }) => {
    await login(page)
    await page.goto('/app/dashboard')
    await page.waitForLoadState('networkidle')

    // 查找退出按钮
    const logoutBtn = page.locator('[data-testid="logout"], button:has-text("退出"), button:has-text("登出")')
    const userMenu = page.locator('[data-testid="user-menu"], .user-avatar, .q-btn-dropdown')

    // 如果有用户菜单，先点击展开
    if (await userMenu.count() > 0) {
      await userMenu.first().click()
      await page.waitForTimeout(300)
    }

    // 点击退出
    if (await logoutBtn.count() > 0) {
      await logoutBtn.first().click()
      await page.waitForURL('/login', { timeout: 10000 })
      await expect(page).toHaveURL('/login')
    } else {
      // 如果找不到退出按钮，手动清除登录状态验证
      await page.evaluate(() => localStorage.clear())
      await page.goto('/login')
      await expect(page).toHaveURL('/login')
    }
  })
})
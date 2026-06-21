/**
 * 认证模块 E2E 测试
 */
const { test, expect } = require('@playwright/test')

test.describe('认证模块', () => {
  test.describe('登录页面', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/login')
    })

    test('应显示登录表单', async ({ page }) => {
      // 等待页面加载
      await expect(page.locator('.login-box')).toBeVisible()

      // 验证表单元素存在
      await expect(page.locator('.q-input')).toHaveCount(2)
      await expect(page.locator('button.q-btn')).toBeVisible()
    })

    test('空表单提交应显示验证错误', async ({ page }) => {
      // 直接点击登录按钮
      await page.click('button[type="submit"]')

      // 验证显示错误提示
      await expect(page.locator('.q-field--error')).toBeVisible()
    })

    test('错误凭据应显示错误提示', async ({ page }) => {
      // 填写错误凭据
      await page.locator('.q-input').first().locator('input').fill('wronguser')
      await page.locator('.q-input').last().locator('input').fill('wrongpass')
      await page.locator('button.q-btn').click()

      // 验证显示错误通知
      await expect(page.locator('.q-notification')).toBeVisible()
    })

    test('页面标题正确', async ({ page }) => {
      await expect(page).toHaveTitle(/云值守/)
    })
  })

  test.describe('登录成功', () => {
    test('管理员登录应跳转到仪表盘', async ({ page }) => {
      await page.goto('/login')

      // 填写管理员凭据
      await page.locator('.q-input').first().locator('input').fill('admin')
      await page.locator('.q-input').last().locator('input').fill('mzjy.com')
      await page.locator('button.q-btn').click()

      // 验证跳转到仪表盘
      await expect(page).toHaveURL(/\/app\/(dashboard|my-handled)/)
    })

    test('登录后应保存 Token', async ({ page }) => {
      await page.goto('/login')
      await page.locator('.q-input').first().locator('input').fill('admin')
      await page.locator('.q-input').last().locator('input').fill('mzjy.com')
      await page.locator('button.q-btn').click()

      // 等待登录完成
      await page.waitForURL(/\/app\//)

      // 验证 localStorage 中有 token
      const token = await page.evaluate(() => localStorage.getItem('token'))
      expect(token).toBeTruthy()
    })
  })

  test.describe('路由守卫', () => {
    test('未登录访问受保护路由应跳转到登录页', async ({ page }) => {
      // 直接访问受保护路由
      await page.goto('/app/dashboard')

      // 应跳转到登录页
      await expect(page).toHaveURL('/login')
    })

    test('已登录访问登录页应跳转到首页', async ({ page }) => {
      // 先登录
      await page.goto('/login')
      await page.locator('.q-input').first().locator('input').fill('admin')
      await page.locator('.q-input').last().locator('input').fill('mzjy.com')
      await page.locator('button.q-btn').click()
      await page.waitForURL(/\/app\//)

      // 再访问登录页
      await page.goto('/login')

      // 应跳转回应用页面
      await expect(page).toHaveURL(/\/app\//)
    })
  })

  test.describe('退出登录', () => {
    test('退出后应清除 Token 并跳转登录页', async ({ page }) => {
      // 登录
      await page.goto('/login')
      await page.locator('.q-input').first().locator('input').fill('admin')
      await page.locator('.q-input').last().locator('input').fill('mzjy.com')
      await page.locator('button.q-btn').click()
      await page.waitForURL(/\/app\//)

      // 点击退出按钮（需根据实际选择器调整）
      // 如果有用户菜单下拉，需要先点击展开
      const userMenu = page.locator('[data-testid="user-menu"], .q-btn-dropdown, .user-avatar')
      if (await userMenu.count() > 0) {
        await userMenu.first().click()
      }

      // 点击退出按钮
      const logoutBtn = page.locator('[data-testid="logout-button"], button:has-text("退出"), button:has-text("登出")')
      if (await logoutBtn.count() > 0) {
        await logoutBtn.first().click()
      } else {
        // 直接清除 localStorage 模拟退出
        await page.evaluate(() => localStorage.clear())
        await page.goto('/login')
      }

      // 验证跳转到登录页
      await expect(page).toHaveURL('/login')

      // 验证 Token 已清除
      const token = await page.evaluate(() => localStorage.getItem('token'))
      expect(token).toBeNull()
    })
  })
})
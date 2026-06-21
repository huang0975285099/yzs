/**
 * 认证测试固件
 * 用于复用登录状态，提高测试效率
 */
const { test: base } = require('@playwright/test')

/**
 * 认证固件扩展
 */
exports.test = base.extend({
  // 已认证的页面
  authenticatedPage: async ({ page }, use) => {
    // 执行登录
    await page.goto('/login')
    await page.locator('.q-input').first().locator('input').fill('admin')
    await page.locator('.q-input').last().locator('input').fill('mzjy.com')
    await page.locator('button.q-btn').click()
    await page.waitForURL(/\/app\//)

    // 保存存储状态到文件
    await page.context().storageState({ path: 'tests/.auth/admin.json' })

    // 传递给测试
    await use(page)
  },

  // 操作员身份
  operatorPage: async ({ page }, use) => {
    await page.goto('/login')
    await page.locator('.q-input').first().locator('input').fill('operator')
    await page.locator('.q-input').last().locator('input').fill('operator123')
    await page.locator('button.q-btn').click()
    await page.waitForURL(/\/app\/my-handled/)
    await use(page)
  },

  // 质检员身份
  inspectorPage: async ({ page }, use) => {
    await page.goto('/login')
    await page.locator('.q-input').first().locator('input').fill('inspector')
    await page.locator('.q-input').last().locator('input').fill('inspector123')
    await page.locator('button.q-btn').click()
    await page.waitForURL(/\/app\//)
    await use(page)
  },
})

exports.expect = require('@playwright/test').expect
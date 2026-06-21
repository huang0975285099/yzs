/**
 * 测试辅助函数
 */
const { expect } = require('@playwright/test')

/**
 * 登录辅助函数
 * @param {Page} page - Playwright Page 对象
 * @param {string} username - 用户名
 * @param {string} password - 密码
 */
async function login(page, username, password) {
  await page.goto('/login')
  await page.fill('input[type="text"]', username)
  await page.fill('input[type="password"]', password)
  await page.click('button[type="submit"]')
  await page.waitForURL(/\/app\//)
}

/**
 * 等待表格加载完成
 * @param {Page} page - Playwright Page 对象
 */
async function waitForTable(page) {
  await page.waitForSelector('.q-table tbody tr, .q-table__bottom--nodata', {
    state: 'visible',
    timeout: 10000,
  })
}

/**
 * 获取表格行数
 * @param {Page} page - Playwright Page 对象
 * @returns {Promise<number>} 行数
 */
async function getTableRowCount(page) {
  return await page.locator('.q-table tbody tr').count()
}

/**
 * 点击表格排序
 * @param {Page} page - Playwright Page 对象
 * @param {string} columnName - 列名
 */
async function sortTableBy(page, columnName) {
  const header = page.locator(`.q-table th:has-text("${columnName}")`)
  if (await header.count() > 0) {
    await header.click()
  }
}

/**
 * 等待通知消息
 * @param {Page} page - Playwright Page 对象
 * @returns {Promise<string>} 通知文本
 */
async function waitForNotification(page) {
  const notification = page.locator('.q-notification')
  await notification.waitFor({ state: 'visible', timeout: 5000 })
  return await notification.textContent()
}

/**
 * 检查是否显示成功通知
 * @param {Page} page - Playwright Page 对象
 */
async function expectSuccessNotification(page) {
  const notification = page.locator('.q-notification')
  await expect(notification).toBeVisible()
}

/**
 * 检查是否显示错误通知
 * @param {Page} page - Playwright Page 对象
 */
async function expectErrorNotification(page) {
  const notification = page.locator('.q-notification')
  await expect(notification).toBeVisible()
}

/**
 * 响应式视口预设
 */
const viewports = {
  mobile: { width: 375, height: 667 },   // iPhone SE
  mobileLarge: { width: 414, height: 896 }, // iPhone XR
  tablet: { width: 768, height: 1024 },  // iPad
  tabletLarge: { width: 1024, height: 768 }, // iPad Pro
  desktop: { width: 1440, height: 900 }, // 桌面端
  desktopLarge: { width: 1920, height: 1080 }, // 大屏桌面
}

/**
 * 设置视口并导航
 * @param {Page} page - Playwright Page 对象
 * @param {object} viewport - 视口尺寸
 * @param {string} url - 目标 URL
 */
async function setViewportAndNavigate(page, viewport, url) {
  await page.setViewportSize(viewport)
  await page.goto(url)
}

/**
 * 模拟 API 响应
 * @param {Page} page - Playwright Page 对象
 * @param {string} urlPattern - URL 匹配模式
 * @param {object} response - 响应数据
 * @param {number} status - HTTP 状态码
 */
async function mockApi(page, urlPattern, response, status = 200) {
  await page.route(urlPattern, async (route) => {
    await route.fulfill({
      status,
      contentType: 'application/json',
      body: JSON.stringify(response),
    })
  })
}

/**
 * 清除所有 Mock
 * @param {Page} page - Playwright Page 对象
 */
async function clearMocks(page) {
  await page.unrouteAll({ behavior: 'ignoreErrors' })
}

/**
 * 等待 API 请求完成
 * @param {Page} page - Playwright Page 对象
 * @param {string} urlPattern - URL 匹配模式
 */
async function waitForApi(page, urlPattern) {
  await page.waitForResponse(urlPattern)
}

/**
 * 截图保存
 * @param {Page} page - Playwright Page 对象
 * @param {string} filename - 文件名
 */
async function saveScreenshot(page, filename) {
  await page.screenshot({
    path: `tests/screenshots/${filename}.png`,
    fullPage: true,
  })
}

module.exports = {
  login,
  waitForTable,
  getTableRowCount,
  sortTableBy,
  waitForNotification,
  expectSuccessNotification,
  expectErrorNotification,
  viewports,
  setViewportAndNavigate,
  mockApi,
  clearMocks,
  waitForApi,
  saveScreenshot,
}
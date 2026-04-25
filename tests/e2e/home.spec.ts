import { test, expect } from '@playwright/test'

test.describe('DevCleaner E2E Tests', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  test('homepage loads correctly', async ({ page }) => {
    // 检查标题
    await expect(page.locator('h1')).toContainText('DevCleaner')
  })

  test('scan button exists', async ({ page }) => {
    // 检查扫描按钮
    const scanButton = page.getByRole('button', { name: /扫描|scan/i })
    await expect(scanButton).toBeVisible()
  })

  test('settings button navigates to settings', async ({ page }) => {
    // 点击设置按钮
    await page.getByRole('button', { name: /设置|settings/i }).click()
    // 应该跳转到设置页面
    await expect(page).toHaveURL(/settings/)
  })

  test('tool cards are displayed', async ({ page }) => {
    // 等待工具卡片加载
    await page.waitForSelector('.tool-card', { timeout: 5000 })
    const cards = await page.locator('.tool-card').count()
    expect(cards).toBeGreaterThan(0)
  })
})

test.describe('Settings Page', () => {
  test('settings page loads', async ({ page }) => {
    await page.goto('/settings')
    await expect(page.locator('h2')).toContainText('设置|Settings')
  })
})

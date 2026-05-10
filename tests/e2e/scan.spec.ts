import { test, expect } from '@playwright/test'

test.describe('扫描页面测试', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/scan')
    await page.waitForLoadState('networkidle')
  })

  test('扫描页面加载成功', async ({ page }) => {
    await expect(page).toHaveTitle(/DevCleaner/i)
  })

  test('显示扫描标题', async ({ page }) => {
    const title = page.locator('h1, h2').first()
    await expect(title).toBeVisible()
  })

  test('显示扫描工具列表', async ({ page }) => {
    await page.waitForTimeout(1000)
    const toolList = page.locator('.tool-list, .scan-list, [class*="tool"]')
    await expect(toolList.first()).toBeVisible()
  })

  test('显示扫描全部按钮', async ({ page }) => {
    const scanAllButton = page.locator('button:has-text("扫描全部"), a-button:has-text("扫描全部")')
    await expect(scanAllButton.first()).toBeVisible()
  })

  test('可以选中单个工具', async ({ page }) => {
    await page.waitForTimeout(500)
    const toolItem = page.locator('.tool-item, .scan-item, [class*="tool-item"]').first()
    if (await toolItem.count() > 0) {
      await toolItem.click()
    }
  })

  test('点击扫描全部开始扫描', async ({ page }) => {
    const scanAllButton = page.locator('button:has-text("扫描全部"), a-button:has-text("扫描全部")').first()
    await scanAllButton.click()
    
    await page.waitForTimeout(500)
    // 按钮应该显示加载状态
    const loadingButton = page.locator('button:has-text("扫描中"), a-button:has-text("扫描中")')
    await expect(loadingButton.first()).toBeVisible({ timeout: 3000 })
  })

  test('显示返回按钮', async ({ page }) => {
    const backButton = page.locator('button:has-text("返回"), .back-btn')
    await expect(backButton.first()).toBeVisible()
  })
})

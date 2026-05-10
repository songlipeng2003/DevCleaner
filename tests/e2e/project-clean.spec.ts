import { test, expect } from '@playwright/test'

test.describe('项目清理页面测试', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/project-clean')
    await page.waitForLoadState('networkidle')
  })

  test('项目清理页面加载成功', async ({ page }) => {
    await expect(page).toHaveTitle(/DevCleaner/i)
  })

  test('显示页面标题', async ({ page }) => {
    const title = page.locator('h1, h2').first()
    await expect(title).toBeVisible()
  })

  test('显示扫描路径配置', async ({ page }) => {
    await page.waitForTimeout(500)
    const configSection = page.locator('.config, .settings, [class*="config"]')
    await expect(configSection.first()).toBeVisible({ timeout: 3000 })
  })

  test('显示扫描按钮', async ({ page }) => {
    const scanButton = page.locator('button:has-text("扫描"), a-button:has-text("扫描")')
    await expect(scanButton.first()).toBeVisible()
  })

  test('点击扫描按钮开始扫描', async ({ page }) => {
    const scanButton = page.locator('button:has-text("扫描"), a-button:has-text("扫描")').first()
    await scanButton.click()
    
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

test.describe('项目清理配置测试', () => {
  test('可以修改扫描路径', async ({ page }) => {
    await page.goto('/project-clean')
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(500)
    
    const pathInput = page.locator('input[type="text"], .path-input')
    const count = await pathInput.count()
    if (count > 0) {
      await pathInput.first().clear()
      await pathInput.first().fill('~/Projects')
    }
  })

  test('显示深度设置', async ({ page }) => {
    await page.goto('/project-clean')
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(500)
    
    const depthSetting = page.locator('text=深度, [class*="depth"]')
    await expect(depthSetting.first()).toBeVisible({ timeout: 3000 })
  })
})

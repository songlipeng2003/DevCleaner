import { test, expect } from '@playwright/test'

test.describe('项目清理页面测试', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/projects')
    await page.waitForLoadState('networkidle')
  })

  test('项目页面加载成功', async ({ page }) => {
    await expect(page).toHaveTitle(/DevCleaner/i)
  })

  test('页面主体可见', async ({ page }) => {
    const body = page.locator('body')
    await expect(body).toBeVisible()
  })
})

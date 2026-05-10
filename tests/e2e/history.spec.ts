import { test, expect } from '@playwright/test'

test.describe('历史记录页面测试', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/history')
    await page.waitForLoadState('networkidle')
  })

  test('历史页面加载成功', async ({ page }) => {
    await expect(page).toHaveTitle(/DevCleaner/i)
  })

  test('显示历史标题', async ({ page }) => {
    const title = page.locator('h1, h2')
    await expect(title.first()).toBeVisible()
  })

  test('显示过滤器选项', async ({ page }) => {
    await page.waitForTimeout(500)
    const filterSection = page.locator('.filter, .tabs, [class*="filter"]')
    await expect(filterSection.first()).toBeVisible({ timeout: 3000 })
  })

  test('可以切换过滤器', async ({ page }) => {
    await page.waitForTimeout(500)
    // 查找过滤器选项
    const filterOptions = page.locator('.tab-item, .filter-item, button')
    const count = await filterOptions.count()
    if (count > 1) {
      await filterOptions.nth(1).click()
      await page.waitForTimeout(300)
    }
  })

  test('显示统计数据', async ({ page }) => {
    await page.waitForTimeout(500)
    const statsSection = page.locator('.stats, .statistics, [class*="stat"]')
    await expect(statsSection.first()).toBeVisible({ timeout: 3000 })
  })

  test('显示返回按钮', async ({ page }) => {
    const backButton = page.locator('button:has-text("返回"), .back-btn')
    await expect(backButton.first()).toBeVisible()
  })
})

test.describe('历史记录功能测试', () => {
  test('切换到今日视图', async ({ page }) => {
    await page.goto('/history')
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(500)
    
    const dayFilter = page.locator('button:has-text("今日"), .tab-item:has-text("今日")')
    await dayFilter.first().click()
    
    await page.waitForTimeout(300)
  })

  test('切换到本周视图', async ({ page }) => {
    await page.goto('/history')
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(500)
    
    const weekFilter = page.locator('button:has-text("本周"), .tab-item:has-text("本周")')
    await weekFilter.first().click()
    
    await page.waitForTimeout(300)
  })

  test('切换到全部视图', async ({ page }) => {
    await page.goto('/history')
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(500)
    
    const allFilter = page.locator('button:has-text("全部"), .tab-item:has-text("全部")')
    await allFilter.first().click()
    
    await page.waitForTimeout(300)
  })
})

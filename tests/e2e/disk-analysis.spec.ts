import { test, expect } from '@playwright/test'

test.describe('磁盘分析页面测试', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/disk-analysis')
    await page.waitForLoadState('networkidle')
  })

  test('磁盘分析页面加载成功', async ({ page }) => {
    await expect(page).toHaveTitle(/DevCleaner/i)
  })

  test('显示页面标题', async ({ page }) => {
    const title = page.locator('h1, h2').first()
    await expect(title).toBeVisible()
  })

  test('显示分析按钮', async ({ page }) => {
    const analyzeButton = page.locator('button:has-text("分析"), a-button:has-text("分析")')
    await expect(analyzeButton.first()).toBeVisible()
  })

  test('点击分析按钮开始分析', async ({ page }) => {
    const analyzeButton = page.locator('button:has-text("分析"), a-button:has-text("分析")').first()
    await analyzeButton.click()
    
    await page.waitForTimeout(500)
    // 应该显示加载状态
    const loadingIndicator = page.locator('.loading, .spinning, [class*="loading"]')
    await expect(loadingIndicator.first()).toBeVisible({ timeout: 3000 })
  })

  test('显示返回按钮', async ({ page }) => {
    const backButton = page.locator('button:has-text("返回"), .back-btn')
    await expect(backButton.first()).toBeVisible()
  })
})

test.describe('磁盘分析结果测试', () => {
  test('显示分类统计', async ({ page }) => {
    await page.goto('/disk-analysis')
    await page.waitForLoadState('networkidle')
    
    // 点击分析
    const analyzeButton = page.locator('button:has-text("分析"), a-button:has-text("分析")').first()
    await analyzeButton.click()
    
    // 等待分析完成
    await page.waitForTimeout(3000)
    
    const categorySection = page.locator('.categories, .category-list, [class*="category"]')
    await expect(categorySection.first()).toBeVisible({ timeout: 10000 })
  })

  test('显示缓存趋势', async ({ page }) => {
    await page.goto('/disk-analysis')
    await page.waitForLoadState('networkidle')
    await page.waitForTimeout(500)
    
    const trendSection = page.locator('.trends, .trend-chart, [class*="trend"]')
    await expect(trendSection.first()).toBeVisible({ timeout: 5000 })
  })
})

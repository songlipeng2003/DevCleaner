import { test, expect } from '@playwright/test'

test.describe('DevCleaner 首页测试', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
  })

  test('页面标题正确', async ({ page }) => {
    await expect(page).toHaveTitle(/DevCleaner/i)
  })

  test('页面主体内容可见', async ({ page }) => {
    const body = page.locator('body')
    await expect(body).toBeVisible()
  })
})

test.describe('扫描功能测试', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
  })

  test('扫描页面可访问', async ({ page }) => {
    await page.goto('/scan')
    await page.waitForLoadState('networkidle')
    const body = page.locator('body')
    await expect(body).toBeVisible()
  })
})

test.describe('设置页面测试', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/settings')
    await page.waitForLoadState('networkidle')
  })

  test('设置页面加载成功', async ({ page }) => {
    const body = page.locator('body')
    await expect(body).toBeVisible()
  })
})

test.describe('导航测试', () => {
  test('首页可以访问', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    await expect(page).toHaveTitle(/DevCleaner/i)
  })

  test('扫描页面可以访问', async ({ page }) => {
    await page.goto('/scan')
    await page.waitForLoadState('networkidle')
    const body = page.locator('body')
    await expect(body).toBeVisible()
  })

  test('设置页面可以访问', async ({ page }) => {
    await page.goto('/settings')
    await page.waitForLoadState('networkidle')
    const body = page.locator('body')
    await expect(body).toBeVisible()
  })

  test('项目页面可以访问', async ({ page }) => {
    await page.goto('/projects')
    await page.waitForLoadState('networkidle')
    const body = page.locator('body')
    await expect(body).toBeVisible()
  })

  test('历史页面可以访问', async ({ page }) => {
    await page.goto('/history')
    await page.waitForLoadState('networkidle')
    const body = page.locator('body')
    await expect(body).toBeVisible()
  })

  test('分析页面可以访问', async ({ page }) => {
    await page.goto('/analysis')
    await page.waitForLoadState('networkidle')
    const body = page.locator('body')
    await expect(body).toBeVisible()
  })
})

test.describe('页面响应式测试', () => {
  test('桌面端视图正常', async ({ page }) => {
    await page.setViewportSize({ width: 1920, height: 1080 })
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    
    const body = page.locator('body')
    await expect(body).toBeVisible()
  })

  test('平板视图正常', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 })
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    
    const body = page.locator('body')
    await expect(body).toBeVisible()
  })

  test('移动端视图正常', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    
    const body = page.locator('body')
    await expect(body).toBeVisible()
  })
})

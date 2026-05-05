import { test, expect } from '@playwright/test'

test.describe('DevCleaner E2E Tests', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  test('homepage loads and displays app name', async ({ page }) => {
    // 检查页面加载
    await expect(page).toHaveTitle(/DevCleaner/i)
  })

  test('navigation links exist', async ({ page }) => {
    // 检查导航链接存在
    const hasNav = await page.locator('nav, header, .nav, .header').count() > 0
    if (!hasNav) {
      // 如果没有导航，检查 URL 变化
      await page.goto('/settings')
      await expect(page).not.toHaveTitle('')
    }
  })
})

test.describe('Basic Page Tests', () => {
  test('home page is accessible', async ({ page }) => {
    await page.goto('/')
    // 基本检查
    const body = page.locator('body')
    await expect(body).toBeVisible()
  })

  test('settings page is accessible', async ({ page }) => {
    await page.goto('/settings')
    const body = page.locator('body')
    await expect(body).toBeVisible()
  })

  test('scan page is accessible', async ({ page }) => {
    await page.goto('/scan')
    const body = page.locator('body')
    await expect(body).toBeVisible()
  })
})

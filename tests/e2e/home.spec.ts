import { test, expect } from '@playwright/test'

test.describe('DevCleaner E2E Tests', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
  })

  test('homepage loads correctly', async ({ page }) => {
    // 检查标题
    await expect(page.locator('h1')).toContainText('DevCleaner')
    // 检查版本号
    await expect(page.locator('.version')).toContainText('v')
  })

  test('disk usage information is displayed', async ({ page }) => {
    // 检查磁盘使用情况卡片
    await expect(page.locator('.disk-card')).toBeVisible()
    // 检查进度条
    await expect(page.locator('.ant-progress')).toBeVisible()
    // 检查可用空间统计
    await expect(page.locator('.ant-statistic-title')).toContainText('可用空间')
  })

  test('scan button exists and works', async ({ page }) => {
    // 检查扫描按钮
    const scanButton = page.getByRole('button', { name: /开始扫描|scan/i })
    await expect(scanButton).toBeVisible()
    
    // 点击扫描按钮
    await scanButton.click()
    // 检查扫描状态（可能显示扫描中或完成）
    await expect(page.locator('.ant-spin')).toBeVisible({ timeout: 5000 })
  })

  test('tool cards are displayed with correct information', async ({ page }) => {
    // 等待工具卡片加载
    await page.waitForSelector('.tool-card', { timeout: 10000 })
    const cards = await page.locator('.tool-card').count()
    expect(cards).toBeGreaterThan(0)
    
    // 检查每个卡片包含工具名称和大小
    for (let i = 0; i < Math.min(cards, 3); i++) {
      const card = page.locator('.tool-card').nth(i)
      await expect(card.locator('.tool-name')).not.toBeEmpty()
      await expect(card.locator('.tool-size')).toBeVisible()
    }
  })

  test('tool toggle switch works', async ({ page }) => {
    // 等待工具卡片加载
    await page.waitForSelector('.tool-card', { timeout: 10000 })
    // 找到第一个工具的开关
    const firstSwitch = page.locator('.tool-card .ant-switch').first()
    await expect(firstSwitch).toBeVisible()
    
    // 获取初始状态
    const isChecked = await firstSwitch.isChecked()
    // 点击切换
    await firstSwitch.click()
    // 等待状态变化
    await expect(firstSwitch).toBeChecked({ checked: !isChecked })
  })

  test('tool detail drawer opens', async ({ page }) => {
    // 等待工具卡片加载
    await page.waitForSelector('.tool-card', { timeout: 10000 })
    // 点击第一个工具卡片
    await page.locator('.tool-card').first().click()
    // 检查抽屉打开
    await expect(page.locator('.ant-drawer')).toBeVisible({ timeout: 5000 })
    // 检查抽屉标题
    await expect(page.locator('.ant-drawer-title')).not.toBeEmpty()
  })

  test('settings button navigates to settings page', async ({ page }) => {
    // 点击设置按钮
    await page.getByRole('button', { name: /设置|settings/i }).click()
    // 应该跳转到设置页面
    await expect(page).toHaveURL(/settings/)
  })
})

test.describe('Settings Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/settings')
  })

  test('settings page loads with all sections', async ({ page }) => {
    await expect(page.locator('h2')).toContainText('设置')
    // 检查主题设置
    await expect(page.locator('label:has-text("主题模式")')).toBeVisible()
    // 检查白名单设置
    await expect(page.locator('label:has-text("白名单路径")')).toBeVisible()
    // 检查自动扫描设置
    await expect(page.locator('label:has-text("自动扫描")')).toBeVisible()
  })

  test('theme switch works', async ({ page }) => {
    const themeSwitch = page.locator('.settings-section .ant-switch').first()
    await expect(themeSwitch).toBeVisible()
    
    // 获取初始状态
    const isChecked = await themeSwitch.isChecked()
    // 点击切换
    await themeSwitch.click()
    // 等待状态变化
    await expect(themeSwitch).toBeChecked({ checked: !isChecked })
  })

  test('scan view page works', async ({ page }) => {
    await page.goto('/scan')
    await expect(page.locator('h2')).toContainText('扫描结果')
    // 检查扫描结果列表
    await expect(page.locator('.ant-list')).toBeVisible()
  })
})

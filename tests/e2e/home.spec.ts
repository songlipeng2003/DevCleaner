import { test, expect } from '@playwright/test'

test.describe('DevCleaner 首页测试', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
    // 等待页面加载
    await page.waitForLoadState('networkidle')
  })

  test('页面标题正确', async ({ page }) => {
    await expect(page).toHaveTitle(/DevCleaner/i)
  })

  test('显示 DevCleaner 标题', async ({ page }) => {
    const title = page.locator('h1').first()
    await expect(title).toContainText('DevCleaner')
  })

  test('显示版本信息', async ({ page }) => {
    const version = page.locator('.version')
    await expect(version).toBeVisible()
  })

  test('显示磁盘使用情况', async ({ page }) => {
    const diskChart = page.locator('.disk-chart-container, .disk-chart, [class*="disk"]')
    // 等待磁盘数据加载
    await page.waitForTimeout(1000)
    await expect(diskChart.first()).toBeVisible()
  })

  test('显示磁盘百分比', async ({ page }) => {
    // 等待数据加载
    await page.waitForTimeout(1000)
    const percentText = page.locator('.disk-percentage, [class*="percentage"]')
    await expect(percentText.first()).toBeVisible()
  })

  test('显示已使用和可用空间', async ({ page }) => {
    await page.waitForTimeout(1000)
    const usedLabel = page.locator('text=已用, text=已使用')
    const freeLabel = page.locator('text=可用, text=空闲')
    await expect(usedLabel.first()).toBeVisible()
    await expect(freeLabel.first()).toBeVisible()
  })

  test('显示开始扫描按钮', async ({ page }) => {
    const scanButton = page.locator('button:has-text("扫描"), a-button:has-text("扫描")')
    await expect(scanButton.first()).toBeVisible()
  })

  test('显示刷新按钮', async ({ page }) => {
    const refreshButton = page.locator('button:has-text("刷新"), a-button:has-text("刷新")')
    await expect(refreshButton.first()).toBeVisible()
  })

  test('显示设置按钮', async ({ page }) => {
    const settingsButton = page.locator('button:has-text("设置"), a-button:has-text("设置")')
    await expect(settingsButton.first()).toBeVisible()
  })

  test('显示工具列表区域', async ({ page }) => {
    // 等待工具数据加载
    await page.waitForTimeout(1000)
    const toolSection = page.locator('.tool-card, .tools-section, [class*="tool"]')
    // 工具区域应该存在
    await expect(toolSection.first()).toBeVisible()
  })

  test('显示统计信息区域', async ({ page }) => {
    await page.waitForTimeout(500)
    const statsSection = page.locator('.stats, .statistics, [class*="stat"]')
    await expect(statsSection.first()).toBeVisible()
  })
})

test.describe('扫描功能测试', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
  })

  test('点击开始扫描按钮', async ({ page }) => {
    const scanButton = page.locator('button:has-text("扫描"), a-button:has-text("扫描")').first()
    await scanButton.click()
    
    // 应该显示加载状态或扫描进度
    await page.waitForTimeout(500)
    // 检查按钮变为扫描中状态
    const loadingButton = page.locator('button:has-text("扫描中"), a-button:has-text("扫描中")')
    await expect(loadingButton.first()).toBeVisible({ timeout: 5000 })
  })

  test('点击刷新按钮重新加载工具', async ({ page }) => {
    const refreshButton = page.locator('button:has-text("刷新"), a-button:has-text("刷新")').first()
    await refreshButton.click()
    
    // 等待刷新完成
    await page.waitForTimeout(1000)
    // 页面应该仍然正常显示
    await expect(page.locator('.home, [class*="home"]').first()).toBeVisible()
  })
})

test.describe('设置页面测试', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/settings')
    await page.waitForLoadState('networkidle')
  })

  test('设置页面加载成功', async ({ page }) => {
    const settingsTitle = page.locator('h1, h2:has-text("设置")')
    await expect(settingsTitle.first()).toBeVisible()
  })

  test('显示返回按钮', async ({ page }) => {
    const backButton = page.locator('button:has-text("返回"), .back-btn')
    await expect(backButton.first()).toBeVisible()
  })

  test('显示保存按钮', async ({ page }) => {
    const saveButton = page.locator('button:has-text("保存"), a-button:has-text("保存")')
    await expect(saveButton.first()).toBeVisible()
  })

  test('显示主题设置选项', async ({ page }) => {
    const themeSection = page.locator('text=外观, .theme-selector')
    await expect(themeSection.first()).toBeVisible()
  })

  test('主题选项可以点击', async ({ page }) => {
    const themeOptions = page.locator('.theme-option')
    const count = await themeOptions.count()
    if (count > 0) {
      await themeOptions.first().click()
      await expect(themeOptions.first()).toHaveClass(/active/)
    }
  })

  test('点击保存按钮', async ({ page }) => {
    const saveButton = page.locator('button:has-text("保存"), a-button:has-text("保存")').first()
    await saveButton.click()
    
    // 等待保存完成
    await page.waitForTimeout(500)
    // 设置页面应该仍然正常显示
    await expect(page.locator('.settings').first()).toBeVisible()
  })

  test('点击返回按钮回到首页', async ({ page }) => {
    const backButton = page.locator('button:has-text("返回"), .back-btn').first()
    await backButton.click()
    
    // 应该回到首页
    await page.waitForTimeout(500)
    await expect(page).toHaveURL(/\//)
  })
})

test.describe('导航测试', () => {
  test('通过首页设置按钮进入设置页面', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    
    const settingsButton = page.locator('button:has-text("设置"), a-button:has-text("设置")').first()
    await settingsButton.click()
    
    await page.waitForTimeout(500)
    await expect(page).toHaveURL(/\/settings/)
  })

  test('通过设置页面返回按钮回到首页', async ({ page }) => {
    await page.goto('/settings')
    await page.waitForLoadState('networkidle')
    
    const backButton = page.locator('button:has-text("返回"), .back-btn').first()
    await backButton.click()
    
    await page.waitForTimeout(500)
    await expect(page).toHaveURL(/\//)
  })

  test('直接从 URL 访问设置页面', async ({ page }) => {
    await page.goto('/settings')
    await expect(page).toHaveTitle(/DevCleaner/i)
  })

  test('直接从 URL 访问扫描页面', async ({ page }) => {
    await page.goto('/scan')
    await expect(page).toHaveTitle(/DevCleaner/i)
  })
})

test.describe('主题切换测试', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')
  })

  test('显示主题切换按钮', async ({ page }) => {
    const themeToggle = page.locator('.theme-toggle-btn, button[class*="theme"]')
    await expect(themeToggle.first()).toBeVisible()
  })

  test('点击主题切换按钮切换主题', async ({ page }) => {
    const themeToggle = page.locator('.theme-toggle-btn').first()
    await themeToggle.click()
    
    // 等待主题切换
    await page.waitForTimeout(300)
    
    // 检查 html 标签的 class 变化
    const html = page.locator('html')
    const classList = await html.getAttribute('class')
    // 应该包含 dark 或 light
    expect(classList).toMatch(/dark|light/)
  })
})

test.describe('页面响应式测试', () => {
  test('桌面端视图正常', async ({ page }) => {
    await page.setViewportSize({ width: 1920, height: 1080 })
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    
    await expect(page.locator('.home').first()).toBeVisible()
  })

  test('平板视图正常', async ({ page }) => {
    await page.setViewportSize({ width: 768, height: 1024 })
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    
    await expect(page.locator('.home').first()).toBeVisible()
  })

  test('移动端视图正常', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/')
    await page.waitForLoadState('networkidle')
    
    await expect(page.locator('.home').first()).toBeVisible()
  })
})

test.describe('错误处理测试', () => {
  test('网络错误时页面仍然可访问', async ({ page }) => {
    // 模拟离线状态
    await page.context().setOffline(true)
    
    await page.goto('/')
    await page.waitForLoadState('domcontentloaded')
    
    // 页面应该仍然可以加载基本结构
    const body = page.locator('body')
    await expect(body).toBeVisible()
    
    // 恢复网络
    await page.context().setOffline(false)
  })
})

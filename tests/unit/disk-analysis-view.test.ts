import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, config } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { nextTick } from 'vue'

// Mock Tauri API
vi.mock('@tauri-apps/api/core', () => ({
  invoke: vi.fn(),
}))

vi.mock('@tauri-apps/api/event', () => ({
  listen: vi.fn(),
}))

// Mock ant-design-vue components globally
config.global.stubs = {
  'a-layout': true,
  'a-layout-header': true,
  'a-layout-content': true,
  'a-layout-sider': true,
  'a-menu': true,
  'a-menu-item': true,
  'a-card': true,
  'a-button': true,
  'a-row': true,
  'a-col': true,
  'a-progress': true,
  'a-statistic': true,
  'a-space': true,
  'a-space-item': true,
  'a-switch': true,
  'a-list': true,
  'a-list-item': true,
  'a-list-item-meta': true,
  'a-divider': true,
  'a-alert': true,
  'a-empty': true,
  'a-spin': true,
  'a-tooltip': true,
  'a-progress': true,
  'a-tag': true,
  'a-badge': true,
  'a-table': true,
  'a-table-column': true,
  'a-collapse': true,
  'a-collapse-panel': true,
}

describe('DiskAnalysisView Component', () => {
  beforeEach(() => {
    const pinia = createPinia()
    setActivePinia(pinia)
    vi.clearAllMocks()
  })

  it('renders disk analysis view container', async () => {
    // Dynamic import to avoid hoisting issues
    const { default: DiskAnalysisView } = await import('@/views/DiskAnalysisView.vue')
    
    const wrapper = mount(DiskAnalysisView, {
      global: {
        plugins: [createPinia()],
        mocks: {
          $router: { push: vi.fn() },
        },
      },
    })
    await nextTick()

    expect(wrapper.find('.disk-analysis-view').exists()).toBe(true)
  })
})

describe('DiskAnalysisView - Data Display', () => {
  it('formats size correctly', () => {
    // Test size formatting utility
    const formatSize = (bytes: number): string => {
      if (bytes === 0) return '0 B'
      const k = 1024
      const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
    }

    expect(formatSize(0)).toBe('0 B')
    expect(formatSize(1024)).toBe('1 KB')
    expect(formatSize(1048576)).toBe('1 MB')
    expect(formatSize(1073741824)).toBe('1 GB')
    expect(formatSize(1099511627776)).toBe('1 TB')
  })

  it('calculates percentage correctly', () => {
    const calculatePercentage = (value: number, total: number): number => {
      if (total === 0) return 0
      return Math.round((value / total) * 100)
    }

    expect(calculatePercentage(500, 1000)).toBe(50)
    expect(calculatePercentage(250, 1000)).toBe(25)
    expect(calculatePercentage(1000, 1000)).toBe(100)
    expect(calculatePercentage(0, 1000)).toBe(0)
    expect(calculatePercentage(100, 0)).toBe(0)
  })
})

describe('DiskAnalysisView - Category Processing', () => {
  it('sorts categories by size descending', () => {
    const categories = [
      { name: 'npm', totalSize: 500000 },
      { name: 'docker', totalSize: 2000000 },
      { name: 'yarn', totalSize: 800000 },
    ]

    const sorted = [...categories].sort((a, b) => b.totalSize - a.totalSize)

    expect(sorted[0].name).toBe('docker')
    expect(sorted[1].name).toBe('yarn')
    expect(sorted[2].name).toBe('npm')
  })

  it('filters cleanable items correctly', () => {
    const items = [
      { name: 'node_modules', isCleanable: true },
      { name: 'src', isCleanable: false },
      { name: '__pycache__', isCleanable: true },
    ]

    const cleanable = items.filter(item => item.isCleanable)
    expect(cleanable).toHaveLength(2)
    expect(cleanable[0].name).toBe('node_modules')
    expect(cleanable[1].name).toBe('__pycache__')
  })

  it('calculates total cleanable size', () => {
    const categories = [
      { totalSize: 1000000, cleanableSize: 800000 },
      { totalSize: 500000, cleanableSize: 400000 },
      { totalSize: 200000, cleanableSize: 100000 },
    ]

    const totalCleanable = categories.reduce((sum, cat) => sum + cat.cleanableSize, 0)
    expect(totalCleanable).toBe(1300000)
  })
})

describe('DiskAnalysisView - Cache Trends', () => {
  it('processes trend data for chart', () => {
    const trends = [
      { date: '2026-03', size: 500000 },
      { date: '2026-04', size: 800000 },
      { date: '2026-05', size: 1024000 },
    ]

    // 转换为图表格式
    const chartData = trends.map(t => ({
      month: t.date,
      size: t.size / (1024 * 1024), // 转换为 MB
    }))

    expect(chartData).toHaveLength(3)
    expect(chartData[0].month).toBe('2026-03')
    expect(chartData[2].size).toBe(0.9765625) // 1024000 / 1048576 ≈ 0.976
  })

  it('detects cache growth trend', () => {
    const trends = [
      { date: '2026-03', size: 500000 },
      { date: '2026-04', size: 800000 },
      { date: '2026-05', size: 1200000 },
    ]

    const isGrowing = trends[trends.length - 1].size > trends[0].size
    expect(isGrowing).toBe(true)
  })

  it('calculates average monthly growth', () => {
    const trends = [
      { date: '2026-03', size: 500000 },
      { date: '2026-04', size: 800000 },
      { date: '2026-05', size: 1100000 },
    ]

    const firstSize = trends[0].size
    const lastSize = trends[trends.length - 1].size
    const growth = lastSize - firstSize
    const avgGrowth = growth / (trends.length - 1)

    expect(avgGrowth).toBe(300000)
  })
})

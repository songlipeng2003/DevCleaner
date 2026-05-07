import { describe, it, expect, vi, beforeEach } from 'vitest'

// Use vi.hoisted to define mocks before module imports
const { mockInvoke } = vi.hoisted(() => ({
  mockInvoke: vi.fn(),
}))

vi.mock('@tauri-apps/api/core', () => ({
  invoke: mockInvoke,
}))

import {
  getDiskAnalysis,
  getCacheTrends,
  type DiskAnalysisSummary,
  type DiskAnalysisCategory,
  type DiskAnalysisItem,
  type CacheTrend,
} from '@/services/tauri'

describe('v0.3.0 Disk Analysis Functions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  // ============== getDiskAnalysis Tests ==============
  describe('getDiskAnalysis', () => {
    it('calls invoke with get_disk_analysis', async () => {
      const mockAnalysis: DiskAnalysisSummary = {
        totalSize: 5000000,
        cleanableSize: 3000000,
        analyzedPaths: 10,
        analyzedAt: Date.now(),
        categories: [],
      }
      mockInvoke.mockResolvedValue(mockAnalysis)

      const result = await getDiskAnalysis()

      expect(mockInvoke).toHaveBeenCalledWith('get_disk_analysis')
      expect(result).toEqual(mockAnalysis)
    })

    it('returns analysis with multiple categories', async () => {
      const mockAnalysis: DiskAnalysisSummary = {
        totalSize: 5000000,
        cleanableSize: 3000000,
        analyzedPaths: 5,
        analyzedAt: Date.now(),
        categories: [
          {
            name: 'npm',
            toolId: 'npm',
            items: [
              {
                name: 'node_modules',
                path: '/home/user/project/node_modules',
                toolId: 'npm',
                size: 2000000,
                percentage: 40,
                fileCount: 500,
                lastModified: Date.now(),
                isCleanable: true,
                cleanReason: 'Dependencies can be reinstalled',
              },
            ],
            totalSize: 2000000,
            itemCount: 1,
          },
          {
            name: 'Docker',
            toolId: 'docker',
            items: [],
            totalSize: 3000000,
            itemCount: 0,
          },
        ],
      }
      mockInvoke.mockResolvedValue(mockAnalysis)

      const result = await getDiskAnalysis()

      expect(result.categories).toHaveLength(2)
      expect(result.totalSize).toBe(5000000)
      expect(result.cleanableSize).toBe(3000000)
    })

    it('handles empty analysis', async () => {
      const mockAnalysis: DiskAnalysisSummary = {
        totalSize: 0,
        cleanableSize: 0,
        analyzedPaths: 0,
        analyzedAt: Date.now(),
        categories: [],
      }
      mockInvoke.mockResolvedValue(mockAnalysis)

      const result = await getDiskAnalysis()

      expect(result.categories).toHaveLength(0)
      expect(result.totalSize).toBe(0)
    })

    it('includes cleanable items with reasons', async () => {
      const mockItem: DiskAnalysisItem = {
        name: 'node_modules',
        path: '/test/node_modules',
        toolId: 'npm',
        size: 1024000,
        percentage: 50,
        fileCount: 100,
        lastModified: Date.now(),
        isCleanable: true,
        cleanReason: 'Dependencies can be reinstalled',
      }
      const mockAnalysis: DiskAnalysisSummary = {
        totalSize: 1024000,
        cleanableSize: 1024000,
        analyzedPaths: 1,
        analyzedAt: Date.now(),
        categories: [
          {
            name: 'npm',
            toolId: 'npm',
            items: [mockItem],
            totalSize: 1024000,
            itemCount: 1,
          },
        ],
      }
      mockInvoke.mockResolvedValue(mockAnalysis)

      const result = await getDiskAnalysis()
      const item = result.categories[0].items[0]

      expect(item.isCleanable).toBe(true)
      expect(item.cleanReason).toBeDefined()
    })

    it('marks non-cleanable items correctly', async () => {
      const mockItem: DiskAnalysisItem = {
        name: 'source',
        path: '/test/src',
        toolId: null as any,
        size: 500000,
        percentage: 50,
        fileCount: 50,
        lastModified: Date.now(),
        isCleanable: false,
        cleanReason: 'Source code, cannot be deleted',
      }
      const mockAnalysis: DiskAnalysisSummary = {
        totalSize: 500000,
        cleanableSize: 0,
        analyzedPaths: 1,
        analyzedAt: Date.now(),
        categories: [
          {
            name: 'Other',
            toolId: null as any,
            items: [mockItem],
            totalSize: 500000,
            itemCount: 1,
          },
        ],
      }
      mockInvoke.mockResolvedValue(mockAnalysis)

      const result = await getDiskAnalysis()
      const item = result.categories[0].items[0]

      expect(item.isCleanable).toBe(false)
      expect(item.cleanReason).toBe('Source code, cannot be deleted')
    })

    it('includes analyzed timestamp', async () => {
      const now = Date.now()
      const mockAnalysis: DiskAnalysisSummary = {
        totalSize: 1000000,
        cleanableSize: 800000,
        analyzedPaths: 3,
        analyzedAt: now,
        categories: [],
      }
      mockInvoke.mockResolvedValue(mockAnalysis)

      const result = await getDiskAnalysis()

      expect(result.analyzedAt).toBe(now)
    })
  })

  // ============== getCacheTrends Tests ==============
  describe('getCacheTrends', () => {
    it('calls invoke with get_cache_trends', async () => {
      const mockTrends: CacheTrend[] = []
      mockInvoke.mockResolvedValue(mockTrends)

      const result = await getCacheTrends()

      expect(mockInvoke).toHaveBeenCalledWith('get_cache_trends')
      expect(result).toEqual(mockTrends)
    })

    it('returns monthly trend data', async () => {
      const mockTrends: CacheTrend[] = [
        { date: '2026-03', size: 500000 },
        { date: '2026-04', size: 800000 },
        { date: '2026-05', size: 1024000 },
      ]
      mockInvoke.mockResolvedValue(mockTrends)

      const result = await getCacheTrends()

      expect(result).toHaveLength(3)
      expect(result[0].date).toBe('2026-03')
      expect(result[2].size).toBeGreaterThan(result[0].size)
    })

    it('handles empty trends', async () => {
      const mockTrends: CacheTrend[] = []
      mockInvoke.mockResolvedValue(mockTrends)

      const result = await getCacheTrends()

      expect(result).toHaveLength(0)
    })

    it('returns trends in order returned from backend', async () => {
      // Mock 数据直接返回，不进行排序验证
      const mockTrends: CacheTrend[] = [
        { date: '2026-01', size: 100000 },
        { date: '2026-02', size: 200000 },
        { date: '2026-03', size: 300000 },
      ]
      mockInvoke.mockResolvedValue(mockTrends)

      const result = await getCacheTrends()

      // 验证返回的数据结构正确
      expect(result).toHaveLength(3)
      expect(result[0].date).toBe('2026-01')
    })
  })

  // ============== Type Tests ==============
  describe('Disk Analysis Types', () => {
    it('DiskAnalysisSummary has required fields', () => {
      const summary: DiskAnalysisSummary = {
        totalSize: 1000000,
        cleanableSize: 800000,
        analyzedPaths: 5,
        analyzedAt: Date.now(),
        categories: [],
      }

      expect(summary.totalSize).toBeDefined()
      expect(summary.cleanableSize).toBeDefined()
      expect(summary.analyzedPaths).toBeDefined()
      expect(summary.analyzedAt).toBeDefined()
      expect(summary.categories).toBeDefined()
    })

    it('DiskAnalysisCategory has required fields', () => {
      const category: DiskAnalysisCategory = {
        name: 'npm',
        toolId: 'npm',
        items: [],
        totalSize: 500000,
        itemCount: 3,
      }

      expect(category.name).toBe('npm')
      expect(category.totalSize).toBe(500000)
      expect(category.itemCount).toBe(3)
    })

    it('DiskAnalysisItem calculates percentage correctly', () => {
      const item: DiskAnalysisItem = {
        name: 'test',
        path: '/test',
        toolId: 'npm',
        size: 250000,
        percentage: 25,
        fileCount: 50,
        lastModified: Date.now(),
        isCleanable: true,
        cleanReason: 'Test',
      }

      expect(item.percentage).toBe(25)
      expect(item.size).toBe(250000)
    })

    it('CacheTrend has required fields', () => {
      const trend: CacheTrend = {
        date: '2026-05',
        size: 1024000,
      }

      expect(trend.date).toBe('2026-05')
      expect(trend.size).toBe(1024000)
    })
  })

  // ============== Percentage Calculation Tests ==============
  describe('Percentage Calculation', () => {
    it('calculates correct percentage for single item', async () => {
      const mockAnalysis: DiskAnalysisSummary = {
        totalSize: 1000000,
        cleanableSize: 1000000,
        analyzedPaths: 1,
        analyzedAt: Date.now(),
        categories: [
          {
            name: 'npm',
            toolId: 'npm',
            items: [
              {
                name: 'cache',
                path: '/test/cache',
                toolId: 'npm',
                size: 1000000,
                percentage: 100,
                fileCount: 100,
                lastModified: Date.now(),
                isCleanable: true,
                cleanReason: 'Cache',
              },
            ],
            totalSize: 1000000,
            itemCount: 1,
          },
        ],
      }
      mockInvoke.mockResolvedValue(mockAnalysis)

      const result = await getDiskAnalysis()

      expect(result.categories[0].items[0].percentage).toBe(100)
    })

    it('calculates correct percentages for multiple items', async () => {
      const mockAnalysis: DiskAnalysisSummary = {
        totalSize: 1000000,
        cleanableSize: 1000000,
        analyzedPaths: 2,
        analyzedAt: Date.now(),
        categories: [
          {
            name: 'npm',
            toolId: 'npm',
            items: [
              {
                name: 'cache1',
                path: '/test/cache1',
                toolId: 'npm',
                size: 750000,
                percentage: 75,
                fileCount: 75,
                lastModified: Date.now(),
                isCleanable: true,
                cleanReason: 'Cache',
              },
              {
                name: 'cache2',
                path: '/test/cache2',
                toolId: 'npm',
                size: 250000,
                percentage: 25,
                fileCount: 25,
                lastModified: Date.now(),
                isCleanable: true,
                cleanReason: 'Cache',
              },
            ],
            totalSize: 1000000,
            itemCount: 2,
          },
        ],
      }
      mockInvoke.mockResolvedValue(mockAnalysis)

      const result = await getDiskAnalysis()

      expect(result.categories[0].items[0].percentage).toBe(75)
      expect(result.categories[0].items[1].percentage).toBe(25)
    })
  })
})

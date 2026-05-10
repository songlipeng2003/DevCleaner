import { describe, it, expect, vi, beforeEach } from 'vitest'

// Use vi.hoisted to define mocks before module imports
const { mockInvoke } = vi.hoisted(() => ({
  mockInvoke: vi.fn(),
}))

vi.mock('@tauri-apps/api/core', () => ({
  invoke: mockInvoke,
}))

import {
  getCleanPreview,
  scanProjects,
  cleanPaths,
  getCleanHistory,
  recordCleanHistory,
  exportCleanReport,
  type CleanPreview,
  type ProjectScanResult,
  type CleanHistory,
  type CleanStrategy,
} from '@/services/tauri'

describe('v0.2.0 Tauri Service Functions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  // ============== getCleanPreview Tests ==============
  describe('getCleanPreview', () => {
    it('calls invoke with basic parameters', async () => {
      const mockPreview: CleanPreview = {
        toolId: 'npm',
        toolName: 'npm',
        paths: [],
        totalSize: 1024000,
        riskLevel: 'safe',
        recommendations: ['Remove node_modules older than 30 days'],
      }
      mockInvoke.mockResolvedValue(mockPreview)

      const result = await getCleanPreview('npm', ['~/.npm'])

      expect(mockInvoke).toHaveBeenCalledWith('get_clean_preview', {
        toolId: 'npm',
        paths: ['~/.npm'],
        strategy: undefined,
        timeThreshold: undefined,
        sizeThreshold: undefined,
      })
      expect(result).toEqual(mockPreview)
    })

    it('calls invoke with all strategy parameters', async () => {
      const mockPreview: CleanPreview = {
        toolId: 'npm',
        toolName: 'npm',
        paths: [
          {
            path: '~/.npm/cache',
            files: [],
            size: 500000,
            oldestFile: Date.now() - 90 * 24 * 60 * 60 * 1000,
            newestFile: Date.now() - 30 * 24 * 60 * 60 * 1000,
          },
        ],
        totalSize: 500000,
        riskLevel: 'moderate',
        recommendations: ['Consider using time-based cleanup'],
      }
      mockInvoke.mockResolvedValue(mockPreview)

      const result = await getCleanPreview('npm', ['~/.npm'], 'time', 30, 1024)

      expect(mockInvoke).toHaveBeenCalledWith('get_clean_preview', {
        toolId: 'npm',
        paths: ['~/.npm'],
        strategy: 'time',
        timeThreshold: 30,
        sizeThreshold: 1024,
      })
      expect(result).toEqual(mockPreview)
      expect(result.riskLevel).toBe('moderate')
    })

    it('handles different strategy types', async () => {
      const strategies: CleanStrategy[] = ['time', 'size', 'selective', 'safe', 'deep']
      const mockPreview: CleanPreview = {
        toolId: 'npm',
        toolName: 'npm',
        paths: [],
        totalSize: 0,
        riskLevel: 'safe',
        recommendations: [],
      }

      for (const strategy of strategies) {
        mockInvoke.mockResolvedValue({ ...mockPreview, strategy })
        const result = await getCleanPreview('npm', [], strategy)
        expect(result).toBeDefined()
      }
    })
  })

  // ============== scanProjects Tests ==============
  describe('scanProjects', () => {
    it('calls invoke with basic scan paths', async () => {
      const mockResults: ProjectScanResult[] = [
        {
          name: 'my-project',
          path: '/home/user/projects/my-project',
          projectType: 'node',
          size: 102400000,
          fileNum: 500,
          lastModified: Date.now(),
          cleanableItems: [
            {
              id: 'node_modules',
              name: 'node_modules',
              path: '/home/user/projects/my-project/node_modules',
              itemType: 'node_modules',
              size: 90000000,
              fileNum: 400,
              lastModified: Date.now() - 60 * 24 * 60 * 60 * 1000,
              cleanable: true,
              reason: 'Dependencies can be reinstalled',
            },
          ],
          riskLevel: 'safe',
        },
      ]
      mockInvoke.mockResolvedValue(mockResults)

      const result = await scanProjects('/home/user/projects')

      expect(mockInvoke).toHaveBeenCalledWith('scan_projects', {
        basePath: '/home/user/projects',
        maxDepth: undefined,
      })
      expect(result).toEqual(mockResults)
      expect(result[0].cleanableItems).toHaveLength(1)
    })

    it('calls invoke with maxDepth parameter', async () => {
      const mockResults: ProjectScanResult[] = []
      mockInvoke.mockResolvedValue(mockResults)

      await scanProjects('/home/user/projects', 3)

      expect(mockInvoke).toHaveBeenCalledWith('scan_projects', {
        basePath: '/home/user/projects',
        maxDepth: 3,
      })
    })

    it('handles multiple scan paths', async () => {
      const mockResults: ProjectScanResult[] = []
      mockInvoke.mockResolvedValue(mockResults)

      await scanProjects('/home/user/projects')

      expect(mockInvoke).toHaveBeenCalledWith('scan_projects', {
        basePath: '/home/user/projects',
        maxDepth: undefined,
      })
    })

    it('handles empty results', async () => {
      const mockResults: ProjectScanResult[] = []
      mockInvoke.mockResolvedValue(mockResults)

      const result = await scanProjects(['/empty/path'])

      expect(result).toEqual([])
      expect(mockInvoke).toHaveBeenCalled()
    })
  })

  // ============== cleanPaths Tests ==============
  describe('cleanPaths', () => {
    it('calls invoke with paths array', async () => {
      const mockResult = {
        tool_id: 'custom',
        cleaned: 2048000,
        failed: [],
        file_num: 100,
      }
      mockInvoke.mockResolvedValue(mockResult)

      const result = await cleanPaths(['/path/to/clean1', '/path/to/clean2'])

      expect(mockInvoke).toHaveBeenCalledWith('clean_paths', {
        paths: ['/path/to/clean1', '/path/to/clean2'],
      })
      expect(result).toEqual(mockResult)
    })

    it('handles partial cleanup with failures', async () => {
      const mockResult = {
        tool_id: 'custom',
        cleaned: 1024000,
        failed: ['/path/to/clean2: Permission denied'],
        file_num: 50,
      }
      mockInvoke.mockResolvedValue(mockResult)

      const result = await cleanPaths(['/path/to/clean1', '/path/to/clean2'])

      expect(result.cleaned).toBe(1024000)
      expect(result.failed.length).toBe(1)
      expect(result.file_num).toBe(50)
    })

    it('handles empty paths array', async () => {
      const mockResult = {
        tool_id: 'custom',
        cleaned: 0,
        failed: [],
        file_num: 0,
      }
      mockInvoke.mockResolvedValue(mockResult)

      const result = await cleanPaths([])

      expect(result.cleaned).toBe(0)
      expect(result.file_num).toBe(0)
    })
  })

  // ============== getCleanHistory Tests ==============
  describe('getCleanHistory', () => {
    it('calls invoke without filter (defaults to all)', async () => {
      const mockHistory: CleanHistory = {
        items: [
          {
            id: '1',
            toolId: 'npm',
            toolName: 'npm',
            size: 1024000,
            fileNum: 100,
            timestamp: Date.now(),
            paths: ['~/.npm'],
          },
        ],
        totalCleaned: 1024000,
        totalCount: 1,
        monthlyStats: [
          { month: '2026-05', cleaned: 1024000, count: 1 },
        ],
      }
      mockInvoke.mockResolvedValue(mockHistory)

      const result = await getCleanHistory()

      expect(mockInvoke).toHaveBeenCalledWith('get_clean_history', { filter: undefined })
      expect(result).toEqual(mockHistory)
    })

    it('calls invoke with day filter', async () => {
      const mockHistory: CleanHistory = {
        items: [],
        totalCleaned: 0,
        totalCount: 0,
        monthlyStats: [],
      }
      mockInvoke.mockResolvedValue(mockHistory)

      await getCleanHistory('day')

      expect(mockInvoke).toHaveBeenCalledWith('get_clean_history', { filter: 'day' })
    })

    it('calls invoke with week filter', async () => {
      mockInvoke.mockResolvedValue({
        items: [],
        totalCleaned: 0,
        totalCount: 0,
        monthlyStats: [],
      })

      await getCleanHistory('week')

      expect(mockInvoke).toHaveBeenCalledWith('get_clean_history', { filter: 'week' })
    })

    it('calls invoke with month filter', async () => {
      mockInvoke.mockResolvedValue({
        items: [],
        totalCleaned: 0,
        totalCount: 0,
        monthlyStats: [],
      })

      await getCleanHistory('month')

      expect(mockInvoke).toHaveBeenCalledWith('get_clean_history', { filter: 'month' })
    })

    it('handles history with multiple items', async () => {
      const mockHistory: CleanHistory = {
        items: [
          { id: '1', toolId: 'npm', toolName: 'npm', size: 1024000, fileNum: 100, timestamp: Date.now(), paths: [] },
          { id: '2', toolId: 'yarn', toolName: 'yarn', size: 2048000, fileNum: 200, timestamp: Date.now() - 86400000, paths: [] },
          { id: '3', toolId: 'docker', toolName: 'docker', size: 5120000, fileNum: 50, timestamp: Date.now() - 172800000, paths: [] },
        ],
        totalCleaned: 8192000,
        totalCount: 3,
        monthlyStats: [
          { month: '2026-05', cleaned: 8192000, count: 3 },
        ],
      }
      mockInvoke.mockResolvedValue(mockHistory)

      const result = await getCleanHistory()

      expect(result.items).toHaveLength(3)
      expect(result.totalCleaned).toBe(8192000)
      expect(result.totalCount).toBe(3)
    })
  })

  // ============== recordCleanHistory Tests ==============
  describe('recordCleanHistory', () => {
    it('calls invoke with all parameters', async () => {
      mockInvoke.mockResolvedValue(undefined)

      await recordCleanHistory('npm', 'npm', 1024000, 100, ['~/.npm'])

      expect(mockInvoke).toHaveBeenCalledWith('record_clean_history', {
        toolId: 'npm',
        toolName: 'npm',
        size: 1024000,
        fileNum: 100,
        paths: ['~/.npm'],
      })
    })

    it('handles multiple paths', async () => {
      mockInvoke.mockResolvedValue(undefined)

      await recordCleanHistory('npm', 'npm', 2048000, 200, ['~/.npm', '~/.npm/_cacache'])

      expect(mockInvoke).toHaveBeenCalledWith('record_clean_history', {
        toolId: 'npm',
        toolName: 'npm',
        size: 2048000,
        fileNum: 200,
        paths: ['~/.npm', '~/.npm/_cacache'],
      })
    })

    it('returns void on success', async () => {
      mockInvoke.mockResolvedValue(undefined)

      const result = await recordCleanHistory('npm', 'npm', 1024, 10, [])

      expect(result).toBeUndefined()
    })
  })

  // ============== exportCleanReport Tests ==============
  describe('exportCleanReport', () => {
    it('calls invoke with json format', async () => {
      const mockReport = JSON.stringify({
        exportDate: new Date().toISOString(),
        totalCleaned: 1024000,
        items: [],
      })
      mockInvoke.mockResolvedValue(mockReport)

      const result = await exportCleanReport('json')

      expect(mockInvoke).toHaveBeenCalledWith('export_clean_report', { format: 'json' })
      expect(result).toBe(mockReport)
    })

    it('calls invoke with csv format', async () => {
      const mockReport = 'Tool,Size,Files,Timestamp\nnpm,1024000,100,2026-05-06'
      mockInvoke.mockResolvedValue(mockReport)

      const result = await exportCleanReport('csv')

      expect(mockInvoke).toHaveBeenCalledWith('export_clean_report', { format: 'csv' })
      expect(result).toBe(mockReport)
    })

    it('handles empty report', async () => {
      const mockReport = JSON.stringify({
        exportDate: new Date().toISOString(),
        totalCleaned: 0,
        items: [],
      })
      mockInvoke.mockResolvedValue(mockReport)

      const result = await exportCleanReport('json')

      expect(result).toBeDefined()
    })
  })
})

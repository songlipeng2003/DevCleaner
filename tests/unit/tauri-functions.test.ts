import { describe, it, expect, vi, beforeEach } from 'vitest'

// Use vi.hoisted to define mocks before module imports
const { mockInvoke, mockListen } = vi.hoisted(() => ({
  mockInvoke: vi.fn(),
  mockListen: vi.fn(),
}))

vi.mock('@tauri-apps/api/core', () => ({
  invoke: mockInvoke,
}))

vi.mock('@tauri-apps/api/event', () => ({
  listen: mockListen,
}))

import {
  getToolList,
  getToolInfo,
  scanTool,
  scanAllTools,
  cleanTool,
  getSettings,
  saveSettings,
  getDiskUsage,
  openPath,
  getVersion,
  previewTool,
  getUsageStats,
  recordClean,
  onScanComplete,
} from '@/services/tauri'

describe('Tauri Service Functions', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getToolList', () => {
    it('calls invoke with get_tool_list', async () => {
      const mockTools = [{ id: 'npm', name: 'npm', paths: [], enabled: true }]
      mockInvoke.mockResolvedValue(mockTools)

      const result = await getToolList()

      expect(mockInvoke).toHaveBeenCalledWith('get_tool_list')
      expect(result).toEqual(mockTools)
    })
  })

  describe('getToolInfo', () => {
    it('calls invoke with get_tool_info and toolId', async () => {
      const mockTool = { id: 'npm', name: 'npm', paths: [], enabled: true }
      mockInvoke.mockResolvedValue(mockTool)

      const result = await getToolInfo('npm')

      expect(mockInvoke).toHaveBeenCalledWith('get_tool_info', { toolId: 'npm' })
      expect(result).toEqual(mockTool)
    })
  })

  describe('scanTool', () => {
    it('calls invoke with scan_tool and toolId', async () => {
      const mockResults = [{ tool_id: 'npm', path: '~/.npm', size: 1024, file_num: 5, last_modified: 0 }]
      mockInvoke.mockResolvedValue(mockResults)

      const result = await scanTool('npm')

      expect(mockInvoke).toHaveBeenCalledWith('scan_tool', { toolId: 'npm' })
      expect(result).toEqual(mockResults)
    })
  })

  describe('scanAllTools', () => {
    it('calls invoke with scan_all_tools when no callback provided', async () => {
      const mockResults = [{ tool_id: 'npm', path: '~/.npm', size: 1024, file_num: 5, last_modified: 0 }]
      mockInvoke.mockResolvedValue(mockResults)

      const result = await scanAllTools()

      expect(mockInvoke).toHaveBeenCalledWith('scan_all_tools')
      expect(result).toEqual(mockResults)
    })

    it('sets up progress listener when callback provided', async () => {
      const mockResults: never[] = []
      const mockUnlisten = vi.fn()
      mockInvoke.mockResolvedValue(mockResults)
      mockListen.mockResolvedValue(mockUnlisten)

      const progressCallback = vi.fn()
      const result = await scanAllTools(progressCallback)

      expect(mockListen).toHaveBeenCalledWith('scan-progress', expect.any(Function))
      expect(result).toEqual(mockResults)
    })

    it('calls unlisten after scan completes', async () => {
      const mockUnlisten = vi.fn()
      mockInvoke.mockResolvedValue([])
      mockListen.mockResolvedValue(mockUnlisten)

      await scanAllTools(vi.fn())

      expect(mockUnlisten).toHaveBeenCalled()
    })

    it('calls unlisten even when invoke throws', async () => {
      const mockUnlisten = vi.fn()
      mockInvoke.mockRejectedValue(new Error('scan error'))
      mockListen.mockResolvedValue(mockUnlisten)

      await expect(scanAllTools(vi.fn())).rejects.toThrow('scan error')
      expect(mockUnlisten).toHaveBeenCalled()
    })

    it('does not set up listener when no callback provided', async () => {
      mockInvoke.mockResolvedValue([])

      await scanAllTools()

      expect(mockListen).not.toHaveBeenCalled()
    })
  })

  describe('onScanComplete', () => {
    it('listens for scan-complete event', async () => {
      const mockUnlisten = vi.fn()
      mockListen.mockResolvedValue(mockUnlisten)

      const callback = vi.fn()
      const unlisten = await onScanComplete(callback)

      expect(mockListen).toHaveBeenCalledWith('scan-complete', expect.any(Function))
      expect(unlisten).toBe(mockUnlisten)
    })
  })

  describe('cleanTool', () => {
    it('calls invoke with clean_tool, toolId, and paths', async () => {
      const mockResult = { tool_id: 'npm', cleaned: 1024, failed: [], file_num: 5 }
      mockInvoke.mockResolvedValue(mockResult)

      const result = await cleanTool('npm', ['~/.npm'])

      expect(mockInvoke).toHaveBeenCalledWith('clean_tool', { toolId: 'npm', paths: ['~/.npm'] })
      expect(result).toEqual(mockResult)
    })
  })

  describe('getSettings', () => {
    it('calls invoke with get_settings', async () => {
      const mockSettings = {
        threshold: 100,
        whitelist: [],
        autoScan: false,
        scanInterval: 7,
        theme: 'auto',
      }
      mockInvoke.mockResolvedValue(mockSettings)

      const result = await getSettings()

      expect(mockInvoke).toHaveBeenCalledWith('get_settings')
      expect(result).toEqual(mockSettings)
    })
  })

  describe('saveSettings', () => {
    it('calls invoke with save_settings and settings object', async () => {
      mockInvoke.mockResolvedValue(undefined)

      const settingsToSave = {
        threshold: 200,
        whitelist: ['/path'],
        autoScan: true,
        scanInterval: 14,
        theme: 'dark' as const,
      }

      await saveSettings(settingsToSave)

      expect(mockInvoke).toHaveBeenCalledWith('save_settings', { settings: settingsToSave })
    })
  })

  describe('getDiskUsage', () => {
    it('calls invoke with get_disk_usage', async () => {
      const mockUsage = { total: 1000, used: 500, free: 500 }
      mockInvoke.mockResolvedValue(mockUsage)

      const result = await getDiskUsage()

      expect(mockInvoke).toHaveBeenCalledWith('get_disk_usage')
      expect(result).toEqual(mockUsage)
    })
  })

  describe('openPath', () => {
    it('calls invoke with open_path and path', async () => {
      mockInvoke.mockResolvedValue(undefined)

      await openPath('/home/user/.npm')

      expect(mockInvoke).toHaveBeenCalledWith('open_path', { path: '/home/user/.npm' })
    })
  })

  describe('getVersion', () => {
    it('calls invoke with get_version', async () => {
      const mockVersion = { version: '1.0.0', build: '2024' }
      mockInvoke.mockResolvedValue(mockVersion)

      const result = await getVersion()

      expect(mockInvoke).toHaveBeenCalledWith('get_version')
      expect(result).toEqual(mockVersion)
    })
  })

  describe('previewTool', () => {
    it('calls invoke with preview_tool, toolId, and paths', async () => {
      const mockItems = [{ path: '~/.npm', size: 1024, fileNum: 5, lastModified: 0 }]
      mockInvoke.mockResolvedValue(mockItems)

      const result = await previewTool('npm', ['~/.npm'])

      expect(mockInvoke).toHaveBeenCalledWith('preview_tool', { toolId: 'npm', paths: ['~/.npm'] })
      expect(result).toEqual(mockItems)
    })
  })

  describe('getUsageStats', () => {
    it('calls invoke with get_usage_stats', async () => {
      const mockStats = {
        totalCleaned: 5000,
        cleanCount: 3,
        lastClean: Date.now(),
        cleanHistory: [],
      }
      mockInvoke.mockResolvedValue(mockStats)

      const result = await getUsageStats()

      expect(mockInvoke).toHaveBeenCalledWith('get_usage_stats')
      expect(result).toEqual(mockStats)
    })
  })

  describe('recordClean', () => {
    it('calls invoke with record_clean and all parameters', async () => {
      mockInvoke.mockResolvedValue(undefined)

      await recordClean('npm', 'npm', 2048, 10)

      expect(mockInvoke).toHaveBeenCalledWith('record_clean', {
        toolId: 'npm',
        toolName: 'npm',
        size: 2048,
        fileNum: 10,
      })
    })
  })
})

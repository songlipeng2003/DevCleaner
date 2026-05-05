import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useToolStore } from '@/stores/tools'
import { useSettingsStore } from '@/stores/settings'
import { createPinia, setActivePinia } from 'pinia'

// Mock Tauri service
vi.mock('@/services/tauri', () => ({
  getToolList: vi.fn(),
  scanTool: vi.fn(),
  scanAllTools: vi.fn(),
  cleanTool: vi.fn(),
  getSettings: vi.fn(),
  saveSettings: vi.fn(),
  openPath: vi.fn(),
  getDiskUsage: vi.fn(),
  getUsageStats: vi.fn(),
  recordClean: vi.fn(),
  previewTool: vi.fn(),
  getToolInfo: vi.fn(),
  onScanComplete: vi.fn(),
  getVersion: vi.fn(),
}))

import * as tauriService from '@/services/tauri'

describe('Tool Store Async Actions', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  describe('fetchTools', () => {
    it('fetches tools and updates state on success', async () => {
      const mockTools = [
        { id: 'npm', name: 'npm', paths: ['~/.npm'], enabled: true },
        { id: 'yarn', name: 'yarn', paths: ['~/.yarn'], enabled: true },
      ]
      vi.mocked(tauriService.getToolList).mockResolvedValue(mockTools)

      const store = useToolStore()
      expect(store.isLoading).toBe(false)

      const promise = store.fetchTools()
      expect(store.isLoading).toBe(true)

      await promise

      expect(store.isLoading).toBe(false)
      expect(store.tools).toEqual(mockTools)
      expect(store.error).toBe(null)
    })

    it('sets error message on failure', async () => {
      vi.mocked(tauriService.getToolList).mockRejectedValue(new Error('Network error'))

      const store = useToolStore()
      await store.fetchTools()

      expect(store.isLoading).toBe(false)
      expect(store.error).toBe('Network error')
      expect(store.tools).toEqual([])
    })

    it('sets fallback error message for non-Error rejection', async () => {
      vi.mocked(tauriService.getToolList).mockRejectedValue('unknown error')

      const store = useToolStore()
      await store.fetchTools()

      expect(store.error).toBe('获取工具列表失败')
    })

    it('clears previous error on new fetch attempt', async () => {
      vi.mocked(tauriService.getToolList)
        .mockRejectedValueOnce(new Error('first error'))
        .mockResolvedValueOnce([])

      const store = useToolStore()
      await store.fetchTools()
      expect(store.error).toBe('first error')

      await store.fetchTools()
      expect(store.error).toBe(null)
    })
  })

  describe('scanTool', () => {
    it('scans a single tool and updates scan results', async () => {
      const mockResults = [
        { tool_id: 'npm', path: '~/.npm', size: 1024, file_num: 5, last_modified: Date.now() },
      ]
      vi.mocked(tauriService.scanTool).mockResolvedValue(mockResults)

      const store = useToolStore()
      store.scanResults = [
        { tool_id: 'yarn', path: '~/.yarn', size: 2048, file_num: 3, last_modified: Date.now() },
      ]

      const results = await store.scanTool('npm')

      expect(results).toEqual(mockResults)
      expect(store.scanResults).toHaveLength(2)
      expect(store.scanResults.find(r => r.tool_id === 'npm')).toBeDefined()
      expect(store.scanResults.find(r => r.tool_id === 'yarn')).toBeDefined()
    })

    it('replaces previous results for the same tool', async () => {
      const oldResults = [
        { tool_id: 'npm', path: '~/.npm/old', size: 500, file_num: 2, last_modified: Date.now() },
      ]
      const newResults = [
        { tool_id: 'npm', path: '~/.npm/new', size: 1024, file_num: 5, last_modified: Date.now() },
      ]

      vi.mocked(tauriService.scanTool).mockResolvedValue(newResults)

      const store = useToolStore()
      store.scanResults = oldResults

      await store.scanTool('npm')

      const npmResults = store.scanResults.filter(r => r.tool_id === 'npm')
      expect(npmResults).toHaveLength(1)
      expect(npmResults[0].path).toBe('~/.npm/new')
    })
  })

  describe('scanAllTools', () => {
    it('scans all tools and returns results', async () => {
      const mockResults = [
        { tool_id: 'npm', path: '~/.npm', size: 1024, file_num: 5, last_modified: Date.now() },
        { tool_id: 'yarn', path: '~/.yarn', size: 2048, file_num: 3, last_modified: Date.now() },
      ]
      vi.mocked(tauriService.scanAllTools).mockResolvedValue(mockResults)

      const store = useToolStore()
      expect(store.isScanning).toBe(false)

      const promise = store.scanAllTools()
      expect(store.isScanning).toBe(true)

      const results = await promise

      expect(results).toEqual(mockResults)
      expect(store.scanResults).toEqual(mockResults)
      expect(store.isScanning).toBe(false)
      expect(store.error).toBe(null)
    })

    it('passes progress callback to tauri service', async () => {
      vi.mocked(tauriService.scanAllTools).mockResolvedValue([])

      const store = useToolStore()
      const progressCallback = vi.fn()

      await store.scanAllTools(progressCallback)

      expect(tauriService.scanAllTools).toHaveBeenCalledWith(progressCallback)
    })

    it('sets error and rethrows on failure', async () => {
      vi.mocked(tauriService.scanAllTools).mockRejectedValue(new Error('Scan failed'))

      const store = useToolStore()

      await expect(store.scanAllTools()).rejects.toThrow('Scan failed')

      expect(store.isScanning).toBe(false)
      expect(store.error).toBe('Scan failed')
    })

    it('sets fallback error for non-Error rejection', async () => {
      vi.mocked(tauriService.scanAllTools).mockRejectedValue('some error')

      const store = useToolStore()

      await expect(store.scanAllTools()).rejects.toBe('some error')
      expect(store.error).toBe('扫描失败')
    })

    it('resets isScanning even on error', async () => {
      vi.mocked(tauriService.scanAllTools).mockRejectedValue(new Error('err'))

      const store = useToolStore()
      await expect(store.scanAllTools()).rejects.toThrow()

      expect(store.isScanning).toBe(false)
    })
  })

  describe('cleanTool', () => {
    it('cleans tool and removes results from scan list', async () => {
      const mockCleanResult = { tool_id: 'npm', cleaned: 2048, failed: [], file_num: 5 }
      vi.mocked(tauriService.cleanTool).mockResolvedValue(mockCleanResult)
      vi.mocked(tauriService.recordClean).mockResolvedValue(undefined)

      const store = useToolStore()
      store.tools = [{ id: 'npm', name: 'npm', paths: ['~/.npm'], enabled: true }]
      store.scanResults = [
        { tool_id: 'npm', path: '~/.npm', size: 2048, file_num: 5, last_modified: Date.now() },
        { tool_id: 'yarn', path: '~/.yarn', size: 1024, file_num: 3, last_modified: Date.now() },
      ]

      const result = await store.cleanTool('npm', ['~/.npm'])

      expect(result).toEqual(mockCleanResult)
      expect(store.scanResults.find(r => r.tool_id === 'npm')).toBeUndefined()
      expect(store.scanResults.find(r => r.tool_id === 'yarn')).toBeDefined()
    })

    it('calls recordClean when cleaned size is greater than 0', async () => {
      vi.mocked(tauriService.cleanTool).mockResolvedValue({
        tool_id: 'npm', cleaned: 1024, failed: [], file_num: 3
      })
      vi.mocked(tauriService.recordClean).mockResolvedValue(undefined)

      const store = useToolStore()
      store.tools = [{ id: 'npm', name: 'npm', paths: ['~/.npm'], enabled: true }]

      await store.cleanTool('npm', ['~/.npm'])

      expect(tauriService.recordClean).toHaveBeenCalledWith('npm', 'npm', 1024, 3)
    })

    it('does not call recordClean when cleaned is 0', async () => {
      vi.mocked(tauriService.cleanTool).mockResolvedValue({
        tool_id: 'npm', cleaned: 0, failed: [], file_num: 0
      })
      vi.mocked(tauriService.recordClean).mockResolvedValue(undefined)

      const store = useToolStore()
      store.tools = [{ id: 'npm', name: 'npm', paths: ['~/.npm'], enabled: true }]

      await store.cleanTool('npm', ['~/.npm'])

      expect(tauriService.recordClean).not.toHaveBeenCalled()
    })

    it('does not call recordClean when tool is not found', async () => {
      vi.mocked(tauriService.cleanTool).mockResolvedValue({
        tool_id: 'unknown', cleaned: 1024, failed: [], file_num: 1
      })
      vi.mocked(tauriService.recordClean).mockResolvedValue(undefined)

      const store = useToolStore()
      store.tools = []

      await store.cleanTool('unknown', [])

      expect(tauriService.recordClean).not.toHaveBeenCalled()
    })

    it('sets error and rethrows on failure', async () => {
      vi.mocked(tauriService.cleanTool).mockRejectedValue(new Error('Clean error'))

      const store = useToolStore()
      store.tools = []

      await expect(store.cleanTool('npm', [])).rejects.toThrow('Clean error')
      expect(store.error).toBe('Clean error')
    })

    it('sets fallback error for non-Error rejection on clean', async () => {
      vi.mocked(tauriService.cleanTool).mockRejectedValue('oops')

      const store = useToolStore()
      store.tools = []

      await expect(store.cleanTool('npm', [])).rejects.toBe('oops')
      expect(store.error).toBe('清理失败')
    })
  })

  describe('openPath', () => {
    it('delegates to tauri service', async () => {
      vi.mocked(tauriService.openPath).mockResolvedValue(undefined)

      const store = useToolStore()
      await store.openPath('/some/path')

      expect(tauriService.openPath).toHaveBeenCalledWith('/some/path')
    })
  })

  describe('updateToolEnabled', () => {
    it('updates tool enabled state (alias for toggleTool)', () => {
      const store = useToolStore()
      store.tools = [{ id: 'npm', name: 'npm', paths: [], enabled: true }]

      store.updateToolEnabled('npm', false)
      expect(store.tools[0].enabled).toBe(false)

      store.updateToolEnabled('npm', true)
      expect(store.tools[0].enabled).toBe(true)
    })

    it('does nothing for unknown tool id', () => {
      const store = useToolStore()
      store.tools = [{ id: 'npm', name: 'npm', paths: [], enabled: true }]

      store.updateToolEnabled('nonexistent', false)
      expect(store.tools[0].enabled).toBe(true)
    })
  })
})

describe('Settings Store Async Actions', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  describe('fetchSettings', () => {
    it('fetches settings and updates state on success', async () => {
      const mockSettings = {
        threshold: 200,
        whitelist: ['/some/path'],
        autoScan: true,
        scanInterval: 14,
        theme: 'dark' as const,
      }
      vi.mocked(tauriService.getSettings).mockResolvedValue(mockSettings)

      const store = useSettingsStore()
      expect(store.isLoading).toBe(false)

      const promise = store.fetchSettings()
      expect(store.isLoading).toBe(true)

      await promise

      expect(store.isLoading).toBe(false)
      expect(store.settings).toEqual(mockSettings)
      expect(store.error).toBe(null)
    })

    it('falls back to defaults and sets error on failure', async () => {
      vi.mocked(tauriService.getSettings).mockRejectedValue(new Error('Fetch failed'))

      const store = useSettingsStore()
      store.settings.threshold = 999

      await store.fetchSettings()

      expect(store.isLoading).toBe(false)
      expect(store.error).toBe('Fetch failed')
      expect(store.settings.threshold).toBe(100)
      expect(store.settings.whitelist).toEqual([])
    })

    it('sets fallback error message for non-Error rejection', async () => {
      vi.mocked(tauriService.getSettings).mockRejectedValue('connection lost')

      const store = useSettingsStore()
      await store.fetchSettings()

      expect(store.error).toBe('获取设置失败')
    })

    it('resets isLoading even on error', async () => {
      vi.mocked(tauriService.getSettings).mockRejectedValue(new Error('err'))

      const store = useSettingsStore()
      await store.fetchSettings()

      expect(store.isLoading).toBe(false)
    })
  })

  describe('saveSettings', () => {
    it('saves partial settings and merges with current state', async () => {
      vi.mocked(tauriService.saveSettings).mockResolvedValue(undefined)

      const store = useSettingsStore()
      store.settings.threshold = 100

      expect(store.isSaving).toBe(false)

      const promise = store.saveSettings({ threshold: 250 })
      expect(store.isSaving).toBe(true)

      await promise

      expect(store.isSaving).toBe(false)
      expect(store.settings.threshold).toBe(250)
      expect(store.error).toBe(null)
    })

    it('calls tauri service with merged settings', async () => {
      vi.mocked(tauriService.saveSettings).mockResolvedValue(undefined)

      const store = useSettingsStore()
      store.settings = {
        threshold: 100,
        whitelist: [],
        autoScan: false,
        scanInterval: 7,
        theme: 'auto',
      }

      await store.saveSettings({ autoScan: true, scanInterval: 14 })

      expect(tauriService.saveSettings).toHaveBeenCalledWith({
        threshold: 100,
        whitelist: [],
        autoScan: true,
        scanInterval: 14,
        theme: 'auto',
      })
    })

    it('sets error and rethrows on failure', async () => {
      vi.mocked(tauriService.saveSettings).mockRejectedValue(new Error('Save failed'))

      const store = useSettingsStore()

      await expect(store.saveSettings({})).rejects.toThrow('Save failed')
      expect(store.error).toBe('Save failed')
    })

    it('sets fallback error for non-Error rejection', async () => {
      vi.mocked(tauriService.saveSettings).mockRejectedValue('disk full')

      const store = useSettingsStore()

      await expect(store.saveSettings({})).rejects.toBe('disk full')
      expect(store.error).toBe('保存设置失败')
    })

    it('resets isSaving even on error', async () => {
      vi.mocked(tauriService.saveSettings).mockRejectedValue(new Error('err'))

      const store = useSettingsStore()
      await expect(store.saveSettings({})).rejects.toThrow()

      expect(store.isSaving).toBe(false)
    })
  })
})

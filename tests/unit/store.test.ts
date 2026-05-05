import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useToolStore } from '@/stores/tools'
import { useSettingsStore } from '@/stores/settings'
import { createPinia, setActivePinia } from 'pinia'
import type { ScanResult, ToolInfo } from '@/types'

// Mock Tauri API
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
  previewTool: vi.fn()
}))

describe('Tool Store Tests', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  describe('Tool Store State Management', () => {
    it('initializes with empty state', () => {
      const store = useToolStore()
      expect(store.tools).toEqual([])
      expect(store.scanResults).toEqual([])
      expect(store.isScanning).toBe(false)
      expect(store.isLoading).toBe(false)
      expect(store.error).toBe(null)
    })

    it('computes enabled tools correctly', () => {
      const store = useToolStore()
      store.tools = [
        { id: 'npm', name: 'npm', paths: [], enabled: true },
        { id: 'yarn', name: 'yarn', paths: [], enabled: false },
        { id: 'pnpm', name: 'pnpm', paths: [], enabled: true }
      ]
      expect(store.enabledTools).toHaveLength(2)
      expect(store.enabledTools.every(t => t.enabled)).toBe(true)
    })

    it('computes total cache size correctly', () => {
      const store = useToolStore()
      store.scanResults = [
        { tool_id: 'npm', path: '/path1', size: 1024, file_num: 1, last_modified: Date.now() },
        { tool_id: 'yarn', path: '/path2', size: 2048, file_num: 2, last_modified: Date.now() }
      ]
      expect(store.totalCacheSize).toBe(3072)
    })
  })

  describe('Tool Store Actions', () => {
    it('formats file size correctly', () => {
      const store = useToolStore()
      expect(store.formatSize(0)).toBe('0 B')
      expect(store.formatSize(1024)).toBe('1 KB')
      expect(store.formatSize(1048576)).toBe('1 MB')
      expect(store.formatSize(1073741824)).toBe('1 GB')
      expect(store.formatSize(1099511627776)).toBe('1 TB')
    })

    it('toggles tool enabled state', () => {
      const store = useToolStore()
      store.tools = [
        { id: 'npm', name: 'npm', paths: [], enabled: true }
      ]
      store.toggleTool('npm', false)
      expect(store.tools[0].enabled).toBe(false)
      store.toggleTool('npm', true)
      expect(store.tools[0].enabled).toBe(true)
    })

    it('gets tool results correctly', () => {
      const store = useToolStore()
      store.scanResults = [
        { tool_id: 'npm', path: '/path1', size: 1024, file_num: 1, last_modified: Date.now() },
        { tool_id: 'yarn', path: '/path2', size: 2048, file_num: 2, last_modified: Date.now() },
        { tool_id: 'npm', path: '/path3', size: 512, file_num: 3, last_modified: Date.now() }
      ]
      const npmResults = store.getToolResults('npm')
      expect(npmResults).toHaveLength(2)
      expect(npmResults.every(r => r.tool_id === 'npm')).toBe(true)
    })

    it('gets tool size correctly', () => {
      const store = useToolStore()
      store.scanResults = [
        { tool_id: 'npm', path: '/path1', size: 1024, file_num: 1, last_modified: Date.now() },
        { tool_id: 'npm', path: '/path2', size: 2048, file_num: 2, last_modified: Date.now() }
      ]
      expect(store.getToolSize('npm')).toBe(3072)
      expect(store.getToolSize('yarn')).toBe(0)
    })
  })
})

describe('Settings Store Tests', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  describe('Settings Store State Management', () => {
    it('initializes with default settings', () => {
      const store = useSettingsStore()
      expect(store.settings.threshold).toBe(100)
      expect(store.settings.whitelist).toEqual([])
      expect(store.settings.autoScan).toBe(false)
      expect(store.settings.scanInterval).toBe(7)
      expect(store.settings.theme).toBe('auto')
    })

    it('adds whitelist path correctly', () => {
      const store = useSettingsStore()
      store.addWhitelist('/important/path')
      expect(store.settings.whitelist).toContain('/important/path')
    })

    it('does not add duplicate whitelist paths', () => {
      const store = useSettingsStore()
      store.addWhitelist('/important/path')
      store.addWhitelist('/important/path')
      expect(store.settings.whitelist).toHaveLength(1)
    })

    it('removes whitelist path correctly', () => {
      const store = useSettingsStore()
      store.addWhitelist('/path1')
      store.addWhitelist('/path2')
      store.removeWhitelist('/path1')
      expect(store.settings.whitelist).not.toContain('/path1')
      expect(store.settings.whitelist).toContain('/path2')
    })

    it('resets settings to defaults', () => {
      const store = useSettingsStore()
      store.settings.threshold = 200
      store.settings.whitelist = ['/test']
      store.resetSettings()
      expect(store.settings.threshold).toBe(100)
      expect(store.settings.whitelist).toEqual([])
    })
  })
})

describe('Utility Functions', () => {
  describe('File Size Formatting', () => {
    it('handles edge cases', () => {
      const formatSize = (bytes: number): string => {
        if (bytes === 0) return '0 B'
        const k = 1024
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
        const i = Math.floor(Math.log(bytes) / Math.log(k))
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
      }

      expect(formatSize(1)).toBe('1 B')
      expect(formatSize(999)).toBe('999 B')
      expect(formatSize(1000)).toBe('1000 B')
      expect(formatSize(1023)).toBe('1023 B')
      expect(formatSize(1024)).toBe('1 KB')
    })
  })

  describe('Path Validation', () => {
    it('detects path traversal attempts', () => {
      const isSafePath = (path: string): boolean => {
        return !path.includes('..')
      }

      expect(isSafePath('/home/user/.npm')).toBe(true)
      expect(isSafePath('/home/user/../../etc')).toBe(false)
      expect(isSafePath('../secret')).toBe(false)
    })

    it('validates absolute paths', () => {
      const isAbsolutePath = (path: string): boolean => {
        return path.startsWith('/') || path.startsWith('~')
      }

      expect(isAbsolutePath('/home/user/.npm')).toBe(true)
      expect(isAbsolutePath('~/.npm')).toBe(true)
      expect(isAbsolutePath('relative/path')).toBe(false)
    })
  })
})

describe('Error Handling', () => {
  it('handles missing tool gracefully', () => {
    const store = useToolStore()
    store.tools = [
      { id: 'npm', name: 'npm', paths: [], enabled: true }
    ]
    const result = store.getToolResults('nonexistent')
    expect(result).toEqual([])
  })

  it('handles empty scan results', () => {
    const store = useToolStore()
    store.tools = [] // 重置工具列表
    expect(store.totalCacheSize).toBe(0)
    expect(store.enabledTools).toHaveLength(0)
  })
})

describe('Integration Tests', () => {
  it('handles complete scan-clean workflow', async () => {
    const toolStore = useToolStore()
    const settingsStore = useSettingsStore()

    // Setup
    toolStore.tools = [
      { id: 'npm', name: 'npm', paths: ['~/.npm'], enabled: true }
    ]
    toolStore.scanResults = [
      {
        tool_id: 'npm',
        path: '~/.npm',
        size: 1048576, // 1 MB
        file_num: 100,
        last_modified: Date.now()
      }
    ]

    // Verify setup
    expect(toolStore.getToolSize('npm')).toBe(1048576)
    expect(toolStore.formatSize(1048576)).toBe('1 MB')
    expect(toolStore.enabledTools).toHaveLength(1)

    // Simulate clean
    toolStore.scanResults = []

    // Verify clean
    expect(toolStore.getToolSize('npm')).toBe(0)
  })

  it('handles whitelist functionality correctly', () => {
    const settingsStore = useSettingsStore()
    const toolStore = useToolStore()

    // Add path to whitelist
    settingsStore.addWhitelist('/important/project/.npm')

    // Verify whitelist
    expect(settingsStore.settings.whitelist).toContain('/important/project/.npm')

    // Simulate checking if path is whitelisted
    const isWhitelisted = (path: string, whitelist: string[]): boolean => {
      return whitelist.some(w => path.startsWith(w))
    }

    expect(isWhitelisted('/important/project/.npm/cache', settingsStore.settings.whitelist)).toBe(true)
    expect(isWhitelisted('/other/project/.npm/cache', settingsStore.settings.whitelist)).toBe(false)
  })
})
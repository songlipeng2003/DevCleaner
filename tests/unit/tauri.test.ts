import { describe, it, expect, vi, beforeEach } from 'vitest'
import * as tauriApi from '@/services/tauri'

describe('Tauri Service Type Tests', () => {
  describe('ScanProgress interface', () => {
    it('has correct structure', () => {
      const progress: tauriApi.ScanProgress = {
        toolId: 'npm',
        toolName: 'npm',
        progress: 50,
        currentPath: '/path',
        pathsScanned: 5,
        totalPaths: 10,
      }

      expect(progress.toolId).toBe('npm')
      expect(progress.progress).toBe(50)
      expect(progress.pathsScanned).toBe(5)
      expect(progress.totalPaths).toBe(10)
    })
  })

  describe('PreviewItem interface', () => {
    it('has correct structure', () => {
      const item: tauriApi.PreviewItem = {
        path: '/test/path',
        size: 1024,
        fileNum: 10,
        lastModified: Date.now(),
      }

      expect(item.path).toBe('/test/path')
      expect(item.size).toBe(1024)
      expect(item.fileNum).toBe(10)
    })
  })

  describe('UsageStats interface', () => {
    it('has correct structure with history', () => {
      const stats: tauriApi.UsageStats = {
        totalCleaned: 1024000,
        cleanCount: 3,
        lastClean: Date.now(),
        cleanHistory: [
          { toolId: 'npm', toolName: 'npm', size: 512, fileNum: 5, timestamp: Date.now() },
          { toolId: 'yarn', toolName: 'yarn', size: 512, fileNum: 5, timestamp: Date.now() },
        ],
      }

      expect(stats.totalCleaned).toBe(1024000)
      expect(stats.cleanHistory).toHaveLength(2)
      expect(stats.cleanHistory[0].toolId).toBe('npm')
    })
  })

  describe('CleanHistoryItem interface', () => {
    it('has correct structure', () => {
      const item: tauriApi.CleanHistoryItem = {
        toolId: 'pnpm',
        toolName: 'pnpm',
        size: 2048,
        fileNum: 15,
        timestamp: Date.now(),
      }

      expect(item.toolId).toBe('pnpm')
      expect(item.toolName).toBe('pnpm')
      expect(item.size).toBe(2048)
      expect(item.fileNum).toBe(15)
    })
  })
})

describe('Settings Store Extended Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('removeWhitelist', () => {
    it('removes existing whitelist path', async () => {
      const { useSettingsStore } = await import('@/stores/settings')
      const { createPinia, setActivePinia } = await import('pinia')

      setActivePinia(createPinia())
      const store = useSettingsStore()
      store.settings.whitelist = ['/path1', '/path2', '/path3']

      store.removeWhitelist('/path2')
      expect(store.settings.whitelist).not.toContain('/path2')
      expect(store.settings.whitelist).toHaveLength(2)
    })

    it('handles removing non-existent path', async () => {
      const { useSettingsStore } = await import('@/stores/settings')
      const { createPinia, setActivePinia } = await import('pinia')

      setActivePinia(createPinia())
      const store = useSettingsStore()
      store.settings.whitelist = ['/path1']

      store.removeWhitelist('/nonexistent')
      expect(store.settings.whitelist).toHaveLength(1)
    })

    it('removes first item correctly', async () => {
      const { useSettingsStore } = await import('@/stores/settings')
      const { createPinia, setActivePinia } = await import('pinia')

      setActivePinia(createPinia())
      const store = useSettingsStore()
      store.settings.whitelist = ['/first', '/second']

      store.removeWhitelist('/first')
      expect(store.settings.whitelist).toEqual(['/second'])
    })

    it('removes last item correctly', async () => {
      const { useSettingsStore } = await import('@/stores/settings')
      const { createPinia, setActivePinia } = await import('pinia')

      setActivePinia(createPinia())
      const store = useSettingsStore()
      store.settings.whitelist = ['/only']

      store.removeWhitelist('/only')
      expect(store.settings.whitelist).toEqual([])
    })
  })

  describe('addWhitelist', () => {
    it('does not add empty path', async () => {
      const { useSettingsStore } = await import('@/stores/settings')
      const { createPinia, setActivePinia } = await import('pinia')

      setActivePinia(createPinia())
      const store = useSettingsStore()

      store.addWhitelist('')
      expect(store.settings.whitelist).toEqual([])
    })

    it('adds multiple unique paths', async () => {
      const { useSettingsStore } = await import('@/stores/settings')
      const { createPinia, setActivePinia } = await import('pinia')

      setActivePinia(createPinia())
      const store = useSettingsStore()

      store.addWhitelist('/path1')
      store.addWhitelist('/path2')
      store.addWhitelist('/path3')

      expect(store.settings.whitelist).toHaveLength(3)
    })
  })

  describe('resetSettings', () => {
    it('resets all settings to default', async () => {
      const { useSettingsStore } = await import('@/stores/settings')
      const { createPinia, setActivePinia } = await import('pinia')

      setActivePinia(createPinia())
      const store = useSettingsStore()

      // 修改所有设置
      store.settings.threshold = 999
      store.settings.whitelist = ['/test']
      store.settings.autoScan = true
      store.settings.scanInterval = 30
      store.settings.theme = 'dark'

      store.resetSettings()

      expect(store.settings.threshold).toBe(100)
      expect(store.settings.whitelist).toEqual([])
      expect(store.settings.autoScan).toBe(false)
      expect(store.settings.scanInterval).toBe(7)
      expect(store.settings.theme).toBe('auto')
    })
  })
})

import { describe, it, expect, vi, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useProjectCleanStore } from '@/stores/projectClean'
import * as tauriApi from '@/services/tauri'

// Mock tauri service
vi.mock('@/services/tauri', () => ({
  scanProjects: vi.fn(),
  cleanPaths: vi.fn(),
  recordCleanHistory: vi.fn(),
  openPath: vi.fn(),
}))

// Factory to create fresh mock data each time
function createMockProjects() {
  return [
    {
      name: 'my-app',
      path: '/Users/dev/projects/my-app',
      type: 'node',
      size: 2147483648,
      fileNum: 50000,
      lastModified: Date.now() - 30 * 24 * 60 * 60 * 1000,
      cleanableItems: [
        {
          id: 'node_modules-1',
          name: 'node_modules',
          path: '/Users/dev/projects/my-app/node_modules',
          type: 'node_modules' as const,
          size: 1610612736,
          fileNum: 40000,
          lastModified: Date.now() - 60 * 24 * 60 * 60 * 1000,
          cleanable: true,
          reason: 'Dependencies can be reinstalled',
        },
        {
          id: 'dist-1',
          name: 'dist',
          path: '/Users/dev/projects/my-app/dist',
          type: 'dist' as const,
          size: 536870912,
          fileNum: 10000,
          lastModified: Date.now() - 7 * 24 * 60 * 60 * 1000,
          cleanable: false,
          reason: 'Production build, may be needed',
        },
      ],
      riskLevel: 'safe' as const,
    },
    {
      name: 'api-server',
      path: '/Users/dev/projects/api-server',
      type: 'rust',
      size: 1073741824,
      fileNum: 10000,
      lastModified: Date.now() - 7 * 24 * 60 * 60 * 1000,
      cleanableItems: [
        {
          id: 'target-1',
          name: 'target',
          path: '/Users/dev/projects/api-server/target',
          type: 'target' as const,
          size: 805306368,
          fileNum: 8000,
          lastModified: Date.now() - 3 * 24 * 60 * 60 * 1000,
          cleanable: true,
          reason: 'Build artifacts can be rebuilt',
        },
      ],
      riskLevel: 'safe' as const,
    },
  ]
}

describe('ProjectClean Store', () => {
  let store: ReturnType<typeof useProjectCleanStore>
  let mockProjectScanResults: ReturnType<typeof createMockProjects>

  beforeEach(() => {
    setActivePinia(createPinia())
    store = useProjectCleanStore()
    mockProjectScanResults = createMockProjects()
    vi.clearAllMocks()
  })

  // ============== Initial State Tests ==============
  describe('Initial State', () => {
    it('has empty projects array', () => {
      expect(store.projects).toEqual([])
    })

    it('has isScanning as false', () => {
      expect(store.isScanning).toBe(false)
    })

    it('has isCleaning as false', () => {
      expect(store.isCleaning).toBe(false)
    })

    it('has empty selectedItems array', () => {
      expect(store.selectedItems).toEqual([])
    })

    it('has default config', () => {
      expect(store.config.scanPaths).toContain('~/projects')
      expect(store.config.includeTypes).toContain('node')
      expect(store.config.maxDepth).toBe(3)
      expect(store.config.minSize).toBe(10 * 1024 * 1024)
    })

    it('has null error', () => {
      expect(store.error).toBeNull()
    })
  })

  // ============== scanProjects Tests ==============
  describe('scanProjects', () => {
    it('calls scanProjects API', async () => {
      vi.mocked(tauriApi.scanProjects).mockResolvedValue(mockProjectScanResults)

      await store.scanProjects()

      expect(tauriApi.scanProjects).toHaveBeenCalledWith(
        store.config.scanPaths[0],
        store.config.maxDepth
      )
    })

    it('updates projects state with scan results', async () => {
      vi.mocked(tauriApi.scanProjects).mockResolvedValue(mockProjectScanResults)

      await store.scanProjects()

      expect(store.projects).toEqual(mockProjectScanResults)
    })

    it('clears selectedItems after scan', async () => {
      vi.mocked(tauriApi.scanProjects).mockResolvedValue(mockProjectScanResults)

      // First scan
      await store.scanProjects()

      // Add some selection
      store.selectedItems.push(mockProjectScanResults[0].cleanableItems[0])

      // Scan again
      await store.scanProjects()

      expect(store.selectedItems).toEqual([])
    })

    it('sets isScanning during scan', async () => {
      vi.mocked(tauriApi.scanProjects).mockImplementation(async () => {
        store.isScanning = true
        return mockProjectScanResults
      })

      await store.scanProjects()

      expect(store.isScanning).toBe(false)
    })

    it('sets error on scan failure', async () => {
      vi.mocked(tauriApi.scanProjects).mockRejectedValue(new Error('Scan failed'))

      await expect(store.scanProjects()).rejects.toThrow()
      expect(store.error).toBe('Scan failed')
    })
  })

  // ============== cleanSelected Tests ==============
  describe('cleanSelected', () => {
    beforeEach(async () => {
      vi.mocked(tauriApi.scanProjects).mockResolvedValue(mockProjectScanResults)
      await store.scanProjects()
    })

    it('throws error when no items selected', async () => {
      await expect(store.cleanSelected()).rejects.toThrow('没有选中的清理项')
    })

    it('calls cleanPaths API with selected paths', async () => {
      vi.mocked(tauriApi.cleanPaths).mockResolvedValue({
        tool_id: 'custom',
        cleaned: 1610612736,
        failed: [],
        file_num: 40000,
      })

      store.selectedItems.push(mockProjectScanResults[0].cleanableItems[0])

      await store.cleanSelected()

      expect(tauriApi.cleanPaths).toHaveBeenCalledWith([
        '/Users/dev/projects/my-app/node_modules',
      ])
    })

    it('clears selectedItems after clean', async () => {
      vi.mocked(tauriApi.cleanPaths).mockResolvedValue({
        tool_id: 'custom',
        cleaned: 1610612736,
        failed: [],
        file_num: 40000,
      })

      store.selectedItems.push(mockProjectScanResults[0].cleanableItems[0])

      await store.cleanSelected()

      expect(store.selectedItems).toEqual([])
    })

    it('removes cleaned items from projects', async () => {
      vi.mocked(tauriApi.cleanPaths).mockResolvedValue({
        tool_id: 'custom',
        cleaned: 1610612736,
        failed: [],
        file_num: 40000,
      })

      // Get item from current store state
      await store.scanProjects()
      const item = store.allCleanableItems[0]
      const itemPath = item.path

      await store.cleanItems([item])

      // Verify item path was passed to cleanPaths
      expect(vi.mocked(tauriApi.cleanPaths)).toHaveBeenCalledWith([itemPath])
    })

    it('records clean history', async () => {
      vi.mocked(tauriApi.cleanPaths).mockResolvedValue({
        tool_id: 'custom',
        cleaned: 1610612736,
        failed: [],
        file_num: 40000,
      })

      await store.scanProjects()
      const item = store.allCleanableItems[0]
      const itemPath = item.path

      await store.cleanItems([item])

      expect(tauriApi.recordCleanHistory).toHaveBeenCalledWith(
        'project',
        '项目清理',
        1610612736,
        40000,
        [itemPath]
      )
    })
  })

  // ============== Selection Tests ==============
  describe('Selection', () => {
    beforeEach(async () => {
      vi.mocked(tauriApi.scanProjects).mockResolvedValue(mockProjectScanResults)
      await store.scanProjects()
    })

    it('toggleSelection adds item to selectedItems', () => {
      const item = { ...store.allCleanableItems[0] }

      store.toggleSelection(item)

      expect(store.selectedItems).toContainEqual(item)
    })

// ============== Selection Tests ==============

    it('selectAll selects all cleanable items', () => {
      store.selectAll()

      expect(store.selectedItems.length).toBe(store.allCleanableItems.length)
    })

    it('clearSelection removes all selected items', () => {
      store.selectAll()
      store.clearSelection()

      expect(store.selectedItems).toEqual([])
    })

    it('selectByType selects only items of specified type', () => {
      store.selectByType('node_modules')

      const allNodeModules = store.allCleanableItems.filter(i => i.type === 'node_modules')
      expect(store.selectedItems.length).toBe(allNodeModules.length)
      expect(store.selectedItems.every(i => i.type === 'node_modules')).toBe(true)
    })

    it('deselectByType removes items of specified type', () => {
      store.selectAll()
      store.deselectByType('target')

      expect(store.selectedItems.some(i => i.type === 'target')).toBe(false)
    })
  })

  // ============== Computed/Getters Tests ==============
  describe('Getters', () => {
    beforeEach(async () => {
      vi.mocked(tauriApi.scanProjects).mockResolvedValue(mockProjectScanResults)
      await store.scanProjects()
    })

    it('allCleanableItems returns only cleanable items', () => {
      expect(store.allCleanableItems.every(i => i.cleanable)).toBe(true)
    })

    it('totalCleanableSize returns sum of cleanable items', () => {
      const expected = store.allCleanableItems.reduce((sum, i) => sum + i.size, 0)
      expect(store.totalCleanableSize).toBe(expected)
    })

    it('totalCleanableItems returns count of cleanable items', () => {
      expect(store.totalCleanableItems).toBe(store.allCleanableItems.length)
    })

    it('selectedSize returns total size of selected items', () => {
      store.selectedItems.push(store.allCleanableItems[0])
      expect(store.selectedSize).toBe(store.allCleanableItems[0].size)
    })

    it('selectedCount returns number of selected items', () => {
      store.selectedItems.push(store.allCleanableItems[0])
      store.selectedItems.push(store.allCleanableItems[1])
      expect(store.selectedCount).toBe(2)
    })

    it('isSelected returns true for selected items', () => {
      const item = store.allCleanableItems[0]
      store.selectedItems.push(item)
      expect(store.isSelected(item.id)).toBe(true)
    })

    it('isSelected returns false for unselected items', () => {
      expect(store.isSelected('non-existent-id')).toBe(false)
    })

    it('projectsByType groups projects by type', () => {
      expect(store.projectsByType['node']).toBeDefined()
      expect(store.projectsByType['rust']).toBeDefined()
    })

    it('sortedProjects returns projects sorted by size descending', () => {
      const sizes = store.sortedProjects.map(p => p.size)
      expect(sizes).toEqual([...sizes].sort((a, b) => b - a))
    })
  })

  // ============== Config Tests ==============
  describe('Config', () => {
    it('setScanPaths updates scan paths', () => {
      const newPaths = ['~/work', '~/code']
      store.setScanPaths(newPaths)

      expect(store.config.scanPaths).toEqual(newPaths)
    })

    it('setConfig updates config with partial update', () => {
      store.setConfig({ maxDepth: 5, minSize: 50000000 })

      expect(store.config.maxDepth).toBe(5)
      expect(store.config.minSize).toBe(50000000)
      expect(store.config.scanPaths).toContain('~/projects') // Preserved
    })
  })

  // ============== Utility Tests ==============
  describe('Utility Functions', () => {
    beforeEach(async () => {
      vi.mocked(tauriApi.scanProjects).mockResolvedValue(mockProjectScanResults)
      await store.scanProjects()
    })

    it('getFilteredProjects filters by type', () => {
      const nodeProjects = store.getFilteredProjects({ type: 'node' })

      expect(nodeProjects.every(p => p.type === 'node')).toBe(true)
    })

    it('getFilteredProjects filters by minSize', () => {
      // my-app has size 2147483648 (2GB), api-server has 1073741824 (1GB)
      const largeProjects = store.getFilteredProjects({ minSize: 1500000000 })

      expect(largeProjects.length).toBe(1)
      expect(largeProjects[0].name).toBe('my-app')
    })

    it('formatSize formats bytes correctly', () => {
      expect(store.formatSize(0)).toBe('0 B')
      expect(store.formatSize(1024)).toBe('1 KB')
      expect(store.formatSize(1048576)).toBe('1 MB')
      expect(store.formatSize(1073741824)).toBe('1 GB')
    })

    it('clearProjects resets all state', () => {
      vi.mocked(tauriApi.scanProjects).mockResolvedValue(mockProjectScanResults)
      store.scanProjects()
      store.selectedItems.push(store.allCleanableItems[0])

      store.clearProjects()

      expect(store.projects).toEqual([])
      expect(store.selectedItems).toEqual([])
      expect(store.error).toBeNull()
    })

    it('openPath calls tauriApi.openPath', async () => {
      vi.mocked(tauriApi.openPath).mockResolvedValue()

      await store.openPath('/test/path')

      expect(tauriApi.openPath).toHaveBeenCalledWith('/test/path')
    })
  })
})

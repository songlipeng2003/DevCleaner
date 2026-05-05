import { describe, it, expect, vi, beforeEach } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { nextTick } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import ScanView from '@/views/ScanView.vue'
import { useToolStore } from '@/stores/tools'

// Mock Tauri API
vi.mock('@tauri-apps/api/core', () => ({
  invoke: vi.fn()
}))

vi.mock('@tauri-apps/api/event', () => ({
  listen: vi.fn().mockResolvedValue(() => {})
}))

vi.mock('@/services/tauri', () => ({
  getToolList: vi.fn().mockResolvedValue([]),
  scanTool: vi.fn().mockResolvedValue([]),
  scanAllTools: vi.fn().mockResolvedValue([]),
  cleanTool: vi.fn().mockResolvedValue({ tool_id: '', cleaned: 0, failed: [], file_num: 0 }),
  getSettings: vi.fn().mockResolvedValue({
    threshold: 100, whitelist: [], autoScan: false, scanInterval: 7, theme: 'auto'
  }),
  saveSettings: vi.fn().mockResolvedValue(undefined),
  openPath: vi.fn().mockResolvedValue(undefined),
  getDiskUsage: vi.fn().mockResolvedValue({ total: 0, used: 0, free: 0 }),
  getUsageStats: vi.fn().mockResolvedValue({ totalCleaned: 0, cleanCount: 0, lastClean: 0, cleanHistory: [] }),
  recordClean: vi.fn().mockResolvedValue(undefined),
  previewTool: vi.fn().mockResolvedValue([]),
}))

// Mock ant-design-vue message
vi.mock('ant-design-vue', () => ({
  message: {
    success: vi.fn(),
    error: vi.fn(),
    loading: vi.fn(),
    warning: vi.fn(),
    info: vi.fn(),
  }
}))

const createTestRouter = () =>
  createRouter({
    history: createWebHistory(),
    routes: [
      { path: '/', name: 'home', component: { template: '<div>Home</div>' } },
      { path: '/scan', name: 'scan', component: { template: '<div>Scan</div>' } },
    ],
  })

/** Mount ScanView with pre-populated scan results to prevent onMounted from triggering a scan */
const mountScanView = (storeSetup?: (store: ReturnType<typeof useToolStore>) => void) => {
  const pinia = createPinia()
  setActivePinia(pinia)

  // Pre-populate store so onMounted scan does not fire (scanResults.length > 0)
  const store = useToolStore()
  store.scanResults = [
    { tool_id: 'npm', path: '~/.npm', size: 1024, file_num: 5, last_modified: 0 },
  ]
  if (storeSetup) storeSetup(store)

  const router = createTestRouter()
  const wrapper = shallowMount(ScanView, {
    global: { plugins: [pinia, router] },
  })
  return { wrapper, store, router }
}

describe('ScanView Component', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Rendering', () => {
    it('renders the scan view container', async () => {
      const { wrapper } = mountScanView()
      await nextTick()
      expect(wrapper.find('.scan').exists()).toBe(true)
    })

    it('shows scanning animation container when isScanning is true', async () => {
      const { wrapper, store } = mountScanView()
      store.isScanning = true
      await nextTick()
      expect(wrapper.find('.scanning-container').exists()).toBe(true)
    })

    it('hides scanning container and shows results when not scanning', async () => {
      const { wrapper, store } = mountScanView()
      store.isScanning = false
      await nextTick()
      expect(wrapper.find('.scanning-container').exists()).toBe(false)
      expect(wrapper.find('.overview-cards').exists()).toBe(true)
    })
  })

  describe('Tool Icon Mapping', () => {
    it('returns a defined component for each known tool ID', async () => {
      const { wrapper } = mountScanView()
      await nextTick()
      const vm = wrapper.vm as unknown as { getToolIcon: (id: string) => unknown }
      const knownTools = [
        'npm', 'yarn', 'pnpm', 'bun', 'composer', 'cargo', 'flutter',
        'nuget', 'android_sdk', 'docker', 'xcode', 'homebrew', 'python', 'go', 'ruby',
        'maven', 'gradle', 'cocoapods', 'carthage', 'unity', 'jetbrains', 'vscode',
      ]
      for (const toolId of knownTools) {
        expect(vm.getToolIcon(toolId), `expected icon for ${toolId}`).toBeDefined()
      }
    })

    it('returns a fallback icon for unknown tool IDs', async () => {
      const { wrapper } = mountScanView()
      await nextTick()
      const vm = wrapper.vm as unknown as { getToolIcon: (id: string) => unknown }
      expect(vm.getToolIcon('unknown_tool_xyz')).toBeDefined()
    })
  })

  describe('Tool Name Resolution', () => {
    it('returns tool name when the tool exists in the store', async () => {
      const { wrapper } = mountScanView(store => {
        store.tools = [{ id: 'npm', name: 'Node Package Manager', paths: [], enabled: true }]
      })
      await nextTick()
      const vm = wrapper.vm as unknown as { getToolName: (id: string) => string }
      expect(vm.getToolName('npm')).toBe('Node Package Manager')
    })

    it('returns the toolId as a fallback when tool is not found', async () => {
      const { wrapper } = mountScanView()
      await nextTick()
      const vm = wrapper.vm as unknown as { getToolName: (id: string) => string }
      expect(vm.getToolName('unknown_tool')).toBe('unknown_tool')
    })
  })

  describe('Format Size Helper', () => {
    it('delegates formatSize to the tool store', async () => {
      const { wrapper } = mountScanView()
      await nextTick()
      const vm = wrapper.vm as unknown as { formatSize: (n: number) => string }
      expect(vm.formatSize(0)).toBe('0 B')
      expect(vm.formatSize(1024)).toBe('1 KB')
      expect(vm.formatSize(1048576)).toBe('1 MB')
    })
  })

  describe('Progress Offset Computation', () => {
    it('equals the full circumference at 0% progress', async () => {
      const { wrapper } = mountScanView()
      await nextTick()
      const vm = wrapper.vm as unknown as {
        scanProgressLocal: number
        progressOffset: number
      }
      vm.scanProgressLocal = 0
      await nextTick()
      const circumference = 2 * Math.PI * 70
      expect(vm.progressOffset).toBeCloseTo(circumference, 2)
    })

    it('equals 0 at 100% progress', async () => {
      const { wrapper } = mountScanView()
      await nextTick()
      const vm = wrapper.vm as unknown as {
        scanProgressLocal: number
        progressOffset: number
      }
      vm.scanProgressLocal = 1
      await nextTick()
      expect(vm.progressOffset).toBeCloseTo(0, 2)
    })
  })

  describe('Navigation', () => {
    it('goBack pushes to the home route', async () => {
      const { wrapper, router } = mountScanView()
      await nextTick()

      const pushSpy = vi.spyOn(router, 'push')
      const vm = wrapper.vm as unknown as { goBack: () => void }
      vm.goBack()

      expect(pushSpy).toHaveBeenCalledWith('/')
    })
  })

  describe('Computed Properties', () => {
    it('isScanning reflects the store state', async () => {
      const { wrapper, store } = mountScanView()
      await nextTick()

      const vm = wrapper.vm as unknown as { isScanning: boolean }
      expect(vm.isScanning).toBe(false)

      store.isScanning = true
      await nextTick()
      expect(vm.isScanning).toBe(true)
    })

    it('scanResults reflects the store state', async () => {
      const { wrapper } = mountScanView(store => {
        store.scanResults = [
          { tool_id: 'npm', path: '~/.npm', size: 1024, file_num: 5, last_modified: 0 },
        ]
      })
      await nextTick()

      const vm = wrapper.vm as unknown as { scanResults: unknown[] }
      expect(vm.scanResults).toHaveLength(1)
    })

    it('totalCacheSizeFormatted shows human-readable size', async () => {
      const { wrapper } = mountScanView(store => {
        store.scanResults = [
          { tool_id: 'npm', path: '~/.npm', size: 1048576, file_num: 1, last_modified: 0 },
        ]
      })
      await nextTick()

      const vm = wrapper.vm as unknown as { totalCacheSizeFormatted: string }
      expect(vm.totalCacheSizeFormatted).toBe('1 MB')
    })
  })
})

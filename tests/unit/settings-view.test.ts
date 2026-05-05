import { describe, it, expect, vi, beforeEach } from 'vitest'
import { shallowMount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { nextTick } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import SettingsView from '@/views/SettingsView.vue'
import { useSettingsStore } from '@/stores/settings'

// Mock Tauri API
vi.mock('@tauri-apps/api/core', () => ({
  invoke: vi.fn()
}))

vi.mock('@tauri-apps/api/event', () => ({
  listen: vi.fn().mockResolvedValue(() => {})
}))

vi.mock('@/services/tauri', () => ({
  getSettings: vi.fn().mockResolvedValue({
    threshold: 100,
    whitelist: [],
    autoScan: false,
    scanInterval: 7,
    theme: 'auto',
  }),
  saveSettings: vi.fn().mockResolvedValue(undefined),
  getToolList: vi.fn().mockResolvedValue([]),
  scanTool: vi.fn().mockResolvedValue([]),
  scanAllTools: vi.fn().mockResolvedValue([]),
  cleanTool: vi.fn().mockResolvedValue({ tool_id: '', cleaned: 0, failed: [], file_num: 0 }),
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
      { path: '/settings', name: 'settings', component: { template: '<div>Settings</div>' } },
    ],
  })

const mountSettingsView = (storeSetup?: (store: ReturnType<typeof useSettingsStore>) => void) => {
  const pinia = createPinia()
  setActivePinia(pinia)

  // Get store before mounting so we can spy on fetchSettings
  // (prevents onMounted from overwriting store state)
  const store = useSettingsStore()
  vi.spyOn(store, 'fetchSettings').mockResolvedValue(undefined)

  if (storeSetup) storeSetup(store)

  const router = createTestRouter()
  const wrapper = shallowMount(SettingsView, {
    global: { plugins: [pinia, router] },
  })
  return { wrapper, store, router }
}

describe('SettingsView Component', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Rendering', () => {
    it('renders the settings view container', async () => {
      const { wrapper } = mountSettingsView()
      await nextTick()
      expect(wrapper.find('.settings').exists()).toBe(true)
    })

    it('has three theme options: dark, light, auto', async () => {
      const { wrapper } = mountSettingsView()
      await nextTick()
      const vm = wrapper.vm as unknown as {
        themeOptions: Array<{ value: string; label: string }>
      }
      expect(vm.themeOptions).toHaveLength(3)
      expect(vm.themeOptions.map(o => o.value)).toEqual(['dark', 'light', 'auto'])
    })
  })

  describe('Navigation', () => {
    it('goBack pushes to the home route', async () => {
      const { wrapper, router } = mountSettingsView()
      await nextTick()

      const pushSpy = vi.spyOn(router, 'push')
      const vm = wrapper.vm as unknown as { goBack: () => void }
      vm.goBack()

      expect(pushSpy).toHaveBeenCalledWith('/')
    })
  })

  describe('Whitelist Management', () => {
    it('addWhitelist trims the input, adds it to the store, and resets the field', async () => {
      const { wrapper, store } = mountSettingsView()
      await nextTick()

      const vm = wrapper.vm as unknown as {
        newWhitelist: string
        addWhitelist: () => void
      }

      vm.newWhitelist = '  /my/path  '
      vm.addWhitelist()
      await nextTick()

      expect(store.settings.whitelist).toContain('/my/path')
      expect(vm.newWhitelist).toBe('')
    })

    it('addWhitelist does nothing for an empty or whitespace-only input', async () => {
      const { wrapper, store } = mountSettingsView()
      await nextTick()

      const vm = wrapper.vm as unknown as {
        newWhitelist: string
        addWhitelist: () => void
      }

      vm.newWhitelist = '   '
      vm.addWhitelist()
      await nextTick()

      expect(store.settings.whitelist).toHaveLength(0)
    })

    it('removeWhitelist removes the given path from the store', async () => {
      const { wrapper, store } = mountSettingsView(s => {
        s.settings.whitelist = ['/path1', '/path2']
      })
      await nextTick()

      const vm = wrapper.vm as unknown as { removeWhitelist: (path: string) => void }
      vm.removeWhitelist('/path1')
      await nextTick()

      expect(store.settings.whitelist).not.toContain('/path1')
      expect(store.settings.whitelist).toContain('/path2')
    })
  })

  describe('Settings Computed Properties', () => {
    it('settings reflects the store state', async () => {
      const { wrapper, store } = mountSettingsView()
      store.settings.threshold = 250
      await nextTick()

      const vm = wrapper.vm as unknown as { settings: { threshold: number } }
      expect(vm.settings.threshold).toBe(250)
    })

    it('isLoading reflects the store state', async () => {
      const { wrapper, store } = mountSettingsView()
      store.isLoading = true
      await nextTick()

      const vm = wrapper.vm as unknown as { isLoading: boolean }
      expect(vm.isLoading).toBe(true)
    })

    it('isSaving reflects the store state', async () => {
      const { wrapper, store } = mountSettingsView()
      store.isSaving = true
      await nextTick()

      const vm = wrapper.vm as unknown as { isSaving: boolean }
      expect(vm.isSaving).toBe(true)
    })
  })

  describe('Save Settings', () => {
    it('calls store saveSettings and clears error on success', async () => {
      const { wrapper, store } = mountSettingsView()
      const saveSpy = vi.spyOn(store, 'saveSettings').mockResolvedValue(undefined)
      await nextTick()

      const vm = wrapper.vm as unknown as {
        saveSettings: () => Promise<void>
        error: string | null
      }
      await vm.saveSettings()

      expect(saveSpy).toHaveBeenCalled()
      expect(vm.error).toBe(null)
    })

    it('sets error message when saveSettings throws an Error', async () => {
      const { wrapper, store } = mountSettingsView()
      vi.spyOn(store, 'saveSettings').mockRejectedValue(new Error('Save failed'))
      await nextTick()

      const vm = wrapper.vm as unknown as {
        saveSettings: () => Promise<void>
        error: string | null
      }
      await vm.saveSettings()

      expect(vm.error).toBe('Save failed')
    })

    it('sets fallback error message for non-Error rejection', async () => {
      const { wrapper, store } = mountSettingsView()
      vi.spyOn(store, 'saveSettings').mockRejectedValue('connection lost')
      await nextTick()

      const vm = wrapper.vm as unknown as {
        saveSettings: () => Promise<void>
        error: string | null
      }
      await vm.saveSettings()

      expect(vm.error).toBe('保存设置失败')
    })
  })
})

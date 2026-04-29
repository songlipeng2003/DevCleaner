import { describe, it, expect, beforeEach } from 'vitest'
import { mount, config } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { nextTick } from 'vue'
import HomeView from '@/views/HomeView.vue'

// Mock Tauri API
vi.mock('@tauri-apps/api/core', () => ({
  invoke: vi.fn()
}))

// Mock ant-design-vue components globally
config.global.stubs = {
  'a-layout': true,
  'a-layout-header': true,
  'a-layout-content': true,
  'a-layout-footer': true,
  'a-card': true,
  'a-button': true,
  'a-row': true,
  'a-col': true,
  'a-progress': true,
  'a-statistic': true,
  'a-space': true,
  'a-switch': true,
  'a-list': true,
  'a-list-item': true,
  'a-list-item-meta': true,
  'a-divider': true,
  'a-alert': true,
  'a-empty': true,
  'a-spin': true,
  'a-descriptions': true,
  'a-descriptions-item': true,
  'a-drawer': true,
}

describe('DevCleaner Tests', () => {
  beforeEach(() => {
    // Create and activate a pinia instance for each test
    const pinia = createPinia()
    setActivePinia(pinia)
  })

  describe('HomeView Component', () => {
    it('renders correctly', async () => {
      const wrapper = mount(HomeView, {
        global: {
          plugins: [createPinia()],
        }
      })
      await nextTick()
      
      expect(wrapper.find('.home').exists()).toBe(true)
    })
  })

  describe('Format Size Utility', () => {
    it('formats bytes correctly', () => {
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
    })
  })

  describe('Tool ID Validation', () => {
    const validToolIds = ['npm', 'yarn', 'pnpm', 'docker', 'xcode', 'homebrew', 'python', 'go', 'ruby', 'maven', 'gradle', 'cocoapods', 'carthage', 'unity']
    
    it('contains all expected tool IDs', () => {
      expect(validToolIds).toContain('npm')
      expect(validToolIds).toContain('docker')
      expect(validToolIds).toContain('xcode')
      expect(validToolIds).toContain('homebrew')
      expect(validToolIds).toContain('maven')
      expect(validToolIds).toContain('gradle')
      expect(validToolIds).toContain('cocoapods')
      expect(validToolIds).toContain('carthage')
      expect(validToolIds).toContain('unity')
    })

    it('has 14 tools defined', () => {
      expect(validToolIds.length).toBe(14)
    })
  })
})

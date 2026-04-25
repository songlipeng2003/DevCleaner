import { describe, it, expect, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import HomeView from '@/views/HomeView.vue'

// Mock Tauri API
vi.mock('@tauri-apps/api/core', () => ({
  invoke: vi.fn()
}))

describe('DevCleaner Tests', () => {
  describe('HomeView Component', () => {
    it('renders correctly', () => {
      const wrapper = mount(HomeView, {
        global: {
          stubs: {
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
            'a-divider': true,
          }
        }
      })
      
      expect(wrapper.find('.home').exists()).toBe(true)
      expect(wrapper.find('h1').text()).toBe('DevCleaner')
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
    const validToolIds = ['npm', 'yarn', 'pnpm', 'docker', 'xcode', 'homebrew', 'python', 'go', 'ruby', 'maven', 'gradle', 'cocoapods', 'unity']
    
    it('contains all expected tool IDs', () => {
      expect(validToolIds).toContain('npm')
      expect(validToolIds).toContain('docker')
      expect(validToolIds).toContain('xcode')
      expect(validToolIds).toContain('homebrew')
    })

    it('has 13 tools defined', () => {
      expect(validToolIds.length).toBe(13)
    })
  })
})

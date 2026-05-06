import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Settings } from '@/types'
import * as tauriApi from '@/services/tauri'

export const useSettingsStore = defineStore('settings', () => {
  // 默认设置
  const defaultSettings: Settings = {
    threshold: 100,
    whitelist: [],
    autoScan: false,
    scanInterval: 7,
    theme: 'auto',
    shortcutsEnabled: true,
  }

  // 状态
  const settings = ref<Settings>({ ...defaultSettings })
  const isLoading = ref(false)
  const isSaving = ref(false)
  const error = ref<string | null>(null)

  // 获取设置
  async function fetchSettings() {
    isLoading.value = true
    error.value = null
    try {
      settings.value = await tauriApi.getSettings()
    } catch (e) {
      error.value = e instanceof Error ? e.message : '获取设置失败'
      settings.value = { ...defaultSettings }
    } finally {
      isLoading.value = false
    }
  }

  // 保存设置
  async function saveSettings(newSettings: Partial<Settings>) {
    isSaving.value = true
    error.value = null
    try {
      const updated = { ...settings.value, ...newSettings }
      await tauriApi.saveSettings(updated)
      settings.value = updated
    } catch (e) {
      error.value = e instanceof Error ? e.message : '保存设置失败'
      throw e
    } finally {
      isSaving.value = false
    }
  }

  // 重置设置
  function resetSettings() {
    settings.value = { ...defaultSettings }
  }

  // 添加白名单路径
  function addWhitelist(path: string) {
    if (path && !settings.value.whitelist.includes(path)) {
      settings.value.whitelist.push(path)
    }
  }

  // 移除白名单路径
  function removeWhitelist(path: string) {
    settings.value.whitelist = settings.value.whitelist.filter(p => p !== path)
  }

  return {
    settings,
    isLoading,
    isSaving,
    error,
    fetchSettings,
    saveSettings,
    resetSettings,
    addWhitelist,
    removeWhitelist,
  }
})

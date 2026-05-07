import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ToolInfo, ScanResult } from '@/types'
import * as tauriApi from '@/services/tauri'
import type { ScanProgress as TauriScanProgress } from '@/services/tauri'

export const useToolStore = defineStore('tools', () => {
  // 状态
  const tools = ref<ToolInfo[]>([])
  const scanResults = ref<ScanResult[]>([])
  const isScanning = ref(false)
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // 计算属性
  const enabledTools = computed(() => tools.value.filter(t => t.enabled))
  const totalCacheSize = computed(() => 
    scanResults.value.reduce((sum, r) => sum + r.size, 0)
  )

  // 获取工具列表
  async function fetchTools() {
    isLoading.value = true
    error.value = null
    try {
      tools.value = await tauriApi.getToolList()
    } catch (e) {
      error.value = e instanceof Error ? e.message : '获取工具列表失败'
    } finally {
      isLoading.value = false
    }
  }

  // 扫描单个工具
  async function scanTool(toolId: string) {
    const results = await tauriApi.scanTool(toolId)
    scanResults.value = [...scanResults.value.filter(r => r.tool_id !== toolId), ...results]
    return results
  }

  // 扫描所有启用的工具（带进度回调）
  async function scanAllTools(onProgress?: (progress: TauriScanProgress) => void) {
    isScanning.value = true
    error.value = null

    try {
      const results = await tauriApi.scanAllTools(onProgress)
      scanResults.value = results
      return results
    } catch (e) {
      error.value = e instanceof Error ? e.message : '扫描失败'
      throw e
    } finally {
      isScanning.value = false
    }
  }

  // 清理工具缓存
  async function cleanTool(toolId: string, paths: string[]) {
    try {
      const result = await tauriApi.cleanTool(toolId, paths)
      // 清理成功后，从扫描结果中移除
      scanResults.value = scanResults.value.filter(r => r.tool_id !== toolId)
      
      // 记录清理统计
      const tool = tools.value.find(t => t.id === toolId)
      if (tool && result.cleaned > 0) {
        await tauriApi.recordClean(toolId, tool.name, result.cleaned, result.file_num)
        // 同时记录清理历史
        await tauriApi.recordCleanHistory(toolId, tool.name, result.cleaned, result.file_num, paths)
      }
      
      return result
    } catch (e) {
      error.value = e instanceof Error ? e.message : '清理失败'
      throw e
    }
  }

  // 获取工具扫描结果
  function getToolResults(toolId: string) {
    return scanResults.value.filter(r => r.tool_id === toolId)
  }

  // 获取工具缓存大小
  function getToolSize(toolId: string) {
    return getToolResults(toolId).reduce((sum, r) => sum + r.size, 0)
  }

  // 格式化大小
  function formatSize(bytes: number): string {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  // 打开路径
  async function openPath(path: string) {
    return tauriApi.openPath(path)
  }

  // 切换工具启用状态
  function toggleTool(toolId: string, enabled: boolean) {
    const tool = tools.value.find(t => t.id === toolId)
    if (tool) {
      tool.enabled = enabled
    }
  }

  // 更新工具启用状态（批量）
  function updateToolEnabled(toolId: string, enabled: boolean) {
    toggleTool(toolId, enabled)
  }

  return {
    // 状态
    tools,
    scanResults,
    isScanning,
    isLoading,
    error,
    // 计算属性
    enabledTools,
    totalCacheSize,
    // 方法
    fetchTools,
    scanTool,
    scanAllTools,
    cleanTool,
    getToolResults,
    getToolSize,
    formatSize,
    openPath,
    toggleTool,
    updateToolEnabled,
  }
})

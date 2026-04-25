import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ToolInfo, ScanResult, ScanProgress } from '@/types'
import * as tauriApi from '@/services/tauri'

export const useToolStore = defineStore('tools', () => {
  // 状态
  const tools = ref<ToolInfo[]>([])
  const scanResults = ref<ScanResult[]>([])
  const scanProgress = ref<Map<string, ScanProgress>>(new Map())
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
    scanProgress.value.set(toolId, {
      tool_id: toolId,
      status: 'scanning',
      progress: 0,
    })

    try {
      const results = await tauriApi.scanTool(toolId)
      scanResults.value = [...scanResults.value.filter(r => r.tool_id !== toolId), ...results]
      scanProgress.value.set(toolId, {
        tool_id: toolId,
        status: 'completed',
        progress: 100,
      })
      return results
    } catch (e) {
      scanProgress.value.set(toolId, {
        tool_id: toolId,
        status: 'error',
        progress: 0,
      })
      throw e
    }
  }

  // 扫描所有启用的工具
  async function scanAllTools() {
    isScanning.value = true
    error.value = null

    // 初始化进度
    for (const tool of enabledTools.value) {
      scanProgress.value.set(tool.id, {
        tool_id: tool.id,
        status: 'pending',
        progress: 0,
      })
    }

    try {
      const results = await tauriApi.scanAllTools()
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

  return {
    // 状态
    tools,
    scanResults,
    scanProgress,
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
  }
})

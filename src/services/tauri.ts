import { invoke } from '@tauri-apps/api/core'
import { listen, type UnlistenFn } from '@tauri-apps/api/event'
import type { ScanResult, ToolInfo, CleanResult, Settings } from '@/types'

// 扫描进度类型
export interface ScanProgress {
  toolId: string
  toolName: string
  progress: number
  currentPath: string
  pathsScanned: number
  totalPaths: number
}

// 获取所有支持的工具
export async function getToolList(): Promise<ToolInfo[]> {
  return invoke('get_tool_list')
}

// 获取单个工具信息
export async function getToolInfo(toolId: string): Promise<ToolInfo> {
  return invoke('get_tool_info', { toolId })
}

// 扫描指定工具
export async function scanTool(toolId: string): Promise<ScanResult[]> {
  return invoke('scan_tool', { toolId })
}

// 扫描所有工具（带进度回调）
export async function scanAllTools(onProgress?: (progress: ScanProgress) => void): Promise<ScanResult[]> {
  // 监听扫描进度事件
  let unlisten: UnlistenFn | null = null
  if (onProgress) {
    unlisten = await listen<ScanProgress>('scan-progress', (event) => {
      onProgress(event.payload)
    })
  }

  try {
    const result = await invoke<ScanResult[]>('scan_all_tools')
    return result
  } finally {
    if (unlisten) {
      unlisten()
    }
  }
}

// 监听扫描完成事件
export async function onScanComplete(callback: (progress: ScanProgress) => void): Promise<UnlistenFn> {
  return listen<ScanProgress>('scan-complete', (event) => {
    callback(event.payload)
  })
}

// 清理指定工具的缓存
export async function cleanTool(toolId: string, paths: string[]): Promise<CleanResult> {
  return invoke('clean_tool', { toolId, paths })
}

// 获取设置
export async function getSettings(): Promise<Settings> {
  return invoke('get_settings')
}

// 保存设置
export async function saveSettings(settings: Settings): Promise<void> {
  return invoke('save_settings', { settings })
}

// 获取磁盘使用情况
export async function getDiskUsage(): Promise<{ total: number; used: number; free: number }> {
  return invoke('get_disk_usage')
}

// 打开路径（文件管理器）
export async function openPath(path: string): Promise<void> {
  return invoke('open_path', { path })
}

// 获取版本信息
export async function getVersion(): Promise<{ version: string; build: string }> {
  return invoke('get_version')
}

// 预览清理项
export interface PreviewItem {
  path: string
  size: number
  fileNum: number
  lastModified: number
}

export async function previewTool(toolId: string, paths: string[]): Promise<PreviewItem[]> {
  return invoke('preview_tool', { toolId, paths })
}

// 使用统计
export interface UsageStats {
  totalCleaned: number
  cleanCount: number
  lastClean: number
  cleanHistory: CleanHistoryItem[]
}

export interface CleanHistoryItem {
  toolId: string
  toolName: string
  size: number
  fileNum: number
  timestamp: number
}

export async function getUsageStats(): Promise<UsageStats> {
  return invoke('get_usage_stats')
}

export async function recordClean(toolId: string, toolName: string, size: number, fileNum: number): Promise<void> {
  return invoke('record_clean', { toolId, toolName, size, fileNum })
}

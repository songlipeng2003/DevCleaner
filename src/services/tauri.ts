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
  const result = await invoke<{ version: string; build: string }>('get_version')
  return result
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

// ============== v0.2.0 新增 API ==============

// 清理预览 - 文件信息
export interface PreviewFile {
  name: string
  path: string
  size: number
  modified: number
  isSafe: boolean
  reason?: string
}

// 清理预览 - 路径信息
export interface PreviewPath {
  path: string
  files: PreviewFile[]
  size: number
  oldestFile: number
  newestFile: number
}

// 清理预览 - 完整预览
export interface CleanPreview {
  toolId: string
  toolName: string
  paths: PreviewPath[]
  totalSize: number
  riskLevel: 'safe' | 'moderate' | 'careful'
  recommendations: string[]
}

// 清理策略
export type CleanStrategy = 'time' | 'size' | 'selective' | 'safe' | 'deep'

// 获取增强的清理预览
export async function getCleanPreview(
  toolId: string,
  paths: string[],
  strategy?: CleanStrategy,
  timeThreshold?: number,
  sizeThreshold?: number
): Promise<CleanPreview> {
  return invoke('get_clean_preview', {
    toolId,
    paths,
    strategy,
    timeThreshold,
    sizeThreshold,
  })
}

// 项目扫描结果
export interface CleanableItem {
  id: string
  name: string
  path: string
  itemType: string
  size: number
  fileNum: number
  lastModified: number
  cleanable: boolean
  reason: string
}

export interface ProjectScanResult {
  name: string
  path: string
  projectType: string
  size: number
  fileNum: number
  lastModified: number
  cleanableItems: CleanableItem[]
  riskLevel: 'safe' | 'moderate' | 'careful'
}

// 扫描项目目录
export async function scanProjects(
  scanPaths: string[],
  maxDepth?: number
): Promise<ProjectScanResult[]> {
  return invoke('scan_projects', { scanPaths, maxDepth })
}

// 清理指定路径
export async function cleanPaths(paths: string[]): Promise<CleanResult> {
  return invoke('clean_paths', { paths })
}

// 清理历史 - 项目信息
export interface CleanHistoryItemV2 {
  id: string
  toolId: string
  toolName: string
  size: number
  fileNum: number
  timestamp: number
  paths: string[]
}

// 月度统计
export interface MonthlyStat {
  month: string
  cleaned: number
  count: number
}

// 清理历史
export interface CleanHistory {
  items: CleanHistoryItemV2[]
  totalCleaned: number
  totalCount: number
  monthlyStats: MonthlyStat[]
}

// 获取清理历史
export async function getCleanHistory(
  filter?: 'day' | 'week' | 'month' | 'all'
): Promise<CleanHistory> {
  return invoke('get_clean_history', { filter })
}

// 记录清理历史
export async function recordCleanHistory(
  toolId: string,
  toolName: string,
  size: number,
  fileNum: number,
  paths: string[]
): Promise<void> {
  return invoke('record_clean_history', { toolId, toolName, size, fileNum, paths })
}

// 导出清理报告
export async function exportCleanReport(
  format: 'json' | 'csv'
): Promise<string> {
  return invoke('export_clean_report', { format })
}

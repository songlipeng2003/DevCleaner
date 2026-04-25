import { invoke } from '@tauri-apps/api/core'
import type { ScanResult, ToolInfo, CleanResult, Settings } from '@/types'

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

// 扫描所有工具
export async function scanAllTools(): Promise<ScanResult[]> {
  return invoke('scan_all_tools')
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

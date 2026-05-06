// DevTool 开发工具类型
export interface DevTool {
  id: string
  name: string
  icon: string
  description: string
}

// ScanResult 扫描结果
export interface ScanResult {
  tool_id: string
  path: string
  size: number
  file_num: number
  last_modified: number
}

// ToolInfo 工具信息
export interface ToolInfo {
  id: string
  name: string
  paths: string[]
  enabled: boolean
}

// CleanResult 清理结果
export interface CleanResult {
  tool_id: string
  cleaned: number
  failed: string[]
  file_num: number
}

// ScanProgress 扫描进度
export interface ScanProgress {
  tool_id: string
  status: 'pending' | 'scanning' | 'completed' | 'error'
  progress: number
  current_path?: string
}

// Settings 应用设置
export interface Settings {
  threshold: number
  whitelist: string[]
  autoScan: boolean
  scanInterval: number
  theme: 'light' | 'dark' | 'auto'
  shortcutsEnabled: boolean
}

// API 响应类型
export interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: string
}

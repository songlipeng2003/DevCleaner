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

// ============== v0.2.0 新增类型 ==============

// 清理策略类型
export type CleanStrategy = 'time' | 'size' | 'selective' | 'safe' | 'deep'

// 清理预览 - 预览文件信息
export interface PreviewFile {
  name: string
  path: string
  size: number
  modified: number
  isSafe: boolean
  reason?: string
}

// 清理预览 - 预览路径信息
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

// 项目扫描结果
export interface ProjectScanResult {
  name: string
  path: string
  type?: string
  projectType?: string
  size: number
  fileNum: number
  lastModified: number
  cleanableItems: CleanableItem[]
  riskLevel: 'safe' | 'moderate' | 'careful'
}

// 可清理项目
export interface CleanableItem {
  id: string
  name: string
  path: string
  type?: string
  itemType?: string
  size: number
  fileNum: number
  lastModified: number
  cleanable: boolean
  reason: string
}

// 清理历史记录
export interface CleanHistoryItem {
  id: string
  toolId: string
  toolName: string
  size: number
  fileNum: number
  timestamp: number
  paths: string[]
}

// 清理统计
export interface CleanStats {
  totalCleaned: number
  cleanCount: number
  lastClean: number
  cleanHistory: CleanHistoryItem[]
  monthlyStats: MonthlyStats[]
}

// 月度统计
export interface MonthlyStats {
  month: string
  cleaned: number
  count: number
}

// 项目清理配置
export interface ProjectCleanConfig {
  scanPaths: string[]
  includeTypes: string[]
  excludePatterns: string[]
  maxDepth: number
  minSize: number
}

// 清理策略配置
export interface CleanStrategyConfig {
  strategy: CleanStrategy
  timeThreshold?: number      // 天数（按时间清理）
  sizeThreshold?: number      // 字节（按大小清理）
  selectedPaths?: string[]    // 选中的路径（选择性清理）
  excludePatterns?: string[]  // 排除模式
}

// 磁盘使用分析
export interface DiskAnalysis {
  category: string
  items: DiskAnalysisItem[]
  totalSize: number
}

export interface DiskAnalysisItem {
  name: string
  toolId?: string
  size: number
  percentage: number
  lastModified: number
}

// 缓存趋势数据
export interface CacheTrend {
  date: string
  size: number
}

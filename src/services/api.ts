import axios from 'axios'
import type { ScanResult, ToolInfo, CleanResult, Settings } from '@/types'

const API_BASE = 'http://localhost:8080/api'

const api = axios.create({
  baseURL: API_BASE,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// DevTool 相关的 API
export const toolApi = {
  // 获取工具列表
  list: async (): Promise<ToolInfo[]> => {
    const { data } = await api.get('/tools')
    return data
  },

  // 获取工具详情
  get: async (toolId: string): Promise<ToolInfo> => {
    const { data } = await api.get(`/tools/${toolId}`)
    return data
  },

  // 获取工具扫描结果
  scan: async (toolId: string): Promise<ScanResult[]> => {
    const { data } = await api.post(`/tools/${toolId}/scan`)
    return data
  },

  // 清理工具缓存
  clean: async (toolId: string, paths: string[]): Promise<CleanResult> => {
    const { data } = await api.post(`/tools/${toolId}/clean`, { paths })
    return data
  },
}

// Settings 相关的 API
export const settingsApi = {
  // 获取设置
  get: async (): Promise<Settings> => {
    const { data } = await api.get('/settings')
    return data
  },

  // 保存设置
  save: async (settings: Settings): Promise<void> => {
    await api.put('/settings', settings)
  },
}

// System 相关的 API
export const systemApi = {
  // 获取磁盘使用情况
  diskUsage: async () => {
    const { data } = await api.get('/system/disk')
    return data
  },

  // 获取版本信息
  version: async () => {
    const { data } = await api.get('/system/version')
    return data
  },
}

export default api

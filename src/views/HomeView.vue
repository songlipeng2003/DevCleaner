<template>
  <div class="home">
    <!-- 极光背景装饰 -->
    <div class="hero-glow hero-glow-1" />
    <div class="hero-glow hero-glow-2" />
    
    <a-layout class="layout">
      <a-layout-header class="header">
        <div class="header-left">
          <div class="logo-icon">
            <Zap :size="20" />
          </div>
          <h1>DevCleaner</h1>
          <span class="version">v{{ version }}{{ buildType ? '-' + buildType : '' }}</span>
        </div>
        <div class="header-right">
          <button
            class="theme-toggle-btn"
            @click="toggleTheme"
          >
            <Moon
              v-if="isDark"
              :size="18"
            />
            <Sun
              v-else
              :size="18"
            />
          </button>
          <a-button @click="openSettings">
            <template #icon>
              <SettingOutlined />
            </template>
            设置
          </a-button>
        </div>
      </a-layout-header>
      
      <a-layout-content class="content">
        <!-- 磁盘使用情况 -->
        <div class="hero-section">
          <div class="hero-content">
            <div class="hero-badge fade-in-up">
              <Sparkles :size="14" />
              <span>智能清理 · 安全无忧</span>
            </div>
            
            <h1 class="hero-title fade-in-up fade-in-up-delay-1">
              释放磁盘空间<br>
              <span class="gradient-text">让开发更流畅</span>
            </h1>
            
            <p class="hero-description fade-in-up fade-in-up-delay-2">
              智能扫描并清理 Node.js、Python、Docker 等开发工具的缓存文件，
              帮助您轻松释放宝贵的磁盘空间。
            </p>
            
            <div class="hero-actions fade-in-up fade-in-up-delay-3">
              <a-button
                type="primary"
                size="large"
                :loading="isScanning"
                @click="startScan"
              >
                <template #icon>
                  <ScanOutlined />
                </template>
                {{ isScanning ? '扫描中...' : '开始扫描' }}
              </a-button>
              <a-button
                size="large"
                @click="refreshTools"
              >
                <template #icon>
                  <ReloadOutlined />
                </template>
                刷新
              </a-button>
            </div>
          </div>
          
          <!-- 磁盘图表 -->
          <div class="hero-visual fade-in-up fade-in-up-delay-2">
            <div class="disk-chart-container glass-card">
              <div class="disk-chart-wrapper">
                <svg
                  class="disk-chart"
                  viewBox="0 0 200 200"
                >
                  <circle
                    cx="100"
                    cy="100"
                    r="70"
                    fill="none"
                    stroke="var(--aurora-border)"
                    stroke-width="10"
                  />
                  <circle
                    class="progress-ring__circle"
                    cx="100"
                    cy="100"
                    r="70"
                    fill="none"
                    stroke="url(#diskGradient)"
                    stroke-width="10"
                    :stroke-dasharray="circumference"
                    :stroke-dashoffset="circumference - (diskPercent / 100) * circumference"
                  />
                  <defs>
                    <linearGradient
                      id="diskGradient"
                      x1="0%"
                      y1="0%"
                      x2="100%"
                      y2="100%"
                    >
                      <stop
                        offset="0%"
                        stop-color="#667eea"
                      />
                      <stop
                        offset="100%"
                        stop-color="#00d9ff"
                      />
                    </linearGradient>
                  </defs>
                </svg>
                
                <div class="disk-chart-center">
                  <HardDrive
                    :size="24"
                    class="disk-icon"
                  />
                  <div class="disk-percentage gradient-text">
                    {{ diskPercent }}%
                  </div>
                  <div class="disk-label">
                    已使用
                  </div>
                </div>
              </div>
              
              <div class="disk-stats">
                <div class="disk-stat">
                  <span class="disk-stat-value">{{ formatSize(diskUsage.used) }}</span>
                  <span class="disk-stat-label">已用</span>
                </div>
                <div class="disk-stat">
                  <span class="disk-stat-value">{{ formatSize(diskUsage.free) }}</span>
                  <span class="disk-stat-label">可用</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <!-- 扫描进度 -->
        <div
          v-if="isScanning"
          class="scan-progress-container glass-card"
        >
          <div class="scan-progress-header">
            <ScanOutlined class="scan-icon spinning" />
            <span class="scan-title">正在扫描: {{ scanProgress.toolName }}</span>
            <span class="scan-percent gradient-text">{{ Math.round(scanProgress.progress * 100) }}%</span>
          </div>
          <a-progress
            :percent="Math.round(scanProgress.progress * 100)"
            :show-info="false"
            :stroke-color="{ '0%': '#667eea', '100%': '#00d9ff' }"
          />
        </div>
        
        <!-- 扫描摘要 & 统计卡片 -->
        <div
          v-else-if="scanResults.length > 0 || stats.totalCleaned > 0"
          class="summary-section"
        >
          <!-- 清理统计卡片 (v0.2.0 新增) -->
          <div
            v-if="stats.totalCleaned > 0"
            class="stats-cards"
          >
            <div class="stat-card glass-card">
              <div class="stat-icon-wrapper">
                <DatabaseOutlined class="stat-icon" />
              </div>
              <div class="stat-info">
                <div class="stat-value">
                  {{ formatSize(stats.totalCleaned) }}
                </div>
                <div class="stat-label">
                  总计已清理
                </div>
              </div>
            </div>
            <div class="stat-card glass-card">
              <div class="stat-icon-wrapper">
                <CheckCircleOutlined class="stat-icon" />
              </div>
              <div class="stat-info">
                <div class="stat-value">
                  {{ stats.cleanCount }}
                </div>
                <div class="stat-label">
                  清理次数
                </div>
              </div>
            </div>
            <div class="stat-card glass-card">
              <div class="stat-icon-wrapper">
                <History class="stat-icon" />
              </div>
              <div class="stat-info">
                <div class="stat-value">
                  {{ formatSize(totalCacheSize) }}
                </div>
                <div class="stat-label">
                  当前缓存
                </div>
              </div>
            </div>
          </div>

          <!-- 扫描摘要 -->
          <div
            v-if="scanResults.length > 0"
            class="scan-summary glass-card"
          >
            <span>发现 {{ scanResults.length }} 处缓存，共 {{ formatSize(totalCacheSize) }}</span>
          </div>
        </div>

        <!-- 快速操作 (v0.2.0 新增) -->
        <div class="quick-actions glass-card">
          <div class="quick-actions-header">
            <Zap :size="16" />
            <span>快速操作</span>
          </div>
          <div class="quick-actions-grid">
            <button
              class="quick-action-btn"
              :disabled="isScanning"
              @click="startScan"
            >
              <ScanOutlined :size="20" />
              <span>开始扫描</span>
            </button>
            <button
              class="quick-action-btn"
              :disabled="scanResults.length === 0"
              @click="quickCleanAll"
            >
              <DeleteOutlined :size="20" />
              <span>快速清理</span>
            </button>
            <button
              class="quick-action-btn"
              @click="goToProjects"
            >
              <FolderSearch :size="20" />
              <span>项目清理</span>
            </button>
            <button
              class="quick-action-btn"
              @click="goToHistory"
            >
              <History :size="20" />
              <span>清理历史</span>
            </button>
            <button
              class="quick-action-btn"
              @click="goToAnalysis"
            >
              <PieChart :size="20" />
              <span>磁盘分析</span>
            </button>
          </div>
        </div>
        
        <!-- 智能推荐 -->
        <a-card
          v-if="recommendations.length > 0 && !isScanning"
          class="recommendation-card glass-card"
        >
          <template #title>
            <div class="recommendation-title">
              <BulbOutlined />
              <span>智能推荐</span>
            </div>
          </template>
          <a-list
            :data-source="recommendations"
            item-layout="horizontal"
          >
            <template #renderItem="{ item }">
              <a-list-item>
                <template #actions>
                  <a-button
                    type="primary"
                    size="small"
                    @click="quickClean(item)"
                  >
                    清理 {{ formatSize(item.size) }}
                  </a-button>
                </template>
                <a-list-item-meta>
                  <template #title>
                    <component
                      :is="getToolIcon(item.toolId)"
                      :size="16"
                      :stroke-width="1.5"
                      style="margin-right: 8px; vertical-align: middle;"
                    />
                    {{ item.toolName }}
                  </template>
                  <template #description>
                    {{ item.reason }}
                  </template>
                </a-list-item-meta>
              </a-list-item>
            </template>
          </a-list>
        </a-card>

        <!-- 开发工具列表 -->
        <a-alert
          v-if="error"
          type="error"
          :message="error"
          show-icon
          closable
          style="margin-bottom: 16px"
          @close="error = null"
        />
        
        <div class="section-header">
          <div class="section-title-group">
            <h2 class="section-title">
              开发工具
            </h2>
            <p class="section-subtitle">
              点击工具卡片查看详情
            </p>
          </div>
        </div>
        
        <a-spin :spinning="toolStore.isLoading">
          <div
            v-if="enabledTools.length > 0"
            class="tools-grid"
          >
            <div
              v-for="tool in enabledTools"
              :key="tool.id"
              class="tool-card"
              :class="{ 'has-cache': getToolSize(tool.id) > 0 }"
              @click="showToolDetail(tool)"
            >
              <div class="tool-header">
                <div class="tool-icon-wrapper">
                  <component
                    :is="getToolIcon(tool.id)"
                    :size="28"
                    :stroke-width="1.5"
                  />
                </div>
                <a-switch 
                  v-model:checked="tool.enabled" 
                  size="small" 
                  @click.stop
                  @change="toggleTool(tool)"
                />
              </div>
              <div class="tool-info">
                <div class="tool-name">
                  {{ tool.name }}
                </div>
                <div
                  class="tool-size"
                  :class="{ 'has-size': getToolSize(tool.id) > 0 }"
                >
                  {{ getToolSize(tool.id) > 0 ? formatSize(getToolSize(tool.id)) : '无缓存' }}
                </div>
                <div class="tool-paths">
                  {{ tool.paths.length }} 个路径
                </div>
              </div>
              <div class="tool-footer">
                <a-button 
                  type="text" 
                  size="small" 
                  :disabled="getToolSize(tool.id) === 0"
                  @click.stop="cleanTool(tool)"
                >
                  清理
                </a-button>
                <a-button
                  type="text"
                  size="small"
                  @click.stop="openToolPath(tool)"
                >
                  打开
                </a-button>
              </div>
            </div>
          </div>
          
          <a-empty
            v-else-if="!toolStore.isLoading"
            description="暂无启用的开发工具"
          />
        </a-spin>
      </a-layout-content>
      
      <a-layout-footer class="footer">
        <a-space>
          <span>DevCleaner v{{ version }}{{ buildType ? '-' + buildType : '' }}</span>
          <a-divider type="vertical" />
          <a
            href="https://github.com/songlipeng2003/DevCleaner"
            target="_blank"
          >GitHub</a>
        </a-space>
      </a-layout-footer>
    </a-layout>

    <!-- 工具详情抽屉 -->
    <a-drawer
      v-model:open="drawerVisible"
      :title="selectedTool?.name"
      width="400"
      placement="right"
    >
      <template v-if="selectedTool">
        <a-descriptions
          :column="1"
          bordered
        >
          <a-descriptions-item label="工具 ID">
            {{ selectedTool.id }}
          </a-descriptions-item>
          <a-descriptions-item label="路径数量">
            {{ selectedTool.paths.length }}
          </a-descriptions-item>
          <a-descriptions-item label="当前缓存">
            {{ formatSize(getToolSize(selectedTool.id)) }}
          </a-descriptions-item>
        </a-descriptions>

        <a-divider>缓存路径</a-divider>
        
        <a-list 
          size="small" 
          :data-source="selectedTool.paths"
        >
          <template #renderItem="{ item }">
            <a-list-item>
              <template #actions>
                <a-button
                  type="link"
                  size="small"
                  @click="openPath(item)"
                >
                  打开
                </a-button>
              </template>
              <a-list-item-meta>
                <template #title>
                  {{ item }}
                </template>
              </a-list-item-meta>
            </a-list-item>
          </template>
        </a-list>

        <a-divider>操作</a-divider>
        <a-space
          direction="vertical"
          style="width: 100%"
        >
          <a-button
            block
            type="primary"
            @click="scanTool(selectedTool)"
          >
            扫描此工具
          </a-button>
          <a-button 
            block 
            danger 
            :disabled="getToolSize(selectedTool.id) === 0"
            @click="cleanTool(selectedTool)"
          >
            清理缓存
          </a-button>
        </a-space>
      </template>
    </a-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import { 
  SettingOutlined, 
  ScanOutlined, 
  ReloadOutlined,
  BulbOutlined,
  DatabaseOutlined,
  CheckCircleOutlined,
  DeleteOutlined,
} from '@ant-design/icons-vue'
import {
  Package,
  Sparkles,
  Folder,
  FolderSearch,
  Cookie,
  Gem,
  Box,
  Wind,
  Smartphone,
  Bug,
  Apple,
  Beer,
  CircleDot,
  BookOpen,
  Cog,
  Gamepad2,
  Code2,
  Laptop,
  HardDrive,
  Zap,
  Moon,
  Sun,
  History,
  PieChart,
} from 'lucide-vue-next'
import { useToolStore } from '@/stores/tools'
import { useSettingsStore } from '@/stores/settings'
import { getDiskUsage, type ScanProgress, getUsageStats, type UsageStats, getVersion } from '@/services/tauri'
import {
  trackAppStarted,
  trackPageView,
  trackScanStart,
  trackScanComplete,
  trackCleanComplete,
  trackSettingChange,
} from '@/services/analytics'
import * as tauriApi from '@/services/tauri'
import type { ToolInfo } from '@/types'
import type { Component } from 'vue'

// 智能推荐项
interface Recommendation {
  toolId: string
  toolName: string
  size: number
  reason: string
}

const router = useRouter()
const toolStore = useToolStore()
const settingsStore = useSettingsStore()

const version = ref('')
const buildType = ref('')
const drawerVisible = ref(false)
const selectedTool = ref<ToolInfo | null>(null)
const diskRefreshTimer = ref<ReturnType<typeof setTimeout> | null>(null)
const error = ref<string | null>(null)

// 扫描进度状态
const scanProgress = ref<ScanProgress>({
  toolId: '',
  toolName: '',
  progress: 0,
  currentPath: '',
  pathsScanned: 0,
  totalPaths: 0
})

// 智能推荐
const recommendations = ref<Recommendation[]>([])

// 统计数据 (v0.2.0)
const stats = ref({
  totalCleaned: 0,
  cleanCount: 0,
})

const diskUsage = ref({
  total: 0,
  used: 0,
  free: 0
})

const diskPercent = computed(() => {
  if (diskUsage.value.total <= 0) return 0
  return Math.round((diskUsage.value.used / diskUsage.value.total) * 100)
})

const circumference = 2 * Math.PI * 70

const enabledTools = computed(() => toolStore.tools.filter(t => t.enabled))
const scanResults = computed(() => toolStore.scanResults)
const totalCacheSize = computed(() => toolStore.totalCacheSize)
const isScanning = computed(() => toolStore.isScanning)

const isDark = computed(() => {
  return settingsStore.settings.theme === 'dark' ||
    (settingsStore.settings.theme === 'auto' && window.matchMedia('(prefers-color-scheme: dark)').matches)
})

// 工具图标映射
const toolIcons: Record<string, Component> = {
  npm: Package,
  yarn: Sparkles,
  pnpm: Folder,
  bun: Cookie,
  composer: Gem,
  cargo: Box,
  flutter: Wind,
  nuget: Package,
  android_sdk: Smartphone,
  docker: Bug,
  xcode: Apple,
  homebrew: Beer,
  python: CircleDot,
  go: CircleDot,
  ruby: Gem,
  maven: BookOpen,
  gradle: Cog,
  cocoapods: Box,
  carthage: Gamepad2,
  unity: Gamepad2,
  jetbrains: Code2,
  vscode: Laptop,
}

function getToolIcon(toolId: string) {
  return toolIcons[toolId] || HardDrive
}

function formatSize(bytes: number): string {
  return toolStore.formatSize(bytes)
}

function getToolSize(toolId: string): number {
  return toolStore.getToolSize(toolId)
}

function toggleTheme() {
  const newTheme = isDark.value ? 'light' : 'dark'
  settingsStore.settings.theme = newTheme
  settingsStore.saveSettings({ theme: newTheme })
  trackSettingChange('theme', newTheme)
}

async function refreshTools() {
  await toolStore.fetchTools()
  await fetchDiskUsage()
}

async function startScan() {
  trackScanStart('all')
  try {
    await toolStore.scanAllTools((progress) => {
      scanProgress.value = progress
    })
    message.success(`扫描完成，共发现 ${scanResults.value.length} 处缓存`)
    await generateRecommendations()
    trackScanComplete('all', scanResults.value.length)
  } catch (e) {
    error.value = e instanceof Error ? e.message : '扫描失败'
    message.error('扫描失败')
  }
}

async function scanTool(tool: ToolInfo) {
  try {
    await toolStore.scanTool(tool.id)
    message.success(`${tool.name} 扫描完成`)
  } catch (e) {
    error.value = e instanceof Error ? e.message : '扫描失败'
    message.error('扫描失败')
  }
}

function showToolDetail(tool: ToolInfo) {
  selectedTool.value = tool
  drawerVisible.value = true
}

function toggleTool(tool: ToolInfo) {
  toolStore.toggleTool(tool.id, tool.enabled)
  message.info(`${tool.name} 已${tool.enabled ? '启用' : '禁用'}`)
}

async function cleanTool(tool: ToolInfo) {
  const size = getToolSize(tool.id)
  
  message.loading({ content: '正在预览...', key: 'preview' })
  try {
    const previewItems = await tauriApi.previewTool(tool.id, tool.paths)
    message.loading({ content: '', key: 'preview' })
    
    if (previewItems.length === 0) {
      message.warning('没有找到可清理的缓存')
      return
    }
    
    const previewContent = previewItems.map(item => 
      `${item.path}\n${item.fileNum} 个文件 · ${formatSize(item.size)}`
    ).join('\n\n')
    
    Modal.confirm({
      title: `确认清理 ${tool.name}`,
      content: `确定要清理以下缓存吗？这将释放约 ${formatSize(size)} 磁盘空间。\n\n${previewContent}`,
      okText: '确认清理',
      okType: 'danger',
      cancelText: '取消',
      async onOk() {
        try {
          await toolStore.cleanTool(tool.id, tool.paths)
          message.success(`${tool.name} 清理完成`)
          trackCleanComplete(tool.id, size)
        } catch (e) {
          message.error('清理失败')
        }
      }
    })
  } catch (e) {
    message.loading({ content: '', key: 'preview' })
    error.value = e instanceof Error ? e.message : '预览失败'
    message.error('预览失败')
  }
}

function openToolPath(tool: ToolInfo) {
  if (tool.paths.length > 0) {
    toolStore.openPath(tool.paths[0]).catch(() => {
      message.error('无法打开路径')
    })
  }
}

function openPath(path: string) {
  toolStore.openPath(path).catch(() => {
    message.error('无法打开路径')
  })
}

function openSettings() {
  router.push('/settings')
}

// v0.2.0 新增导航方法
function goToProjects() {
  trackPageView('ProjectCleanView')
  router.push('/projects')
}

function goToHistory() {
  trackPageView('HistoryView')
  router.push('/history')
}

function goToAnalysis() {
  trackPageView('DiskAnalysisView')
  router.push('/analysis')
}

// 快速清理所有
async function quickCleanAll() {
  const toolsWithCache = enabledTools.value.filter(t => getToolSize(t.id) > 0)
  if (toolsWithCache.length === 0) {
    message.warning('没有可清理的缓存')
    return
  }

  Modal.confirm({
    title: '确认清理',
    content: `确定要清理 ${toolsWithCache.length} 个工具的缓存吗？这将释放约 ${formatSize(totalCacheSize.value)} 磁盘空间。`,
    okText: '确认清理',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      for (const tool of toolsWithCache) {
        try {
          await toolStore.cleanTool(tool.id, tool.paths)
        } catch (e) {
          // 忽略单个错误
        }
      }
      message.success('清理完成')
      recommendations.value = []
    }
  })
}

// 获取统计数据
async function fetchStats() {
  try {
    const usageStats = await getUsageStats()
    stats.value = {
      totalCleaned: usageStats.totalCleaned,
      cleanCount: usageStats.cleanCount,
    }
  } catch (e) {
    // ignore
  }
}

async function fetchDiskUsage() {
  try {
    const usage = await getDiskUsage()
    diskUsage.value = usage
  } catch (error) {
    // eslint-disable-next-line no-console
    console.warn('获取磁盘使用情况失败:', error)
  }
}

onMounted(async () => {
  // 获取版本信息
  try {
    const { version: ver, build } = await getVersion()
    version.value = ver
    buildType.value = build
  } catch (error) {
    version.value = '0.1.0'
    // eslint-disable-next-line no-console
    console.warn('获取版本失败:', error)
  }

  // 追踪页面浏览
  trackAppStarted()
  trackPageView('HomeView')

  await toolStore.fetchTools()
  await fetchDiskUsage()
  await fetchStats()
  await checkAutoScan()

  diskRefreshTimer.value = setInterval(async () => {
    await fetchDiskUsage()
  }, 30000)

  // 监听快捷键事件
  window.addEventListener('devcleaner:scan', handleShortcutScan as ShortcutHandler)
  window.addEventListener('devcleaner:refresh', handleShortcutRefresh as ShortcutHandler)
})

// 快捷键处理：开始扫描
function handleShortcutScan() {
  if (!isScanning.value) {
    startScan()
  }
}

// 快捷键处理：刷新
function handleShortcutRefresh() {
  refreshTools()
}

type ShortcutHandler = (event: Event) => void

async function checkAutoScan() {
  const settings = settingsStore.settings
  if (!settings.autoScan) return

  const lastScanKey = 'devcleaner:lastScan'
  let lastScan: string | null = null
  try {
    lastScan = localStorage.getItem(lastScanKey)
  } catch (e) {
    return
  }

  const now = Date.now()
  const intervalMs = settings.scanInterval * 24 * 60 * 60 * 1000

  if (!lastScan || (now - parseInt(lastScan)) > intervalMs) {
    message.info('自动扫描中...')
    await startScan()
    try {
      localStorage.setItem(lastScanKey, now.toString())
    } catch (e) {
      // ignore
    }
  }
}

async function generateRecommendations() {
  if (scanResults.value.length === 0) {
    recommendations.value = []
    return
  }

  const diskPct = diskUsage.value.total > 0 
    ? (diskUsage.value.used / diskUsage.value.total) * 100 
    : 0
  
  const threshold = settingsStore.settings.threshold * 1024 * 1024 * 1024
  
  let stats: UsageStats = { totalCleaned: 0, cleanCount: 0, lastClean: 0, cleanHistory: [] }
  try {
    stats = await getUsageStats()
  } catch (e) {
    // ignore
  }

  const recs: Recommendation[] = []
  
  for (const result of scanResults.value) {
    const tool = toolStore.tools.find(t => t.id === result.tool_id)
    if (!tool) continue

    let reason = ''
    
    if (diskPct > 80 && result.size > 100 * 1024 * 1024) {
      reason = `磁盘空间不足 (${diskPct.toFixed(0)}%)，推荐清理大缓存`
    } else if (result.size > threshold) {
      reason = `缓存超过阈值 (${formatSize(result.size)} > ${formatSize(threshold)})`
    } else if (result.last_modified > 0) {
      const daysSinceModified = (Date.now() - result.last_modified * 1000) / (1000 * 60 * 60 * 24)
      if (daysSinceModified > 30) {
        reason = `缓存超过 ${Math.floor(daysSinceModified)} 天未清理`
      }
    } else {
      const recentCleans = stats.cleanHistory.filter(h => 
        h.toolId === result.tool_id && 
        Date.now() - h.timestamp * 1000 < 7 * 24 * 60 * 60 * 1000
      ).length
      if (recentCleans >= 2 && result.size > 50 * 1024 * 1024) {
        reason = `频繁使用工具，建议定期清理`
      }
    }

    if (reason && result.size > 0) {
      recs.push({
        toolId: result.tool_id,
        toolName: tool.name,
        size: result.size,
        reason
      })
    }
  }

  recommendations.value = recs
    .sort((a, b) => b.size - a.size)
    .slice(0, 3)
}

async function quickClean(item: Recommendation) {
  const tool = toolStore.tools.find(t => t.id === item.toolId)
  if (tool) {
    await cleanTool(tool)
  }
}

onUnmounted(() => {
  if (diskRefreshTimer.value) {
    clearInterval(diskRefreshTimer.value)
    diskRefreshTimer.value = null
  }
  // 移除快捷键事件监听
  window.removeEventListener('devcleaner:scan', handleShortcutScan as ShortcutHandler)
  window.removeEventListener('devcleaner:refresh', handleShortcutRefresh as ShortcutHandler)
})
</script>

<style scoped>
.home {
  height: 100vh;
  position: relative;
  overflow: hidden;
}

/* 极光背景装饰 */
.hero-glow {
  position: fixed;
  width: 600px;
  height: 600px;
  border-radius: 50%;
  filter: blur(120px);
  opacity: 0.3;
  pointer-events: none;
  z-index: 0;
}

.hero-glow-1 {
  top: -200px;
  right: -100px;
  background: var(--aurora-primary);
}

.hero-glow-2 {
  bottom: -200px;
  left: -100px;
  background: var(--aurora-secondary);
}

.layout {
  height: 100%;
  background: transparent;
  position: relative;
  z-index: 1;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--aurora-bg-glass);
  backdrop-filter: blur(20px);
  padding: 0 16px;
  border-bottom: 1px solid var(--aurora-border);
  height: 56px;
  min-height: 56px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.logo-icon {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--aurora-gradient-hero);
  border-radius: var(--aurora-radius-md);
  color: white;
}

.header h1 {
  font-size: 18px;
  margin: 0;
  font-weight: 700;
  background: var(--aurora-text-gradient);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.version {
  font-size: 11px;
  color: var(--aurora-text-tertiary);
  background: var(--aurora-bg-glass);
  padding: 2px 6px;
  border-radius: var(--aurora-radius-sm);
  font-weight: 500;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.theme-toggle-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--aurora-bg-glass);
  border: 1px solid var(--aurora-border);
  border-radius: var(--aurora-radius-md);
  color: var(--aurora-text-secondary);
  cursor: pointer;
  transition: all var(--aurora-transition-fast);
}

.theme-toggle-btn:hover {
  color: var(--aurora-text-primary);
  border-color: var(--aurora-border-light);
}

.content {
  padding: 16px 24px;
  overflow-y: auto;
  max-height: calc(100vh - 56px);
}

/* 英雄区域 */
.hero-section {
  display: flex;
  flex-direction: row;
  gap: 32px;
  align-items: center;
  margin-bottom: 24px;
  overflow: hidden;
}

.hero-content {
  flex: 1;
  min-width: 0;
}

.hero-visual {
  flex-shrink: 0;
  display: flex;
  justify-content: center;
  align-items: center;
}

.hero-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: var(--aurora-bg-glass);
  border: 1px solid var(--aurora-border-light);
  border-radius: var(--aurora-radius-full);
  font-size: 12px;
  font-weight: 500;
  color: var(--aurora-secondary);
  margin-bottom: 10px;
}

.hero-title {
  font-size: 28px;
  font-weight: 800;
  line-height: 1.2;
  letter-spacing: -0.5px;
  margin-bottom: 10px;
}

.hero-description {
  font-size: 14px;
  color: var(--aurora-text-secondary);
  line-height: 1.5;
  max-width: 380px;
  margin-bottom: 16px;
}

.hero-actions {
  display: flex;
  gap: 10px;
}

/* 磁盘图表 */
.disk-chart-container {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 16px;
  overflow: hidden;
  width: 180px;
}

.disk-chart-wrapper {
  position: relative;
  width: 120px;
  height: 120px;
}

.disk-chart {
  width: 100%;
  height: 100%;
  transform: rotate(-90deg);
}

.progress-ring__circle {
  transition: stroke-dashoffset 1s ease-out;
}

.disk-chart-center {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.disk-icon {
  color: var(--aurora-primary);
}

.disk-percentage {
  font-size: 18px;
  font-weight: 700;
}

.disk-label {
  font-size: 10px;
  color: var(--aurora-text-tertiary);
}

.disk-stats {
  display: flex;
  gap: 20px;
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid var(--aurora-border);
}

.disk-stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.disk-stat-value {
  font-size: 13px;
  font-weight: 600;
}

.disk-stat-label {
  font-size: 11px;
  color: var(--aurora-text-tertiary);
}

/* 扫描进度 */
.scan-progress-container {
  padding: 16px 20px;
  margin-bottom: 24px;
}

.scan-progress-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.scan-icon {
  color: var(--aurora-primary);
}

.scan-icon.spinning {
  animation: spin 1s linear infinite;
}

.scan-title {
  flex: 1;
  font-size: 14px;
  color: var(--aurora-text-secondary);
}

.scan-percent {
  font-weight: 700;
  font-size: 14px;
}

/* 扫描摘要 */
.scan-summary {
  display: inline-block;
  padding: 8px 16px;
  margin-bottom: 24px;
  font-size: 14px;
  color: var(--aurora-text-primary);
}

/* 统计卡片 (v0.2.0 新增) */
.summary-section {
  margin-bottom: 16px;
}

.stats-cards {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 12px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
}

.stat-icon-wrapper {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--aurora-bg-glass);
  border-radius: var(--aurora-radius-md);
}

.stat-icon {
  font-size: 20px;
  color: var(--aurora-primary);
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 18px;
  font-weight: 700;
  color: var(--aurora-text-primary);
}

.stat-label {
  font-size: 12px;
  color: var(--aurora-text-tertiary);
  margin-top: 2px;
}

/* 快速操作 (v0.2.0 新增) */
.quick-actions {
  margin-bottom: 24px;
  padding: 16px;
}

.quick-actions-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  font-weight: 600;
  color: var(--aurora-primary);
}

.quick-actions-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.quick-action-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 16px 12px;
  background: var(--aurora-bg-card);
  border: 1px solid var(--aurora-border);
  border-radius: var(--aurora-radius-md);
  color: var(--aurora-text-secondary);
  cursor: pointer;
  transition: all var(--aurora-transition-fast);
}

.quick-action-btn:hover:not(:disabled) {
  background: var(--aurora-bg-glass);
  border-color: var(--aurora-primary);
  color: var(--aurora-primary);
  transform: translateY(-2px);
}

.quick-action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.quick-action-btn span {
  font-size: 13px;
  font-weight: 500;
}

/* 推荐卡片 */
.recommendation-card {
  margin-bottom: 24px;
}

.recommendation-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--aurora-primary);
  font-weight: 600;
}

/* 工具区域标题 */
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
}

.section-title {
  font-size: 24px;
  font-weight: 700;
  margin-bottom: 4px;
  background: var(--aurora-text-gradient);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.section-subtitle {
  color: var(--aurora-text-tertiary);
  font-size: 14px;
}

/* 工具卡片 */
.tools-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 16px;
}

.tool-card {
  position: relative;
  padding: 20px;
  background: var(--aurora-bg-card);
  backdrop-filter: blur(20px);
  border: 1px solid var(--aurora-border);
  border-radius: var(--aurora-radius-lg);
  cursor: pointer;
  transition: all var(--aurora-transition-normal);
}

.tool-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: var(--aurora-gradient-hero);
  opacity: 0;
  transition: opacity var(--aurora-transition-normal);
  border-radius: var(--aurora-radius-lg) var(--aurora-radius-lg) 0 0;
}

.tool-card:hover {
  transform: translateY(-4px);
  border-color: var(--aurora-border-light);
  box-shadow: var(--aurora-shadow-card), 0 0 30px var(--aurora-primary-glow);
}

.tool-card:hover::before {
  opacity: 1;
}

.tool-card.has-cache {
  border-color: var(--aurora-primary);
  box-shadow: var(--aurora-shadow-card), var(--aurora-shadow-glow);
}

.tool-card.has-cache::before {
  opacity: 1;
}

.tool-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
}

.tool-icon-wrapper {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--aurora-bg-glass);
  border-radius: var(--aurora-radius-md);
  color: var(--aurora-primary);
}

.tool-info {
  text-align: center;
  padding: 12px 0;
}

.tool-name {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 4px;
  color: var(--aurora-text-primary);
}

.tool-size {
  font-size: 14px;
  color: var(--aurora-text-tertiary);
}

.tool-size.has-size {
  color: var(--aurora-primary);
  font-weight: 600;
}

.tool-paths {
  font-size: 12px;
  color: var(--aurora-text-tertiary);
  margin-top: 4px;
}

.tool-footer {
  display: flex;
  justify-content: center;
  gap: 8px;
  padding-top: 12px;
  border-top: 1px solid var(--aurora-border);
}

/* 页脚 */
.footer {
  text-align: center;
  background: var(--aurora-bg-glass);
  border-top: 1px solid var(--aurora-border);
  padding: 12px 24px;
  color: var(--aurora-text-tertiary);
}

/* 响应式 */
@media (max-width: 768px) {
  .content {
    padding: 12px;
  }

  .hero-section {
    gap: 16px;
  }

  .hero-title {
    font-size: 22px;
  }

  .hero-description {
    font-size: 13px;
    max-width: 280px;
  }

  .disk-chart-container {
    width: 140px;
    padding: 12px;
  }

  .disk-chart-wrapper {
    width: 100px;
    height: 100px;
  }

  .disk-percentage {
    font-size: 16px;
  }

  .tools-grid {
    grid-template-columns: 1fr;
  }

  .stats-cards {
    grid-template-columns: 1fr;
  }

  .quick-actions-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>

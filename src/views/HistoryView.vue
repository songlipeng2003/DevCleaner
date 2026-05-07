<template>
  <div class="history-view">
    <!-- 极光背景装饰 -->
    <div class="hero-glow hero-glow-1" />
    <div class="hero-glow hero-glow-2" />

    <a-layout class="layout">
      <a-layout-header class="header">
        <div class="header-left">
          <button class="back-btn" @click="goBack">
            <ArrowLeft :size="20" />
          </button>
          <div class="logo-icon">
            <History :size="20" />
          </div>
          <h1>清理历史</h1>
        </div>
        <div class="header-right">
          <a-dropdown>
            <a-button>
              <template #icon>
                <DownloadOutlined />
              </template>
              导出报告
            </a-button>
            <template #overlay>
              <a-menu @click="handleExport">
                <a-menu-item key="json">JSON 格式</a-menu-item>
                <a-menu-item key="csv">CSV 格式</a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
          <a-button @click="openSettings">
            <template #icon>
              <SettingOutlined />
            </template>
            设置
          </a-button>
        </div>
      </a-layout-header>

      <a-layout-content class="content">
        <!-- 统计卡片 -->
        <div class="stats-grid">
          <a-card class="stat-card glass-card">
            <div class="stat-icon">
              <DatabaseOutlined />
            </div>
            <div class="stat-content">
              <div class="stat-value gradient-text">{{ formatSize(stats.totalCleaned) }}</div>
              <div class="stat-label">总计已清理</div>
            </div>
          </a-card>

          <a-card class="stat-card glass-card">
            <div class="stat-icon">
              <CheckCircleOutlined />
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ stats.totalCount }}</div>
              <div class="stat-label">清理次数</div>
            </div>
          </a-card>

          <a-card class="stat-card glass-card">
            <div class="stat-icon">
              <CalendarOutlined />
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ lastCleanDate }}</div>
              <div class="stat-label">上次清理</div>
            </div>
          </a-card>

          <a-card class="stat-card glass-card">
            <div class="stat-icon">
              <BarChartOutlined />
            </div>
            <div class="stat-content">
              <div class="stat-value">{{ avgSizePerClean }}</div>
              <div class="stat-label">平均每次清理</div>
            </div>
          </a-card>
        </div>

        <!-- 筛选器 -->
        <a-card class="filter-card glass-card">
          <a-radio-group v-model:value="filterType" button-style="solid" @change="handleFilterChange">
            <a-radio-button value="all">全部</a-radio-button>
            <a-radio-button value="day">今天</a-radio-button>
            <a-radio-button value="week">本周</a-radio-button>
            <a-radio-button value="month">本月</a-radio-button>
          </a-radio-group>
        </a-card>

        <!-- 月度趋势图 -->
        <a-card class="trend-card glass-card" v-if="monthlyStats.length > 0">
          <template #title>
            <div class="card-title">
              <LineChartOutlined />
              <span>清理趋势</span>
            </div>
          </template>
          <div class="trend-chart">
            <div
              v-for="stat in monthlyStats"
              :key="stat.month"
              class="trend-bar"
            >
              <div
                class="trend-bar-fill"
                :style="{ height: getBarHeight(stat.cleaned) + '%' }"
              />
              <div class="trend-bar-label">{{ stat.month }}</div>
              <div class="trend-bar-value">{{ formatSize(stat.cleaned) }}</div>
            </div>
          </div>
        </a-card>

        <!-- 历史记录列表 -->
        <a-card class="history-card glass-card">
          <template #title>
            <div class="card-title">
              <TableOutlined />
              <span>清理记录</span>
            </div>
          </template>

          <a-spin :spinning="isLoading">
            <a-table
              :columns="columns"
              :data-source="historyItems"
              :pagination="{ pageSize: 10 }"
              row-key="id"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'toolName'">
                  <div class="tool-cell">
                    <component :is="getToolIcon(record.toolId)" :size="16" style="margin-right: 8px;" />
                    {{ record.toolName }}
                  </div>
                </template>
                <template v-else-if="column.key === 'size'">
                  <span class="size-cell">{{ formatSize(record.size) }}</span>
                </template>
                <template v-else-if="column.key === 'timestamp'">
                  {{ formatDate(record.timestamp) }}
                </template>
                <template v-else-if="column.key === 'paths'">
                  <a-tooltip>
                    <template #title>
                      <div v-for="(path, idx) in record.paths" :key="idx" class="path-item">
                        {{ path }}
                      </div>
                    </template>
                    <a-button type="link" size="small">
                      {{ record.paths.length }} 个路径
                    </a-button>
                  </a-tooltip>
                </template>
              </template>
            </a-table>
          </a-spin>

          <a-empty v-if="!isLoading && historyItems.length === 0" description="暂无清理记录" />
        </a-card>
      </a-layout-content>
    </a-layout>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  SettingOutlined,
  DownloadOutlined,
  DatabaseOutlined,
  CheckCircleOutlined,
  CalendarOutlined,
  BarChartOutlined,
  LineChartOutlined,
  TableOutlined,
} from '@ant-design/icons-vue'
import {
  History,
  HardDrive,
  Package,
  Box,
  ArrowLeft,
} from 'lucide-vue-next'
import * as tauriApi from '@/services/tauri'
import type { CleanHistory, CleanHistoryItemV2, MonthlyStat } from '@/services/tauri'

const router = useRouter()

// 状态
const isLoading = ref(false)
const filterType = ref('all')
const historyItems = ref<CleanHistoryItemV2[]>([])
const monthlyStats = ref<MonthlyStat[]>([])
const stats = ref({
  totalCleaned: 0,
  totalCount: 0,
  lastClean: 0,
  avgSize: 0,
})

// 表格列定义
const columns = [
  { title: '时间', key: 'timestamp', dataIndex: 'timestamp' },
  { title: '工具', key: 'toolName', dataIndex: 'toolName' },
  { title: '清理大小', key: 'size', dataIndex: 'size' },
  { title: '文件数', key: 'fileNum', dataIndex: 'fileNum' },
  { title: '路径', key: 'paths', dataIndex: 'paths' },
]

// 计算属性
const lastCleanDate = computed(() => {
  if (stats.value.lastClean === 0) return '暂无记录'
  return formatDate(stats.value.lastClean)
})

const avgSizePerClean = computed(() => {
  if (stats.value.totalCount === 0) return '0 B'
  const avg = stats.value.totalCleaned / stats.value.totalCount
  return formatSize(avg)
})

// 格式化
function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

function formatDate(timestamp: number): string {
  const date = new Date(timestamp * 1000)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))

  if (days === 0) return '今天'
  if (days === 1) return '昨天'
  if (days < 7) return `${days} 天前`

  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  })
}

function getBarHeight(size: number): number {
  if (monthlyStats.value.length === 0) return 0
  const max = Math.max(...monthlyStats.value.map(s => s.cleaned))
  if (max === 0) return 0
  return Math.round((size / max) * 100)
}

// 获取工具图标
function getToolIcon(toolId: string) {
  switch (toolId) {
    case 'npm':
    case 'yarn':
    case 'pnpm':
    case 'bun':
      return Package
    case 'cargo':
    case 'rust':
      return Box
    default:
      return HardDrive
  }
}

// 获取历史数据
async function fetchHistory() {
  console.log('fetchHistory called, filter:', filterType.value)
  isLoading.value = true
  try {
    console.log('Calling tauriApi.getCleanHistory...')
    const result: CleanHistory = await tauriApi.getCleanHistory(filterType.value as any)
    console.log('getCleanHistory result:', result)
    historyItems.value = result.items
    monthlyStats.value = result.monthlyStats
    stats.value.totalCleaned = result.totalCleaned
    stats.value.totalCount = result.totalCount
    stats.value.lastClean = result.items.length > 0 ? result.items[0].timestamp : 0
  } catch (error) {
    console.error('获取历史记录失败:', error)
    message.error('获取历史记录失败')
  } finally {
    isLoading.value = false
  }
}

// 筛选
function handleFilterChange() {
  fetchHistory()
}

// 导出
async function handleExport({ key }: { key: string }) {
  try {
    const format = key as 'json' | 'csv'
    const content = await tauriApi.exportCleanReport(format)

    // 创建下载
    const blob = new Blob([content], { type: format === 'json' ? 'application/json' : 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `clean-report.${format}`
    a.click()
    URL.revokeObjectURL(url)

    message.success('导出成功')
  } catch (error) {
    message.error('导出失败')
  }
}

// 导航
function goBack() {
  router.push('/')
}

function openSettings() {
  router.push('/settings')
}

// 初始化
onMounted(() => {
  console.log('HistoryView mounted, fetching history...')
  try {
    fetchHistory()
  } catch (error) {
    console.error('HistoryView 初始化失败:', error)
    message.error('页面加载失败')
  }
})
</script>

<style scoped>
.history-view {
  height: 100vh;
  position: relative;
  overflow: hidden;
}

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
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.back-btn {
  width: 36px;
  height: 36px;
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

.back-btn:hover {
  color: var(--aurora-text-primary);
  border-color: var(--aurora-border-light);
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

.header-right {
  display: flex;
  gap: 8px;
}

.content {
  padding: 16px 24px;
  overflow-y: auto;
  max-height: calc(100vh - 56px);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 16px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
}

.stat-icon {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--aurora-bg-glass);
  border-radius: var(--aurora-radius-md);
  font-size: 24px;
  color: var(--aurora-primary);
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
}

.stat-value.gradient-text {
  background: var(--aurora-text-gradient);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.stat-label {
  font-size: 12px;
  color: var(--aurora-text-tertiary);
  margin-top: 4px;
}

.filter-card {
  margin-bottom: 16px;
}

.trend-card {
  margin-bottom: 16px;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.trend-chart {
  display: flex;
  justify-content: space-around;
  align-items: flex-end;
  height: 200px;
  padding: 16px 0;
}

.trend-bar {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 60px;
  height: 100%;
}

.trend-bar-fill {
  width: 40px;
  background: var(--aurora-gradient-hero);
  border-radius: var(--aurora-radius-sm) var(--aurora-radius-sm) 0 0;
  min-height: 4px;
  transition: height 0.3s ease;
}

.trend-bar-label {
  font-size: 11px;
  color: var(--aurora-text-tertiary);
  margin-top: 8px;
}

.trend-bar-value {
  font-size: 10px;
  color: var(--aurora-text-secondary);
  margin-top: 4px;
}

.history-card {
  margin-bottom: 16px;
}

.tool-cell {
  display: flex;
  align-items: center;
}

.size-cell {
  color: var(--aurora-primary);
  font-weight: 600;
}

.path-item {
  padding: 2px 0;
  word-break: break-all;
  font-size: 12px;
}

@media (max-width: 1024px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>

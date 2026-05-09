<template>
  <div class="disk-analysis-view">
    <!-- 极光背景装饰 -->
    <div class="hero-glow hero-glow-1" />
    <div class="hero-glow hero-glow-2" />

    <a-layout class="layout">
      <a-layout-header class="header">
        <div class="header-left">
          <button
            class="back-btn"
            @click="goBack"
          >
            <ArrowLeft :size="20" />
          </button>
          <div class="logo-icon">
            <PieChart :size="20" />
          </div>
          <h1>磁盘使用分析</h1>
        </div>
        <div class="header-right">
          <a-button
            :loading="loading"
            @click="refreshAnalysis"
          >
            <template #icon>
              <ReloadOutlined />
            </template>
            刷新
          </a-button>
        </div>
      </a-layout-header>

      <a-layout-content class="content">
        <!-- 加载状态 -->
        <div
          v-if="loading"
          class="loading-state"
        >
          <a-spin
            size="large"
            tip="正在分析磁盘使用情况..."
          />
        </div>

        <!-- 分析结果 -->
        <template v-else-if="analysisData">
          <!-- 概览卡片 -->
          <a-row
            :gutter="16"
            class="overview-row"
          >
            <a-col
              :xs="24"
              :sm="12"
              :md="6"
            >
              <a-card class="stat-card glass-card">
                <a-statistic
                  title="总占用空间"
                  :value="formatSize(analysisData.totalSize)"
                  :value-style="{ color: '#1890ff' }"
                >
                  <template #prefix>
                    <DatabaseOutlined />
                  </template>
                </a-statistic>
              </a-card>
            </a-col>
            <a-col
              :xs="24"
              :sm="12"
              :md="6"
            >
              <a-card class="stat-card glass-card">
                <a-statistic
                  title="可清理空间"
                  :value="formatSize(analysisData.cleanableSize)"
                  :value-style="{ color: '#52c41a' }"
                >
                  <template #prefix>
                    <DeleteOutlined />
                  </template>
                </a-statistic>
              </a-card>
            </a-col>
            <a-col
              :xs="24"
              :sm="12"
              :md="6"
            >
              <a-card class="stat-card glass-card">
                <a-statistic
                  title="分析目录数"
                  :value="analysisData.analyzedPaths"
                >
                  <template #prefix>
                    <FolderOutlined />
                  </template>
                </a-statistic>
              </a-card>
            </a-col>
            <a-col
              :xs="24"
              :sm="12"
              :md="6"
            >
              <a-card class="stat-card glass-card">
                <a-statistic
                  title="清理比例"
                  :value="cleanablePercentage + '%'"
                  :value-style="{ color: cleanablePercentage > 50 ? '#faad14' : '#52c41a' }"
                >
                  <template #prefix>
                    <PieChartOutlined />
                  </template>
                </a-statistic>
              </a-card>
            </a-col>
          </a-row>

          <!-- 分类分析 -->
          <a-card class="categories-card glass-card">
            <template #title>
              <Space>
                <PieChartOutlined />
                <span>分类占用分析</span>
              </Space>
            </template>

            <a-collapse
              v-model:active-key="activeCategories"
              accordion
            >
              <a-collapse-panel
                v-for="category in sortedCategories"
                :key="category.name"
                :header="`${category.name} - ${formatSize(category.totalSize)}`"
              >
                <template #extra>
                  <a-tag :color="getCategoryColor(category.totalSize)">
                    {{ formatSize(category.totalSize) }}
                  </a-tag>
                </template>

                <a-table
                  :data-source="category.items"
                  :pagination="false"
                  size="small"
                >
                  <a-table-column
                    title="名称"
                    data-index="name"
                  />
                  <a-table-column
                    title="路径"
                    data-index="path"
                  />
                  <a-table-column
                    title="大小"
                    data-index="size"
                  >
                    <template #default="{ record }">
                      {{ formatSize(record.size) }}
                    </template>
                  </a-table-column>
                  <a-table-column
                    title="占比"
                    data-index="percentage"
                    width="100"
                  >
                    <template #default="{ record }">
                      <a-progress
                        :percent="record.percentage"
                        :stroke-color="record.isCleanable ? '#52c41a' : '#1890ff'"
                        size="small"
                      />
                    </template>
                  </a-table-column>
                  <a-table-column
                    title="状态"
                    data-index="isCleanable"
                    width="100"
                  >
                    <template #default="{ record }">
                      <a-tag :color="record.isCleanable ? 'green' : 'blue'">
                        {{ record.isCleanable ? '可清理' : '保留' }}
                      </a-tag>
                    </template>
                  </a-table-column>
                  <a-table-column
                    title="操作"
                    width="120"
                  >
                    <template #default="{ record }">
                      <a-space>
                        <a-tooltip :title="record.path">
                          <a-button
                            size="small"
                            @click="openPath(record.path)"
                          >
                            <FolderOpen :size="14" />
                          </a-button>
                        </a-tooltip>
                        <a-tooltip title="打开目录">
                          <a-button
                            size="small"
                            type="primary"
                            ghost
                            @click="openInExplorer(record.path)"
                          >
                            <ExternalLink :size="14" />
                          </a-button>
                        </a-tooltip>
                      </a-space>
                    </template>
                  </a-table-column>
                </a-table>
              </a-collapse-panel>
            </a-collapse>
          </a-card>

          <!-- 缓存趋势 -->
          <a-card class="trends-card glass-card">
            <template #title>
              <Space>
                <LineChartOutlined />
                <span>缓存趋势</span>
              </Space>
            </template>

            <div
              v-if="trendData.length > 0"
              class="trend-chart"
            >
              <div class="trend-bars">
                <div
                  v-for="trend in trendData"
                  :key="trend.date"
                  class="trend-bar-container"
                >
                  <div
                    class="trend-bar"
                    :style="{ height: getBarHeight(trend.size) + '%' }"
                  >
                    <span class="trend-value">{{ formatSize(trend.size) }}</span>
                  </div>
                  <span class="trend-date">{{ trend.date }}</span>
                </div>
              </div>
            </div>
            <a-empty
              v-else
              description="暂无趋势数据"
            />
          </a-card>
        </template>

        <!-- 空状态 -->
        <a-empty
          v-else
          description="暂无分析数据"
        />
      </a-layout-content>
    </a-layout>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  FolderOutlined,
  ReloadOutlined,
  DatabaseOutlined,
  DeleteOutlined,
  PieChartOutlined,
  LineChartOutlined,
} from '@ant-design/icons-vue'
import {
  ArrowLeft,
  PieChart,
  FolderOpen,
  ExternalLink,
} from 'lucide-vue-next'
import {
  getDiskAnalysis,
  getCacheTrends,
  type DiskAnalysisSummary,
  type CacheTrend,
} from '@/services/tauri'

const router = useRouter()

// 状态
const loading = ref(false)
const analysisData = ref<DiskAnalysisSummary | null>(null)
const trendData = ref<CacheTrend[]>([])
const activeCategories = ref<string[]>([])

// 计算属性
const cleanablePercentage = computed(() => {
  if (!analysisData.value || analysisData.value.totalSize === 0) return 0
  return Math.round(
    (analysisData.value.cleanableSize / analysisData.value.totalSize) * 100
  )
})

const sortedCategories = computed(() => {
  if (!analysisData.value) return []
  return [...analysisData.value.categories].sort(
    (a, b) => b.totalSize - a.totalSize
  )
})

// 方法
const formatSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const getCategoryColor = (size: number): string => {
  const maxSize = sortedCategories.value[0]?.totalSize || 1
  const ratio = size / maxSize
  if (ratio > 0.5) return 'red'
  if (ratio > 0.3) return 'orange'
  return 'blue'
}

const getBarHeight = (size: number): number => {
  if (trendData.value.length === 0) return 0
  const maxSize = Math.max(...trendData.value.map((t) => t.size))
  if (maxSize === 0) return 0
  return Math.max(10, (size / maxSize) * 100)
}

const goBack = () => {
  router.push('/')
}

const refreshAnalysis = async () => {
  await loadAnalysis()
}

const loadAnalysis = async () => {
  loading.value = true
  try {
    const [analysis, trends] = await Promise.all([
      getDiskAnalysis(),
      getCacheTrends(),
    ])
    analysisData.value = analysis
    trendData.value = trends

    // 默认展开第一个分类
    if (analysis.categories.length > 0) {
      activeCategories.value = [analysis.categories[0].name]
    }
  } catch (error) {
    console.error('Failed to load analysis:', error)
    message.error('加载分析数据失败')
  } finally {
    loading.value = false
  }
}

const openPath = (path: string) => {
  // 复制路径到剪贴板
  navigator.clipboard.writeText(path)
  message.success('路径已复制到剪贴板')
}

const openInExplorer = async (path: string) => {
  try {
    const { openPath: openPathApi } = await import('@/services/tauri')
    await openPathApi(path)
  } catch (error) {
    console.error('Failed to open path:', error)
    message.error('打开目录失败')
  }
}

// 生命周期
onMounted(() => {
  loadAnalysis()
})
</script>

<style scoped>
.disk-analysis-view {
  min-height: 100vh;
  background: linear-gradient(135deg, #0a0e27 0%, #1a1f4e 100%);
  position: relative;
  overflow: hidden;
}

.hero-glow {
  position: absolute;
  width: 600px;
  height: 600px;
  border-radius: 50%;
  filter: blur(120px);
  opacity: 0.15;
  pointer-events: none;
}

.hero-glow-1 {
  top: -200px;
  left: -200px;
  background: #1890ff;
}

.hero-glow-2 {
  bottom: -200px;
  right: -200px;
  background: #722ed1;
}

.layout {
  min-height: 100vh;
  background: transparent;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(255, 255, 255, 0.03);
  backdrop-filter: blur(20px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  padding: 0 24px;
  height: 64px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-left h1 {
  color: #fff;
  font-size: 20px;
  font-weight: 600;
  margin: 0;
}

.back-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.05);
  color: #fff;
  cursor: pointer;
  transition: all 0.3s;
}

.back-btn:hover {
  background: rgba(255, 255, 255, 0.1);
}

.logo-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, #1890ff 0%, #722ed1 100%);
  border-radius: 10px;
  color: #fff;
}

.content {
  padding: 24px;
  position: relative;
  z-index: 1;
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 400px;
}

.overview-row {
  margin-bottom: 24px;
}

.stat-card {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  transition: all 0.3s;
}

.stat-card:hover {
  transform: translateY(-2px);
  border-color: rgba(255, 255, 255, 0.1);
}

.categories-card,
.trends-card {
  margin-bottom: 24px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 12px;
}

.trend-chart {
  padding: 20px 0;
}

.trend-bars {
  display: flex;
  align-items: flex-end;
  justify-content: space-around;
  height: 200px;
  padding: 0 20px;
}

.trend-bar-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
  max-width: 80px;
}

.trend-bar {
  width: 40px;
  background: linear-gradient(180deg, #1890ff 0%, #722ed1 100%);
  border-radius: 8px 8px 0 0;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 8px;
  transition: height 0.5s ease;
  min-height: 20px;
}

.trend-value {
  font-size: 10px;
  color: #fff;
  white-space: nowrap;
}

.trend-date {
  margin-top: 8px;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.6);
}

.glass-card {
  background: rgba(255, 255, 255, 0.03) !important;
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.05);
}

:deep(.ant-card-head) {
  border-bottom-color: rgba(255, 255, 255, 0.05);
}

:deep(.ant-collapse-header) {
  color: rgba(255, 255, 255, 0.85) !important;
}

:deep(.ant-table) {
  background: transparent;
}

:deep(.ant-table-thead > tr > th) {
  background: rgba(255, 255, 255, 0.05);
  color: rgba(255, 255, 255, 0.85);
  border-bottom-color: rgba(255, 255, 255, 0.05);
}

:deep(.ant-table-tbody > tr > td) {
  color: rgba(255, 255, 255, 0.65);
  border-bottom-color: rgba(255, 255, 255, 0.03);
}

:deep(.ant-table-tbody > tr:hover > td) {
  background: rgba(255, 255, 255, 0.02);
}

:deep(.ant-empty-description) {
  color: rgba(255, 255, 255, 0.45);
}
</style>

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
                  :value-style="{ color: 'var(--aurora-primary)' }"
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
          <div class="section-header">
            <div class="section-title-group">
              <h2 class="section-title">
                分类占用分析
              </h2>
              <p class="section-subtitle">
                点击展开查看各类别的详细占用情况
              </p>
            </div>
          </div>

          <a-collapse
            v-model:active-key="activeCategories"
            class="categories-collapse"
          >
            <a-collapse-panel
              v-for="category in sortedCategories"
              :key="category.name"
              :header="category.name"
            >
              <template #extra>
                <a-tag :color="getCategoryColor(category.totalSize)">
                  {{ formatSize(category.totalSize) }}
                </a-tag>
              </template>

              <div class="category-info">
                <div class="category-bar">
                  <div
                    class="category-bar-fill"
                    :style="{ width: (category.totalSize / sortedCategories[0].totalSize * 100) + '%' }"
                  />
                </div>
                <div class="category-details">
                  <div
                    v-for="item in category.items"
                    :key="item.path"
                    class="detail-item"
                  >
                    <div class="detail-header">
                      <span class="detail-name">{{ item.name }}</span>
                      <span class="detail-size">{{ formatSize(item.size) }}</span>
                    </div>
                    <div
                      class="detail-path"
                      :title="item.path"
                      @click="copyPath(item.path)"
                    >
                      <FolderOpen :size="12" />
                      <span>{{ item.path }}</span>
                    </div>
                    <div class="detail-footer">
                      <a-tag
                        :color="item.isCleanable ? 'green' : 'blue'"
                        size="small"
                      >
                        {{ item.isCleanable ? '可清理' : '保留' }}
                      </a-tag>
                      <a-button
                        type="text"
                        size="small"
                        @click="openInExplorer(item.path)"
                      >
                        <ExternalLink :size="14" />
                        打开
                      </a-button>
                    </div>
                  </div>
                </div>
              </div>
            </a-collapse-panel>
          </a-collapse>

          <!-- 缓存趋势 -->
          <div
            class="section-header"
            style="margin-top: 24px;"
          >
            <div class="section-title-group">
              <h2 class="section-title">
                缓存趋势
              </h2>
              <p class="section-subtitle">
                清理历史的趋势变化
              </p>
            </div>
          </div>

          <div class="trends-card glass-card">
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
          </div>
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

const copyPath = (path: string) => {
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
  min-height: 56px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.back-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: var(--aurora-radius-md);
  background: var(--aurora-bg-glass);
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
  align-items: center;
  gap: 10px;
}

.content {
  padding: 16px 24px;
  overflow-y: auto;
  max-height: calc(100vh - 56px);
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
  background: var(--aurora-bg-card);
  backdrop-filter: blur(20px);
  border: 1px solid var(--aurora-border);
  border-radius: var(--aurora-radius-lg);
  transition: all var(--aurora-transition-normal);
}

.stat-card:hover {
  transform: translateY(-4px);
  border-color: var(--aurora-border-light);
  box-shadow: var(--aurora-shadow-card), 0 0 30px var(--aurora-primary-glow);
}

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

.categories-collapse {
  background: var(--aurora-bg-card);
  border: 1px solid var(--aurora-border);
  border-radius: var(--aurora-radius-lg);
  overflow: hidden;
}

.categories-collapse :deep(.ant-collapse-item) {
  border-bottom: 1px solid var(--aurora-border);
}

.categories-collapse :deep(.ant-collapse-header) {
  padding: 16px 20px !important;
  background: var(--aurora-bg-glass);
  color: var(--aurora-text-primary) !important;
  font-weight: 600;
}

.categories-collapse :deep(.ant-collapse-content-box) {
  padding: 0 !important;
}

.category-info {
  padding: 16px 20px;
}

.category-bar {
  height: 8px;
  background: var(--aurora-bg-glass);
  border-radius: 4px;
  overflow: hidden;
  margin-bottom: 16px;
}

.category-bar-fill {
  height: 100%;
  background: var(--aurora-gradient-hero);
  border-radius: 4px;
  transition: width 0.5s ease;
}

.category-details {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.detail-item {
  background: var(--aurora-bg-glass);
  border: 1px solid var(--aurora-border);
  border-radius: var(--aurora-radius-md);
  padding: 12px 16px;
  transition: all var(--aurora-transition-fast);
}

.detail-item:hover {
  border-color: var(--aurora-border-light);
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.detail-name {
  font-weight: 600;
  color: var(--aurora-text-primary);
}

.detail-size {
  color: var(--aurora-primary);
  font-weight: 600;
}

.detail-path {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--aurora-text-tertiary);
  cursor: pointer;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 8px;
}

.detail-path:hover {
  color: var(--aurora-primary);
}

.detail-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.trends-card {
  background: var(--aurora-bg-card);
  border: 1px solid var(--aurora-border);
  border-radius: var(--aurora-radius-lg);
  padding: 20px;
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
  background: var(--aurora-gradient-hero);
  border-radius: var(--aurora-radius-md) var(--aurora-radius-md) 0 0;
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
  color: var(--aurora-text-tertiary);
}

/* 响应式 */
@media (max-width: 768px) {
  .content {
    padding: 12px;
  }

  .section-title {
    font-size: 20px;
  }

  .overview-row {
    margin-bottom: 16px;
  }
}
</style>

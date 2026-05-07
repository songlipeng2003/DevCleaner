<template>
  <div class="project-clean">
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
            <FolderSearch :size="20" />
          </div>
          <h1>项目级清理</h1>
        </div>
        <div class="header-right">
          <a-button @click="openSettings">
            <template #icon>
              <SettingOutlined />
            </template>
            设置
          </a-button>
        </div>
      </a-layout-header>

      <a-layout-content class="content">
        <!-- 配置区域 -->
        <a-card class="config-card glass-card">
          <a-form layout="vertical">
            <a-form-item label="扫描目录">
              <a-input-group compact>
                <a-input
                  v-model:value="scanPath"
                  placeholder="输入项目目录路径，如 ~/Projects"
                  style="width: calc(100% - 100px)"
                />
                <a-button
                  type="primary"
                  @click="selectFolder"
                >
                  <FolderOpen :size="16" />
                  选择
                </a-button>
              </a-input-group>
              <div class="config-tips">
                <InfoCircleOutlined /> 支持输入绝对路径或包含 ~ 的路径
              </div>
            </a-form-item>

            <a-row :gutter="16">
              <a-col :span="12">
                <a-form-item label="项目类型">
                  <a-checkbox-group v-model:value="selectedTypes">
                    <a-row>
                      <a-col
                        v-for="type in projectTypes"
                        :key="type.value"
                        :span="12"
                      >
                        <a-checkbox :value="type.value">
                          {{ type.label }}
                        </a-checkbox>
                      </a-col>
                    </a-row>
                  </a-checkbox-group>
                </a-form-item>
              </a-col>
              <a-col :span="12">
                <a-form-item label="扫描深度">
                  <a-slider
                    v-model:value="maxDepth"
                    :min="1"
                    :max="5"
                    :marks="depthMarks"
                  />
                </a-form-item>
              </a-col>
            </a-row>

            <div class="config-actions">
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
            </div>
          </a-form>
        </a-card>

        <!-- 扫描进度 -->
        <div
          v-if="isScanning"
          class="scan-progress-container glass-card"
        >
          <div class="scan-progress-header">
            <LoadingOutlined class="scan-icon spinning" />
            <span class="scan-title">正在扫描项目...</span>
          </div>
          <a-progress
            :percent="scanProgress"
            :show-info="false"
            :stroke-color="{ '0%': '#667eea', '100%': '#00d9ff' }"
          />
        </div>

        <!-- 扫描结果 -->
        <div
          v-if="!isScanning && projects.length > 0"
          class="results-section"
        >
          <div class="results-summary glass-card">
            <div class="summary-item">
              <span class="summary-label">发现项目</span>
              <span class="summary-value">{{ projects.length }} 个</span>
            </div>
            <div class="summary-item">
              <span class="summary-label">可清理空间</span>
              <span class="summary-value highlight">{{ formatSize(totalCleanableSize) }}</span>
            </div>
            <div class="summary-item">
              <span class="summary-label">高风险项目</span>
              <span
                class="summary-value"
                :class="{ 'risk': riskyProjects > 0 }"
              >{{ riskyProjects }} 个</span>
            </div>
          </div>

          <!-- 快速操作 -->
          <div class="quick-actions">
            <a-button @click="selectAll">
              <CheckOutlined /> 全选安全项目
            </a-button>
            <a-button @click="deselectAll">
              <StopOutlined /> 取消全选
            </a-button>
            <a-button
              type="primary"
              danger
              :disabled="selectedItems.length === 0"
              @click="cleanSelected"
            >
              <DeleteOutlined /> 清理选中 ({{ selectedItems.length }})
            </a-button>
          </div>

          <!-- 项目列表 -->
          <a-list
            :data-source="projects"
            :loading="isLoading"
            item-layout="horizontal"
            class="project-list"
          >
            <template #renderItem="{ item }">
              <a-list-item class="project-item glass-card">
                <template #actions>
                  <a-checkbox
                    :checked="isItemSelected(item)"
                    :disabled="item.cleanableItems.length === 0"
                    @change="toggleItem(item)"
                  />
                </template>
                <a-list-item-meta>
                  <template #avatar>
                    <div
                      class="project-icon"
                      :class="getProjectIconClass(item)"
                    >
                      <component
                        :is="getProjectIcon(item)"
                        :size="24"
                      />
                    </div>
                  </template>
                  <template #title>
                    <div class="project-title">
                      <span class="project-name">{{ item.name }}</span>
                      <a-tag :color="getRiskColor(item.riskLevel)">
                        {{ item.riskLevel === 'safe' ? '安全' : item.riskLevel === 'moderate' ? '谨慎' : '危险' }}
                      </a-tag>
                    </div>
                  </template>
                  <template #description>
                    <div class="project-info">
                      <span class="project-path">{{ item.path }}</span>
                      <span class="project-meta">
                        {{ item.projectType }} · {{ formatSize(item.size) }} · {{ item.fileNum }} 个文件
                      </span>
                    </div>
                  </template>
                </a-list-item-meta>

                <!-- 可清理项目 -->
                <div
                  v-if="item.cleanableItems.length > 0"
                  class="cleanable-items"
                >
                  <div class="cleanable-header">
                    <span class="cleanable-label">可清理项</span>
                    <span class="cleanable-size">{{ formatSize(item.size) }}</span>
                  </div>
                  <div class="cleanable-list">
                    <div
                      v-for="cleanable in item.cleanableItems"
                      :key="cleanable.id"
                      class="cleanable-item"
                    >
                      <Folder :size="14" />
                      <span class="cleanable-name">{{ cleanable.name }}</span>
                      <span class="cleanable-info">{{ formatSize(cleanable.size) }}</span>
                      <a-tooltip :title="cleanable.reason">
                        <InfoCircleOutlined class="cleanable-tip" />
                      </a-tooltip>
                    </div>
                  </div>
                </div>

                <div
                  v-else
                  class="no-cleanable"
                >
                  <CheckCircleOutlined /> 无可清理项
                </div>
              </a-list-item>
            </template>
          </a-list>
        </div>

        <!-- 空状态 -->
        <a-empty
          v-if="!isScanning && hasSearched && projects.length === 0"
          description="未找到可清理的项目"
        />

        <!-- 初始状态 -->
        <div
          v-if="!isScanning && !hasSearched"
          class="initial-state"
        >
          <div class="initial-icon">
            <FolderSearch :size="64" />
          </div>
          <h2>项目级清理</h2>
          <p>扫描指定目录下的开发项目，查找可清理的缓存文件</p>
          <div class="feature-list">
            <div class="feature-item">
              <Nodejs :size="20" />
              <span>Node.js (node_modules, dist)</span>
            </div>
            <div class="feature-item">
              <Cpu :size="20" />
              <span>Rust (target)</span>
            </div>
            <div class="feature-item">
              <Box :size="20" />
              <span>Python (__pycache__, .venv)</span>
            </div>
            <div class="feature-item">
              <Coffee :size="20" />
              <span>Java (Maven/Gradle)</span>
            </div>
          </div>
        </div>
      </a-layout-content>
    </a-layout>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import {
  SettingOutlined,
  ScanOutlined,
  ArrowLeft,
  FolderOpen,
  CheckOutlined,
  StopOutlined,
  DeleteOutlined,
  InfoCircleOutlined,
  CheckCircleOutlined,
  LoadingOutlined,
  Folder,
} from '@ant-design/icons-vue'
import {
  FolderSearch,
  Box,
  Coffee,
  Cpu,
} from 'lucide-vue-next'
import * as tauriApi from '@/services/tauri'
import type { ProjectScanResult } from '@/services/tauri'

const router = useRouter()

// 配置状态
const scanPath = ref('')
const selectedTypes = ref<string[]>(['nodejs', 'python', 'rust', 'java'])
const maxDepth = ref(3)
const depthMarks = { 1: '1层', 2: '2层', 3: '3层', 4: '4层', 5: '5层' }

// 项目类型
const projectTypes = [
  { value: 'nodejs', label: 'Node.js' },
  { value: 'python', label: 'Python' },
  { value: 'rust', label: 'Rust' },
  { value: 'go', label: 'Go' },
  { value: 'java', label: 'Java' },
  { value: 'dotnet', label: '.NET' },
  { value: 'ruby', label: 'Ruby' },
]

// 扫描状态
const isScanning = ref(false)
const isLoading = ref(false)
const hasSearched = ref(false)
const scanProgress = ref(0)
const projects = ref<ProjectScanResult[]>([])
const selectedItems = ref<Set<string>>(new Set())

// 计算属性
const totalCleanableSize = computed(() => {
  return projects.value.reduce((sum, p) => sum + p.size, 0)
})

const riskyProjects = computed(() => {
  return projects.value.filter(p => p.riskLevel !== 'safe').length
})

// 格式化大小
function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 获取项目图标
function getProjectIcon(item: ProjectScanResult) {
  switch (item.projectType) {
    case 'Node.js': return Box
    case 'Python': return Cpu
    case 'Rust': return Box
    case 'Go': return Coffee
    case 'Maven':
    case 'Gradle': return Coffee
    default: return Folder
  }
}

function getProjectIconClass(item: ProjectScanResult) {
  switch (item.projectType) {
    case 'Node.js': return 'icon-nodejs'
    case 'Python': return 'icon-python'
    case 'Rust': return 'icon-rust'
    case 'Go': return 'icon-go'
    default: return 'icon-default'
  }
}

function getRiskColor(risk: string) {
  switch (risk) {
    case 'safe': return 'green'
    case 'moderate': return 'orange'
    case 'careful': return 'red'
    default: return 'default'
  }
}

// 选择操作
function isItemSelected(item: ProjectScanResult) {
  return selectedItems.value.has(item.path)
}

function toggleItem(item: ProjectScanResult) {
  if (selectedItems.value.has(item.path)) {
    selectedItems.value.delete(item.path)
  } else {
    selectedItems.value.add(item.path)
  }
}

function selectAll() {
  projects.value.forEach(p => {
    if (p.cleanableItems.length > 0 && p.riskLevel === 'safe') {
      selectedItems.value.add(p.path)
    }
  })
  message.success('已选中所有安全项目')
}

function deselectAll() {
  selectedItems.value.clear()
  message.info('已取消全选')
}

// 开始扫描
async function startScan() {
  if (!scanPath.value.trim()) {
    message.warning('请输入扫描目录')
    return
  }

  isScanning.value = true
  hasSearched.value = true
  scanProgress.value = 0
  projects.value = []
  selectedItems.value.clear()

  try {
    // 模拟进度
    const progressInterval = setInterval(() => {
      if (scanProgress.value < 90) {
        scanProgress.value += Math.random() * 10
      }
    }, 200)

    const results = await tauriApi.scanProjects([scanPath.value], maxDepth.value)

    clearInterval(progressInterval)
    scanProgress.value = 100

    projects.value = results
    message.success(`扫描完成，发现 ${results.length} 个项目`)
  } catch (error) {
    message.error('扫描失败: ' + (error as Error).message)
  } finally {
    isScanning.value = false
  }
}

// 清理选中项目
async function cleanSelected() {
  if (selectedItems.value.size === 0) {
    message.warning('请先选择要清理的项目')
    return
  }

  const paths: string[] = []
  projects.value.forEach(p => {
    if (selectedItems.value.has(p.path)) {
      p.cleanableItems.forEach(item => {
        paths.push(item.path)
      })
    }
  })

  Modal.confirm({
    title: '确认清理',
    content: `确定要清理选中的 ${selectedItems.value.size} 个项目的缓存吗？这将释放约 ${formatSize(paths.reduce((sum, path) => {
      const project = projects.value.find(p => p.path === path)
      return sum + (project?.size || 0)
    }, 0))} 磁盘空间。`,
    okText: '确认清理',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await tauriApi.cleanPaths(paths)
        message.success('清理完成')

        // 移除已清理的项目
        projects.value = projects.value.filter(p => !selectedItems.value.has(p.path))
        selectedItems.value.clear()
      } catch (error) {
        message.error('清理失败: ' + (error as Error).message)
      }
    }
  })
}

// 选择文件夹
async function selectFolder() {
  // 简化版本：直接使用输入的路径
  if (!scanPath.value.trim()) {
    message.warning('请输入目录路径')
  } else {
    message.info('将扫描: ' + scanPath.value)
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
  // 如果用户有 Projects 目录，设置默认路径
  const defaultPath = process.env.HOME ? `${process.env.HOME}/Projects` : ''
  if (defaultPath) {
    scanPath.value = defaultPath
  }
})
</script>

<style scoped>
.project-clean {
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

.content {
  padding: 16px 24px;
  overflow-y: auto;
  max-height: calc(100vh - 56px);
}

.config-card {
  margin-bottom: 16px;
}

.config-tips {
  margin-top: 8px;
  font-size: 12px;
  color: var(--aurora-text-tertiary);
}

.config-actions {
  display: flex;
  justify-content: center;
  padding-top: 16px;
}

.scan-progress-container {
  padding: 16px 20px;
  margin-bottom: 16px;
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

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.scan-title {
  flex: 1;
  font-size: 14px;
  color: var(--aurora-text-secondary);
}

.results-summary {
  display: flex;
  justify-content: space-around;
  padding: 16px;
  margin-bottom: 16px;
}

.summary-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.summary-label {
  font-size: 12px;
  color: var(--aurora-text-tertiary);
}

.summary-value {
  font-size: 20px;
  font-weight: 700;
}

.summary-value.highlight {
  color: var(--aurora-primary);
}

.summary-value.risk {
  color: #faad14;
}

.quick-actions {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}

.project-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.project-item {
  padding: 16px;
}

.project-icon {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--aurora-radius-md);
  background: var(--aurora-bg-glass);
  color: var(--aurora-primary);
}

.project-icon.icon-nodejs { color: #3c873a; }
.project-icon.icon-python { color: #3776ab; }
.project-icon.icon-rust { color: #dea584; }
.project-icon.icon-go { color: #00add8; }

.project-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.project-name {
  font-weight: 600;
}

.project-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.project-path {
  font-size: 12px;
  color: var(--aurora-text-tertiary);
  word-break: break-all;
}

.project-meta {
  font-size: 12px;
  color: var(--aurora-text-secondary);
}

.cleanable-items {
  margin-top: 12px;
  padding: 12px;
  background: var(--aurora-bg-glass);
  border-radius: var(--aurora-radius-md);
}

.cleanable-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.cleanable-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--aurora-text-secondary);
}

.cleanable-size {
  font-size: 12px;
  font-weight: 600;
  color: var(--aurora-primary);
}

.cleanable-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.cleanable-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  background: var(--aurora-bg-card);
  border-radius: var(--aurora-radius-sm);
  font-size: 12px;
}

.cleanable-name {
  flex: 1;
}

.cleanable-info {
  color: var(--aurora-text-tertiary);
}

.cleanable-tip {
  color: var(--aurora-text-tertiary);
  cursor: pointer;
}

.no-cleanable {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--aurora-text-tertiary);
  font-size: 14px;
}

.initial-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 64px 24px;
  text-align: center;
}

.initial-icon {
  width: 120px;
  height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--aurora-bg-glass);
  border-radius: 50%;
  color: var(--aurora-primary);
  margin-bottom: 24px;
}

.initial-state h2 {
  font-size: 24px;
  font-weight: 700;
  margin-bottom: 8px;
  background: var(--aurora-text-gradient);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.initial-state p {
  color: var(--aurora-text-secondary);
  margin-bottom: 24px;
}

.feature-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: var(--aurora-bg-glass);
  border-radius: var(--aurora-radius-md);
  color: var(--aurora-text-secondary);
  font-size: 14px;
}
</style>

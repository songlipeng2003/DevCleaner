<template>
  <div class="home">
    <a-layout class="layout">
      <a-layout-header class="header">
        <div class="header-left">
          <h1>DevCleaner</h1>
          <span class="version">v{{ version }}</span>
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
        <!-- 磁盘使用情况 -->
        <a-card
          class="disk-card"
          :bordered="false"
        >
          <a-row
            :gutter="16"
            align="middle"
          >
            <a-col :span="20">
              <div class="disk-info">
                <span>磁盘使用: {{ formatSize(diskUsage.used) }} / {{ formatSize(diskUsage.total) }}</span>
                <a-progress 
                  :percent="diskPercent" 
                  :stroke-color="diskPercent > 90 ? '#ff4d4f' : diskPercent > 70 ? '#faad14' : '#52c41a'"
                />
              </div>
            </a-col>
            <a-col :span="4">
              <div class="free-space">
                <div class="free-space-label">可用空间</div>
                <div class="free-space-value">{{ formatSize(diskUsage.free) }}</div>
              </div>
            </a-col>
          </a-row>
        </a-card>

        <!-- 扫描控制 -->
        <div class="scan-actions">
          <a-space>
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
            <a-button @click="refreshTools">
              <template #icon>
                <ReloadOutlined />
              </template>
              刷新
            </a-button>
          </a-space>
          <!-- 扫描进度 -->
          <div
            v-if="isScanning"
            class="scan-progress"
          >
            <div class="progress-info">
              <span class="progress-tool">{{ scanProgress.toolName }}</span>
              <span class="progress-percent">{{ Math.round(scanProgress.progress * 100) }}%</span>
            </div>
            <a-progress
              :percent="Math.round(scanProgress.progress * 100)"
              :show-info="false"
              :stroke-color="'#1890ff'"
            />
          </div>
          <div
            v-else-if="scanResults.length > 0"
            class="scan-summary"
          >
            <span>发现 {{ scanResults.length }} 处缓存，共 {{ formatSize(totalCacheSize) }}</span>
          </div>
        </div>

        <!-- 智能推荐 -->
        <a-card
          v-if="recommendations.length > 0 && !isScanning"
          class="recommendation-card"
          :bordered="false"
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
                    <component :is="getToolIcon(item.toolId)" :size="16" :stroke-width="1.5" style="margin-right: 8px" />
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
          v-if="toolStore.error"
          type="error"
          :message="toolStore.error"
          show-icon
          closable
          style="margin-bottom: 16px"
          @close="toolStore.error = null"
        />
        
        <a-spin :spinning="toolStore.isLoading">
          <a-row
            v-if="enabledTools.length > 0"
            :gutter="[16, 16]"
            class="tools-grid"
          >
            <a-col
              v-for="tool in enabledTools"
              :key="tool.id"
              :xs="24"
              :sm="12"
              :md="8"
              :lg="6"
            >
              <a-card 
                class="tool-card" 
                :class="{ 'has-cache': getToolSize(tool.id) > 0 }"
                hoverable
                @click="showToolDetail(tool)"
              >
                <div class="tool-header">
                  <div class="tool-icon">
                    <component :is="getToolIcon(tool.id)" :size="32" :stroke-width="1.5" />
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
                <div
                  class="tool-actions"
                  @click.stop
                >
                  <a-button 
                    type="text" 
                    size="small" 
                    :disabled="getToolSize(tool.id) === 0"
                    @click="cleanTool(tool)"
                  >
                    清理
                  </a-button>
                  <a-button
                    type="text"
                    size="small"
                    @click="openToolPath(tool)"
                  >
                    打开
                  </a-button>
                </div>
              </a-card>
            </a-col>
          </a-row>
          
          <a-empty
            v-else-if="!toolStore.isLoading"
            description="暂无启用的开发工具"
          />
        </a-spin>
      </a-layout-content>
      
      <a-layout-footer class="footer">
        <a-space>
          <span>DevCleaner v0.1.0-alpha</span>
          <a-divider type="vertical" />
          <a
            href="https://github.com/devcleaner/devcleaner"
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
  BulbOutlined
} from '@ant-design/icons-vue'
import {
  Package,
  Sparkles,
  Folder,
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
} from 'lucide-vue-next'
import { useToolStore } from '@/stores/tools'
import { useSettingsStore } from '@/stores/settings'
import { getDiskUsage, type ScanProgress, getUsageStats } from '@/services/tauri'
import * as tauriApi from '@/services/tauri'
import type { ToolInfo } from '@/types'

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

const version = ref('0.1.0')
const drawerVisible = ref(false)
const selectedTool = ref<ToolInfo | null>(null)
const diskRefreshTimer = ref<ReturnType<typeof setTimeout> | null>(null)

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

const diskUsage = ref({
  total: 0,
  used: 0,
  free: 0
})

const diskPercent = computed(() => {
  if (diskUsage.value.total <= 0) return 0
  return Math.round((diskUsage.value.used / diskUsage.value.total) * 100)
})

const enabledTools = computed(() => toolStore.tools.filter(t => t.enabled))
const scanResults = computed(() => toolStore.scanResults)
const totalCacheSize = computed(() => toolStore.totalCacheSize)
const isScanning = computed(() => toolStore.isScanning)

// 工具图标映射
const toolIcons: Record<string, any> = {
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

async function refreshTools() {
  await toolStore.fetchTools()
  await fetchDiskUsage()
}

async function startScan() {
  try {
    await toolStore.scanAllTools((progress) => {
      scanProgress.value = progress
    })
    message.success(`扫描完成，共发现 ${scanResults.value.length} 处缓存`)
    // 生成智能推荐
    await generateRecommendations()
  } catch (e) {
    message.error('扫描失败')
  }
}

async function scanTool(tool: ToolInfo) {
  try {
    await toolStore.scanTool(tool.id)
    message.success(`${tool.name} 扫描完成`)
  } catch (e) {
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
  
  // 先预览
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
        } catch (e) {
          message.error('清理失败')
        }
      }
    })
  } catch (e) {
    message.loading({ content: '', key: 'preview' })
    message.error('预览失败')
  }
}

function openToolPath(tool: ToolInfo) {
  if (tool.paths.length > 0) {
    toolStore.openPath?.(tool.paths[0])
  }
}

function openPath(path: string) {
  toolStore.openPath?.(path)
}

function openSettings() {
  router.push('/settings')
}

async function fetchDiskUsage() {
  try {
    const usage = await getDiskUsage()
    diskUsage.value = usage
  } catch (error) {
    console.error('获取磁盘使用情况失败:', error)
    message.error('获取磁盘使用情况失败')
  }
}

onMounted(async () => {
  await toolStore.fetchTools()
  await fetchDiskUsage()
  await checkAutoScan()
  
  // 每30秒刷新磁盘使用情况
  diskRefreshTimer.value = setInterval(async () => {
    await fetchDiskUsage()
  }, 30000)
})

// 检查是否需要自动扫描
async function checkAutoScan() {
  const settings = settingsStore.settings
  if (!settings.autoScan) return
  
  // 从 localStorage 获取上次扫描时间
  const lastScanKey = 'devcleaner:lastScan'
  const lastScan = localStorage.getItem(lastScanKey)
  const now = Date.now()
  const intervalMs = settings.scanInterval * 24 * 60 * 60 * 1000
  
  if (!lastScan || (now - parseInt(lastScan)) > intervalMs) {
    message.info('自动扫描中...')
    await startScan()
    localStorage.setItem(lastScanKey, now.toString())
  }
}

// 生成智能推荐
async function generateRecommendations() {
  if (scanResults.value.length === 0) {
    recommendations.value = []
    return
  }

  const diskPercent = diskUsage.value.total > 0 
    ? (diskUsage.value.used / diskUsage.value.total) * 100 
    : 0
  
  const threshold = settingsStore.settings.threshold * 1024 * 1024 * 1024 // GB to bytes
  
  // 获取使用统计
  let stats = { cleanHistory: [] as any[] }
  try {
    stats = await getUsageStats()
  } catch (e) {
    // 忽略错误
  }

  const recs: Recommendation[] = []
  
  for (const result of scanResults.value) {
    const tool = toolStore.tools.find(t => t.id === result.tool_id)
    if (!tool) continue

    let reason = ''
    
    // 1. 磁盘空间不足时，优先推荐大缓存
    if (diskPercent > 80 && result.size > 100 * 1024 * 1024) {
      reason = `磁盘空间不足 (${diskPercent.toFixed(0)}%)，推荐清理大缓存`
    }
    // 2. 超过阈值的大缓存
    else if (result.size > threshold) {
      reason = `缓存超过阈值 (${formatSize(result.size)} > ${formatSize(threshold)})`
    }
    // 3. 长时间未清理 (> 30天)
    else if (result.last_modified > 0) {
      const daysSinceModified = (Date.now() - result.last_modified * 1000) / (1000 * 60 * 60 * 24)
      if (daysSinceModified > 30) {
        reason = `缓存超过 ${Math.floor(daysSinceModified)} 天未清理`
      }
    }
    // 4. 经常使用但缓存大的工具
    else {
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

  // 按大小排序，取前 3
  recommendations.value = recs
    .sort((a, b) => b.size - a.size)
    .slice(0, 3)
}

// 快速清理推荐项
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
})
</script>

<style scoped>
.home {
  height: 100vh;
  background: linear-gradient(135deg, var(--nature-primary-color) 0%, var(--nature-secondary-color) 100%);
  position: relative;
}

.layout {
  height: 100%;
  background: transparent;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: rgba(255, 255, 255, 0.95);
  padding: 0 24px;
  box-shadow: var(--nature-box-shadow);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid var(--nature-border-color);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header h1 {
  font-size: 20px;
  margin: 0;
  color: var(--nature-text-primary);
  font-weight: 600;
}

.version {
  font-size: 12px;
  color: var(--nature-text-tertiary);
  background: var(--nature-bg-hover);
  padding: 2px 8px;
  border-radius: var(--nature-border-radius-sm);
  font-weight: 500;
}

.content {
  padding: 24px;
  overflow-y: auto;
}

.disk-card {
  margin-bottom: 24px;
  border-radius: var(--nature-border-radius-base);
  box-shadow: var(--nature-box-shadow);
  background: var(--nature-bg-surface);
  border: 1px solid var(--nature-border-color);
  transition: box-shadow 0.3s ease;
}

.disk-card:hover {
  box-shadow: var(--nature-box-shadow-hover);
}

.disk-info {
  font-size: 14px;
  color: var(--nature-text-secondary);
}

.free-space {
  text-align: right;
}

.free-space-label {
  font-size: 12px;
  color: var(--nature-text-tertiary);
  margin-bottom: 4px;
}

.free-space-value {
  font-size: 18px;
  font-weight: 600;
  color: var(--nature-text-primary);
}

.scan-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.scan-summary {
  color: var(--nature-text-inverse);
  font-size: 14px;
  background: rgba(0, 0, 0, 0.2);
  padding: 4px 12px;
  border-radius: var(--nature-border-radius-sm);
}

.scan-progress {
  flex: 1;
  max-width: 400px;
  margin-left: 16px;
  background: rgba(255, 255, 255, 0.95);
  padding: 8px 16px;
  border-radius: var(--nature-border-radius-base);
  box-shadow: var(--nature-box-shadow);
}

.progress-info {
  display: flex;
  justify-content: space-between;
  margin-bottom: 4px;
  font-size: 12px;
  color: var(--nature-text-secondary);
}

.progress-tool {
  font-weight: 500;
}

.progress-percent {
  color: var(--nature-primary-color);
  font-weight: 600;
}

.recommendation-card {
  margin-bottom: 24px;
  border-radius: var(--nature-border-radius-base);
  box-shadow: var(--nature-box-shadow);
  background: linear-gradient(135deg, rgba(24, 144, 255, 0.1) 0%, rgba(82, 196, 26, 0.1) 100%);
  border: 1px solid var(--nature-border-color);
}

.recommendation-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--nature-primary-color);
  font-weight: 600;
}

.tools-grid {
  margin-top: 16px;
}

.tool-card {
  border-radius: var(--nature-border-radius-base);
  transition: all 0.3s ease;
  background: var(--nature-bg-surface);
  border: 1px solid var(--nature-border-color);
  box-shadow: var(--nature-box-shadow);
  height: 100%;
}

.tool-card:hover {
  box-shadow: var(--nature-box-shadow-hover);
  transform: translateY(-2px);
}

.tool-card.has-cache {
  border: 2px solid var(--nature-primary-color);
}

.tool-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.tool-icon {
  color: var(--nature-primary-color);
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.1));
}

.tool-info {
  text-align: center;
  padding: 8px 0;
}

.tool-name {
  font-size: 16px;
  font-weight: 600;
  color: var(--nature-text-primary);
  margin-bottom: 4px;
}

.tool-size {
  font-size: 14px;
  color: var(--nature-text-tertiary);
}

.tool-size.has-size {
  color: var(--nature-primary-color);
  font-weight: 600;
}

.tool-paths {
  font-size: 12px;
  color: var(--nature-text-tertiary);
  margin-top: 4px;
}

.tool-actions {
  display: flex;
  justify-content: center;
  gap: 8px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--nature-border-color-split);
}

.footer {
  text-align: center;
  background: rgba(255, 255, 255, 0.9);
  color: var(--nature-text-secondary);
  border-top: 1px solid var(--nature-border-color);
  padding: 12px 24px;
}
</style>

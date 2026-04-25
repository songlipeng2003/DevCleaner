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
            <template #icon><SettingOutlined /></template>
            设置
          </a-button>
        </div>
      </a-layout-header>
      
      <a-layout-content class="content">
        <!-- 磁盘使用情况 -->
        <a-card class="disk-card" :bordered="false">
          <a-row :gutter="16" align="middle">
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
              <a-statistic 
                title="可用空间" 
                :value="diskUsage.free" 
                :formatter="formatSize" 
              />
            </a-col>
          </a-row>
        </a-card>

        <!-- 扫描控制 -->
        <div class="scan-actions">
          <a-space>
            <a-button type="primary" size="large" @click="startScan" :loading="isScanning">
              <template #icon><ScanOutlined /></template>
              {{ isScanning ? '扫描中...' : '开始扫描' }}
            </a-button>
            <a-button @click="refreshTools">
              <template #icon><ReloadOutlined /></template>
              刷新
            </a-button>
          </a-space>
          <div class="scan-summary" v-if="scanResults.length > 0">
            <span>发现 {{ scanResults.length }} 处缓存，共 {{ formatSize(totalCacheSize) }}</span>
          </div>
        </div>

        <!-- 开发工具列表 -->
        <a-row :gutter="[16, 16]" class="tools-grid">
          <a-col :xs="24" :sm="12" :md="8" :lg="6" v-for="tool in enabledTools" :key="tool.id">
            <a-card 
              class="tool-card" 
              :class="{ 'has-cache': getToolSize(tool.id) > 0 }"
              hoverable
              @click="showToolDetail(tool)"
            >
              <div class="tool-header">
                <div class="tool-icon">{{ getToolIcon(tool.id) }}</div>
                <a-switch 
                  v-model:checked="tool.enabled" 
                  size="small" 
                  @click.stop
                  @change="toggleTool(tool)"
                />
              </div>
              <div class="tool-info">
                <div class="tool-name">{{ tool.name }}</div>
                <div class="tool-size" :class="{ 'has-size': getToolSize(tool.id) > 0 }">
                  {{ getToolSize(tool.id) > 0 ? formatSize(getToolSize(tool.id)) : '无缓存' }}
                </div>
                <div class="tool-paths">{{ tool.paths.length }} 个路径</div>
              </div>
              <div class="tool-actions" @click.stop>
                <a-button 
                  type="text" 
                  size="small" 
                  :disabled="getToolSize(tool.id) === 0"
                  @click="cleanTool(tool)"
                >
                  清理
                </a-button>
                <a-button type="text" size="small" @click="openToolPath(tool)">
                  打开
                </a-button>
              </div>
            </a-card>
          </a-col>
        </a-row>
      </a-layout-content>
      
      <a-layout-footer class="footer">
        <a-space>
          <span>DevCleaner v0.1.0-alpha</span>
          <a-divider type="vertical" />
          <a href="https://github.com/devcleaner/devcleaner" target="_blank">GitHub</a>
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
        <a-descriptions :column="1" bordered>
          <a-descriptions-item label="工具 ID">{{ selectedTool.id }}</a-descriptions-item>
          <a-descriptions-item label="路径数量">{{ selectedTool.paths.length }}</a-descriptions-item>
          <a-descriptions-item label="当前缓存">{{ formatSize(getToolSize(selectedTool.id)) }}</a-descriptions-item>
        </a-descriptions>

        <a-divider>缓存路径</a-divider>
        
        <a-list 
          size="small" 
          :data-source="selectedTool.paths"
        >
          <template #renderItem="{ item }">
            <a-list-item>
              <template #actions>
                <a-button type="link" size="small" @click="openPath(item)">打开</a-button>
              </template>
              <a-list-item-meta>
                <template #title>{{ item }}</template>
              </a-list-item-meta>
            </a-list-item>
          </template>
        </a-list>

        <a-divider>操作</a-divider>
        <a-space direction="vertical" style="width: 100%">
          <a-button block type="primary" @click="scanTool(selectedTool)">
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
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import { 
  SettingOutlined, 
  ScanOutlined, 
  ReloadOutlined 
} from '@ant-design/icons-vue'
import { useToolStore } from '@/stores/tools'
import { getDiskUsage } from '@/services/tauri'
import type { ToolInfo } from '@/types'

const router = useRouter()
const toolStore = useToolStore()

const version = ref('0.1.0')
const drawerVisible = ref(false)
const selectedTool = ref<ToolInfo | null>(null)

const diskUsage = ref({
  total: 0,
  used: 0,
  free: 0
})

const diskPercent = computed(() => {
  if (diskUsage.value.total === 0) return 0
  return Math.round((diskUsage.value.used / diskUsage.value.total) * 100)
})

const enabledTools = computed(() => toolStore.tools.filter(t => t.enabled))
const scanResults = computed(() => toolStore.scanResults)
const totalCacheSize = computed(() => toolStore.totalCacheSize)
const isScanning = computed(() => toolStore.isScanning)

const toolIcons: Record<string, string> = {
  npm: '📦',
  yarn: '🧶',
  pnpm: '📁',
  docker: '🐳',
  xcode: '🍎',
  homebrew: '🍺',
  python: '🐍',
  go: '🔵',
  ruby: '💎',
  maven: '📚',
  gradle: '⚙️',
  cocoapods: '🫘',
  carthage: '🐴',
  unity: '🎮',
}

function getToolIcon(toolId: string): string {
  return toolIcons[toolId] || '🔧'
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
    await toolStore.scanAllTools()
    message.success(`扫描完成，共发现 ${scanResults.value.length} 处缓存`)
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
  message.info(`${tool.name} 已${tool.enabled ? '启用' : '禁用'}`)
}

function cleanTool(tool: ToolInfo) {
  const size = getToolSize(tool.id)
  Modal.confirm({
    title: `确认清理 ${tool.name}`,
    content: `确定要清理 ${tool.name} 的缓存吗？这将释放约 ${formatSize(size)} 磁盘空间。`,
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
})
</script>

<style scoped>
.home {
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
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
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header h1 {
  font-size: 20px;
  margin: 0;
  color: #333;
}

.version {
  font-size: 12px;
  color: #999;
  background: #f5f5f5;
  padding: 2px 8px;
  border-radius: 4px;
}

.content {
  padding: 24px;
  overflow-y: auto;
}

.disk-card {
  margin-bottom: 24px;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}

.disk-info {
  font-size: 14px;
  color: #666;
}

.scan-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.scan-summary {
  color: #fff;
  font-size: 14px;
}

.tools-grid {
  margin-top: 16px;
}

.tool-card {
  border-radius: 12px;
  transition: all 0.3s;
  background: rgba(255, 255, 255, 0.95);
}

.tool-card.has-cache {
  border: 2px solid #1890ff;
}

.tool-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.tool-icon {
  font-size: 32px;
}

.tool-info {
  text-align: center;
  padding: 8px 0;
}

.tool-name {
  font-size: 16px;
  font-weight: 600;
  color: #333;
  margin-bottom: 4px;
}

.tool-size {
  font-size: 14px;
  color: #999;
}

.tool-size.has-size {
  color: #1890ff;
  font-weight: 600;
}

.tool-paths {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

.tool-actions {
  display: flex;
  justify-content: center;
  gap: 8px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
}

.footer {
  text-align: center;
  background: rgba(255, 255, 255, 0.9);
  color: #666;
}
</style>

<template>
  <div class="scan">
    <a-layout class="layout">
      <a-layout-header class="header">
        <a-button @click="goBack">返回</a-button>
        <h2>{{ isScanning ? '扫描中...' : '扫描完成' }}</h2>
        <div>
          <a-button v-if="!isScanning" type="primary" @click="startScan">重新扫描</a-button>
        </div>
      </a-layout-header>
      
      <a-layout-content class="content">
        <a-alert
          v-if="error"
          type="error"
          :message="error"
          show-icon
          closable
          style="margin-bottom: 16px"
          @close="error = null"
        />
        
        <!-- 总体进度 -->
        <a-card title="扫描进度" :bordered="false" style="margin-bottom: 24px">
          <a-progress 
            :percent="overallProgress" 
            :status="isScanning ? 'active' : (error ? 'exception' : 'success')"
            :stroke-color="progressColor"
          />
          <div class="progress-stats">
            <span>已扫描: {{ completedToolsCount }} / {{ totalToolsCount }} 个工具</span>
            <span>发现缓存: {{ totalCacheSizeFormatted }}</span>
          </div>
        </a-card>
        
        <!-- 工具扫描列表 -->
        <a-card title="工具扫描详情" :bordered="false">
          <a-list 
            class="scan-list" 
            :data-source="scanningTools"
            :loading="isScanning && scanningTools.length === 0"
          >
            <template #renderItem="{ item }">
              <a-list-item>
                <template #actions>
                  <a-tag :color="getStatusColor(item.status)">
                    {{ getStatusText(item.status) }}
                  </a-tag>
                  <span>{{ item.sizeFormatted }}</span>
                </template>
                <a-list-item-meta :title="item.name" :description="item.paths?.join(', ') || '无路径'">
                  <template #avatar>
                    <div class="tool-icon">{{ getToolIcon(item.id) }}</div>
                  </template>
                </a-list-item-meta>
                <div v-if="item.status === 'scanning' && item.currentPath" class="current-path">
                  正在扫描: {{ item.currentPath }}
                </div>
                <a-progress 
                  v-if="item.status === 'scanning'"
                  :percent="item.progress" 
                  size="small"
                  style="margin-top: 8px"
                />
              </a-list-item>
            </template>
          </a-list>
        </a-card>
        
        <!-- 扫描结果摘要 -->
        <a-card v-if="!isScanning && scanResults.length > 0" title="扫描结果摘要" :bordered="false" style="margin-top: 24px">
          <a-descriptions :column="2" bordered>
            <a-descriptions-item label="扫描工具数">{{ completedToolsCount }}</a-descriptions-item>
            <a-descriptions-item label="发现缓存路径">{{ scanResults.length }}</a-descriptions-item>
            <a-descriptions-item label="总缓存大小">{{ totalCacheSizeFormatted }}</a-descriptions-item>
            <a-descriptions-item label="可释放空间">{{ totalCacheSizeFormatted }}</a-descriptions-item>
          </a-descriptions>
          <div style="margin-top: 16px; text-align: center">
            <a-space>
              <a-button type="primary" @click="goBack">返回主页</a-button>
              <a-button @click="startScan">重新扫描</a-button>
            </a-space>
          </div>
        </a-card>
      </a-layout-content>
    </a-layout>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { useToolStore } from '@/stores/tools'
import type { ScanProgress } from '@/types'

const router = useRouter()
const toolStore = useToolStore()

const error = ref<string | null>(null)
const scanning = ref(false)

const isScanning = computed(() => toolStore.isScanning)
const scanResults = computed(() => toolStore.scanResults)
const scanProgress = computed(() => toolStore.scanProgress)
const enabledTools = computed(() => toolStore.enabledTools)
const totalCacheSize = computed(() => toolStore.totalCacheSize)

const totalToolsCount = computed(() => enabledTools.value.length)
const completedToolsCount = computed(() => {
  const progress = scanProgress.value
  return Array.from(progress.values()).filter(p => 
    p.status === 'completed' || p.status === 'error'
  ).length
})

const overallProgress = computed(() => {
  if (totalToolsCount.value === 0) return 0
  return Math.round((completedToolsCount.value / totalToolsCount.value) * 100)
})

const progressColor = computed(() => {
  if (error.value) return '#ff4d4f'
  if (isScanning.value) return '#1890ff'
  return '#52c41a'
})

const totalCacheSizeFormatted = computed(() => toolStore.formatSize(totalCacheSize.value))

const scanningTools = computed(() => {
  return enabledTools.value.map(tool => {
    const progress = scanProgress.value.get(tool.id) || {
      tool_id: tool.id,
      status: 'pending' as const,
      progress: 0,
      current_path: undefined
    }
    
    const toolResults = toolStore.getToolResults(tool.id)
    const toolSize = toolResults.reduce((sum, r) => sum + r.size, 0)
    
    return {
      id: tool.id,
      name: tool.name,
      paths: tool.paths,
      status: progress.status,
      progress: progress.progress,
      currentPath: progress.current_path,
      size: toolSize,
      sizeFormatted: toolStore.formatSize(toolSize)
    }
  })
})

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

function getStatusColor(status: ScanProgress['status']): string {
  switch (status) {
    case 'completed': return 'success'
    case 'scanning': return 'processing'
    case 'error': return 'error'
    case 'pending': return 'default'
  }
}

function getStatusText(status: ScanProgress['status']): string {
  switch (status) {
    case 'completed': return '完成'
    case 'scanning': return '扫描中'
    case 'error': return '失败'
    case 'pending': return '等待中'
  }
}

async function startScan() {
  error.value = null
  scanning.value = true
  try {
    await toolStore.scanAllTools()
    message.success('扫描完成')
  } catch (e) {
    error.value = e instanceof Error ? e.message : '扫描失败'
    message.error('扫描失败')
  } finally {
    scanning.value = false
  }
}

function goBack() {
  router.push('/')
}

onMounted(async () => {
  // 如果还没有扫描结果，自动开始扫描
  if (scanResults.value.length === 0 && !isScanning.value) {
    await startScan()
  }
})
</script>

<style scoped>
.scan {
  height: 100vh;
  background: #f0f2f5;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
  padding: 0 24px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}

.content {
  padding: 24px;
  margin: 24px;
  background: #fff;
  border-radius: 8px;
  overflow-y: auto;
}

.progress-stats {
  display: flex;
  justify-content: space-between;
  margin-top: 12px;
  color: #666;
  font-size: 14px;
}

.tool-icon {
  font-size: 24px;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f5f5;
  border-radius: 8px;
}

.current-path {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.scan-list {
  max-height: 400px;
  overflow-y: auto;
}
</style>

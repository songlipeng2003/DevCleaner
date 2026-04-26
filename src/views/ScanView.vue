<template>
  <div class="scan">
    <a-layout class="layout">
      <a-layout-header class="header">
        <a-button @click="goBack">
          返回
        </a-button>
        <h2>{{ isScanning ? '扫描中...' : '扫描结果' }}</h2>
        <div>
          <a-button
            v-if="!isScanning"
            type="primary"
            @click="startScan"
          >
            重新扫描
          </a-button>
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
        
        <!-- 扫描结果摘要 -->
        <a-card
          title="扫描结果摘要"
          :bordered="false"
          style="margin-bottom: 24px"
        >
          <a-descriptions
            :column="2"
            bordered
          >
            <a-descriptions-item label="扫描工具数">
              {{ enabledTools.length }}
            </a-descriptions-item>
            <a-descriptions-item label="发现缓存路径">
              {{ scanResults.length }}
            </a-descriptions-item>
            <a-descriptions-item label="总缓存大小">
              {{ totalCacheSizeFormatted }}
            </a-descriptions-item>
            <a-descriptions-item label="可释放空间">
              {{ totalCacheSizeFormatted }}
            </a-descriptions-item>
          </a-descriptions>
        </a-card>
        
        <!-- 扫描结果列表 -->
        <a-card
          v-if="scanResults.length > 0"
          title="缓存详情"
          :bordered="false"
        >
          <a-list 
            class="scan-list" 
            :data-source="scanResults"
            :loading="isScanning"
          >
            <template #renderItem="{ item }">
              <a-list-item>
                <template #actions>
                  <span class="cache-size">{{ formatSize(item.size) }}</span>
                </template>
                <a-list-item-meta
                  :title="item.tool_id"
                  :description="`${item.file_num} 个文件`"
                >
                  <template #avatar>
                    <div class="tool-icon">
                      <component :is="getToolIcon(item.tool_id)" :size="24" :stroke-width="1.5" />
                    </div>
                  </template>
                </a-list-item-meta>
              </a-list-item>
            </template>
          </a-list>
        </a-card>
        
        <a-empty
          v-else-if="!isScanning"
          description="暂无缓存"
          style="margin-top: 48px"
        />
        
        <div style="margin-top: 24px; text-align: center">
          <a-space>
            <a-button
              type="primary"
              @click="goBack"
            >
              返回主页
            </a-button>
            <a-button @click="startScan">
              {{ isScanning ? '扫描中...' : '重新扫描' }}
            </a-button>
          </a-space>
        </div>
      </a-layout-content>
    </a-layout>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
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

const router = useRouter()
const toolStore = useToolStore()

const error = ref<string | null>(null)

const isScanning = computed(() => toolStore.isScanning)
const scanResults = computed(() => toolStore.scanResults)
const enabledTools = computed(() => toolStore.enabledTools)
const totalCacheSize = computed(() => toolStore.totalCacheSize)

const totalCacheSizeFormatted = computed(() => toolStore.formatSize(totalCacheSize.value))

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

async function startScan() {
  error.value = null
  try {
    await toolStore.scanAllTools()
    message.success('扫描完成')
  } catch (e) {
    error.value = e instanceof Error ? e.message : '扫描失败'
    message.error('扫描失败')
  }
}

function goBack() {
  router.push('/')
}
</script>

<style scoped>
.scan {
  height: 100vh;
  background: var(--nature-bg-body);
  position: relative;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--nature-bg-surface);
  padding: 0 24px;
  box-shadow: var(--nature-box-shadow);
  border-bottom: 1px solid var(--nature-border-color);
  backdrop-filter: blur(10px);
}

.content {
  padding: 24px;
  margin: 24px;
  background: var(--nature-bg-surface);
  border-radius: var(--nature-border-radius-base);
  overflow-y: auto;
  border: 1px solid var(--nature-border-color);
  box-shadow: var(--nature-box-shadow);
}

.tool-icon {
  color: var(--nature-primary-color);
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--nature-bg-hover);
  border-radius: var(--nature-border-radius-sm);
}

.cache-size {
  color: var(--nature-primary-color);
  font-weight: 600;
}

.scan-list {
  max-height: 400px;
  overflow-y: auto;
}

/* 自定义卡片样式 */
:deep(.ant-card) {
  background: var(--nature-bg-surface);
  border: 1px solid var(--nature-border-color);
  border-radius: var(--nature-border-radius-base);
  box-shadow: var(--nature-box-shadow);
}

:deep(.ant-card:hover) {
  box-shadow: var(--nature-box-shadow-hover);
}
</style>

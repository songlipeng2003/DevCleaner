<template>
  <div class="scan">
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
            <ArrowLeft :size="18" />
            <span>返回</span>
          </button>
        </div>
        
        <div class="header-center">
          <h2 :class="{ 'gradient-text': !isScanning }">
            {{ isScanning ? '扫描中...' : '扫描结果' }}
          </h2>
        </div>
        
        <div class="header-right">
          <a-button
            v-if="!isScanning"
            type="primary"
            @click="startScan"
          >
            <template #icon>
              <ScanOutlined />
            </template>
            重新扫描
          </a-button>
          <a-button
            v-else
            loading
          >
            扫描中...
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
          style="margin-bottom: 24px"
          @close="error = null"
        />
        
        <!-- 扫描中动画 -->
        <div
          v-if="isScanning"
          class="scanning-container"
        >
          <div class="scanning-visual glass-card">
            <div class="scanning-ring">
              <svg viewBox="0 0 160 160">
                <circle
                  cx="80"
                  cy="80"
                  r="70"
                  fill="none"
                  stroke="var(--aurora-border)"
                  stroke-width="8"
                />
                <circle
                  class="scan-progress"
                  cx="80"
                  cy="80"
                  r="70"
                  fill="none"
                  stroke="url(#scanGradient)"
                  stroke-width="8"
                  :stroke-dasharray="circumference"
                  :stroke-dashoffset="progressOffset"
                />
                <defs>
                  <linearGradient
                    id="scanGradient"
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
              <div class="scan-center">
                <ScanOutlined
                  :size="36"
                  class="scan-icon spinning"
                />
              </div>
            </div>
            <h3 class="scanning-title">
              正在扫描
            </h3>
            <p class="scanning-tool">
              {{ currentToolName }}
            </p>
          </div>
        </div>
        
        <!-- 扫描结果 -->
        <template v-else>
          <!-- 结果摘要 -->
          <div class="overview-cards">
            <div class="overview-card glass-card">
              <div class="overview-icon">
                <Package :size="24" />
              </div>
              <div class="overview-info">
                <span class="overview-value gradient-text">{{ enabledTools.length }}</span>
                <span class="overview-label">扫描工具</span>
              </div>
            </div>
            
            <div class="overview-card glass-card">
              <div class="overview-icon">
                <FolderOpen :size="24" />
              </div>
              <div class="overview-info">
                <span class="overview-value gradient-text">{{ scanResults.length }}</span>
                <span class="overview-label">发现缓存</span>
              </div>
            </div>
            
            <div class="overview-card glass-card">
              <div class="overview-icon danger">
                <HardDrive :size="24" />
              </div>
              <div class="overview-info">
                <span class="overview-value gradient-text">{{ totalCacheSizeFormatted }}</span>
                <span class="overview-label">可释放空间</span>
              </div>
            </div>
          </div>
          
          <!-- 缓存详情列表 -->
          <a-card
            v-if="scanResults.length > 0"
            class="results-card glass-card"
            title="缓存详情"
          >
            <a-list 
              class="scan-list" 
              :data-source="scanResults"
            >
              <template #renderItem="{ item }">
                <a-list-item>
                  <template #actions>
                    <span class="cache-size gradient-text">{{ formatSize(item.size) }}</span>
                  </template>
                  <a-list-item-meta
                    :title="getToolName(item.tool_id)"
                    :description="`${item.file_num} 个文件 · ${item.paths?.length || 0} 个路径`"
                  >
                    <template #avatar>
                      <div class="tool-icon">
                        <component
                          :is="getToolIcon(item.tool_id)"
                          :size="24"
                          :stroke-width="1.5"
                        />
                      </div>
                    </template>
                  </a-list-item-meta>
                </a-list-item>
              </template>
            </a-list>
          </a-card>
          
          <a-empty
            v-else
            description="暂无缓存"
            style="margin-top: 48px"
          />
          
          <div class="actions-bar">
            <a-button @click="goBack">
              返回主页
            </a-button>
            <a-button
              type="primary"
              @click="startScan"
            >
              重新扫描
            </a-button>
          </div>
        </template>
      </a-layout-content>
    </a-layout>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import type { Component } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { ScanOutlined } from '@ant-design/icons-vue'
import {
  ArrowLeft,
  Package,
  FolderOpen,
  HardDrive,
  Package as PackageIcon,
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
} from 'lucide-vue-next'
import { useToolStore } from '@/stores/tools'
import {
  trackPageView,
  trackScanStart,
  trackScanComplete,
} from '@/services/analytics'

const router = useRouter()
const toolStore = useToolStore()

const error = ref<string | null>(null)

const isScanning = computed(() => toolStore.isScanning)
const scanResults = computed(() => toolStore.scanResults)
const enabledTools = computed(() => toolStore.enabledTools)
const totalCacheSize = computed(() => toolStore.totalCacheSize)
const totalCacheSizeFormatted = computed(() => toolStore.formatSize(totalCacheSize.value))

const circumference = 2 * Math.PI * 70
const scanProgressLocal = ref(0)
const currentToolName = ref('')

const progressOffset = computed(() => {
  return circumference - (scanProgressLocal.value * circumference)
})

// 工具图标映射
const toolIcons: Record<string, Component> = {
  npm: PackageIcon,
  yarn: Sparkles,
  pnpm: Folder,
  bun: Cookie,
  composer: Gem,
  cargo: Box,
  flutter: Wind,
  nuget: PackageIcon,
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

function getToolName(toolId: string): string {
  const tool = toolStore.tools.find(t => t.id === toolId)
  return tool?.name || toolId
}

function formatSize(bytes: number): string {
  return toolStore.formatSize(bytes)
}

async function startScan() {
  error.value = null
  trackScanStart('all')
  try {
    await toolStore.scanAllTools((progress) => {
      scanProgressLocal.value = progress.progress
      currentToolName.value = progress.toolName
    })
    message.success('扫描完成')
    trackScanComplete('all', scanResults.value.length)
  } catch (e) {
    error.value = e instanceof Error ? e.message : '扫描失败'
    message.error('扫描失败')
  }
}

function goBack() {
  router.push('/')
}

onMounted(async () => {
  trackPageView('ScanView')
  if (scanResults.value.length === 0 && !isScanning.value) {
    await startScan()
  }
})
</script>

<style scoped>
.scan {
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
  left: -200px;
  background: var(--aurora-accent);
}

.hero-glow-2 {
  bottom: -200px;
  right: -200px;
  background: var(--aurora-primary);
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
  padding: 0 24px;
  border-bottom: 1px solid var(--aurora-border);
}

.header-center h2 {
  font-size: 20px;
  font-weight: 600;
  margin: 0;
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: var(--aurora-bg-glass);
  border: 1px solid var(--aurora-border);
  border-radius: var(--aurora-radius-md);
  color: var(--aurora-text-secondary);
  font-size: 14px;
  cursor: pointer;
  transition: all var(--aurora-transition-fast);
}

.back-btn:hover {
  border-color: var(--aurora-border-light);
  color: var(--aurora-text-primary);
}

.content {
  padding: 32px;
  overflow-y: auto;
  max-height: calc(100vh - 64px);
}

/* 扫描中动画 */
.scanning-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 60vh;
}

.scanning-visual {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 48px;
  text-align: center;
}

.scanning-ring {
  position: relative;
  width: 160px;
  height: 160px;
  margin-bottom: 24px;
}

.scanning-ring svg {
  width: 100%;
  height: 100%;
  transform: rotate(-90deg);
}

.scan-progress {
  transition: stroke-dashoffset 0.3s ease;
  stroke-linecap: round;
}

.scan-center {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
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

.scanning-title {
  font-size: 24px;
  font-weight: 700;
  margin-bottom: 8px;
}

.scanning-tool {
  color: var(--aurora-text-secondary);
  font-size: 14px;
}

/* 概览卡片 */
.overview-cards {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.overview-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
}

.overview-icon {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--aurora-bg-glass);
  border-radius: var(--aurora-radius-md);
  color: var(--aurora-primary);
}

.overview-icon.danger {
  color: var(--aurora-danger);
}

.overview-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.overview-value {
  font-size: 24px;
  font-weight: 700;
}

.overview-label {
  font-size: 13px;
  color: var(--aurora-text-tertiary);
}

/* 结果列表 */
.results-card {
  margin-bottom: 24px;
}

.tool-icon {
  width: 44px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--aurora-bg-glass);
  border-radius: var(--aurora-radius-md);
  color: var(--aurora-primary);
}

.cache-size {
  font-weight: 600;
  font-size: 14px;
}

/* 操作栏 */
.actions-bar {
  display: flex;
  justify-content: center;
  gap: 16px;
  padding-top: 16px;
}

/* 响应式 */
@media (max-width: 768px) {
  .content {
    padding: 16px;
  }
  
  .overview-cards {
    grid-template-columns: 1fr;
  }
}
</style>

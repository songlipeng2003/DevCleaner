<template>
  <div class="settings">
    <!-- 极光背景装饰 -->
    <div class="hero-glow hero-glow-1" />
    
    <a-layout class="layout">
      <a-layout-header class="header">
        <div class="header-left">
          <button
            class="back-btn"
            @click="goBack"
          >
            <ArrowLeft :size="16" />
            <span>返回</span>
          </button>
        </div>
        
        <div class="header-center">
          <h2 class="gradient-text">
            设置
          </h2>
        </div>
        
        <div class="header-right">
          <a-button
            type="primary"
            :loading="isSaving"
            @click="saveSettings"
          >
            保存
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
        
        <a-spin :spinning="isLoading">
          <!-- 主题设置 -->
          <section class="settings-section">
            <h3 class="section-title">
              <Palette :size="20" />
              外观
            </h3>
            <p class="section-description">
              自定义应用的外观和配色
            </p>
            
            <div class="theme-selector">
              <button
                v-for="option in themeOptions"
                :key="option.value"
                class="theme-option"
                :class="{ active: settings.theme === option.value }"
                @click="settings.theme = option.value"
              >
                <div class="theme-icon">
                  <component
                    :is="option.icon"
                    :size="24"
                  />
                </div>
                <span class="theme-label">{{ option.label }}</span>
                <div
                  v-if="settings.theme === option.value"
                  class="theme-check"
                >
                  <CheckOutlined />
                </div>
              </button>
            </div>
          </section>
          
          <!-- 扫描设置 -->
          <section class="settings-section">
            <h3 class="section-title">
              <ScanOutlined :size="20" />
              扫描设置
            </h3>
            <p class="section-description">
              配置自动扫描和清理规则
            </p>
            
            <div class="settings-list">
              <div class="setting-item glass-card">
                <div class="setting-info">
                  <div class="setting-icon">
                    <BellOutlined :size="20" />
                  </div>
                  <div class="setting-text">
                    <span class="setting-label">自动扫描</span>
                    <span class="setting-value">开启后自动定期扫描</span>
                  </div>
                </div>
                <a-switch v-model:checked="settings.autoScan" />
              </div>

              <div class="setting-item glass-card">
                <div class="setting-info">
                  <div class="setting-icon">
                    <Keyboard :size="20" />
                  </div>
                  <div class="setting-text">
                    <span class="setting-label">启用快捷键</span>
                    <span class="setting-value">Ctrl/Cmd+S 扫描, Ctrl/Cmd+R 刷新</span>
                  </div>
                </div>
                <a-switch v-model:checked="settings.shortcutsEnabled" />
              </div>
              
              <div
                v-if="settings.autoScan"
                class="setting-item glass-card"
              >
                <div class="setting-info">
                  <div class="setting-icon">
                    <ClockCircleOutlined :size="20" />
                  </div>
                  <div class="setting-text">
                    <span class="setting-label">扫描间隔（天）</span>
                    <span class="setting-value">自动扫描的频率</span>
                  </div>
                </div>
                <a-input-number
                  v-model:value="settings.scanInterval"
                  :min="1"
                  :max="30"
                  style="width: 100px"
                />
              </div>
              
              <div class="setting-item glass-card">
                <div class="setting-info">
                  <div class="setting-icon">
                    <DatabaseOutlined :size="20" />
                  </div>
                  <div class="setting-text">
                    <span class="setting-label">清理阈值</span>
                    <span class="setting-value">当可用空间低于此值时提醒</span>
                  </div>
                </div>
                <div class="threshold-control">
                  <a-slider
                    v-model:value="settings.threshold"
                    :min="10"
                    :max="500"
                    :step="5"
                    style="width: 200px"
                  />
                  <span class="threshold-value gradient-text">{{ settings.threshold }} GB</span>
                </div>
              </div>
            </div>
          </section>
          
          <!-- 白名单设置 -->
          <section class="settings-section">
            <h3 class="section-title">
              <Shield :size="20" />
              白名单
            </h3>
            <p class="section-description">
              设置永远不会清理的文件或文件夹
            </p>
            
            <div class="whitelist-card glass-card">
              <div
                v-if="settings.whitelist.length === 0"
                class="whitelist-empty"
              >
                <InfoCircleOutlined :size="20" />
                <p>暂无白名单项</p>
              </div>
              
              <div
                v-else
                class="whitelist-items"
              >
                <div 
                  v-for="(item, index) in settings.whitelist" 
                  :key="index"
                  class="whitelist-item"
                >
                  <FolderOpen
                    :size="16"
                    class="whitelist-icon"
                  />
                  <span class="whitelist-path">{{ item }}</span>
                  <a-button
                    type="text"
                    danger
                    size="small"
                    @click="removeWhitelist(item)"
                  >
                    删除
                  </a-button>
                </div>
              </div>
              
              <div class="add-whitelist">
                <a-input
                  v-model:value="newWhitelist"
                  placeholder="添加排除路径"
                  style="flex: 1"
                />
                <a-button
                  type="primary"
                  :disabled="!newWhitelist.trim()"
                  @click="addWhitelist"
                >
                  添加
                </a-button>
              </div>
            </div>
          </section>
        </a-spin>
      </a-layout-content>
    </a-layout>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  CheckOutlined,
  ScanOutlined,
  BellOutlined,
  ClockCircleOutlined,
  DatabaseOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons-vue'
import {
  ArrowLeft,
  Sun,
  Moon,
  Monitor,
  Palette,
  Shield,
  FolderOpen,
  Keyboard,
} from 'lucide-vue-next'
import { useSettingsStore } from '@/stores/settings'

const router = useRouter()
const settingsStore = useSettingsStore()

const newWhitelist = ref('')
const error = ref<string | null>(null)

const settings = computed(() => settingsStore.settings)
const isLoading = computed(() => settingsStore.isLoading)
const isSaving = computed(() => settingsStore.isSaving)

const themeOptions = [
  { value: 'dark' as const, label: '深色', icon: Moon },
  { value: 'light' as const, label: '浅色', icon: Sun },
  { value: 'auto' as const, label: '跟随系统', icon: Monitor }
]

const goBack = () => {
  router.push('/')
}

const addWhitelist = () => {
  const path = newWhitelist.value.trim()
  if (path) {
    settingsStore.addWhitelist(path)
    newWhitelist.value = ''
  }
}

const removeWhitelist = (path: string) => {
  settingsStore.removeWhitelist(path)
}

const saveSettings = async () => {
  try {
    await settingsStore.saveSettings(settings.value)
    message.success('设置已保存')
    error.value = null
  } catch (e) {
    error.value = e instanceof Error ? e.message : '保存设置失败'
  }
}

onMounted(async () => {
  try {
    await settingsStore.fetchSettings()
    error.value = null
  } catch (e) {
    error.value = e instanceof Error ? e.message : '获取设置失败'
  }
})
</script>

<style scoped>
.settings {
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
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
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
  padding: 0 16px;
  border-bottom: 1px solid var(--aurora-border);
  gap: 12px;
  flex-shrink: 0;
  min-height: 48px;
  height: 48px;
}

.header-center h2 {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background: var(--aurora-bg-glass);
  border: 1px solid var(--aurora-border);
  border-radius: var(--aurora-radius-md);
  color: var(--aurora-text-secondary);
  font-size: 12px;
  cursor: pointer;
  transition: all var(--aurora-transition-fast);
}

.back-btn:hover {
  border-color: var(--aurora-border-light);
  color: var(--aurora-text-primary);
}

.content {
  padding: 16px 12px;
  overflow-y: auto;
  max-height: calc(100vh - 48px);
  width: 100%;
  box-sizing: border-box;
}

/* 设置区块 */
.settings-section {
  margin-bottom: 24px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 15px;
  font-weight: 600;
  margin-bottom: 2px;
  color: var(--aurora-primary);
}

.section-description {
  color: var(--aurora-text-tertiary);
  font-size: 13px;
  margin-bottom: 12px;
}

/* 主题选择器 */
.theme-selector {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.theme-option {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 16px;
  background: var(--aurora-bg-card);
  border: 2px solid var(--aurora-border);
  border-radius: var(--aurora-radius-lg);
  cursor: pointer;
  transition: all var(--aurora-transition-normal);
}

.theme-option:hover {
  border-color: var(--aurora-border-light);
  background: var(--aurora-bg-glass);
}

.theme-option.active {
  border-color: var(--aurora-primary);
  box-shadow: 0 0 30px var(--aurora-primary-glow);
}

.theme-icon {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--aurora-bg-glass);
  border-radius: var(--aurora-radius-md);
  color: var(--aurora-text-secondary);
  transition: all var(--aurora-transition-normal);
}

.theme-option.active .theme-icon {
  background: var(--aurora-gradient-hero);
  color: white;
}

.theme-label {
  font-size: 14px;
  font-weight: 500;
}

.theme-check {
  position: absolute;
  top: 12px;
  right: 12px;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--aurora-gradient-hero);
  border-radius: 50%;
  color: white;
  font-size: 12px;
}

/* 设置列表 */
.settings-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
}

.setting-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.setting-icon {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--aurora-bg-glass);
  border-radius: var(--aurora-radius-md);
  color: var(--aurora-primary);
}

.setting-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.setting-label {
  font-size: 15px;
  font-weight: 500;
}

.setting-value {
  font-size: 13px;
  color: var(--aurora-text-tertiary);
}

.threshold-control {
  display: flex;
  align-items: center;
  gap: 16px;
}

.threshold-value {
  font-size: 16px;
  font-weight: 600;
  min-width: 70px;
}

/* 白名单 */
.whitelist-card {
  padding: 16px;
}

.whitelist-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 16px;
  color: var(--aurora-text-tertiary);
}

.whitelist-items {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 12px;
}

.whitelist-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 10px;
  background: var(--aurora-bg-glass);
  border-radius: var(--aurora-radius-sm);
}

.whitelist-icon {
  color: var(--aurora-text-tertiary);
}

.whitelist-path {
  flex: 1;
  font-family: 'JetBrains Mono', monospace;
  font-size: 13px;
  color: var(--aurora-text-secondary);
}

.add-whitelist {
  display: flex;
  gap: 12px;
}

/* 响应式 */
@media (max-width: 768px) {
  .content {
    padding: 16px;
  }
  
  .theme-selector {
    grid-template-columns: 1fr;
  }
  
  .setting-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }
  
  .threshold-control {
    width: 100%;
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>

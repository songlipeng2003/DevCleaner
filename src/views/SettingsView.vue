<template>
  <div class="settings">
    <a-layout class="layout">
      <a-layout-header class="header">
        <a-button @click="goBack">返回</a-button>
        <h2>设置</h2>
        <div>
          <a-button type="primary" :loading="isSaving" @click="saveSettings">保存</a-button>
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
        
        <a-spin :spinning="isLoading">
          <a-form layout="vertical">
            <a-form-item label="磁盘空间阈值 (GB)">
              <a-slider v-model:value="settings.threshold" :min="10" :max="500" />
              <span>{{ settings.threshold }} GB</span>
            </a-form-item>
            
            <a-form-item label="白名单（排除的路径）">
              <a-list size="small" :data-source="settings.whitelist" bordered>
                <template #renderItem="{ item }">
                  <a-list-item>
                    {{ item }}
                    <template #actions>
                      <a-button type="link" danger size="small" @click="removeWhitelist(item)">删除</a-button>
                    </template>
                  </a-list-item>
                </template>
              </a-list>
              <a-input-group compact style="margin-top: 8px">
                <a-input v-model:value="newWhitelist" style="width: calc(100% - 80px)" placeholder="添加排除路径" />
                <a-button type="primary" @click="addWhitelist" :disabled="!newWhitelist.trim()">添加</a-button>
              </a-input-group>
            </a-form-item>
            
            <a-form-item label="自动扫描">
              <a-switch v-model:checked="settings.autoScan" />
            </a-form-item>
            
            <a-form-item label="扫描间隔（天）" v-if="settings.autoScan">
              <a-input-number v-model:value="settings.scanInterval" :min="1" :max="30" />
            </a-form-item>

            <a-form-item label="主题">
              <a-radio-group v-model:value="settings.theme">
                <a-radio value="light">浅色</a-radio>
                <a-radio value="dark">深色</a-radio>
                <a-radio value="auto">跟随系统</a-radio>
              </a-radio-group>
            </a-form-item>
          </a-form>
        </a-spin>
      </a-layout-content>
    </a-layout>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { useSettingsStore } from '@/stores/settings'
import type { Settings } from '@/types'

const router = useRouter()
const settingsStore = useSettingsStore()

const newWhitelist = ref('')
const error = ref<string | null>(null)

const settings = computed(() => settingsStore.settings)
const isLoading = computed(() => settingsStore.isLoading)
const isSaving = computed(() => settingsStore.isSaving)

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
  max-width: 600px;
  border: 1px solid var(--nature-border-color);
  box-shadow: var(--nature-box-shadow);
}

/* 表单样式 */
:deep(.ant-form-item-label) {
  color: var(--nature-text-primary);
  font-weight: 500;
}

:deep(.ant-slider-track) {
  background-color: var(--nature-primary-color);
}

:deep(.ant-slider:hover .ant-slider-track) {
  background-color: var(--nature-primary-hover);
}

:deep(.ant-switch-checked) {
  background-color: var(--nature-primary-color);
}
</style>

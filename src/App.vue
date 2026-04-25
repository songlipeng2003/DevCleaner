<template>
  <a-config-provider :locale="zhCN" :theme="themeConfig">
    <div class="nature-bg-pattern"></div>
    <router-view />
  </a-config-provider>
</template>

<script setup lang="ts">
import zhCN from 'ant-design-vue/es/locale/zh_CN'
import { computed, watch, onMounted } from 'vue'
import { useSettingsStore } from '@/stores/settings'

const settingsStore = useSettingsStore()

// 计算主题配置
const themeConfig = computed(() => {
  const isDark = settingsStore.settings.theme === 'dark' ||
    (settingsStore.settings.theme === 'auto' && window.matchMedia('(prefers-color-scheme: dark)').matches)
  
  return {
    token: {
      colorPrimary: isDark ? '#66BB6A' : '#388E3C',
      colorSuccess: '#4CAF50',
      colorWarning: '#FF9800',
      colorError: '#F44336',
      colorInfo: '#2196F3',
      colorBgBase: isDark ? '#1A1A1A' : '#F5F5DC',
      colorTextBase: isDark ? '#E0E0E0' : '#3E2723',
      borderRadius: 12,
      fontSize: 14,
      lineHeight: 1.6,
    },
    components: {
      Button: {
        borderRadius: 8,
      },
      Card: {
        borderRadius: 12,
        boxShadow: '0 4px 12px rgba(56, 142, 60, 0.1)',
      },
      Input: {
        borderRadius: 8,
      },
      Slider: {
        trackBg: isDark ? '#66BB6A' : '#388E3C',
        trackHoverBg: isDark ? '#81C784' : '#2E7D32',
      },
      Switch: {
        colorPrimary: isDark ? '#66BB6A' : '#388E3C',
      },
    }
  }
})

// 监听主题变化并更新根元素属性
watch(() => settingsStore.settings.theme, (newTheme) => {
  updateThemeAttribute(newTheme)
})

onMounted(() => {
  updateThemeAttribute(settingsStore.settings.theme)
})

function updateThemeAttribute(theme: 'light' | 'dark' | 'auto') {
  const root = document.documentElement
  if (theme === 'auto') {
    const isDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    root.setAttribute('data-theme', isDark ? 'dark' : 'light')
  } else {
    root.setAttribute('data-theme', theme)
  }
}
</script>

<style>
#app {
  height: 100vh;
  overflow: hidden;
  position: relative;
}
</style>

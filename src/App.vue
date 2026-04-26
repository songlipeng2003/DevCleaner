<template>
  <a-config-provider
    :locale="zhCN"
    :theme="themeConfig"
  >
    <div class="nature-bg-pattern" />
    <router-view />
    <!-- 快捷键提示 -->
    <a-float-button
      ref="floatButtonRef"
      shape="square"
      style="position: fixed; bottom: 80px; right: 24px; background: rgba(0,0,0,0.6);"
      @click="showShortcuts"
    >
      <template #icon>
        <QuestionCircleOutlined />
      </template>
    </a-float-button>
  </a-config-provider>
</template>

<script setup lang="ts">
import zhCN from 'ant-design-vue/es/locale/zh_CN'
import { computed, watch, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import { QuestionCircleOutlined } from '@ant-design/icons-vue'
import { useSettingsStore } from '@/stores/settings'

const router = useRouter()
const route = useRoute()
const settingsStore = useSettingsStore()

// 快捷键映射
const shortcuts: Record<string, { key: string; desc: string }[]> = {
  global: [
    { key: 'Cmd/Ctrl + S', desc: '开始扫描' },
    { key: 'Cmd/Ctrl + R', desc: '刷新' },
    { key: 'Cmd/Ctrl + ,', desc: '设置' },
    { key: 'Esc', desc: '关闭弹窗' },
  ],
  home: [
    { key: 'Enter', desc: '查看详情' },
  ],
  settings: [
    { key: 'Cmd/Ctrl + S', desc: '保存设置' },
  ],
}

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
  // 添加全局快捷键监听
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
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

// 全局快捷键处理
function handleKeydown(e: KeyboardEvent) {
  const isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0
  const modifier = isMac ? e.metaKey : e.ctrlKey
  
  // Cmd/Ctrl + S: 开始扫描 (仅在首页)
  if (modifier && e.key === 's') {
    e.preventDefault()
    if (route.path === '/') {
      // 触发扫描 - 通过自定义事件
      window.dispatchEvent(new CustomEvent('devcleaner:scan'))
      message.info('开始扫描...')
    }
  }
  
  // Cmd/Ctrl + R: 刷新
  if (modifier && e.key === 'r') {
    e.preventDefault()
    window.dispatchEvent(new CustomEvent('devcleaner:refresh'))
    message.info('刷新中...')
  }
  
  // Cmd/Ctrl + ,: 打开设置
  if (modifier && e.key === ',') {
    e.preventDefault()
    if (route.path !== '/settings') {
      router.push('/settings')
    }
  }
  
  // Esc: 关闭弹窗 (使用 Ant Design 的 Modal)
  if (e.key === 'Escape') {
    Modal.destroyAll()
  }
}

// 显示快捷键帮助
function showShortcuts() {
  const allShortcuts = [
    ...shortcuts.global,
    ...(shortcuts[route.path.slice(1) as keyof typeof shortcuts] || []),
  ]
  
  const content = allShortcuts.map((s) => 
    `${s.desc}: ${s.key}`
  ).join('\n')
  
  Modal.info({
    title: '快捷键',
    content: content,
    okText: '关闭',
  })
}
</script>

<style>
#app {
  height: 100vh;
  overflow: hidden;
  position: relative;
}
</style>

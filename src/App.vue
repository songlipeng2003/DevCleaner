<template>
  <a-config-provider :theme="themeConfig">
    <div class="aurora-bg" />
    <router-view />
    <!-- 快捷键提示 -->
    <a-float-button
      v-if="settingsStore.settings.shortcutsEnabled"
      ref="floatButtonRef"
      shape="square"
      style="position: fixed; bottom: 80px; right: 24px; background: var(--aurora-bg-glass); backdrop-filter: blur(20px);"
      @click="showShortcuts"
    >
      <template #icon>
        <QuestionCircleOutlined />
      </template>
    </a-float-button>
  </a-config-provider>
</template>

<script setup lang="ts">
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
    { key: 'Ctrl/Cmd + S', desc: '开始扫描' },
    { key: 'Ctrl/Cmd + R', desc: '刷新' },
    { key: 'Ctrl/Cmd + ,', desc: '打开设置' },
    { key: 'Ctrl/Cmd + /', desc: '显示帮助' },
    { key: 'Esc', desc: '关闭弹窗' },
  ],
  home: [
    { key: 'Enter', desc: '查看详情' },
  ],
  settings: [
    { key: 'Ctrl/Cmd + S', desc: '保存设置' },
  ],
}

// 计算主题配置
const themeConfig = computed(() => {
  const isDark = settingsStore.settings.theme === 'dark' ||
    (settingsStore.settings.theme === 'auto' && window.matchMedia('(prefers-color-scheme: dark)').matches)
  
  return {
    token: {
      colorPrimary: '#667eea',
      colorSuccess: '#00ff88',
      colorWarning: '#ffb347',
      colorError: '#ff6b6b',
      colorInfo: '#00d9ff',
      colorBgBase: isDark ? '#050510' : '#f8fafc',
      colorTextBase: isDark ? '#ffffff' : '#1e293b',
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
      },
      Input: {
        borderRadius: 8,
      },
      Slider: {
        trackBg: '#667eea',
        trackHoverBg: '#818cf8',
      },
      Switch: {
        colorPrimary: '#667eea',
      },
    }
  }
})

// 监听主题变化并更新根元素属性
watch(() => settingsStore.settings.theme, (newTheme) => {
  updateThemeAttribute(newTheme)
})

// 监听快捷键开关变化
watch(() => settingsStore.settings.shortcutsEnabled, (enabled) => {
  if (enabled) {
    message.success('快捷键已启用')
  }
})

onMounted(() => {
  updateThemeAttribute(settingsStore.settings.theme)
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
  // 如果快捷键被禁用，忽略所有快捷键
  if (!settingsStore.settings.shortcutsEnabled) {
    return
  }

  // 忽略输入框中的按键
  const target = e.target as HTMLElement
  if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable) {
    // 允许 Esc 在输入框中关闭弹窗
    if (e.key !== 'Escape') {
      return
    }
  }

  const isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0
  const modifier = isMac ? e.metaKey : e.ctrlKey
  
  // Ctrl/Cmd + S: 开始扫描 / 保存设置
  if (modifier && e.key.toLowerCase() === 's') {
    e.preventDefault()
    if (route.path === '/') {
      window.dispatchEvent(new CustomEvent('devcleaner:scan'))
    } else if (route.path === '/settings') {
      window.dispatchEvent(new CustomEvent('devcleaner:save-settings'))
    }
    return
  }
  
  // Ctrl/Cmd + R: 刷新
  if (modifier && e.key.toLowerCase() === 'r') {
    e.preventDefault()
    window.dispatchEvent(new CustomEvent('devcleaner:refresh'))
    return
  }
  
  // Ctrl/Cmd + ,: 打开设置
  if (modifier && e.key === ',') {
    e.preventDefault()
    if (route.path !== '/settings') {
      router.push('/settings')
    }
    return
  }
  
  // Ctrl/Cmd + /: 显示快捷键帮助
  if (modifier && e.key === '/') {
    e.preventDefault()
    showShortcuts()
    return
  }
  
  // Esc: 关闭弹窗
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

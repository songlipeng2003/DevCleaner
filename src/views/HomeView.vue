<template>
  <div class="home">
    <a-layout class="layout">
      <a-layout-header class="header">
        <h1>DevCleaner</h1>
        <a-button type="primary" size="large" @click="startScan">
          开始扫描
        </a-button>
      </a-layout-header>
      
      <a-layout-content class="content">
        <a-row :gutter="24">
          <a-col :span="8" v-for="tool in supportedTools" :key="tool.id">
            <a-card class="tool-card" hoverable>
              <template #cover>
                <div class="tool-icon">{{ tool.icon }}</div>
              </template>
              <a-card-meta :title="tool.name" :description="tool.description" />
              <div class="tool-size">{{ tool.totalSize }}</div>
            </a-card>
          </a-col>
        </a-row>
      </a-layout-content>
      
      <a-layout-footer class="footer">
        <a-space>
          <span>总占用空间: {{ totalSize }}</span>
          <a-divider type="vertical" />
          <router-link to="/settings">设置</router-link>
        </a-space>
      </a-layout-footer>
    </a-layout>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'

const router = useRouter()

interface DevTool {
  id: string
  name: string
  icon: string
  description: string
  totalSize: string
}

const supportedTools = ref<DevTool[]>([
  { id: 'npm', name: 'npm', icon: '📦', description: 'Node.js 包管理器', totalSize: '0 MB' },
  { id: 'docker', name: 'Docker', icon: '🐳', description: '容器运行时', totalSize: '0 MB' },
  { id: 'xcode', name: 'Xcode', icon: '🍎', description: 'iOS/macOS 开发工具', totalSize: '0 MB' },
  { id: 'homebrew', name: 'Homebrew', icon: '🍺', description: 'macOS 包管理器', totalSize: '0 MB' },
  { id: 'python', name: 'Python', icon: '🐍', description: 'Python 环境', totalSize: '0 MB' },
  { id: 'go', name: 'Go', icon: '🔵', description: 'Go modules', totalSize: '0 MB' },
])

const totalSize = computed(() => '0 MB')

const startScan = () => {
  message.loading('正在扫描...', 1)
  router.push('/scan')
}
</script>

<style scoped>
.home {
  height: 100vh;
  background: #f0f2f5;
}

.layout {
  height: 100%;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fff;
  padding: 0 24px;
}

.header h1 {
  font-size: 20px;
  margin: 0;
}

.content {
  padding: 24px;
  overflow-y: auto;
}

.tool-card {
  margin-bottom: 16px;
}

.tool-icon {
  font-size: 48px;
  text-align: center;
  padding: 24px;
}

.tool-size {
  margin-top: 8px;
  color: #1890ff;
  font-weight: bold;
}

.footer {
  text-align: center;
  background: #fff;
}
</style>

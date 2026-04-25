<template>
  <div class="scan">
    <a-layout class="layout">
      <a-layout-header class="header">
        <a-button @click="goBack">返回</a-button>
        <h2>扫描中...</h2>
        <div></div>
      </a-layout-header>
      
      <a-layout-content class="content">
        <a-progress :percent="scanProgress" status="active" />
        
        <a-list class="scan-list" :data-source="scanningTools">
          <template #renderItem="{ item }">
            <a-list-item>
              <a-list-item-meta :title="item.name" :description="item.path">
                <template #avatar>
                  <a-spin v-if="item.scanning" />
                  <a-icon type="check-circle" v-else-if="item.done" style="color: #52c41a" />
                  <a-icon type="clock-circle" v-else style="color: #d9d9d9" />
                </template>
              </a-list-item-meta>
              <div>{{ item.size }}</div>
            </a-list-item>
          </template>
        </a-list>
      </a-layout-content>
    </a-layout>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const scanProgress = ref(0)

interface ScanningTool {
  name: string
  path: string
  size: string
  scanning: boolean
  done: boolean
}

const scanningTools = ref<ScanningTool[]>([
  { name: 'npm', path: '~/.npm', size: '0 MB', scanning: true, done: false },
  { name: 'Docker', path: '~/Library/Containers/com.docker.docker', size: '0 MB', scanning: false, done: false },
  { name: 'Xcode', path: '~/Library/Developer', size: '0 MB', scanning: false, done: false },
])

const goBack = () => {
  router.push('/')
}
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
}

.content {
  padding: 24px;
  margin: 24px;
  background: #fff;
  border-radius: 8px;
}

.scan-list {
  margin-top: 24px;
}
</style>

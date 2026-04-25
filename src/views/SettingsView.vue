<template>
  <div class="settings">
    <a-layout class="layout">
      <a-layout-header class="header">
        <a-button @click="goBack">返回</a-button>
        <h2>设置</h2>
        <div></div>
      </a-layout-header>
      
      <a-layout-content class="content">
        <a-form layout="vertical">
          <a-form-item label="磁盘空间阈值">
            <a-slider v-model:value="settings.threshold" :min="10" :max="500" />
            <span>{{ settings.threshold }} GB</span>
          </a-form-item>
          
          <a-form-item label="白名单（排除的路径）">
            <a-list size="small" :data-source="settings.whitelist">
              <template #renderItem="{ item }">
                <a-list-item>
                  {{ item }}
                  <template #actions>
                    <a-button type="link" danger size="small">删除</a-button>
                  </template>
                </a-list-item>
              </template>
            </a-list>
            <a-input-group compact style="margin-top: 8px">
              <a-input v-model:value="newWhitelist" style="width: calc(100% - 80px)" placeholder="添加排除路径" />
              <a-button type="primary" @click="addWhitelist">添加</a-button>
            </a-input-group>
          </a-form-item>
          
          <a-form-item label="自动扫描">
            <a-switch v-model:checked="settings.autoScan" />
          </a-form-item>
          
          <a-form-item label="扫描间隔（天）" v-if="settings.autoScan">
            <a-input-number v-model:value="settings.scanInterval" :min="1" :max="30" />
          </a-form-item>
        </a-form>
      </a-layout-content>
    </a-layout>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const newWhitelist = ref('')

const settings = reactive({
  threshold: 100,
  whitelist: [
    '~/Documents',
    '~/Desktop'
  ],
  autoScan: false,
  scanInterval: 7
})

const goBack = () => {
  router.push('/')
}

const addWhitelist = () => {
  if (newWhitelist.value && !settings.whitelist.includes(newWhitelist.value)) {
    settings.whitelist.push(newWhitelist.value)
    newWhitelist.value = ''
  }
}
</script>

<style scoped>
.settings {
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
  max-width: 600px;
}
</style>

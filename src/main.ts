import { createApp } from 'vue'
import { createPinia } from 'pinia'
import Antd from 'ant-design-vue'
import App from './App.vue'
import router from './router'
import 'ant-design-vue/dist/reset.css'
import './assets/theme-aurora.css'
import './assets/global.css'
import { initAptabase } from './services/analytics'

// 初始化 Aptabase 分析服务
// 设置 VITE_APTABASE_APP_KEY 环境变量或在构建时配置
const appKey = (import.meta.env.VITE_APTABASE_APP_KEY as string) || ''
if (appKey) {
  initAptabase(appKey)
}

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(Antd)

app.mount('#app')

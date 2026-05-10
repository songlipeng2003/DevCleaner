/**
 * Aptabase 分析服务
 * 直接调用 Aptabase API 进行事件追踪
 * 文档: https://aptabase.com
 */

interface AptabaseEvent {
  eventName: string
  eventData: Record<string, string | number | boolean>
  platform: string
  sdkVersion: string
  appVersion: string
  locale: string
  osName: string
  osVersion: string
  isDebug: boolean
}

const APTABASE_URL = 'https://cloud.aptabase.com'
const SDK_VERSION = '1.0.0'

function getAppKey(): string {
  // 从环境变量或 localStorage 获取 App Key
  return (import.meta.env.VITE_APTABASE_APP_KEY as string) ||
         localStorage.getItem('aptabase_app_key') ||
         ''
}

function setAppKey(key: string): void {
  localStorage.setItem('aptabase_app_key', key)
}

function getDeviceId(): string {
  let deviceId = localStorage.getItem('aptabase_device_id')
  if (!deviceId) {
    deviceId = crypto.randomUUID()
    localStorage.setItem('aptabase_device_id', deviceId)
  }
  return deviceId
}

function getSessionId(): string {
  let sessionId = sessionStorage.getItem('aptabase_session_id')
  if (!sessionId) {
    sessionId = crypto.randomUUID()
    sessionStorage.setItem('aptabase_session_id', sessionId)
  }
  return sessionId
}

function sendEvent(eventName: string, properties?: Record<string, string | number | boolean>): void {
  const appKey = getAppKey()

  // 如果没有设置 App Key，静默跳过
  if (!appKey) {
    return
  }

  // 在非浏览器环境中静默跳过（如 SSR 或测试环境）
  if (typeof navigator === 'undefined') {
    return
  }

  const event: AptabaseEvent = {
    eventName,
    eventData: properties || {},
    platform: 'web',
    sdkVersion: SDK_VERSION,
    appVersion: '0.1.0',
    locale: navigator.language || 'en',
    osName: 'Unknown',
    osVersion: '',
    isDebug: false,
  }

  // 使用 sendBeacon 发送请求，确保页面关闭时也能发送
  const data = {
    clientId: getDeviceId(),
    sessionId: getSessionId(),
    events: [event],
  }

  try {
    // 检查 sendBeacon 是否可用
    if (typeof navigator.sendBeacon !== 'function') {
      return
    }
    navigator.sendBeacon(
      `${APTABASE_URL}/v0/events?appKey=${encodeURIComponent(appKey)}`,
      new Blob([JSON.stringify(data)], { type: 'application/json' })
    )
  } catch (e) {
    // 静默处理错误，不影响用户使用
  }
}

/**
 * 初始化 Aptabase（需要先调用此方法设置 App Key）
 * @param appKey Aptabase App Key
 */
export function initAptabase(appKey: string): void {
  setAppKey(appKey)
}

/**
 * 检查是否已初始化
 */
export function isAptabaseInitialized(): boolean {
  return !!getAppKey()
}

/**
 * 追踪应用启动事件
 */
export function trackAppStarted(): void {
  sendEvent('app_started')
}

/**
 * 追踪扫描开始事件
 * @param toolType 工具类型 (node, maven, gradle 等)
 */
export function trackScanStart(toolType: string): void {
  sendEvent('scan_start', { tool_type: toolType })
}

/**
 * 追踪扫描完成事件
 * @param toolType 工具类型
 * @param itemCount 扫描到的项目数量
 */
export function trackScanComplete(toolType: string, itemCount: number): void {
  sendEvent('scan_complete', {
    tool_type: toolType,
    item_count: itemCount
  })
}

/**
 * 追踪清理开始事件
 * @param toolType 工具类型
 * @param size 预估清理大小
 */
export function trackCleanStart(toolType: string, size: number): void {
  sendEvent('clean_start', {
    tool_type: toolType,
    size_bytes: size
  })
}

/**
 * 追踪清理完成事件
 * @param toolType 工具类型
 * @param size 实际清理大小
 */
export function trackCleanComplete(toolType: string, size: number): void {
  sendEvent('clean_complete', {
    tool_type: toolType,
    size_bytes: size
  })
}

/**
 * 追踪页面浏览事件
 * @param pageName 页面名称
 */
export function trackPageView(pageName: string): void {
  sendEvent('page_view', { page: pageName })
}

/**
 * 追踪设置变更事件
 * @param settingKey 设置项名称
 * @param value 设置值
 */
export function trackSettingChange(settingKey: string, value: string | number | boolean): void {
  sendEvent('setting_change', {
    key: settingKey,
    value: String(value)
  })
}

/**
 * 追踪错误事件
 * @param errorType 错误类型
 * @param message 错误信息
 */
export function trackError(errorType: string, message: string): void {
  sendEvent('error', {
    type: errorType,
    message
  })
}

/**
 * 通用事件追踪
 * @param eventName 事件名称
 * @param properties 事件属性
 */
export function track(eventName: string, properties?: Record<string, string | number | boolean>): void {
  sendEvent(eventName, properties)
}

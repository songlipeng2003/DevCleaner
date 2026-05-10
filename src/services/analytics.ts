/**
 * Aptabase 分析服务
 * 用于追踪用户行为和事件
 */
import { trackEvent } from '@aptabase/tauri';

/**
 * 追踪应用启动事件
 */
export function trackAppStarted(): void {
  trackEvent('app_started');
}

/**
 * 追踪扫描开始事件
 * @param toolType 工具类型 (node, maven, gradle 等)
 */
export function trackScanStart(toolType: string): void {
  trackEvent('scan_start', { tool_type: toolType });
}

/**
 * 追踪扫描完成事件
 * @param toolType 工具类型
 * @param itemCount 扫描到的项目数量
 */
export function trackScanComplete(toolType: string, itemCount: number): void {
  trackEvent('scan_complete', {
    tool_type: toolType,
    item_count: itemCount
  });
}

/**
 * 追踪清理开始事件
 * @param toolType 工具类型
 * @param size 预估清理大小
 */
export function trackCleanStart(toolType: string, size: number): void {
  trackEvent('clean_start', {
    tool_type: toolType,
    size_bytes: size
  });
}

/**
 * 追踪清理完成事件
 * @param toolType 工具类型
 * @param size 实际清理大小
 */
export function trackCleanComplete(toolType: string, size: number): void {
  trackEvent('clean_complete', {
    tool_type: toolType,
    size_bytes: size
  });
}

/**
 * 追踪页面浏览事件
 * @param pageName 页面名称
 */
export function trackPageView(pageName: string): void {
  trackEvent('page_view', { page: pageName });
}

/**
 * 追踪设置变更事件
 * @param settingKey 设置项名称
 * @param value 设置值
 */
export function trackSettingChange(settingKey: string, value: string | number | boolean): void {
  trackEvent('setting_change', {
    key: settingKey,
    value: String(value)
  });
}

/**
 * 追踪错误事件
 * @param errorType 错误类型
 * @param message 错误信息
 */
export function trackError(errorType: string, message: string): void {
  trackEvent('error', {
    type: errorType,
    message
  });
}

/**
 * 通用事件追踪
 * @param eventName 事件名称
 * @param properties 事件属性
 */
export function track(eventName: string, properties?: Record<string, string | number | boolean>): void {
  trackEvent(eventName, properties);
}

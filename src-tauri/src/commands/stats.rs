use crate::commands::types::*;
use chrono::{Datelike, Utc};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::fs;
use std::sync::Mutex;

// ============== 统计数据结构 ==============

#[derive(Debug, Serialize, Deserialize, Default, Clone)]
pub struct UsageStats {
    #[serde(rename = "totalCleaned")]
    pub total_cleaned: i64,
    #[serde(rename = "totalFiles")]
    pub total_files: i32,
    #[serde(rename = "cleanCount")]
    pub clean_count: i32,
    #[serde(rename = "lastClean")]
    pub last_clean: Option<i64>,
    #[serde(rename = "cleanByTool")]
    pub clean_by_tool: HashMap<String, i64>,
    #[serde(rename = "cleanByMonth")]
    pub clean_by_month: HashMap<String, i64>,
}

pub(crate) static STATS_MUTEX: once_cell::sync::Lazy<Mutex<()>> =
    once_cell::sync::Lazy::new(|| Mutex::new(()));

// ============== 统计相关命令 ==============

#[tauri::command]
pub async fn get_usage_stats() -> Result<UsageStats, String> {
    let _lock = STATS_MUTEX.lock().map_err(|e| e.to_string())?;

    let stats_path = get_stats_path();

    if stats_path.exists() {
        let content = fs::read_to_string(&stats_path).map_err(|e| e.to_string())?;
        serde_json::from_str(&content).map_err(|e| e.to_string())
    } else {
        Ok(UsageStats::default())
    }
}

#[tauri::command]
pub async fn record_clean(
    _tool_id: String,
    tool_name: String,
    size: i64,
    file_count: i32,
) -> Result<(), String> {
    let _lock = STATS_MUTEX.lock().map_err(|e| e.to_string())?;

    let stats_path = get_stats_path();
    let mut stats = if stats_path.exists() {
        let content = fs::read_to_string(&stats_path).map_err(|e| e.to_string())?;
        serde_json::from_str(&content).unwrap_or_default()
    } else {
        UsageStats::default()
    };

    // 更新统计数据
    stats.total_cleaned += size;
    stats.total_files += file_count;
    stats.clean_count += 1;
    stats.last_clean = Some(Utc::now().timestamp());

    // 按工具统计
    *stats.clean_by_tool.entry(tool_name.clone()).or_insert(0) += size;

    // 按月份统计
    let now = Utc::now();
    let month_key = format!("{}-{:02}", now.year(), now.month());
    *stats.clean_by_month.entry(month_key).or_insert(0) += size;

    // 保存统计
    let content = serde_json::to_string_pretty(&stats).map_err(|e| e.to_string())?;
    fs::write(&stats_path, content).map_err(|e| e.to_string())
}

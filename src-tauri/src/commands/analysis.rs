use crate::commands::types::*;
use chrono::{Datelike, Utc};
use std::collections::HashMap;
use std::path::Path;
use tauri::Emitter;

// ============== 磁盘分析命令 (v0.3.0) ==============

#[tauri::command]
pub async fn get_disk_analysis(
    app: tauri::AppHandle,
) -> Result<DiskAnalysisResult, String> {
    let tools = get_tools();
    let mut category_map: HashMap<String, Vec<DiskAnalysisItem>> = HashMap::new();
    let mut total_size: i64 = 0;
    let mut total_cleanable: i64 = 0;

    // 获取总磁盘使用情况作为基准
    let disk_usage = get_disk_usage().unwrap_or(DiskUsage {
        total: 500_000_000_000, // 默认 500GB
        used: 250_000_000_000,
        free: 250_000_000_000,
    });

    // 扫描每个工具的缓存目录
    let total_paths: usize = tools.iter().map(|t| t.paths.len()).sum();

    let mut paths_scanned: usize = 0;

    for tool in tools {
        let tool_id = tool.tool_id.clone();
        let tool_name = tool.tool_name.clone();

        for path_pattern in tool.paths {
            paths_scanned += 1;
            let expanded = expand_path(&path_pattern.path);

            if expanded.exists() {
                let (size, file_count, last_modified) = scan_directory_size(&expanded);

                if size > 0 {
                    total_size += size;

                    // 判断是否可清理
                    let is_cleanable = is_path_cleanable(&expanded, &tool_id);
                    let clean_reason = if is_cleanable {
                        Some(get_clean_reason(&expanded, &tool_id))
                    } else {
                        None
                    };

                    if is_cleanable {
                        total_cleanable += size;
                    }

                    // 计算占比
                    let percentage = if disk_usage.used > 0 {
                        (size as f64 / disk_usage.used as f64) * 100.0
                    } else {
                        0.0
                    };

                    let item = DiskAnalysisItem {
                        name: expanded
                            .file_name()
                            .map(|n| n.to_string_lossy().to_string())
                            .unwrap_or_default(),
                        path: expanded.to_string_lossy().to_string(),
                        tool_id: Some(tool_id.clone()),
                        size,
                        percentage,
                        file_count,
                        last_modified,
                        is_cleanable,
                        clean_reason,
                    };

                    // 按工具类型分组
                    category_map
                        .entry(tool_name.clone())
                        .or_insert_with(Vec::new)
                        .push(item);
                }
            }

            // 发送进度
            if paths_scanned % 5 == 0 {
                let progress = (paths_scanned as f32 / total_paths as f32).min(1.0);
                let _ = app.emit(
                    "scan-progress",
                    serde_json::json!({
                        "toolId": "analysis",
                        "toolName": "Disk Analysis",
                        "progress": progress,
                        "currentPath": expanded.to_string_lossy(),
                        "pathsScanned": paths_scanned,
                        "totalPaths": total_paths
                    }),
                );
            }
        }
    }

    // 构建分类结果
    let mut categories: Vec<DiskAnalysisCategory> = category_map
        .into_iter()
        .map(|(name, items)| {
            let total_size: i64 = items.iter().map(|i| i.size).sum();
            let item_count = items.len() as i32;

            DiskAnalysisCategory {
                name,
                items,
                total_size,
                item_count,
            }
        })
        .collect();

    // 按大小降序排序
    categories.sort_by_key(|c| std::cmp::Reverse(c.total_size));

    let total_items: i32 = categories.iter().map(|c| c.item_count).sum();

    Ok(DiskAnalysisResult {
        categories,
        total_size,
        cleanable_size: total_cleanable,
        total_items,
        timestamp: Utc::now().timestamp(),
    })
}

#[tauri::command]
pub async fn get_cache_trends(months: Option<i32>) -> Result<Vec<CacheTrend>, String> {
    let _lock = crate::commands::stats::STATS_MUTEX.lock().map_err(|e| e.to_string())?;

    let stats_path = get_stats_path();

    let stats = if stats_path.exists() {
        let content = std::fs::read_to_string(&stats_path).map_err(|e| e.to_string())?;
        serde_json::from_str(&content).unwrap_or_default()
    } else {
        crate::commands::stats::UsageStats::default()
    };

    let month_count = months.unwrap_or(6) as usize;

    // 生成趋势数据
    let now = Utc::now();
    let mut trends: Vec<CacheTrend> = Vec::new();

    for i in 0..month_count {
        let date = if i == 0 {
            format!("{}-{:02}", now.year(), now.month())
        } else {
            // 计算前几个月的日期
            let mut year = now.year();
            let mut month = now.month() as i32 - i as i32;

            while month <= 0 {
                month += 12;
                year -= 1;
            }

            format!("{}-{:02}", year, month)
        };

        let size = stats
            .clean_by_month
            .get(&date)
            .copied()
            .unwrap_or(0);

        trends.push(CacheTrend { date, size });
    }

    // 按日期升序排序
    trends.sort_by_key(|t| t.date.clone());

    Ok(trends)
}

// 辅助函数：判断路径是否可清理
fn is_path_cleanable(path: &Path, tool_id: &str) -> bool {
    // 根据工具类型判断
    match tool_id {
        // 始终可清理
        "npm" | "yarn" | "pnpm" | "node_modules" | "gradle" | "maven" | "cocoapods"
        | "rubygems" | "pip" | "cargo" | "go" | "composer" => true,

        // 需要谨慎的
        "vscode" | "idea" | "android" | "xcode" => {
            // 只清理明确的缓存目录
            let path_str = path.to_string_lossy().to_lowercase();
            path_str.contains("cache")
                || path_str.contains("caches")
                || path_str.contains("derived")
        }

        // 默认不清理
        _ => false,
    }
}

// 辅助函数：获取清理原因
fn get_clean_reason(path: &Path, tool_id: &str) -> String {
    let path_str = path.to_string_lossy().to_lowercase();

    if path_str.contains("node_modules") {
        "Dependencies can be reinstalled with npm/yarn install".to_string()
    } else if path_str.contains("cache") || path_str.contains("caches") {
        "Cache files can be safely removed".to_string()
    } else if path_str.contains("build") || path_str.contains("target") {
        "Build output can be regenerated".to_string()
    } else if path_str.contains("derived") {
        "Derived data can be regenerated by IDE".to_string()
    } else {
        format!("{} cache can be safely cleaned", tool_id)
    }
}

// 获取磁盘使用情况
fn get_disk_usage() -> Result<DiskUsage, String> {
    #[cfg(target_os = "macos")]
    {
        let output = std::process::Command::new("df")
            .args(["-k", "/"])
            .output()
            .map_err(|e| e.to_string())?;

        let output_str = String::from_utf8_lossy(&output.stdout);
        let lines: Vec<&str> = output_str.lines().collect();

        if lines.len() >= 2 {
            let parts: Vec<&str> = lines[1].split_whitespace().collect();
            if parts.len() >= 4 {
                let total_kb: i64 = parts[1].parse().unwrap_or(0);
                let used_kb: i64 = parts[2].parse().unwrap_or(0);
                let free_kb: i64 = parts[3].parse().unwrap_or(0);

                return Ok(DiskUsage {
                    total: total_kb * 1024,
                    used: used_kb * 1024,
                    free: free_kb * 1024,
                });
            }
        }

        Err("Failed to parse disk usage".to_string())
    }

    #[cfg(not(target_os = "macos"))]
    {
        // 其他平台的简化实现
        Ok(DiskUsage {
            total: 500_000_000_000,
            used: 250_000_000_000,
            free: 250_000_000_000,
        })
    }
}

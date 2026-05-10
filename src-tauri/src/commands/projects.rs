use crate::commands::types::*;
use chrono::Utc;
use std::collections::HashMap;
use std::fs;
use std::path::PathBuf;
use std::sync::Mutex;
use tauri::Emitter;

// ============== 项目相关命令 ==============

#[tauri::command]
pub async fn scan_projects(
    app: tauri::AppHandle,
    base_path: String,
) -> Result<ProjectScanResult, String> {
    let expanded = expand_path(&base_path);

    if !expanded.exists() {
        return Err("Base path does not exist".to_string());
    }

    let tools = get_tools();
    let mut projects: Vec<ProjectInfo> = Vec::new();
    let mut total_size: i64 = 0;
    let mut total_count: i32 = 0;

    // 扫描目录下的所有项目
    let entries = std::fs::read_dir(&expanded).map_err(|e| e.to_string())?;

    for (index, entry) in entries.filter_map(|e| e.ok()).enumerate() {
        let path = entry.path();
        if path.is_dir() {
            let project_name = path
                .file_name()
                .map(|n| n.to_string_lossy().to_string())
                .unwrap_or_default();

            // 尝试匹配工具
            for tool in &tools {
                for path_pattern in &tool.paths {
                    if path.to_string_lossy().contains(&path_pattern.path) {
                        let (size, file_count, last_modified) = scan_directory_size(&path);

                        if size > 0 {
                            let project = ProjectInfo {
                                name: project_name.clone(),
                                path: path.to_string_lossy().to_string(),
                                tool_id: tool.tool_id.clone(),
                                size,
                                cache_size: size,
                                last_modified,
                                file_count,
                                cleanable: true,
                                clean_reason: Some(format!(
                                    "This is a {} cache directory",
                                    tool.tool_name
                                )),
                            };

                            projects.push(project.clone());
                            total_size += size;
                            total_count += 1;
                        }
                        break;
                    }
                }
            }

            // 发送进度
            if index % 10 == 0 {
                let _ = app.emit(
                    "scan-progress",
                    serde_json::json!({
                        "toolId": "projects",
                        "toolName": "Project Scan",
                        "progress": index as f32 / 100.0,
                        "currentPath": path.to_string_lossy(),
                        "pathsScanned": index,
                        "totalPaths": 100
                    }),
                );
            }
        }
    }

    Ok(ProjectScanResult {
        projects,
        total_size,
        total_count,
    })
}

#[tauri::command]
pub async fn get_clean_preview(paths: Vec<String>) -> Result<Vec<PreviewItem>, String> {
    let mut previews: Vec<PreviewItem> = Vec::new();

    for path_str in paths {
        let expanded = expand_path(&path_str);
        if expanded.exists() {
            let (size, file_num, last_modified) = scan_directory_size(&expanded);

            if size > 0 {
                previews.push(PreviewItem {
                    path: expanded.to_string_lossy().to_string(),
                    size,
                    file_num,
                    last_modified,
                });
            }
        }
    }

    Ok(previews)
}

#[tauri::command]
pub async fn clean_paths(
    app: tauri::AppHandle,
    paths: Vec<String>,
) -> Result<CleanResult, String> {
    let mut total_cleaned: i64 = 0;
    let mut total_files: i32 = 0;
    let mut failed_paths: Vec<String> = Vec::new();

    for path_str in &paths {
        let expanded = expand_path(path_str);
        if expanded.exists() {
            let (size, file_num, _) = scan_directory_size(&expanded);

            match delete_directory_contents(&expanded) {
                Ok(_) => {
                    total_cleaned += size;
                    total_files += file_num;
                }
                Err(e) => {
                    failed_paths.push(format!("{}: {}", expanded.display(), e));
                }
            }
        }
    }

    // 发送清理完成事件
    let _ = app.emit(
        "clean-complete",
        serde_json::json!({
            "toolId": "custom",
            "cleaned": total_cleaned,
            "fileNum": total_files
        }),
    );

    Ok(CleanResult {
        tool_id: "custom".to_string(),
        cleaned: total_cleaned,
        failed: failed_paths,
        file_num: total_files,
    })
}

// 辅助函数：删除目录内容但保留目录本身
fn delete_directory_contents(path: &PathBuf) -> Result<(), String> {
    if !path.exists() {
        return Ok(());
    }

    for entry in std::fs::read_dir(path).map_err(|e| e.to_string())? {
        let entry = entry.map_err(|e| e.to_string())?;
        let entry_path = entry.path();

        if entry_path.is_dir() {
            std::fs::remove_dir_all(&entry_path).map_err(|e| e.to_string())?;
        } else {
            std::fs::remove_file(&entry_path).map_err(|e| e.to_string())?;
        }
    }

    Ok(())
}

static HISTORY_MUTEX: once_cell::sync::Lazy<Mutex<()>> =
    once_cell::sync::Lazy::new(|| Mutex::new(()));

fn get_history_path() -> std::path::PathBuf {
    let data_path = get_data_path();
    if !data_path.exists() {
        let _ = fs::create_dir_all(&data_path);
    }
    data_path.join("history.json")
}

#[tauri::command]
pub async fn get_clean_history(
    limit: Option<i32>,
) -> Result<Vec<CleanHistory>, String> {
    let _lock = HISTORY_MUTEX.lock().map_err(|e| e.to_string())?;

    let history_path = get_history_path();

    if history_path.exists() {
        let content = fs::read_to_string(&history_path).map_err(|e| e.to_string())?;
        let mut history: Vec<CleanHistory> =
            serde_json::from_str(&content).unwrap_or_default();

        // 按日期降序排序
        history.sort_by_key(|h| std::cmp::Reverse(h.date));

        // 应用限制
        if let Some(l) = limit {
            history.truncate(l as usize);
        }

        Ok(history)
    } else {
        Ok(vec![])
    }
}

#[tauri::command]
pub async fn record_clean_history(
    tool_id: String,
    tool_name: String,
    size: i64,
    file_count: i32,
    paths: Vec<String>,
    note: Option<String>,
) -> Result<(), String> {
    let _lock = HISTORY_MUTEX.lock().map_err(|e| e.to_string())?;

    let history_path = get_history_path();
    let mut history: Vec<CleanHistory> = if history_path.exists() {
        let content = fs::read_to_string(&history_path).map_err(|e| e.to_string())?;
        serde_json::from_str(&content).unwrap_or_default()
    } else {
        vec![]
    };

    let now = Utc::now();
    let clean_history = CleanHistory {
        clean_id: format!("{}-{}", now.timestamp(), tool_id),
        date: now.timestamp(),
        tool_id,
        tool_name,
        size,
        file_count,
        paths,
        note,
    };

    history.push(clean_history);

    // 只保留最近100条记录
    if history.len() > 100 {
        history.sort_by_key(|h| std::cmp::Reverse(h.date));
        history.truncate(100);
    }

    let content = serde_json::to_string_pretty(&history).map_err(|e| e.to_string())?;
    fs::write(&history_path, content).map_err(|e| e.to_string())
}

#[tauri::command]
pub async fn export_clean_report(
    start_date: Option<i64>,
    end_date: Option<i64>,
) -> Result<CleanReport, String> {
    let _lock = HISTORY_MUTEX.lock().map_err(|e| e.to_string())?;

    let history_path = get_history_path();
    let history: Vec<CleanHistory> = if history_path.exists() {
        let content = fs::read_to_string(&history_path).map_err(|e| e.to_string())?;
        serde_json::from_str(&content).unwrap_or_default()
    } else {
        vec![]
    };

    // 过滤日期范围
    let filtered: Vec<CleanHistory> = history
        .into_iter()
        .filter(|h| {
            let date = h.date;
            let after_start = start_date.map(|s| date >= s).unwrap_or(true);
            let before_end = end_date.map(|e| date <= e).unwrap_or(true);
            after_start && before_end
        })
        .collect();

    // 按工具分组
    let mut tool_groups: HashMap<String, Vec<&CleanHistory>> = HashMap::new();
    for item in &filtered {
        tool_groups.entry(item.tool_id.clone()).or_default().push(item);
    }

    let mut items: Vec<ReportItem> = Vec::new();
    let mut total_size: i64 = 0;
    let mut total_files: i32 = 0;

    for (tool_id, group) in tool_groups {
        let tool_name = group.first().map(|h| h.tool_name.clone()).unwrap_or_default();
        let size: i64 = group.iter().map(|h| h.size).sum();
        let file_count: i32 = group.iter().map(|h| h.file_count).sum();
        let paths: Vec<String> = group.iter().flat_map(|h| h.paths.clone()).collect();

        total_size += size;
        total_files += file_count;

        items.push(ReportItem {
            tool_id,
            tool_name,
            size,
            file_count,
            paths,
        });
    }

    // 按清理大小降序排序
    items.sort_by_key(|i| std::cmp::Reverse(i.size));

    let now = Utc::now();
    let report = CleanReport {
        report_id: format!("report-{}", now.timestamp()),
        date: now.format("%Y-%m-%d %H:%M:%S").to_string(),
        total_size,
        total_files,
        items,
    };

    Ok(report)
}

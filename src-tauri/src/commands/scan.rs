use crate::commands::types::*;
use std::path::PathBuf;
use tauri::Emitter;

// ============== 工具扫描命令 ==============

#[tauri::command]
pub async fn get_tool_list() -> Result<Vec<ToolInfo>, String> {
    let tools = get_tools();
    let tool_infos: Vec<ToolInfo> = tools
        .into_iter()
        .map(|t| ToolInfo {
            id: t.tool_id.clone(),
            name: t.tool_name.clone(),
            paths: t.paths.into_iter().map(|p| p.path.clone()).collect(),
            enabled: true,
            description: None,
        })
        .collect();

    Ok(tool_infos)
}

#[tauri::command]
pub async fn get_tool_info(tool_id: String) -> Result<ToolInfo, String> {
    let tools = get_tools();
    let tool = tools
        .into_iter()
        .find(|t| t.tool_id == tool_id)
        .ok_or_else(|| "Tool not found".to_string())?;

    Ok(ToolInfo {
        id: tool.tool_id.clone(),
        name: tool.tool_name.clone(),
        paths: tool.paths.into_iter().map(|p| p.path.clone()).collect(),
        enabled: true,
        description: None,
    })
}

#[tauri::command]
pub async fn scan_tool(
    app: tauri::AppHandle,
    tool_id: String,
) -> Result<ScanResult, String> {
    let tools = get_tools();
    let tool = tools
        .into_iter()
        .find(|t| t.tool_id == tool_id)
        .ok_or_else(|| "Tool not found".to_string())?;

    let paths = tool.paths;
    let total_paths = paths.len() as i32;

    let mut total_size: i64 = 0;
    let mut total_files: i32 = 0;
    let mut latest_modified: i64 = 0;

    for (index, path_pattern) in paths.iter().enumerate() {
        let expanded = expand_path(&path_pattern.path);
        let (size, file_num, last_modified) = scan_directory_size(&expanded);

        total_size += size;
        total_files += file_num;
        if last_modified > latest_modified {
            latest_modified = last_modified;
        }

        // 发送进度
        let progress = ScanProgress {
            tool_id: tool.tool_id.clone(),
            tool_name: tool.tool_name.clone(),
            progress: (index + 1) as f32 / total_paths as f32,
            current_path: expanded.to_string_lossy().to_string(),
            paths_scanned: (index + 1) as i32,
            total_paths,
        };

        let _ = app.emit("scan-progress", &progress);
    }

    Ok(ScanResult {
        tool_id,
        path: paths.first().map(|p| p.path.clone()).unwrap_or_default(),
        size: total_size,
        file_num: total_files,
        last_modified: latest_modified,
        description: Some(tool.tool_name),
    })
}

#[tauri::command]
pub async fn scan_all_tools(
    app: tauri::AppHandle,
) -> Result<Vec<ScanResult>, String> {
    let tools = get_tools();
    let total_tools = tools.len() as i32;

    let mut results: Vec<ScanResult> = Vec::new();

    for (index, tool) in tools.into_iter().enumerate() {
        let paths = tool.paths.clone();
        let tool_id = tool.tool_id.clone();
        let tool_name = tool.tool_name.clone();

        let mut total_size: i64 = 0;
        let mut total_files: i32 = 0;
        let mut latest_modified: i64 = 0;

        for path_pattern in paths {
            let expanded = expand_path(&path_pattern.path);
            let (size, file_num, last_modified) = scan_directory_size(&expanded);

            total_size += size;
            total_files += file_num;
            if last_modified > latest_modified {
                latest_modified = last_modified;
            }
        }

        results.push(ScanResult {
            tool_id: tool_id.clone(),
            path: "".to_string(),
            size: total_size,
            file_num: total_files,
            last_modified: latest_modified,
            description: Some(tool_name.clone()),
        });

        // 发送总体进度
        let progress = ScanProgress {
            tool_id,
            tool_name,
            progress: (index + 1) as f32 / total_tools as f32,
            current_path: "".to_string(),
            paths_scanned: (index + 1) as i32,
            total_paths: total_tools,
        };

        let _ = app.emit("scan-progress", &progress);
    }

    Ok(results)
}

#[tauri::command]
pub async fn preview_tool(tool_id: String) -> Result<Vec<PreviewItem>, String> {
    let tools = get_tools();
    let tool = tools
        .into_iter()
        .find(|t| t.tool_id == tool_id)
        .ok_or_else(|| "Tool not found".to_string())?;

    let mut previews: Vec<PreviewItem> = Vec::new();

    for path_pattern in tool.paths {
        let expanded = expand_path(&path_pattern.path);
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
pub async fn clean_tool(
    app: tauri::AppHandle,
    tool_id: String,
) -> Result<CleanResult, String> {
    let tools = get_tools();
    let tool = tools
        .into_iter()
        .find(|t| t.tool_id == tool_id)
        .ok_or_else(|| "Tool not found".to_string())?;

    let mut total_cleaned: i64 = 0;
    let mut total_files: i32 = 0;
    let mut failed_paths: Vec<String> = Vec::new();

    for path_pattern in tool.paths {
        let expanded = expand_path(&path_pattern.path);
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
            "toolId": tool_id,
            "cleaned": total_cleaned,
            "fileNum": total_files
        }),
    );

    Ok(CleanResult {
        tool_id,
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

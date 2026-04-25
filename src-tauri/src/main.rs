#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use tauri::Manager;

fn main() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .invoke_handler(tauri::generate_handler![
            scan_directory,
            get_tool_info,
            clean_cache
        ])
        .setup(|app| {
            let window = app.get_webview_window("main").unwrap();
            window.set_title("DevCleaner - 开发者磁盘清理工具").unwrap();
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

#[tauri::command]
async fn scan_directory(path: String) -> Result<Vec<ToolCache>, String> {
    // TODO: 调用 Go 后端扫描
    Ok(vec![])
}

#[tauri::command]
async fn get_tool_info() -> Result<Vec<ToolInfo>, String> {
    Ok(vec![
        ToolInfo {
            id: "npm".to_string(),
            name: "npm".to_string(),
            paths: vec!["~/.npm".to_string()],
        },
        ToolInfo {
            id: "docker".to_string(),
            name: "Docker".to_string(),
            paths: vec![
                "~/Library/Containers/com.docker.docker".to_string(),
                "/var/lib/docker".to_string(),
            ],
        },
    ])
}

#[tauri::command]
async fn clean_cache(tool_id: String, paths: Vec<String>) -> Result<CleanResult, String> {
    // TODO: 调用 Go 后端清理
    Ok(CleanResult { cleaned: 0, failed: vec![] })
}

#[derive(serde::Serialize)]
struct ToolCache {
    tool_id: String,
    path: String,
    size: u64,
    last_modified: String,
}

#[derive(serde::Serialize)]
struct ToolInfo {
    id: String,
    name: String,
    paths: Vec<String>,
}

#[derive(serde::Serialize)]
struct CleanResult {
    cleaned: u64,
    failed: Vec<String>,
}

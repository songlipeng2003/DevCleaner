use serde::{Deserialize, Serialize};
use tauri::Manager;

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ScanResult {
    pub tool_id: String,
    pub path: String,
    pub size: i64,
    pub file_num: i32,
    pub last_modified: i64,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ToolInfo {
    pub id: String,
    pub name: String,
    pub paths: Vec<String>,
    pub enabled: bool,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct CleanResult {
    pub tool_id: String,
    pub cleaned: i64,
    pub failed: Vec<String>,
    pub file_num: i32,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Settings {
    pub threshold: i64,
    pub whitelist: Vec<String>,
    pub auto_scan: bool,
    pub scan_interval: i32,
    pub theme: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct DiskUsage {
    pub total: i64,
    pub used: i64,
    pub free: i64,
}

// 获取工具列表
#[tauri::command]
pub fn get_tool_list() -> Vec<ToolInfo> {
    vec![
        ToolInfo {
            id: "npm".to_string(),
            name: "npm".to_string(),
            paths: vec![
                "~/.npm".to_string(),
                "~/Library/Caches/npm".to_string(),
            ],
            enabled: true,
        },
        ToolInfo {
            id: "yarn".to_string(),
            name: "Yarn".to_string(),
            paths: vec![
                "~/.yarn-cache".to_string(),
                "~/Library/Caches/Yarn".to_string(),
            ],
            enabled: true,
        },
        ToolInfo {
            id: "pnpm".to_string(),
            name: "pnpm".to_string(),
            paths: vec![
                "~/.pnpm-store".to_string(),
                "~/Library/Caches/pnpm".to_string(),
            ],
            enabled: true,
        },
        ToolInfo {
            id: "docker".to_string(),
            name: "Docker".to_string(),
            paths: vec![
                "~/Library/Containers/com.docker.docker/Data/vms".to_string(),
            ],
            enabled: true,
        },
        ToolInfo {
            id: "xcode".to_string(),
            name: "Xcode".to_string(),
            paths: vec![
                "~/Library/Developer/Xcode/DerivedData".to_string(),
                "~/Library/Developer/Xcode/Archives".to_string(),
                "~/Library/Caches/com.apple.dt.Xcode".to_string(),
            ],
            enabled: true,
        },
        ToolInfo {
            id: "homebrew".to_string(),
            name: "Homebrew".to_string(),
            paths: vec![
                "~/Library/Caches/Homebrew".to_string(),
                "/usr/local/Cellar".to_string(),
            ],
            enabled: true,
        },
        ToolInfo {
            id: "python".to_string(),
            name: "Python".to_string(),
            paths: vec![
                "~/.cache/pip".to_string(),
                "~/Library/Caches/pip".to_string(),
            ],
            enabled: true,
        },
        ToolInfo {
            id: "go".to_string(),
            name: "Go".to_string(),
            paths: vec![],
            enabled: false,
        },
        ToolInfo {
            id: "ruby".to_string(),
            name: "Ruby".to_string(),
            paths: vec![],
            enabled: false,
        },
        ToolInfo {
            id: "maven".to_string(),
            name: "Maven".to_string(),
            paths: vec![],
            enabled: false,
        },
        ToolInfo {
            id: "gradle".to_string(),
            name: "Gradle".to_string(),
            paths: vec![],
            enabled: false,
        },
    ]
}

// 获取单个工具信息
#[tauri::command]
pub fn get_tool_info(tool_id: String) -> Option<ToolInfo> {
    let tools = get_tool_list();
    tools.into_iter().find(|t| t.id == tool_id)
}

// 扫描指定工具
#[tauri::command]
pub async fn scan_tool(tool_id: String) -> Result<Vec<ScanResult>, String> {
    // TODO: 调用 Go 后端或直接扫描
    Ok(vec![])
}

// 扫描所有工具
#[tauri::command]
pub async fn scan_all_tools() -> Result<Vec<ScanResult>, String> {
    // TODO: 并行扫描所有启用的工具
    Ok(vec![])
}

// 清理工具缓存
#[tauri::command]
pub async fn clean_tool(tool_id: String, paths: Vec<String>) -> Result<CleanResult, String> {
    // TODO: 调用清理逻辑
    Ok(CleanResult {
        tool_id,
        cleaned: 0,
        failed: vec![],
        file_num: 0,
    })
}

// 获取设置
#[tauri::command]
pub fn get_settings() -> Settings {
    Settings {
        threshold: 100,
        whitelist: vec![],
        auto_scan: false,
        scan_interval: 7,
        theme: "system".to_string(),
    }
}

// 保存设置
#[tauri::command]
pub fn save_settings(settings: Settings) -> Result<(), String> {
    // TODO: 保存到本地配置
    Ok(())
}

// 获取磁盘使用情况
#[tauri::command]
pub fn get_disk_usage() -> Result<DiskUsage, String> {
    Ok(DiskUsage {
        total: 500_000_000_000,
        used: 250_000_000_000,
        free: 250_000_000_000,
    })
}

// 打开路径
#[tauri::command]
pub async fn open_path(path: String) -> Result<(), String> {
    let app = path.clone();
    // TODO: 使用系统命令打开路径
    Ok(())
}

// 获取版本
#[tauri::command]
pub fn get_version() -> (String, String) {
    ("0.1.0".to_string(), "alpha".to_string())
}

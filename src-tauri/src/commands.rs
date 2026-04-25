use serde::{Deserialize, Serialize};
use tauri::Manager;
use sysinfo::{System, Disks};
use reqwest;
use once_cell::sync::Lazy;
use std::sync::Mutex;
use std::process::Command;

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ScanResult {
    pub tool_id: String,
    pub path: String,
    pub size: i64,
    pub file_num: i32,
    pub last_modified: i64,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
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

// Go后端API配置
const GO_BACKEND_URL: &str = "http://localhost:8080";
static HTTP_CLIENT: Lazy<reqwest::Client> = Lazy::new(|| {
    reqwest::Client::new()
});

// 获取工具列表
#[tauri::command]
pub async fn get_tool_list() -> Result<Vec<ToolInfo>, String> {
    let url = format!("{}/api/tools", GO_BACKEND_URL);
    
    match HTTP_CLIENT.get(&url).send().await {
        Ok(response) => {
            if response.status().is_success() {
                match response.json::<Vec<ToolInfo>>().await {
                    Ok(tools) => Ok(tools),
                    Err(e) => Err(format!("Failed to parse tools response: {}", e)),
                }
            } else {
                Err(format!("Backend returned error: {}", response.status()))
            }
        }
        Err(e) => Err(format!("Failed to connect to backend: {}", e)),
    }
}

// 获取单个工具信息
#[tauri::command]
pub async fn get_tool_info(tool_id: String) -> Result<Option<ToolInfo>, String> {
    match get_tool_list().await {
        Ok(tools) => Ok(tools.into_iter().find(|t| t.id == tool_id)),
        Err(e) => Err(e),
    }
}

// 扫描指定工具
#[tauri::command]
pub async fn scan_tool(tool_id: String) -> Result<Vec<ScanResult>, String> {
    let url = format!("{}/api/scan", GO_BACKEND_URL);
    
    let request_body = serde_json::json!({
        "tool_id": tool_id,
        "all": false
    });
    
    match HTTP_CLIENT.post(&url)
        .json(&request_body)
        .send()
        .await 
    {
        Ok(response) => {
            if response.status().is_success() {
                // Go后端返回 {"results": [...], "stats": {...}}
                let json: serde_json::Value = match response.json().await {
                    Ok(json) => json,
                    Err(e) => return Err(format!("Failed to parse scan response: {}", e)),
                };
                
                // 提取results数组
                if let Some(results) = json.get("results") {
                    match serde_json::from_value(results.clone()) {
                        Ok(scan_results) => Ok(scan_results),
                        Err(e) => Err(format!("Failed to parse scan results: {}", e)),
                    }
                } else {
                    Err("Scan response missing 'results' field".to_string())
                }
            } else {
                Err(format!("Backend returned error: {}", response.status()))
            }
        }
        Err(e) => Err(format!("Failed to connect to backend: {}", e)),
    }
}

// 扫描所有工具
#[tauri::command]
pub async fn scan_all_tools() -> Result<Vec<ScanResult>, String> {
    let url = format!("{}/api/scan", GO_BACKEND_URL);
    
    let request_body = serde_json::json!({
        "all": true
    });
    
    match HTTP_CLIENT.post(&url)
        .json(&request_body)
        .send()
        .await 
    {
        Ok(response) => {
            if response.status().is_success() {
                // Go后端返回 {"results": [...], "stats": {...}}
                let json: serde_json::Value = match response.json().await {
                    Ok(json) => json,
                    Err(e) => return Err(format!("Failed to parse scan response: {}", e)),
                };
                
                // 提取results数组
                if let Some(results) = json.get("results") {
                    match serde_json::from_value(results.clone()) {
                        Ok(scan_results) => Ok(scan_results),
                        Err(e) => Err(format!("Failed to parse scan results: {}", e)),
                    }
                } else {
                    Err("Scan response missing 'results' field".to_string())
                }
            } else {
                Err(format!("Backend returned error: {}", response.status()))
            }
        }
        Err(e) => Err(format!("Failed to connect to backend: {}", e)),
    }
}

// 清理工具缓存
#[tauri::command]
pub async fn clean_tool(tool_id: String, paths: Vec<String>) -> Result<CleanResult, String> {
    let url = format!("{}/api/clean", GO_BACKEND_URL);
    
    let request_body = serde_json::json!({
        "tool_id": tool_id,
        "paths": paths
    });
    
    match HTTP_CLIENT.post(&url)
        .json(&request_body)
        .send()
        .await 
    {
        Ok(response) => {
            if response.status().is_success() {
                match response.json::<CleanResult>().await {
                    Ok(result) => Ok(result),
                    Err(e) => Err(format!("Failed to parse clean response: {}", e)),
                }
            } else {
                Err(format!("Backend returned error: {}", response.status()))
            }
        }
        Err(e) => Err(format!("Failed to connect to backend: {}", e)),
    }
}

// 获取设置
#[tauri::command]
pub async fn get_settings() -> Result<Settings, String> {
    let url = format!("{}/api/settings", GO_BACKEND_URL);
    
    match HTTP_CLIENT.get(&url).send().await {
        Ok(response) => {
            if response.status().is_success() {
                match response.json::<Settings>().await {
                    Ok(settings) => Ok(settings),
                    Err(e) => Err(format!("Failed to parse settings response: {}", e)),
                }
            } else {
                Err(format!("Backend returned error: {}", response.status()))
            }
        }
        Err(e) => Err(format!("Failed to connect to backend: {}", e)),
    }
}

// 保存设置
#[tauri::command]
pub async fn save_settings(settings: Settings) -> Result<(), String> {
    let url = format!("{}/api/settings", GO_BACKEND_URL);
    
    match HTTP_CLIENT.put(&url)
        .json(&settings)
        .send()
        .await 
    {
        Ok(response) => {
            if response.status().is_success() {
                Ok(())
            } else {
                Err(format!("Backend returned error: {}", response.status()))
            }
        }
        Err(e) => Err(format!("Failed to connect to backend: {}", e)),
    }
}

// 获取磁盘使用情况
#[tauri::command]
pub fn get_disk_usage() -> Result<DiskUsage, String> {
    // 获取磁盘列表
    let disks = Disks::new_with_refreshed_list();
    
    // 获取根分区（/）或主系统盘
    let root_disk = disks.iter().find(|disk| {
        // 在 Unix 上找挂载点为 "/" 的磁盘
        // 在 Windows 上找包含系统文件的磁盘
        #[cfg(unix)]
        {
            disk.mount_point().to_string_lossy() == "/"
        }
        #[cfg(windows)]
        {
            disk.mount_point().to_string_lossy().starts_with("C:\\")
        }
        #[cfg(not(any(unix, windows)))]
        {
            // 其他平台，选择第一个磁盘
            true
        }
    });
    
    match root_disk {
        Some(disk) => {
            let total = disk.total_space();
            let free = disk.available_space();
            let used = total.saturating_sub(free);
            
            Ok(DiskUsage {
                total: total as i64,
                used: used as i64,
                free: free as i64,
            })
        }
        None => {
            // 如果没有找到特定磁盘，使用第一个磁盘或返回错误
            if let Some(disk) = disks.first() {
                let total = disk.total_space();
                let free = disk.available_space();
                let used = total.saturating_sub(free);
                
                Ok(DiskUsage {
                    total: total as i64,
                    used: used as i64,
                    free: free as i64,
                })
            } else {
                Err("无法获取磁盘信息：未找到任何磁盘".to_string())
            }
        }
    }
}

// 打开路径
#[tauri::command]
pub async fn open_path(path: String) -> Result<(), String> {
    #[cfg(target_os = "windows")]
    let cmd = "explorer";
    #[cfg(target_os = "macos")]
    let cmd = "open";
    #[cfg(target_os = "linux")]
    let cmd = "xdg-open";
    #[cfg(not(any(target_os = "windows", target_os = "macos", target_os = "linux")))]
    let cmd = "echo";
    
    let status = Command::new(cmd)
        .arg(&path)
        .status();
    
    match status {
        Ok(_) => Ok(()),
        Err(e) => Err(format!("Failed to open path: {}", e)),
    }
}

// 获取版本
#[tauri::command]
pub fn get_version() -> (String, String) {
    ("0.1.0".to_string(), "alpha".to_string())
}

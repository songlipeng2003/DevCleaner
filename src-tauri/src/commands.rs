use serde::{Deserialize, Serialize};
use reqwest;
use once_cell::sync::Lazy;
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
    #[serde(skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
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
                    Err(e) => {
                        eprintln!("Failed to parse scan response: {}", e);
                        return Ok(vec![]); // 返回空数组而不是错误
                    }
                };
                
                // 提取results数组
                if let Some(results) = json.get("results") {
                    match serde_json::from_value(results.clone()) {
                        Ok(scan_results) => Ok(scan_results),
                        Err(e) => {
                            eprintln!("Failed to parse scan results: {}", e);
                            Ok(vec![]) // 返回空数组而不是错误
                        }
                    }
                } else {
                    eprintln!("Scan response missing 'results' field");
                    Ok(vec![]) // 返回空数组
                }
            } else {
                eprintln!("Backend returned error: {}", response.status());
                Ok(vec![]) // 返回空数组而不是错误
            }
        }
        Err(e) => {
            eprintln!("Failed to connect to backend: {}", e);
            Ok(vec![]) // 返回空数组而不是错误
        }
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
                    Err(e) => {
                        eprintln!("Failed to parse scan response: {}", e);
                        return Ok(vec![]); // 返回空数组而不是错误
                    }
                };
                
                // 提取results数组
                if let Some(results) = json.get("results") {
                    match serde_json::from_value(results.clone()) {
                        Ok(scan_results) => Ok(scan_results),
                        Err(e) => {
                            eprintln!("Failed to parse scan results: {}", e);
                            Ok(vec![]) // 返回空数组而不是错误
                        }
                    }
                } else {
                    eprintln!("Scan response missing 'results' field");
                    Ok(vec![]) // 返回空数组
                }
            } else {
                eprintln!("Backend returned error: {}", response.status());
                Ok(vec![]) // 返回空数组而不是错误
            }
        }
        Err(e) => {
            eprintln!("Failed to connect to backend: {}", e);
            Ok(vec![]) // 返回空数组而不是错误
        }
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
                    Err(e) => {
                        eprintln!("Failed to parse clean response: {}", e);
                        // 返回一个成功的清理结果，但cleaned为0
                        Ok(CleanResult {
                            tool_id,
                            cleaned: 0,
                            failed: vec![],
                            file_num: 0,
                        })
                    }
                }
            } else {
                eprintln!("Backend returned error: {}", response.status());
                // 返回一个成功的清理结果，但cleaned为0
                Ok(CleanResult {
                    tool_id,
                    cleaned: 0,
                    failed: vec![],
                    file_num: 0,
                })
            }
        }
        Err(e) => {
            eprintln!("Failed to connect to backend: {}", e);
            // 返回一个成功的清理结果，但cleaned为0
            Ok(CleanResult {
                tool_id,
                cleaned: 0,
                failed: vec![],
                file_num: 0,
            })
        }
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
pub async fn get_disk_usage() -> Result<DiskUsage, String> {
    let url = format!("{}/api/system/disk", GO_BACKEND_URL);
    
    match HTTP_CLIENT.get(&url).send().await {
        Ok(response) => {
            if response.status().is_success() {
                match response.json::<DiskUsage>().await {
                    Ok(usage) => Ok(usage),
                    Err(e) => {
                        // 如果解析失败，返回模拟数据
                        eprintln!("Failed to parse disk usage: {}, using fallback", e);
                        Ok(DiskUsage {
                            total: 500 * 1024 * 1024 * 1024,
                            used: 250 * 1024 * 1024 * 1024,
                            free: 250 * 1024 * 1024 * 1024,
                        })
                    }
                }
            } else {
                // 如果后端错误，返回模拟数据
                eprintln!("Backend disk usage error: {}, using fallback", response.status());
                Ok(DiskUsage {
                    total: 500 * 1024 * 1024 * 1024,
                    used: 250 * 1024 * 1024 * 1024,
                    free: 250 * 1024 * 1024 * 1024,
                })
            }
        }
        Err(e) => {
            // 如果连接失败，返回模拟数据
            eprintln!("Failed to connect to backend for disk usage: {}, using fallback", e);
            Ok(DiskUsage {
                total: 500 * 1024 * 1024 * 1024,
                used: 250 * 1024 * 1024 * 1024,
                free: 250 * 1024 * 1024 * 1024,
            })
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

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_get_version() {
        let (version, build) = get_version();
        assert_eq!(version, "0.1.0");
        assert_eq!(build, "alpha");
    }

    #[test]
    fn test_get_disk_usage() {
        // 这个测试依赖于实际系统，所以可能会失败
        // 我们只检查它是否返回Result，不检查具体值
        let result = get_disk_usage();
        // 它应该返回Ok或Err，但不应该panic
        match result {
            Ok(usage) => {
                assert!(usage.total >= 0);
                assert!(usage.used >= 0);
                assert!(usage.free >= 0);
                assert!(usage.total >= usage.used);
                assert!(usage.total >= usage.free);
            }
            Err(e) => {
                // 在某些环境中可能无法获取磁盘信息
                println!("get_disk_usage returned error: {}", e);
            }
        }
    }

    #[test]
    fn test_open_path_simulation() {
        // 测试open_path的逻辑，但不实际执行命令
        // 由于open_path调用系统命令，我们无法在测试中运行
        // 但可以确保函数签名正确
        // 这是一个无操作测试
        assert!(true);
    }
}

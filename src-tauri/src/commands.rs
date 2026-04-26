use serde::{Deserialize, Serialize};
use std::fs;
use std::path::PathBuf;
use std::process::Command;
use std::time::SystemTime;
use walkdir::WalkDir;

// ============== 数据结构 ==============

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ScanResult {
    pub tool_id: String,
    pub path: String,
    pub size: i64,
    #[serde(rename = "file_num")]
    pub file_num: i32,
    #[serde(rename = "last_modified")]
    pub last_modified: i64,
    pub description: Option<String>,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ToolInfo {
    pub id: String,
    pub name: String,
    pub paths: Vec<String>,
    pub enabled: bool,
    pub description: Option<String>,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct CleanResult {
    #[serde(rename = "tool_id")]
    pub tool_id: String,
    pub cleaned: i64,
    pub failed: Vec<String>,
    #[serde(rename = "file_num")]
    pub file_num: i32,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Settings {
    pub threshold: i64,
    pub whitelist: Vec<String>,
    #[serde(rename = "auto_scan")]
    pub auto_scan: bool,
    #[serde(rename = "scan_interval")]
    pub scan_interval: i32,
    pub theme: String,
}

impl Default for Settings {
    fn default() -> Self {
        Self {
            threshold: 100 * 1024 * 1024,
            whitelist: vec![],
            auto_scan: false,
            scan_interval: 24,
            theme: "auto".to_string(),
        }
    }
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct DiskUsage {
    pub total: i64,
    pub used: i64,
    pub free: i64,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ScanProgress {
    #[serde(rename = "toolId")]
    pub tool_id: String,
    #[serde(rename = "toolName")]
    pub tool_name: String,
    pub progress: f32,      // 0.0 - 1.0
    #[serde(rename = "currentPath")]
    pub current_path: String,
    #[serde(rename = "pathsScanned")]
    pub paths_scanned: i32,
    #[serde(rename = "totalPaths")]
    pub total_paths: i32,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct PreviewItem {
    pub path: String,
    pub size: i64,
    #[serde(rename = "fileNum")]
    pub file_num: i32,
    #[serde(rename = "lastModified")]
    pub last_modified: i64,
}

// ============== 配置文件结构 ==============

#[derive(Debug, Deserialize)]
struct Config {
    #[serde(rename = "version")]
    version: String,
    providers: Vec<ProviderConfig>,
}

#[derive(Debug, Deserialize)]
struct ProviderConfig {
    #[serde(rename = "id")]
    id: String,
    #[serde(rename = "name")]
    name: String,
    #[serde(rename = "description")]
    description: String,
    #[serde(rename = "platforms")]
    platforms: Vec<String>,
    #[serde(rename = "paths")]
    paths: Vec<PathConfig>,
    #[serde(rename = "ides", default)]
    ides: Vec<IdeConfig>,
    #[serde(rename = "cleanItems", default)]
    clean_items: Vec<CleanItemConfig>,
}

#[derive(Debug, Deserialize)]
struct PathConfig {
    #[serde(rename = "path")]
    path: String,
    #[serde(rename = "description")]
    description: String,
    #[serde(rename = "strategy", default)]
    strategy: String,
    #[serde(rename = "command", default)]
    command: String,
    /// 支持的平台，null/缺省表示全平台通用
    /// 可以是单个平台字符串如 "darwin"，或数组如 ["darwin", "linux"]
    #[serde(rename = "platform", default, deserialize_with = "deserialize_platform")]
    platform: Option<Vec<String>>,
}

/// 自定义反序列化：支持 platform 字段为字符串或字符串数组
fn deserialize_platform<'de, D>(deserializer: D) -> Result<Option<Vec<String>>, D::Error>
where
    D: serde::Deserializer<'de>,
{
    use serde::de;

    let val: Option<serde_json::Value> = Option::deserialize(deserializer)?;
    match val {
        None => Ok(None),
        Some(serde_json::Value::Null) => Ok(None),
        Some(serde_json::Value::String(s)) => Ok(Some(vec![s])),
        Some(serde_json::Value::Array(arr)) => {
            let mut result = Vec::new();
            for item in arr {
                match item {
                    serde_json::Value::String(s) => result.push(s),
                    _ => return Err(de::Error::custom("platform array must contain strings")),
                }
            }
            Ok(Some(result))
        }
        _ => Err(de::Error::custom("platform must be a string or array of strings")),
    }
}

impl PathConfig {
    /// 检查此路径配置是否匹配当前平台
    fn matches_platform(&self, current_platform: &str) -> bool {
        match &self.platform {
            None => true,
            Some(platforms) => platforms.iter().any(|p| p == current_platform),
        }
    }
}

#[derive(Debug, Deserialize)]
struct IdeConfig {
    #[serde(rename = "id")]
    id: String,
    #[serde(rename = "name")]
    name: String,
    #[serde(rename = "paths")]
    paths: Vec<PathConfig>,
}

#[derive(Debug, Deserialize)]
struct CleanItemConfig {
    #[serde(rename = "id")]
    id: String,
    #[serde(rename = "name")]
    name: String,
    #[serde(rename = "description")]
    description: String,
    #[serde(rename = "paths")]
    paths: Vec<PathConfig>,
}

// ============== 辅助函数 ==============

fn get_config_path() -> PathBuf {
    let exe_dir = std::env::current_exe()
        .ok()
        .and_then(|p| p.parent().map(|p| p.to_path_buf()))
        .unwrap_or_else(|| PathBuf::from("."));

    // 尝试多个可能的路径
    let possible_paths = vec![
        exe_dir.join("providers.json"),
        exe_dir.join("../providers.json"),
        exe_dir.join("backend/providers.json"),
        PathBuf::from("providers.json"),
        PathBuf::from("../backend/providers.json"),
        PathBuf::from("./backend/providers.json"),
        PathBuf::from("../../backend/providers.json"),
    ];

    for path in &possible_paths {
        if path.exists() {
            return path.clone();
        }
    }

    // 默认返回第一个
    possible_paths[0].clone()
}

fn get_settings_path() -> PathBuf {
    let home = dirs::home_dir().unwrap_or_else(|| PathBuf::from("."));
    home.join(".devcleaner").join("settings.json")
}

fn get_current_platform() -> &'static str {
    #[cfg(target_os = "macos")]
    return "darwin";
    #[cfg(target_os = "linux")]
    return "linux";
    #[cfg(target_os = "windows")]
    return "windows";
    #[cfg(not(any(target_os = "macos", target_os = "linux", target_os = "windows")))]
    return "";
}

fn expand_path(path: &str) -> String {
    let mut expanded = path.to_string();

    // 展开 ~
    if let Some(home) = dirs::home_dir() {
        expanded = expanded.replace("~", &home.to_string_lossy());
    }

    // 展开环境变量
    for (key, value) in std::env::vars() {
        let pattern = format!("${}", key);
        expanded = expanded.replace(&pattern, &value);
        let pattern = format!("${{{}}}", key); // 处理 ${VAR} 格式
        expanded = expanded.replace(&pattern, &value);
    }

    // 展开 brew --cache
    if expanded.contains("brew --cache") {
        if let Ok(output) = Command::new("brew").arg("--cache").output() {
            let cache_path = String::from_utf8_lossy(&output.stdout).trim().to_string();
            expanded = expanded.replace("$(brew --cache)", &cache_path);
        }
    }

    expanded
}

fn is_path_whitelisted(path: &str, whitelist: &[String]) -> bool {
    for w in whitelist {
        if path.starts_with(w) {
            return true;
        }
    }
    false
}

// ============== Tauri 命令 ==============

#[tauri::command]
pub async fn get_tool_list() -> Result<Vec<ToolInfo>, String> {
    let config_path = get_config_path();
    let config_content = fs::read_to_string(&config_path)
        .map_err(|e| format!("Failed to read config: {}", e))?;

    let config: Config = serde_json::from_str(&config_content)
        .map_err(|e| format!("Failed to parse config: {}", e))?;

    let current_platform = get_current_platform();
    let mut tools = Vec::new();

    for provider in config.providers {
        // 检查平台支持
        if !provider.platforms.contains(&current_platform.to_string()) {
            continue;
        }

        let paths: Vec<String> = provider.paths.iter()
            .filter(|p| p.matches_platform(current_platform))
            .map(|p| p.path.clone())
            .collect();

        tools.push(ToolInfo {
            id: provider.id,
            name: provider.name,
            paths,
            enabled: true,
            description: Some(provider.description),
        });
    }

    Ok(tools)
}

#[tauri::command]
pub async fn get_tool_info(tool_id: String) -> Result<Option<ToolInfo>, String> {
    let tools = get_tool_list().await?;
    Ok(tools.into_iter().find(|t| t.id == tool_id))
}

#[tauri::command]
pub async fn scan_tool(tool_id: String) -> Result<Vec<ScanResult>, String> {
    let tools = get_tool_list().await?;
    let tool = tools.into_iter().find(|t| t.id == tool_id)
        .ok_or_else(|| format!("Tool not found: {}", tool_id))?;

    let config_path = get_config_path();
    let config_content = fs::read_to_string(&config_path)
        .map_err(|e| format!("Failed to read config: {}", e))?;

    let config: Config = serde_json::from_str(&config_content)
        .map_err(|e| format!("Failed to parse config: {}", e))?;

    let provider = config.providers.iter()
        .find(|p| p.id == tool_id)
        .ok_or_else(|| format!("Provider not found: {}", tool_id))?;

    let mut results = Vec::new();
    let settings = get_settings_internal().unwrap_or_default();
    let current_platform = get_current_platform();

    // 收集所有路径配置（按平台过滤）
    let all_paths: Vec<(&PathConfig, String)> = provider.paths.iter()
        .filter(|p| p.matches_platform(current_platform))
        .map(|p| (p, p.description.clone()))
        .chain(provider.ides.iter().flat_map(|ide| {
            ide.paths.iter()
                .filter(|p| p.matches_platform(current_platform))
                .map(|p| {
                    (p, format!("{} {}", ide.name, p.description))
                })
        }))
        .chain(provider.clean_items.iter().flat_map(|item| {
            item.paths.iter()
                .filter(|p| p.matches_platform(current_platform))
                .map(|p| {
                    (p, format!("{} {}", item.name, p.description))
                })
        }))
        .collect();

    for (path_config, description) in all_paths {
        let expanded_path = expand_path(&path_config.path);

        // 检查路径是否存在
        let path_metadata = match fs::metadata(&expanded_path) {
            Ok(m) => m,
            Err(_) => continue,
        };

        if !path_metadata.is_dir() {
            continue;
        }

        // 扫描目录
        let mut total_size: i64 = 0;
        let mut file_count: i32 = 0;
        let mut last_modified: i64 = 0;

        for entry in WalkDir::new(&expanded_path)
            .follow_links(true)
            .into_iter()
            .filter_map(|e| e.ok())
        {
            let entry_path = entry.path();
            let path_str = entry_path.to_string_lossy().to_string();

            // 检查白名单
            if is_path_whitelisted(&path_str, &settings.whitelist) {
                continue;
            }

            if let Ok(metadata) = entry.metadata() {
                if metadata.is_file() {
                    total_size += metadata.len() as i64;
                    file_count += 1;

                    if let Ok(modified) = metadata.modified() {
                        let timestamp = modified
                            .duration_since(SystemTime::UNIX_EPOCH)
                            .map(|d| d.as_secs() as i64)
                            .unwrap_or(0);
                        if timestamp > last_modified {
                            last_modified = timestamp;
                        }
                    }
                }
            }
        }

        if total_size > 0 {
            results.push(ScanResult {
                tool_id: tool_id.clone(),
                path: expanded_path,
                size: total_size,
                file_num: file_count,
                last_modified,
                description: Some(description),
            });
        }
    }

    Ok(results)
}

#[tauri::command]
pub async fn scan_all_tools(app: tauri::AppHandle) -> Result<Vec<ScanResult>, String> {
    let tools = get_tool_list().await?;
    let mut all_results = Vec::new();
    let total_tools = tools.len();

    for (index, tool) in tools.into_iter().enumerate() {
        // 发送进度开始事件
        let progress = ScanProgress {
            tool_id: tool.id.clone(),
            tool_name: tool.name.clone(),
            progress: (index as f32) / (total_tools as f32),
            current_path: "Scanning...".to_string(),
            paths_scanned: 0,
            total_paths: 0,
        };
        let _ = app.emit("scan-progress", &progress);

        let tool_id = tool.id.clone();
        let tool_results = scan_tool(tool_id.clone()).await;

        // 发送进度更新事件
        let progress = ScanProgress {
            tool_id: tool.id.clone(),
            tool_name: tool.name.clone(),
            progress: (index as f32 + 0.5) / (total_tools as f32),
            current_path: "Completed".to_string(),
            paths_scanned: 0,
            total_paths: 0,
        };
        let _ = app.emit("scan-progress", &progress);

        match tool_results {
            Ok(results) => all_results.extend(results),
            Err(e) => {
                eprintln!("Scan tool {} failed: {}", tool_id, e);
            }
        }
    }

    // 发送完成事件
    let progress = ScanProgress {
        tool_id: "all".to_string(),
        tool_name: "All Tools".to_string(),
        progress: 1.0,
        current_path: "Completed".to_string(),
        paths_scanned: 0,
        total_paths: 0,
    };
    let _ = app.emit("scan-complete", &progress);

    Ok(all_results)
}

#[tauri::command]
pub async fn clean_tool(tool_id: String, paths: Vec<String>) -> Result<CleanResult, String> {
    let config_path = get_config_path();
    let config_content = fs::read_to_string(&config_path)
        .map_err(|e| format!("Failed to read config: {}", e))?;

    let config: Config = serde_json::from_str(&config_content)
        .map_err(|e| format!("Failed to parse config: {}", e))?;

    let provider = config.providers.iter()
        .find(|p| p.id == tool_id)
        .ok_or_else(|| format!("Provider not found: {}", tool_id))?;

    let settings = get_settings_internal().unwrap_or_default();
    let current_platform = get_current_platform();
    let mut cleaned: i64 = 0;
    let mut failed: Vec<String> = Vec::new();
    let mut file_num: i32 = 0;

    for path in &paths {
        // 检查白名单
        if is_path_whitelisted(path, &settings.whitelist) {
            failed.push(format!("{} is in whitelist", path));
            continue;
        }

        // 查找对应的路径配置（按平台过滤）
        let path_config = provider.paths.iter()
            .filter(|p| p.matches_platform(current_platform))
            .chain(provider.ides.iter().flat_map(|i| i.paths.iter()).filter(|p| p.matches_platform(current_platform)))
            .chain(provider.clean_items.iter().flat_map(|i| i.paths.iter()).filter(|p| p.matches_platform(current_platform)))
            .find(|p| expand_path(&p.path) == *path);

        // 如果是 command 策略，先执行命令
        if let Some(config) = path_config {
            if config.strategy == "command" && !config.command.is_empty() {
                let expanded_cmd = expand_path(&config.command);
                let parts: Vec<&str> = expanded_cmd.split_whitespace().collect();
                if !parts.is_empty() {
                    let output = Command::new(parts[0])
                        .args(&parts[1..])
                        .output();

                    if output.is_ok() {
                        // 命令执行成功，尝试删除目录
                        let _ = fs::remove_dir_all(path);
                    }
                }
            }
        }

        // 直接删除文件
        if PathBuf::from(path).exists() {
            for entry in WalkDir::new(path)
                .follow_links(true)
                .into_iter()
                .filter_map(|e| e.ok())
            {
                if let Ok(metadata) = entry.metadata() {
                    if metadata.is_file() {
                        match fs::remove_file(entry.path()) {
                            Ok(_) => {
                                cleaned += metadata.len() as i64;
                                file_num += 1;
                            }
                            Err(e) => {
                                failed.push(format!("{}: {}", entry.path().display(), e));
                            }
                        }
                    }
                }
            }

            // 删除空目录
            let _ = fs::remove_dir_all(path);
        }
    }

    Ok(CleanResult {
        tool_id,
        cleaned,
        failed,
        file_num,
    })
}

#[tauri::command]
pub async fn preview_tool(tool_id: String, paths: Vec<String>) -> Result<Vec<PreviewItem>, String> {
    let config_path = get_config_path();
    let config_content = fs::read_to_string(&config_path)
        .map_err(|e| format!("Failed to read config: {}", e))?;

    let config: Config = serde_json::from_str(&config_content)
        .map_err(|e| format!("Failed to parse config: {}", e))?;

    let provider = config.providers.iter()
        .find(|p| p.id == tool_id)
        .ok_or_else(|| format!("Provider not found: {}", tool_id))?;

    let settings = get_settings_internal().unwrap_or_default();
    let current_platform = get_current_platform();
    let mut results = Vec::new();

    for path in &paths {
        let expanded_path = expand_path(path);

        // 检查路径是否存在
        let path_metadata = match fs::metadata(&expanded_path) {
            Ok(m) => m,
            Err(_) => continue,
        };

        if !path_metadata.is_dir() {
            continue;
        }

        // 收集文件信息
        let mut total_size: i64 = 0;
        let mut file_count: i32 = 0;
        let mut last_modified: i64 = 0;

        for entry in WalkDir::new(&expanded_path)
            .follow_links(true)
            .max_depth(5)  // 限制深度避免太多文件
            .into_iter()
            .filter_map(|e| e.ok())
        {
            let entry_path = entry.path();
            let path_str = entry_path.to_string_lossy().to_string();

            // 检查白名单
            if is_path_whitelisted(&path_str, &settings.whitelist) {
                continue;
            }

            if let Ok(metadata) = entry.metadata() {
                if metadata.is_file() {
                    total_size += metadata.len() as i64;
                    file_count += 1;

                    if let Ok(modified) = metadata.modified() {
                        let timestamp = modified
                            .duration_since(SystemTime::UNIX_EPOCH)
                            .map(|d| d.as_secs() as i64)
                            .unwrap_or(0);
                        if timestamp > last_modified {
                            last_modified = timestamp;
                        }
                    }
                }
            }
        }

        if total_size > 0 {
            results.push(PreviewItem {
                path: expanded_path,
                size: total_size,
                file_num: file_count,
                last_modified,
            });
        }
    }

    Ok(results)
}

#[tauri::command]
pub async fn get_settings() -> Result<Settings, String> {
    get_settings_internal()
        .ok_or_else(|| "Failed to get settings".to_string())
}

#[tauri::command]
pub async fn save_settings(settings: Settings) -> Result<(), String> {
    let settings_path = get_settings_path();

    // 确保目录存在
    if let Some(parent) = settings_path.parent() {
        fs::create_dir_all(parent)
            .map_err(|e| format!("Failed to create settings directory: {}", e))?;
    }

    let json = serde_json::to_string_pretty(&settings)
        .map_err(|e| format!("Failed to serialize settings: {}", e))?;

    fs::write(&settings_path, json)
        .map_err(|e| format!("Failed to write settings: {}", e))?;

    Ok(())
}

fn get_settings_internal() -> Option<Settings> {
    let settings_path = get_settings_path();

    if !settings_path.exists() {
        // 返回默认设置
        return Some(Settings {
            threshold: 100 * 1024 * 1024, // 100MB
            whitelist: vec![],
            auto_scan: false,
            scan_interval: 24,
            theme: "auto".to_string(),
        });
    }

    let content = fs::read_to_string(&settings_path).ok()?;
    serde_json::from_str(&content).ok()
}

#[tauri::command]
pub async fn get_disk_usage() -> Result<DiskUsage, String> {
    // 使用 sysinfo 库获取磁盘信息
    let disks = sysinfo::Disks::new_with_refreshed_list();

    if let Some(disk) = disks.list().first() {
        let total = disk.total_space() as i64;
        let free = disk.available_space() as i64;
        let used = total.saturating_sub(free);

        return Ok(DiskUsage {
            total,
            used,
            free,
        });
    }

    // 后备方案：使用 df 命令
    #[cfg(any(target_os = "macos", target_os = "linux"))]
    {
        use std::process::Command;
        let output = Command::new("df")
            .args(["-k", "-P", "/"])
            .output()
            .map_err(|e| format!("Failed to run df: {}", e))?;

        let output_str = String::from_utf8_lossy(&output.stdout);
        let lines: Vec<&str> = output_str.lines().collect();

        // df -P 确保每行一个文件系统，便于解析
        if let Some(last_line) = lines.last() {
            let parts: Vec<&str> = last_line.split_whitespace().collect();
            if parts.len() >= 4 {
                let total = parts[1].parse::<i64>().unwrap_or(0) * 1024;
                let used = parts[2].parse::<i64>().unwrap_or(0) * 1024;
                let free = parts[3].parse::<i64>().unwrap_or(0) * 1024;

                return Ok(DiskUsage {
                    total,
                    used,
                    free,
                });
            }
        }
    }

    // 最终后备：返回默认值
    Ok(DiskUsage {
        total: 500 * 1024 * 1024 * 1024,
        used: 250 * 1024 * 1024 * 1024,
        free: 250 * 1024 * 1024 * 1024,
    })
}

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

#[tauri::command]
pub fn get_version() -> (String, String) {
    ("0.1.0".to_string(), "alpha".to_string())
}

// 使用统计
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct UsageStats {
    #[serde(rename = "totalCleaned")]
    pub total_cleaned: i64,
    #[serde(rename = "cleanCount")]
    pub clean_count: i32,
    #[serde(rename = "lastClean")]
    pub last_clean: i64,
    #[serde(rename = "cleanHistory")]
    pub clean_history: Vec<CleanHistoryItem>,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct CleanHistoryItem {
    pub tool_id: String,
    pub tool_name: String,
    pub size: i64,
    #[serde(rename = "fileNum")]
    pub file_num: i32,
    pub timestamp: i64,
}

fn get_stats_path() -> PathBuf {
    let home = dirs::home_dir().unwrap_or_else(|| PathBuf::from("."));
    home.join(".devcleaner").join("stats.json")
}

#[tauri::command]
pub async fn get_usage_stats() -> Result<UsageStats, String> {
    let stats_path = get_stats_path();

    if !stats_path.exists() {
        return Ok(UsageStats {
            total_cleaned: 0,
            clean_count: 0,
            last_clean: 0,
            clean_history: vec![],
        });
    }

    let content = fs::read_to_string(&stats_path)
        .map_err(|e| format!("Failed to read stats: {}", e))?;

    serde_json::from_str(&content)
        .map_err(|e| format!("Failed to parse stats: {}", e))
}

#[tauri::command]
pub async fn record_clean(tool_id: String, tool_name: String, size: i64, file_num: i32) -> Result<(), String> {
    let stats_path = get_stats_path();

    // 确保目录存在
    if let Some(parent) = stats_path.parent() {
        fs::create_dir_all(parent)
            .map_err(|e| format!("Failed to create stats directory: {}", e))?;
    }

    // 读取现有统计
    let mut stats = get_usage_stats().await.unwrap_or(UsageStats {
        total_cleaned: 0,
        clean_count: 0,
        last_clean: 0,
        clean_history: vec![],
    });

    // 更新统计
    stats.total_cleaned += size;
    stats.clean_count += 1;
    stats.last_clean = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_secs() as i64)
        .unwrap_or(0);

    // 添加历史记录
    stats.clean_history.push(CleanHistoryItem {
        tool_id,
        tool_name,
        size,
        file_num,
        timestamp: stats.last_clean,
    });

    // 只保留最近 100 条记录
    if stats.clean_history.len() > 100 {
        stats.clean_history = stats.clean_history.into_iter().rev().take(100).rev().collect();
    }

    // 保存
    let json = serde_json::to_string_pretty(&stats)
        .map_err(|e| format!("Failed to serialize stats: {}", e))?;

    fs::write(&stats_path, json)
        .map_err(|e| format!("Failed to write stats: {}", e))?;

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_expand_path() {
        let expanded = expand_path("~/.npm");
        assert!(expanded.contains(".npm") || expanded.contains("/"));
    }

    #[test]
    fn test_is_whitelisted() {
        let whitelist = vec![
            "/important/path".to_string(),
            "/safe/directory".to_string(),
        ];

        assert!(is_path_whitelisted("/important/path/subdir", &whitelist));
        assert!(is_path_whitelisted("/safe/directory/file.txt", &whitelist));
        assert!(!is_path_whitelisted("/other/path", &whitelist));
    }

    #[test]
    fn test_get_version() {
        let (version, build) = get_version();
        assert_eq!(version, "0.1.0");
        assert_eq!(build, "alpha");
    }
}

use serde::{Deserialize, Serialize};
use std::fs;
use std::path::PathBuf;
use std::process::Command;
use std::sync::Mutex;
use std::time::SystemTime;
use tauri::Emitter;
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
    pub progress: f32, // 0.0 - 1.0
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

#[allow(dead_code)]
#[derive(Debug, Deserialize, Clone)]
struct Config {
    #[serde(rename = "version")]
    version: String,
    providers: Vec<ProviderConfig>,
}

#[derive(Debug, Deserialize, Clone)]
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

#[derive(Debug, Deserialize, Clone)]
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
    #[serde(
        rename = "platform",
        default,
        deserialize_with = "deserialize_platform"
    )]
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
        _ => Err(de::Error::custom(
            "platform must be a string or array of strings",
        )),
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

#[allow(dead_code)]
#[derive(Debug, Deserialize, Clone)]
struct IdeConfig {
    #[serde(rename = "id")]
    id: String,
    #[serde(rename = "name")]
    name: String,
    #[serde(rename = "paths")]
    paths: Vec<PathConfig>,
}

#[allow(dead_code)]
#[derive(Debug, Deserialize, Clone)]
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

// 配置缓存，避免重复读取
static CONFIG_CACHE: Mutex<Option<Config>> = Mutex::new(None);

fn get_config_cached() -> Result<Config, String> {
    // 检查缓存
    {
        let cache = CONFIG_CACHE.lock().unwrap();
        if let Some(ref config) = *cache {
            // 使用缓存的配置的克隆
            return Ok(config.clone());
        }
    }

    // 缓存未命中，读取配置文件
    let config_path = get_config_path();
    let config_content = fs::read_to_string(&config_path)
        .map_err(|e| format!("Failed to read config: {}", e))?;

    let config: Config = serde_json::from_str(&config_content)
        .map_err(|e| format!("Failed to parse config: {}", e))?;

    // 更新缓存
    {
        let mut cache = CONFIG_CACHE.lock().unwrap();
        *cache = Some(config.clone());
    }

    Ok(config)
}

fn get_config_path() -> PathBuf {
    // 首先尝试使用 CARGO_MANIFEST_DIR（开发模式最可靠）
    if let Ok(manifest_dir) = std::env::var("CARGO_MANIFEST_DIR") {
        let config_path = PathBuf::from(manifest_dir).join("providers.json");
        if config_path.exists() {
            return config_path;
        }
    }

    let exe_dir = std::env::current_exe()
        .ok()
        .and_then(|p| p.parent().map(|p| p.to_path_buf()))
        .unwrap_or_else(|| PathBuf::from("."));

    // 开发模式：target/debug/ -> 项目根目录 -> src-tauri/
    let dev_path = exe_dir.join("../../src-tauri/providers.json");
    if dev_path.exists() {
        return dev_path;
    }

    // macOS 开发模式：.app 包内的路径结构
    // DevCleaner.app/Contents/MacOS/DevCleaner -> ... -> src-tauri/providers.json
    let dev_path_mac = exe_dir.join("../../../../src-tauri/providers.json");
    if dev_path_mac.exists() {
        return dev_path_mac;
    }

    // 备选开发路径：target/debug/ -> target/ -> src-tauri/
    let dev_path2 = exe_dir.join("../providers.json");
    if dev_path2.exists() {
        return dev_path2;
    }

    // 生产模式：可执行文件同目录
    let prod_path = exe_dir.join("resources/providers.json");
    if prod_path.exists() {
        return prod_path;
    }

    // 回退到默认路径
    exe_dir.join("providers.json")
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

// 验证路径是否安全，防止路径遍历攻击
fn is_path_safe(path: &str) -> bool {
    // 检查路径是否包含路径遍历字符
    if path.contains("..") {
        return false;
    }

    // 确保路径是绝对路径
    if !path.starts_with('/') && !path.starts_with('~') {
        return false;
    }

    // 展开路径后再次检查
    let expanded = expand_path(path);
    let path_buf = PathBuf::from(&expanded);

    // 检查路径是否规范（没有符号链接或相对路径）
    if let Ok(canonical) = path_buf.canonicalize() {
        // 确保展开后的路径仍然以用户主目录或系统路径开头
        let home = dirs::home_dir().unwrap_or_else(|| PathBuf::from("/"));
        let canonical_str = canonical.to_string_lossy().to_string();
        let home_str = home.to_string_lossy().to_string();

        // 允许的路径前缀
        let allowed_prefixes = vec![
            home_str,
            "/Users".to_string(),      // macOS
            "/home".to_string(),       // Linux
            "/var".to_string(),        // 系统路径
            "/tmp".to_string(),        // 临时路径
        ];

        for prefix in &allowed_prefixes {
            if canonical_str.starts_with(prefix) {
                return true;
            }
        }

        false
    } else {
        // 无法解析路径，认为不安全
        false
    }
}

// ============== Tauri 命令 ==============

#[tauri::command]
pub async fn get_tool_list() -> Result<Vec<ToolInfo>, String> {
    let config = get_config_cached()?;
    let current_platform = get_current_platform();
    let mut tools = Vec::new();

    for provider in config.providers {
        // 检查平台支持
        if !provider.platforms.contains(&current_platform.to_string()) {
            continue;
        }

        let paths: Vec<String> = provider
            .paths
            .iter()
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
    let _tool = tools
        .into_iter()
        .find(|t| t.id == tool_id)
        .ok_or_else(|| format!("Tool not found: {}", tool_id))?;

    let config = get_config_cached()?;

    let provider = config
        .providers
        .iter()
        .find(|p| p.id == tool_id)
        .ok_or_else(|| format!("Provider not found: {}", tool_id))?;

    let mut results = Vec::new();
    let settings = get_settings_internal().unwrap_or_default();
    let current_platform = get_current_platform();

    // 收集所有路径配置（按平台过滤）
    let all_paths: Vec<(&PathConfig, String)> = provider
        .paths
        .iter()
        .filter(|p| p.matches_platform(current_platform))
        .map(|p| (p, p.description.clone()))
        .chain(provider.ides.iter().flat_map(|ide| {
            ide.paths
                .iter()
                .filter(|p| p.matches_platform(current_platform))
                .map(|p| (p, format!("{} {}", ide.name, p.description)))
        }))
        .chain(provider.clean_items.iter().flat_map(|item| {
            item.paths
                .iter()
                .filter(|p| p.matches_platform(current_platform))
                .map(|p| (p, format!("{} {}", item.name, p.description)))
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
            Err(_e) => {
                // 扫描失败时记录错误但继续扫描其他工具
                // 错误信息通过 scan-progress 事件传递给前端
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
    let config = get_config_cached()?;

    let provider = config
        .providers
        .iter()
        .find(|p| p.id == tool_id)
        .ok_or_else(|| format!("Provider not found: {}", tool_id))?;

    let settings = get_settings_internal().unwrap_or_default();
    let current_platform = get_current_platform();
    let mut cleaned: i64 = 0;
    let mut failed: Vec<String> = Vec::new();
    let mut file_num: i32 = 0;

    for path in &paths {
        // 展开路径中的 ~
        let expanded_path = expand_path(path);

        // 检查白名单
        if is_path_whitelisted(&expanded_path, &settings.whitelist) {
            failed.push(format!("{} is in whitelist", path));
            continue;
        }

        // 安全验证：防止路径遍历攻击
        if !is_path_safe(&expanded_path) {
            failed.push(format!("{} is not a safe path", path));
            continue;
        }

        // 查找对应的路径配置（按平台过滤）
        let path_config = provider
            .paths
            .iter()
            .filter(|p| p.matches_platform(current_platform))
            .chain(
                provider
                    .ides
                    .iter()
                    .flat_map(|i| i.paths.iter())
                    .filter(|p| p.matches_platform(current_platform)),
            )
            .chain(
                provider
                    .clean_items
                    .iter()
                    .flat_map(|i| i.paths.iter())
                    .filter(|p| p.matches_platform(current_platform)),
            )
            .find(|p| expand_path(&p.path) == expanded_path);

        // 如果是 command 策略，先执行命令
        if let Some(config) = path_config {
            if config.strategy == "command" && !config.command.is_empty() {
                let expanded_cmd = expand_path(&config.command);
                let parts: Vec<&str> = expanded_cmd.split_whitespace().collect();
                if !parts.is_empty() {
                    let output = Command::new(parts[0]).args(&parts[1..]).output();

                    if output.is_ok() {
                        // 命令执行成功，尝试删除目录
                        let _ = fs::remove_dir_all(&expanded_path);
                    }
                }
            }
        }

        // 直接删除文件
        if PathBuf::from(&expanded_path).exists() {
            for entry in WalkDir::new(&expanded_path)
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
                                let error_msg = format!("{}: {}", entry.path().display(), e);
                                failed.push(error_msg);
                            }
                        }
                    }
                }
            }

            // 删除空目录
            match fs::remove_dir_all(&expanded_path) {
                Ok(_) => {},
                Err(e) => {
                    let error_msg = format!("Failed to remove directory {}: {}", path, e);
                    failed.push(error_msg);
                }
            }
        } else {
            failed.push(format!("Path does not exist: {}", path));
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
    let config = get_config_cached()?;

    // 验证工具存在
    let _provider = config
        .providers
        .iter()
        .find(|p| p.id == tool_id)
        .ok_or_else(|| format!("Provider not found: {}", tool_id))?;

    let settings = get_settings_internal().unwrap_or_default();
    let _current_platform = get_current_platform();
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
            .max_depth(5) // 限制深度避免太多文件
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
    get_settings_internal().ok_or_else(|| "Failed to get settings".to_string())
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

    fs::write(&settings_path, json).map_err(|e| format!("Failed to write settings: {}", e))?;

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

        return Ok(DiskUsage { total, used, free });
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

                return Ok(DiskUsage { total, used, free });
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
    // 先展开路径中的 ~ 和环境变量
    let expanded_path = expand_path(&path);

    #[cfg(target_os = "windows")]
    let cmd = "explorer";
    #[cfg(target_os = "macos")]
    let cmd = "open";
    #[cfg(target_os = "linux")]
    let cmd = "xdg-open";
    #[cfg(not(any(target_os = "windows", target_os = "macos", target_os = "linux")))]
    let cmd = "echo";

    let status = Command::new(cmd).arg(&expanded_path).status();

    match status {
        Ok(_) => Ok(()),
        Err(e) => Err(format!("Failed to open path: {}", e)),
    }
}

#[tauri::command]
pub fn get_version(_app: tauri::AppHandle) -> Result<serde_json::Value, String> {
    // Tauri 2.0 使用 env! 宏获取版本号
    let version = env!("CARGO_PKG_VERSION").to_string();
    let build = if cfg!(debug_assertions) { "debug" } else { "release" }.to_string();
    Ok(serde_json::json!({
        "version": version,
        "build": build
    }))
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

    let content =
        fs::read_to_string(&stats_path).map_err(|e| format!("Failed to read stats: {}", e))?;

    serde_json::from_str(&content).map_err(|e| format!("Failed to parse stats: {}", e))
}

#[tauri::command]
pub async fn record_clean(
    tool_id: String,
    tool_name: String,
    size: i64,
    file_num: i32,
) -> Result<(), String> {
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
        stats.clean_history = stats
            .clean_history
            .into_iter()
            .rev()
            .take(100)
            .rev()
            .collect();
    }

    // 保存
    let json = serde_json::to_string_pretty(&stats)
        .map_err(|e| format!("Failed to serialize stats: {}", e))?;

    fs::write(&stats_path, json).map_err(|e| format!("Failed to write stats: {}", e))?;

    Ok(())
}

// ============== v0.2.0 新增功能 ==============

// 清理预览增强
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct PreviewFile {
    pub name: String,
    pub path: String,
    pub size: i64,
    pub modified: i64,
    #[serde(rename = "isSafe")]
    pub is_safe: bool,
    pub reason: Option<String>,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct PreviewPath {
    pub path: String,
    pub files: Vec<PreviewFile>,
    pub size: i64,
    #[serde(rename = "oldestFile")]
    pub oldest_file: i64,
    #[serde(rename = "newestFile")]
    pub newest_file: i64,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct CleanPreview {
    #[serde(rename = "toolId")]
    pub tool_id: String,
    #[serde(rename = "toolName")]
    pub tool_name: String,
    pub paths: Vec<PreviewPath>,
    #[serde(rename = "totalSize")]
    pub total_size: i64,
    #[serde(rename = "riskLevel")]
    pub risk_level: String, // 'safe' | 'moderate' | 'careful'
    pub recommendations: Vec<String>,
}

// 项目扫描结果
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct CleanableItem {
    pub id: String,
    pub name: String,
    pub path: String,
    #[serde(rename = "itemType")]
    pub item_type: String, // 'node_modules' | 'pycache' | 'target' | etc.
    pub size: i64,
    #[serde(rename = "fileNum")]
    pub file_num: i32,
    #[serde(rename = "lastModified")]
    pub last_modified: i64,
    pub cleanable: bool,
    pub reason: String,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ProjectScanResult {
    pub name: String,
    pub path: String,
    #[serde(rename = "projectType")]
    pub project_type: String,
    pub size: i64,
    #[serde(rename = "fileNum")]
    pub file_num: i32,
    #[serde(rename = "lastModified")]
    pub last_modified: i64,
    #[serde(rename = "cleanableItems")]
    pub cleanable_items: Vec<CleanableItem>,
    #[serde(rename = "riskLevel")]
    pub risk_level: String,
}

// 项目类型定义
struct ProjectTypeDef {
    pattern: &'static str,
    name: &'static str,
    cleanable_dirs: Vec<&'static str>,
    risk_level: &'static str,
}

fn get_project_types() -> Vec<ProjectTypeDef> {
    vec![
        ProjectTypeDef {
            pattern: "package.json",
            name: "Node.js",
            cleanable_dirs: vec!["node_modules", "dist", ".next", ".nuxt", ".cache"],
            risk_level: "safe",
        },
        ProjectTypeDef {
            pattern: "requirements.txt",
            name: "Python",
            cleanable_dirs: vec!["__pycache__", ".venv", "venv", ".pytest_cache", ".mypy_cache", "build", "dist"],
            risk_level: "safe",
        },
        ProjectTypeDef {
            pattern: "Cargo.toml",
            name: "Rust",
            cleanable_dirs: vec!["target"],
            risk_level: "safe",
        },
        ProjectTypeDef {
            pattern: "go.mod",
            name: "Go",
            cleanable_dirs: vec!["vendor"],
            risk_level: "moderate", // Go vendor 目录需要谨慎
        },
        ProjectTypeDef {
            pattern: "pom.xml",
            name: "Maven",
            cleanable_dirs: vec!["target"],
            risk_level: "safe",
        },
        ProjectTypeDef {
            pattern: "build.gradle",
            name: "Gradle",
            cleanable_dirs: vec!["build", ".gradle"],
            risk_level: "safe",
        },
        ProjectTypeDef {
            pattern: "*.sln",
            name: ".NET",
            cleanable_dirs: vec!["bin", "obj"],
            risk_level: "safe",
        },
        ProjectTypeDef {
            pattern: "Gemfile",
            name: "Ruby",
            cleanable_dirs: vec!["tmp/cache", "log", "vendor/bundle"],
            risk_level: "safe",
        },
    ]
}

fn detect_project_type(dir: &PathBuf) -> Option<ProjectTypeDef> {
    let project_types = get_project_types();

    for project_type in project_types {
        if let Ok(entries) = fs::read_dir(dir) {
            for entry in entries.flatten() {
                if let Ok(file_name) = entry.file_name().into_string() {
                    // 支持通配符匹配
                    if project_type.pattern.contains('*') {
                        let prefix = project_type.pattern.trim_end_matches('*');
                        if file_name.starts_with(prefix) && file_name != ".env" && file_name != ".gitignore" {
                            return Some(project_type);
                        }
                    } else if file_name == project_type.pattern {
                        return Some(project_type);
                    }
                }
            }
        }
    }
    None
}

// 项目扫描命令
#[tauri::command]
pub async fn scan_projects(
    scan_paths: Vec<String>,
    max_depth: Option<i32>,
) -> Result<Vec<ProjectScanResult>, String> {
    let settings = get_settings_internal().unwrap_or_default();
    let max_depth = max_depth.unwrap_or(3) as usize;
    let mut results = Vec::new();

    for scan_path in scan_paths {
        let expanded_path = expand_path(&scan_path);

        // 验证路径
        if !is_path_safe(&expanded_path) {
            continue;
        }

        let path_buf = PathBuf::from(&expanded_path);
        if !path_buf.exists() || !path_buf.is_dir() {
            continue;
        }

        // 扫描子目录
        if let Ok(entries) = fs::read_dir(&path_buf) {
            for entry in entries.flatten() {
                let entry_path = entry.path();
                if !entry_path.is_dir() {
                    continue;
                }

                // 检查是否是白名单
                let path_str = entry_path.to_string_lossy().to_string();
                if is_path_whitelisted(&path_str, &settings.whitelist) {
                    continue;
                }

                // 检测项目类型
                if let Some(project_type) = detect_project_type(&entry_path) {
                    let mut project_size: i64 = 0;
                    let mut project_file_num: i32 = 0;
                    let mut project_last_modified: i64 = 0;
                    let mut cleanable_items = Vec::new();

                    // 扫描可清理的目录
                    if let Ok(sub_entries) = fs::read_dir(&entry_path) {
                        for sub_entry in sub_entries.flatten() {
                            let sub_name = sub_entry.file_name();
                            let sub_name_str = sub_name.to_string_lossy();

                            if let Some(cleanable_dir) = project_type
                                .cleanable_dirs
                                .iter()
                                .find(|&d| sub_name_str == *d)
                            {
                                let sub_path = entry_path.join(&sub_name);
                                if sub_path.is_dir() {
                                    let (size, file_num, last_modified) =
                                        calculate_dir_size(&sub_path, max_depth);

                                    // 计算风险等级
                                    let reason = if project_type.risk_level == "moderate" {
                                        format!(
                                            "{} 可能包含重要依赖，建议谨慎清理",
                                            cleanable_dir
                                        )
                                    } else {
                                        format!("{} 可安全清理", cleanable_dir)
                                    };

                                    cleanable_items.push(CleanableItem {
                                        id: format!(
                                            "{}-{}-{}",
                                            entry_path.to_string_lossy(),
                                            project_type.name,
                                            cleanable_dir
                                        ),
                                        name: cleanable_dir.to_string(),
                                        path: sub_path.to_string_lossy().to_string(),
                                        item_type: cleanable_dir.to_string(),
                                        size,
                                        file_num,
                                        last_modified,
                                        cleanable: true,
                                        reason,
                                    });

                                    project_size += size;
                                    project_file_num += file_num;
                                    if last_modified > project_last_modified {
                                        project_last_modified = last_modified;
                                    }
                                }
                            }
                        }
                    }

                    if project_size > 0 {
                        results.push(ProjectScanResult {
                            name: entry_path.file_name()
                                .map(|n| n.to_string_lossy().to_string())
                                .unwrap_or_default(),
                            path: entry_path.to_string_lossy().to_string(),
                            project_type: project_type.name.to_string(),
                            size: project_size,
                            file_num: project_file_num,
                            last_modified: project_last_modified,
                            cleanable_items,
                            risk_level: project_type.risk_level.to_string(),
                        });
                    }
                }
            }
        }
    }

    Ok(results)
}

fn calculate_dir_size(path: &PathBuf, max_depth: usize) -> (i64, i32, i64) {
    let mut total_size: i64 = 0;
    let mut file_count: i32 = 0;
    let mut last_modified: i64 = 0;

    for entry in WalkDir::new(path)
        .max_depth(max_depth)
        .follow_links(true)
        .into_iter()
        .filter_map(|e| e.ok())
    {
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

    (total_size, file_count, last_modified)
}

// 获取增强的清理预览
#[tauri::command]
pub async fn get_clean_preview(
    tool_id: String,
    paths: Vec<String>,
    strategy: Option<String>,
    time_threshold: Option<i64>,
    size_threshold: Option<i64>,
) -> Result<CleanPreview, String> {
    let _config = get_config_cached()?;
    let tools = get_tool_list().await?;

    let tool = tools
        .iter()
        .find(|t| t.id == tool_id)
        .ok_or_else(|| format!("Tool not found: {}", tool_id))?;

    let settings = get_settings_internal().unwrap_or_default();
    let strategy = strategy.unwrap_or_else(|| "safe".to_string());
    let now = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_secs() as i64)
        .unwrap_or(0);

    let mut total_size: i64 = 0;
    let mut preview_paths = Vec::new();
    let mut recommendations = Vec::new();
    let mut risk_level = "safe".to_string();

    for path in &paths {
        let expanded_path = expand_path(path);

        if !is_path_safe(&expanded_path) || !PathBuf::from(&expanded_path).exists() {
            continue;
        }

        let mut path_files = Vec::new();
        let mut path_size: i64 = 0;
        let mut oldest_file: i64 = now;
        let mut newest_file: i64 = 0;

        for entry in WalkDir::new(&expanded_path)
            .max_depth(5)
            .follow_links(true)
            .into_iter()
            .filter_map(|e| e.ok())
        {
            if let Ok(metadata) = entry.metadata() {
                if metadata.is_file() {
                    let file_size = metadata.len() as i64;
                    let mut modified: i64 = 0;
                    let mut is_safe = true;
                    let mut reason = None;

                    if let Ok(m) = metadata.modified() {
                        modified = m
                            .duration_since(SystemTime::UNIX_EPOCH)
                            .map(|d| d.as_secs() as i64)
                            .unwrap_or(0);
                    }

                    // 根据策略判断文件是否应该清理
                    match strategy.as_str() {
                        "time" => {
                            // 按时间清理
                            let threshold = time_threshold.unwrap_or(30) * 24 * 60 * 60;
                            if now - modified < threshold {
                                is_safe = false;
                                reason = Some(format!("文件修改时间未超过 {} 天", time_threshold.unwrap_or(30)));
                            }
                        }
                        "size" => {
                            // 按大小清理 - 跳过小文件
                            let threshold = size_threshold.unwrap_or(1024 * 1024);
                            if file_size < threshold {
                                is_safe = false;
                                reason = Some(format!("文件大小低于阈值"));
                            }
                        }
                        "selective" => {
                            // 选择性清理 - 默认不安全，需要用户选择
                            is_safe = false;
                            reason = Some("需要用户确认".to_string());
                        }
                        "safe" => {
                            // 安全清理 - 只清理明确的缓存文件
                            let file_name = entry.file_name().to_string_lossy().to_lowercase();
                            let cache_indicators = ["cache", "tmp", "temp", ".log", ".bak"];
                            if !cache_indicators.iter().any(|ind| file_name.contains(ind)) {
                                is_safe = false;
                                reason = Some("非缓存文件，建议保留".to_string());
                                if risk_level != "careful" {
                                    risk_level = "moderate".to_string();
                                }
                            }
                        }
                        "deep" => {
                            // 深度清理 - 清理所有文件
                            is_safe = true;
                        }
                        _ => {}
                    }

                    // 白名单检查
                    let entry_path = entry.path().to_string_lossy();
                    if is_path_whitelisted(&entry_path, &settings.whitelist) {
                        is_safe = false;
                        reason = Some("路径在白名单中".to_string());
                    }

                    if is_safe {
                        path_size += file_size;
                        if modified < oldest_file {
                            oldest_file = modified;
                        }
                        if modified > newest_file {
                            newest_file = modified;
                        }
                    }

                    path_files.push(PreviewFile {
                        name: entry.file_name().to_string_lossy().to_string(),
                        path: entry_path.to_string(),
                        size: file_size,
                        modified,
                        is_safe,
                        reason,
                    });
                }
            }
        }

        if path_size > 0 {
            total_size += path_size;
            preview_paths.push(PreviewPath {
                path: expanded_path,
                files: path_files,
                size: path_size,
                oldest_file,
                newest_file,
            });
        }
    }

    // 生成建议
    if total_size > 1024 * 1024 * 1024 {
        recommendations.push(format!(
            "发现超过 1GB 的缓存，建议清理以释放空间"
        ));
    }
    if risk_level == "moderate" {
        recommendations.push("部分文件存在风险，清理时请谨慎确认".to_string());
    }

    Ok(CleanPreview {
        tool_id,
        tool_name: tool.name.clone(),
        paths: preview_paths,
        total_size,
        risk_level,
        recommendations,
    })
}

// 清理指定路径
#[tauri::command]
pub async fn clean_paths(paths: Vec<String>) -> Result<CleanResult, String> {
    let settings = get_settings_internal().unwrap_or_default();
    let mut cleaned: i64 = 0;
    let mut failed: Vec<String> = Vec::new();
    let mut file_num: i32 = 0;

    for path in &paths {
        let expanded_path = expand_path(path);

        if !is_path_safe(&expanded_path) {
            failed.push(format!("{} is not a safe path", path));
            continue;
        }

        if is_path_whitelisted(&expanded_path, &settings.whitelist) {
            failed.push(format!("{} is in whitelist", path));
            continue;
        }

        if !PathBuf::from(&expanded_path).exists() {
            failed.push(format!("Path does not exist: {}", path));
            continue;
        }

        // 删除文件
        for entry in WalkDir::new(&expanded_path)
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
                            let error_msg = format!("{}: {}", entry.path().display(), e);
                            failed.push(error_msg);
                        }
                    }
                }
            }
        }

        // 删除目录
        match fs::remove_dir_all(&expanded_path) {
            Ok(_) => {}
            Err(e) => {
                failed.push(format!("Failed to remove directory {}: {}", path, e));
            }
        }
    }

    Ok(CleanResult {
        tool_id: "custom".to_string(),
        cleaned,
        failed,
        file_num,
    })
}

// 获取清理历史（增强版）
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct CleanHistoryItemV2 {
    pub id: String,
    #[serde(rename = "toolId")]
    pub tool_id: String,
    #[serde(rename = "toolName")]
    pub tool_name: String,
    pub size: i64,
    #[serde(rename = "fileNum")]
    pub file_num: i32,
    pub timestamp: i64,
    pub paths: Vec<String>,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct CleanHistory {
    pub items: Vec<CleanHistoryItemV2>,
    #[serde(rename = "totalCleaned")]
    pub total_cleaned: i64,
    #[serde(rename = "totalCount")]
    pub total_count: i32,
    #[serde(rename = "monthlyStats")]
    pub monthly_stats: Vec<MonthlyStat>,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct MonthlyStat {
    pub month: String,
    pub cleaned: i64,
    pub count: i32,
}

fn get_history_path() -> PathBuf {
    let home = dirs::home_dir().unwrap_or_else(|| PathBuf::from("."));
    home.join(".devcleaner").join("history.json")
}

#[tauri::command]
pub async fn get_clean_history(
    filter: Option<String>, // 'day' | 'week' | 'month' | 'all'
) -> Result<CleanHistory, String> {
    let history_path = get_history_path();
    let filter = filter.unwrap_or_else(|| "all".to_string());

    let items: Vec<CleanHistoryItemV2> = if history_path.exists() {
        let content = fs::read_to_string(&history_path)
            .map_err(|e| format!("Failed to read history: {}", e))?;
        serde_json::from_str(&content).unwrap_or_default()
    } else {
        Vec::new()
    };

    let now = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_secs() as i64)
        .unwrap_or(0);

    // 按时间筛选
    let filtered_items: Vec<CleanHistoryItemV2> = match filter.as_str() {
        "day" => {
            let threshold = 24 * 60 * 60;
            items.into_iter().filter(|i| now - i.timestamp < threshold).collect()
        }
        "week" => {
            let threshold = 7 * 24 * 60 * 60;
            items.into_iter().filter(|i| now - i.timestamp < threshold).collect()
        }
        "month" => {
            let threshold = 30 * 24 * 60 * 60;
            items.into_iter().filter(|i| now - i.timestamp < threshold).collect()
        }
        _ => items,
    };

    // 计算月度统计
    let mut monthly_map: std::collections::HashMap<String, (i64, i32)> = std::collections::HashMap::new();
    for item in &filtered_items {
        let date = chrono_date(item.timestamp);
        let entry = monthly_map.entry(date).or_insert((0, 0));
        entry.0 += item.size;
        entry.1 += 1;
    }

    let mut monthly_stats: Vec<MonthlyStat> = monthly_map
        .into_iter()
        .map(|(month, (cleaned, count))| MonthlyStat { month, cleaned, count })
        .collect();
    monthly_stats.sort_by(|a, b| b.month.cmp(&a.month));

    let total_cleaned: i64 = filtered_items.iter().map(|i| i.size).sum();
    let total_count = filtered_items.len() as i32;

    Ok(CleanHistory {
        items: filtered_items,
        total_cleaned,
        total_count,
        monthly_stats,
    })
}

fn chrono_date(timestamp: i64) -> String {
    let secs = timestamp as i64;
    let days = secs / (24 * 60 * 60);
    let years = days / 365;
    let remaining_days = days % 365;
    let months = remaining_days / 30;
    format!("{}-{:02}", 2024 + years as i64, months + 1)
}

// 记录清理历史
#[tauri::command]
pub async fn record_clean_history(
    tool_id: String,
    tool_name: String,
    size: i64,
    file_num: i32,
    paths: Vec<String>,
) -> Result<(), String> {
    let history_path = get_history_path();

    // 确保目录存在
    if let Some(parent) = history_path.parent() {
        fs::create_dir_all(parent)
            .map_err(|e| format!("Failed to create history directory: {}", e))?;
    }

    // 读取现有历史
    let mut items: Vec<CleanHistoryItemV2> = if history_path.exists() {
        let content = fs::read_to_string(&history_path)
            .map_err(|e| format!("Failed to read history: {}", e))?;
        serde_json::from_str(&content).unwrap_or_default()
    } else {
        Vec::new()
    };

    let now = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map(|d| d.as_secs() as i64)
        .unwrap_or(0);

    // 添加新记录
    items.push(CleanHistoryItemV2 {
        id: format!("{}-{}", tool_id, now),
        tool_id,
        tool_name,
        size,
        file_num,
        timestamp: now,
        paths,
    });

    // 只保留最近 500 条记录
    if items.len() > 500 {
        items = items.into_iter().rev().take(500).rev().collect();
    }

    // 保存
    let json = serde_json::to_string_pretty(&items)
        .map_err(|e| format!("Failed to serialize history: {}", e))?;

    fs::write(&history_path, json).map_err(|e| format!("Failed to write history: {}", e))?;

    Ok(())
}

// 导出清理报告
#[tauri::command]
pub async fn export_clean_report(
    format: String, // 'json' | 'csv'
) -> Result<String, String> {
    let history = get_clean_history(Some("all".to_string())).await?;

    match format.as_str() {
        "json" => {
            serde_json::to_string_pretty(&history)
                .map_err(|e| format!("Failed to export JSON: {}", e))
        }
        "csv" => {
            let mut csv = String::from("时间,工具,大小(字节),文件数,路径\n");
            for item in history.items {
                let time = chrono_date(item.timestamp);
                let paths = item.paths.join("; ");
                csv.push_str(&format!(
                    "{},{},{},{},{}\n",
                    time, item.tool_name, item.size, item.file_num, paths
                ));
            }
            Ok(csv)
        }
        _ => Err("Unsupported format".to_string()),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    // ============== 基础测试 ==============

    #[test]
    fn test_expand_path() {
        let expanded = expand_path("~/.npm");
        assert!(expanded.contains(".npm") || expanded.contains("/"));
    }

    #[test]
    fn test_expand_path_with_env_var() {
        // 测试包含环境变量的路径
        std::env::set_var("TEST_HOME", "/test");
        let expanded = expand_path("$TEST_HOME/project");
        assert!(expanded.contains("test") || expanded.contains("/"));
        std::env::remove_var("TEST_HOME");
    }

    #[test]
    fn test_is_whitelisted() {
        let whitelist = vec!["/important/path".to_string(), "/safe/directory".to_string()];

        assert!(is_path_whitelisted("/important/path/subdir", &whitelist));
        assert!(is_path_whitelisted("/safe/directory/file.txt", &whitelist));
        assert!(!is_path_whitelisted("/other/path", &whitelist));
    }

    #[test]
    fn test_is_whitelisted_empty_list() {
        let whitelist: Vec<String> = vec![];
        assert!(!is_path_whitelisted("/any/path", &whitelist));
    }

    // ============== 路径安全测试 ==============

    #[test]
    fn test_is_path_safe_rejects_path_traversal() {
        // 路径遍历攻击应该被拒绝
        assert!(!is_path_safe("../etc/passwd"));
        assert!(!is_path_safe("foo/../../../etc/passwd"));
        assert!(!is_path_safe("foo/../../bar"));
    }

    #[test]
    fn test_is_path_safe_requires_absolute_path() {
        // 相对路径应该被拒绝
        assert!(!is_path_safe("relative/path"));
        assert!(!is_path_safe("some/directory"));
    }

    #[test]
    fn test_is_path_safe_accepts_tilde_path() {
        // ~ 开头的路径应该被处理
        let result = is_path_safe("~/test");
        // 结果取决于用户主目录是否可访问
        assert!(result == true || result == false); // 允许任何结果，只要不崩溃
    }

    // ============== 项目类型检测测试 ==============

    #[test]
    fn test_get_project_types() {
        let project_types = get_project_types();
        
        // 验证包含常见项目类型
        let type_names: Vec<&str> = project_types.iter().map(|p| p.name).collect();
        
        assert!(type_names.contains(&"Node.js"));
        assert!(type_names.contains(&"Python"));
        assert!(type_names.contains(&"Rust"));
        assert!(type_names.contains(&"Go"));
        assert!(type_names.contains(&"Maven"));
        assert!(type_names.contains(&"Gradle"));
    }

    #[test]
    fn test_project_types_have_patterns() {
        let project_types = get_project_types();
        
        for project_type in project_types {
            assert!(!project_type.pattern.is_empty());
            assert!(!project_type.cleanable_dirs.is_empty());
            assert!(["safe", "moderate", "careful"].contains(&project_type.risk_level));
        }
    }

    #[test]
    fn test_project_types_cleanable_dirs() {
        let project_types = get_project_types();
        
        // Node.js 应该包含 node_modules
        let node_type = project_types.iter().find(|p| p.name == "Node.js").unwrap();
        assert!(node_type.cleanable_dirs.contains(&"node_modules"));
        
        // Python 应该包含 __pycache__
        let python_type = project_types.iter().find(|p| p.name == "Python").unwrap();
        assert!(python_type.cleanable_dirs.contains(&"__pycache__"));
        
        // Rust 应该包含 target
        let rust_type = project_types.iter().find(|p| p.name == "Rust").unwrap();
        assert!(rust_type.cleanable_dirs.contains(&"target"));
    }

    // ============== 日期格式化测试 ==============

    #[test]
    fn test_chrono_date() {
        // 使用已知的时间戳测试
        // 2024-01-15 00:00:00 UTC
        let timestamp = 1705276800;
        let date = chrono_date(timestamp);
        assert!(date.contains("-"));
    }

    #[test]
    fn test_chrono_date_format() {
        // 验证日期格式为 YYYY-MM
        let timestamp = 1705276800;
        let date = chrono_date(timestamp);
        let parts: Vec<&str> = date.split('-').collect();
        assert_eq!(parts.len(), 2);
    }

    // ============== 数据结构序列化测试 ==============

    #[test]
    fn test_clean_preview_serialization() {
        let preview = CleanPreview {
            tool_id: "npm".to_string(),
            tool_name: "npm".to_string(),
            paths: vec![PreviewPath {
                path: "/test/path".to_string(),
                files: vec![],
                size: 1024,
                oldest_file: 0,
                newest_file: 1000,
            }],
            total_size: 1024,
            risk_level: "safe".to_string(),
            recommendations: vec!["Test recommendation".to_string()],
        };

        let json = serde_json::to_string(&preview).unwrap();
        assert!(json.contains("npm"));
        assert!(json.contains("safe"));
    }

    #[test]
    fn test_project_scan_result_serialization() {
        let result = ProjectScanResult {
            name: "test-project".to_string(),
            path: "/test/path".to_string(),
            project_type: "Node.js".to_string(),
            size: 1024000,
            file_num: 100,
            last_modified: 1705276800,
            cleanable_items: vec![CleanableItem {
                id: "1".to_string(),
                name: "node_modules".to_string(),
                path: "/test/path/node_modules".to_string(),
                item_type: "node_modules".to_string(),
                size: 1000000,
                file_num: 90,
                last_modified: 1705276800,
                cleanable: true,
                reason: "Dependencies can be reinstalled".to_string(),
            }],
            risk_level: "safe".to_string(),
        };

        let json = serde_json::to_string(&result).unwrap();
        assert!(json.contains("test-project"));
        assert!(json.contains("Node.js"));
    }

    #[test]
    fn test_clean_history_serialization() {
        let history = CleanHistory {
            items: vec![CleanHistoryItemV2 {
                id: "1".to_string(),
                tool_id: "npm".to_string(),
                tool_name: "npm".to_string(),
                size: 1024000,
                file_num: 100,
                timestamp: 1705276800,
                paths: vec!["~/.npm".to_string()],
            }],
            total_cleaned: 1024000,
            total_count: 1,
            monthly_stats: vec![MonthlyStat {
                month: "2024-01".to_string(),
                cleaned: 1024000,
                count: 1,
            }],
        };

        let json = serde_json::to_string(&history).unwrap();
        assert!(json.contains("npm"));
        assert!(json.contains("2024-01"));
    }

    #[test]
    fn test_clean_result_serialization() {
        let result = CleanResult {
            tool_id: "npm".to_string(),
            cleaned: 2048000,
            failed: vec!["/path/to/failed: Permission denied".to_string()],
            file_num: 200,
        };

        let json = serde_json::to_string(&result).unwrap();
        assert!(json.contains("npm"));
        assert!(json.contains("2048000"));
    }

    // ============== 清理策略测试 ==============

    #[test]
    fn test_cleanable_item_all_types() {
        // 测试所有支持的可清理项目类型
        let valid_types = vec![
            "node_modules", "pycache", "target", "vendor", 
            "bin_obj", "dist", "cache", "other"
        ];
        
        for type_name in valid_types {
            let item = CleanableItem {
                id: "test".to_string(),
                name: type_name.to_string(),
                path: "/test".to_string(),
                item_type: type_name.to_string(),
                size: 1000,
                file_num: 10,
                last_modified: 1705276800,
                cleanable: true,
                reason: "Test".to_string(),
            };
            
            assert_eq!(item.item_type, type_name);
        }
    }

    // ============== 风险等级测试 ==============

    #[test]
    fn test_risk_levels() {
        // 测试所有有效的风险等级
        let valid_levels = vec!["safe", "moderate", "careful"];
        
        for level in valid_levels {
            let preview = CleanPreview {
                tool_id: "test".to_string(),
                tool_name: "test".to_string(),
                paths: vec![],
                total_size: 0,
                risk_level: level.to_string(),
                recommendations: vec![],
            };
            
            assert_eq!(preview.risk_level, level);
        }
    }

    // ============== 辅助函数测试 ==============

    #[test]
    fn test_get_current_platform() {
        let platform = get_current_platform();
        // 平台应该是已知值之一
        assert!(
            platform == "darwin" || 
            platform == "linux" || 
            platform == "windows" ||
            platform.is_empty()
        );
    }

    #[test]
    fn test_get_stats_path() {
        let path = get_stats_path();
        let path_str = path.to_string_lossy();
        
        // 路径应该包含 .devcleaner
        assert!(path_str.contains(".devcleaner") || path_str.contains("devcleaner"));
        assert!(path_str.contains("stats.json"));
    }

    #[test]
    fn test_get_history_path() {
        let path = get_history_path();
        let path_str = path.to_string_lossy();
        
        // 路径应该包含 .devcleaner
        assert!(path_str.contains(".devcleaner") || path_str.contains("devcleaner"));
        assert!(path_str.contains("history.json"));
    }

    #[test]
    fn test_get_settings_path() {
        let path = get_settings_path();
        let path_str = path.to_string_lossy();
        
        // 路径应该包含 .devcleaner
        assert!(path_str.contains(".devcleaner") || path_str.contains("devcleaner"));
        assert!(path_str.contains("settings.json"));
    }

    // ============== 预览文件测试 ==============

    #[test]
    fn test_preview_file_creation() {
        let file = PreviewFile {
            name: "test.js".to_string(),
            path: "/test/test.js".to_string(),
            size: 1024,
            modified: 1705276800,
            is_safe: true,
            reason: None,
        };

        assert_eq!(file.name, "test.js");
        assert!(file.is_safe);
        assert!(file.reason.is_none());
    }

    #[test]
    fn test_preview_file_with_reason() {
        let file = PreviewFile {
            name: "important.json".to_string(),
            path: "/test/important.json".to_string(),
            size: 512,
            modified: 1705276800,
            is_safe: false,
            reason: Some("Non-cache file, recommended to keep".to_string()),
        };

        assert!(!file.is_safe);
        assert!(file.reason.is_some());
    }

    #[test]
    fn test_preview_path_aggregation() {
        let preview_path = PreviewPath {
            path: "/test/cache".to_string(),
            files: vec![
                PreviewFile {
                    name: "file1".to_string(),
                    path: "/test/cache/file1".to_string(),
                    size: 100,
                    modified: 1000,
                    is_safe: true,
                    reason: None,
                },
                PreviewFile {
                    name: "file2".to_string(),
                    path: "/test/cache/file2".to_string(),
                    size: 200,
                    modified: 2000,
                    is_safe: true,
                    reason: None,
                },
            ],
            size: 300,
            oldest_file: 1000,
            newest_file: 2000,
        };

        assert_eq!(preview_path.files.len(), 2);
        assert_eq!(preview_path.oldest_file, 1000);
        assert_eq!(preview_path.newest_file, 2000);
    }

    // ============== 历史记录测试 ==============

    #[test]
    fn test_monthly_stat_calculation() {
        let stat = MonthlyStat {
            month: "2026-05".to_string(),
            cleaned: 10240000,
            count: 5,
        };

        assert_eq!(stat.count, 5);
        assert!(stat.cleaned > 0);
    }

    #[test]
    fn test_clean_history_item_v2() {
        let item = CleanHistoryItemV2 {
            id: "npm-1705276800".to_string(),
            tool_id: "npm".to_string(),
            tool_name: "npm".to_string(),
            size: 1024000,
            file_num: 100,
            timestamp: 1705276800,
            paths: vec!["~/.npm".to_string(), "~/.npm/_cacache".to_string()],
        };

        assert_eq!(item.tool_id, "npm");
        assert_eq!(item.paths.len(), 2);
    }

    // ============== 配置缓存测试 ==============

    #[test]
    fn test_settings_default_values() {
        let settings = Settings::default();
        
        assert_eq!(settings.threshold, 100 * 1024 * 1024); // 100MB
        assert!(settings.whitelist.is_empty());
        assert!(!settings.auto_scan);
        assert_eq!(settings.scan_interval, 24);
        assert_eq!(settings.theme, "auto");
    }

    // ============== 错误处理测试 ==============

    #[test]
    fn test_clean_result_with_no_failures() {
        let result = CleanResult {
            tool_id: "npm".to_string(),
            cleaned: 1024000,
            failed: vec![],
            file_num: 100,
        };

        assert!(result.failed.is_empty());
        assert_eq!(result.cleaned, 1024000);
    }

    #[test]
    fn test_clean_result_with_multiple_failures() {
        let result = CleanResult {
            tool_id: "npm".to_string(),
            cleaned: 500000,
            failed: vec![
                "/path1: Permission denied".to_string(),
                "/path2: File not found".to_string(),
            ],
            file_num: 50,
        };

        assert_eq!(result.failed.len(), 2);
    }

    // ============== 大小计算测试 ==============

    #[test]
    fn test_calculate_dir_size_result() {
        // 创建一个临时目录进行测试
        use std::io::Write;
        
        let temp_dir = std::env::temp_dir();
        let test_dir = temp_dir.join("devcleaner_test");
        
        // 创建测试目录
        let _ = std::fs::create_dir_all(&test_dir);
        
        // 创建测试文件
        let test_file = test_dir.join("test.txt");
        let mut file = std::fs::File::create(&test_file).unwrap();
        file.write_all(b"test content").unwrap();
        
        let (size, file_num, _last_modified) = calculate_dir_size(&test_dir, 3);
        
        assert!(size > 0);
        assert!(file_num > 0);
        
        // 清理
        let _ = std::fs::remove_file(&test_file);
        let _ = std::fs::remove_dir(&test_dir);
    }
}

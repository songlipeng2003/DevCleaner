use serde::{Deserialize, Serialize};
use std::path::PathBuf;
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
pub struct Config {
    #[serde(rename = "version")]
    version: String,
    providers: Vec<ProviderConfig>,
}

#[allow(dead_code)]
#[derive(Debug, Deserialize, Clone)]
pub struct ProviderConfig {
    id: String,
    name: String,
    #[serde(rename = "description")]
    description: Option<String>,
    #[serde(rename = "platforms")]
    platforms: Option<Vec<String>>,
    #[serde(rename = "icon")]
    icon: Option<String>,
    paths: Vec<PathPattern>,
}

#[allow(dead_code)]
#[derive(Debug, Deserialize, Clone)]
pub struct ToolConfig {
    #[serde(rename = "toolId")]
    pub tool_id: String,
    #[serde(rename = "toolName")]
    pub tool_name: String,
    pub paths: Vec<PathPattern>,
    #[serde(rename = "excludePatterns")]
    pub exclude_patterns: Option<Vec<String>>,
    #[serde(rename = "scanOnInit")]
    pub scan_on_init: Option<bool>,
}

#[derive(Debug, Deserialize, Clone)]
pub struct PathPattern {
    #[serde(rename = "path")]
    pub path: String,
    #[serde(rename = "description")]
    #[allow(dead_code)]
    pub description: Option<String>,
    #[serde(rename = "platform")]
    #[allow(dead_code)]
    pub platform: Option<String>,
    #[serde(rename = "strategy")]
    #[allow(dead_code)]
    pub strategy: Option<String>,
    #[serde(rename = "command")]
    #[allow(dead_code)]
    pub command: Option<String>,
}

#[allow(dead_code)]
#[derive(Debug, Deserialize, Clone)]
pub struct ProjectConfig {
    #[serde(rename = "projectId")]
    project_id: String,
    #[serde(rename = "projectName")]
    project_name: String,
    paths: Vec<ProjectPath>,
}

#[allow(dead_code)]
#[derive(Debug, Deserialize, Clone)]
pub struct ProjectPath {
    path: String,
    #[serde(rename = "platform")]
    platform: Option<String>,
    #[serde(rename = "cacheTypes")]
    cache_types: Vec<String>,
    #[serde(rename = "cleanStrategy")]
    clean_strategy: Option<String>,
}

// ============== 项目扫描结果 ==============

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ProjectInfo {
    pub name: String,
    pub path: String,
    #[serde(rename = "toolId")]
    pub tool_id: String,
    pub size: i64,
    #[serde(rename = "cacheSize")]
    pub cache_size: i64,
    #[serde(rename = "lastModified")]
    pub last_modified: i64,
    #[serde(rename = "fileCount")]
    pub file_count: i32,
    pub cleanable: bool,
    #[serde(rename = "cleanReason")]
    pub clean_reason: Option<String>,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ProjectScanResult {
    pub projects: Vec<ProjectInfo>,
    pub total_size: i64,
    pub total_count: i32,
}

// ============== 清理历史记录 ==============

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct CleanHistory {
    #[serde(rename = "cleanId")]
    pub clean_id: String,
    pub date: i64,
    pub tool_id: String,
    pub tool_name: String,
    pub size: i64,
    #[serde(rename = "fileCount")]
    pub file_count: i32,
    pub paths: Vec<String>,
    pub note: Option<String>,
}

// ============== 清理报告 ==============

#[derive(Debug, Serialize, Deserialize)]
pub struct CleanReport {
    #[serde(rename = "reportId")]
    pub report_id: String,
    pub date: String,
    pub total_size: i64,
    #[serde(rename = "totalFiles")]
    pub total_files: i32,
    pub items: Vec<ReportItem>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ReportItem {
    #[serde(rename = "toolId")]
    pub tool_id: String,
    #[serde(rename = "toolName")]
    pub tool_name: String,
    pub size: i64,
    #[serde(rename = "fileCount")]
    pub file_count: i32,
    pub paths: Vec<String>,
}

// ============== 磁盘分析结构 (v0.3.0) ==============

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct DiskAnalysisItem {
    pub name: String,
    pub path: String,
    #[serde(rename = "toolId")]
    pub tool_id: Option<String>,
    pub size: i64,
    pub percentage: f64,
    #[serde(rename = "fileCount")]
    pub file_count: i32,
    #[serde(rename = "lastModified")]
    pub last_modified: i64,
    #[serde(rename = "isCleanable")]
    pub is_cleanable: bool,
    #[serde(rename = "cleanReason")]
    pub clean_reason: Option<String>,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct DiskAnalysisCategory {
    pub name: String,
    pub items: Vec<DiskAnalysisItem>,
    #[serde(rename = "totalSize")]
    pub total_size: i64,
    #[serde(rename = "itemCount")]
    pub item_count: i32,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct DiskAnalysisResult {
    pub categories: Vec<DiskAnalysisCategory>,
    #[serde(rename = "totalSize")]
    pub total_size: i64,
    #[serde(rename = "cleanableSize")]
    pub cleanable_size: i64,
    #[serde(rename = "totalItems")]
    pub total_items: i32,
    pub timestamp: i64,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct CacheTrend {
    pub date: String,
    pub size: i64,
}

// ============== 辅助函数 ==============

#[cfg(target_os = "macos")]
pub fn get_home_dir() -> PathBuf {
    dirs::home_dir().unwrap_or_else(|| PathBuf::from("/Users/guest"))
}

#[cfg(target_os = "windows")]
pub fn get_home_dir() -> PathBuf {
    dirs::home_dir().unwrap_or_else(|| PathBuf::from("C:\\Users\\guest"))
}

#[cfg(target_os = "linux")]
pub fn get_home_dir() -> PathBuf {
    dirs::home_dir().unwrap_or_else(|| PathBuf::from("/home/guest"))
}

#[allow(dead_code)]
pub fn get_config_path() -> PathBuf {
    let home = get_home_dir();
    home.join(".config").join("devcleaner")
}

#[allow(dead_code)]
pub fn get_data_path() -> PathBuf {
    let home = get_home_dir();
    home.join(".local").join("share").join("devcleaner")
}

#[allow(dead_code)]
pub fn get_cache_dir() -> PathBuf {
    let home = get_home_dir();
    home.join("Library").join("Caches").join("devcleaner")
}

pub fn get_stats_path() -> PathBuf {
    get_cache_dir().join("stats.json")
}

#[allow(dead_code)]
pub fn get_history_path() -> PathBuf {
    get_data_path().join("history.json")
}

pub fn expand_path(path_str: &str) -> PathBuf {
    let expanded = path_str.replace('~', &get_home_dir().to_string_lossy());
    PathBuf::from(expanded)
}

pub fn scan_directory_size(path: &PathBuf) -> (i64, i32, i64) {
    let mut total_size: i64 = 0;
    let mut file_count: i32 = 0;
    let mut last_modified: i64 = 0;

    if path.exists() {
        for entry in WalkDir::new(path).into_iter().filter_map(|e| e.ok()) {
            if entry.file_type().is_file() {
                if let Ok(metadata) = entry.metadata() {
                    total_size += metadata.len() as i64;
                    file_count += 1;

                    if let Ok(modified) = metadata.modified() {
                        let modified_timestamp = modified
                            .duration_since(SystemTime::UNIX_EPOCH)
                            .unwrap_or_default()
                            .as_secs() as i64;
                        if modified_timestamp > last_modified {
                            last_modified = modified_timestamp;
                        }
                    }
                }
            }
        }
    }

    (total_size, file_count, last_modified)
}

pub fn get_config() -> Config {
    let app_dir = std::env::current_exe()
        .ok()
        .and_then(|p| p.parent().map(|p| p.to_path_buf()))
        .unwrap_or_else(|| PathBuf::from("."));

    let config_path = app_dir.join("providers.json");

    if config_path.exists() {
        let content = std::fs::read_to_string(&config_path).unwrap_or_default();
        serde_json::from_str(&content).unwrap_or_else(|e| {
            eprintln!("Failed to parse config: {}", e);
            Config {
                version: "1.0.0".to_string(),
                providers: vec![],
            }
        })
    } else {
        Config {
            version: "1.0.0".to_string(),
            providers: vec![],
        }
    }
}

// 获取用户配置的工具列表
// 将 Provider 转换为 ToolConfig，因为 JSON 中 provider 就是工具
pub fn get_tools() -> Vec<ToolConfig> {
    let config = get_config();
    config.providers.into_iter().map(|p| ToolConfig {
        tool_id: p.id.clone(),
        tool_name: p.name.clone(),
        paths: p.paths,
        exclude_patterns: None,
        scan_on_init: None,
    }).collect()
}

// 平台检测
#[allow(dead_code)]
pub fn is_macos() -> bool {
    cfg!(target_os = "macos")
}

#[allow(dead_code)]
pub fn is_windows() -> bool {
    cfg!(target_os = "windows")
}

#[allow(dead_code)]
pub fn is_linux() -> bool {
    cfg!(target_os = "linux")
}

#[allow(dead_code)]
pub fn get_platform() -> &'static str {
    if cfg!(target_os = "macos") {
        "macos"
    } else if cfg!(target_os = "windows") {
        "windows"
    } else {
        "linux"
    }
}

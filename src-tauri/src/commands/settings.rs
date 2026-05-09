use crate::commands::types::*;
use std::fs;
use std::sync::Mutex;

// ============== 设置相关命令 ==============

static SETTINGS_MUTEX: once_cell::sync::Lazy<Mutex<()>> =
    once_cell::sync::Lazy::new(|| Mutex::new(()));

fn get_settings_path() -> std::path::PathBuf {
    let config_path = get_config_path();
    if !config_path.exists() {
        let _ = fs::create_dir_all(&config_path);
    }
    config_path.join("settings.json")
}

#[tauri::command]
pub async fn get_settings() -> Result<Settings, String> {
    let settings_path = get_settings_path();

    if settings_path.exists() {
        let content = fs::read_to_string(&settings_path).map_err(|e| e.to_string())?;
        serde_json::from_str(&content).map_err(|e| e.to_string())
    } else {
        Ok(Settings::default())
    }
}

#[tauri::command]
pub async fn save_settings(settings: Settings) -> Result<(), String> {
    let _lock = SETTINGS_MUTEX.lock().map_err(|e| e.to_string())?;

    let settings_path = get_settings_path();
    let content = serde_json::to_string_pretty(&settings).map_err(|e| e.to_string())?;
    fs::write(&settings_path, content).map_err(|e| e.to_string())
}

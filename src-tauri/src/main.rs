#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod commands;

use tauri::Manager;

fn main() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .invoke_handler(tauri::generate_handler![
            commands::get_tool_list,
            commands::get_tool_info,
            commands::scan_tool,
            commands::scan_all_tools,
            commands::preview_tool,
            commands::clean_tool,
            commands::get_settings,
            commands::save_settings,
            commands::get_disk_usage,
            commands::open_path,
            commands::get_version,
            commands::get_usage_stats,
            commands::record_clean,
        ])
        .setup(|app| {
            let window = app.get_webview_window("main").unwrap();
            window.set_title("DevCleaner - 开发者磁盘清理工具").unwrap();
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

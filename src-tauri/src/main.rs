#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod commands;

use tauri::Manager;

fn main() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_dialog::init())
        .plugin(tauri_plugin_aptabase::Builder::new(std::env!("APTABASE_KEY")).build())
        .invoke_handler(tauri::generate_handler![
            // 工具扫描命令
            commands::scan::get_tool_list,
            commands::scan::get_tool_info,
            commands::scan::scan_tool,
            commands::scan::scan_all_tools,
            commands::scan::preview_tool,
            commands::scan::clean_tool,
            // 设置命令
            commands::settings::get_settings,
            commands::settings::save_settings,
            // 磁盘命令
            commands::disk::get_disk_usage,
            commands::disk::open_path,
            commands::disk::get_version,
            // 统计命令
            commands::stats::get_usage_stats,
            commands::stats::record_clean,
            // 项目命令
            commands::projects::scan_projects,
            commands::projects::get_clean_preview,
            commands::projects::clean_paths,
            commands::projects::get_clean_history,
            commands::projects::record_clean_history,
            commands::projects::export_clean_report,
            // 分析命令
            commands::analysis::get_disk_analysis,
            commands::analysis::get_cache_trends,
        ])
        .setup(|app| {
            let window = app.get_webview_window("main").unwrap();
            window.set_title("DevCleaner - 开发者磁盘清理工具").unwrap();
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

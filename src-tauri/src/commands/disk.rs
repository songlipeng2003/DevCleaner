use crate::commands::types::*;

// ============== 磁盘使用相关命令 ==============

#[tauri::command]
pub async fn get_disk_usage() -> Result<DiskUsage, String> {
    #[cfg(target_os = "macos")]
    {
        let output = std::process::Command::new("df")
            .args(["-k", "/"])
            .output()
            .map_err(|e| e.to_string())?;

        let output_str = String::from_utf8_lossy(&output.stdout);
        let lines: Vec<&str> = output_str.lines().collect();

        if lines.len() >= 2 {
            let parts: Vec<&str> = lines[1].split_whitespace().collect();
            if parts.len() >= 4 {
                let total_kb: i64 = parts[1].parse().unwrap_or(0);
                let used_kb: i64 = parts[2].parse().unwrap_or(0);
                let free_kb: i64 = parts[3].parse().unwrap_or(0);

                return Ok(DiskUsage {
                    total: total_kb * 1024,
                    used: used_kb * 1024,
                    free: free_kb * 1024,
                });
            }
        }

        Err("Failed to parse disk usage".to_string())
    }

    #[cfg(target_os = "windows")]
    {
        let output = std::process::Command::new("wmic")
            .args(["LogicalDisk", "Where", "DeviceID='C:'", "Get", "Size,FreeSpace", "/format:value"])
            .output()
            .map_err(|e| e.to_string())?;

        let output_str = String::from_utf8_lossy(&output.stdout);

        let mut free_space: i64 = 0;
        let mut total_size: i64 = 0;

        for line in output_str.lines() {
            if line.starts_with("FreeSpace=") {
                free_space = line
                    .trim_start_matches("FreeSpace=")
                    .parse()
                    .unwrap_or(0);
            } else if line.starts_with("Size=") {
                total_size = line.trim_start_matches("Size=").parse().unwrap_or(0);
            }
        }

        if total_size > 0 {
            Ok(DiskUsage {
                total: total_size,
                used: total_size - free_space,
                free: free_space,
            })
        } else {
            Err("Failed to parse disk usage".to_string())
        }
    }

    #[cfg(target_os = "linux")]
    {
        let output = std::process::Command::new("df")
            .args(["-B1", "/"])
            .output()
            .map_err(|e| e.to_string())?;

        let output_str = String::from_utf8_lossy(&output.stdout);
        let lines: Vec<&str> = output_str.lines().collect();

        if lines.len() >= 2 {
            let parts: Vec<&str> = lines[1].split_whitespace().collect();
            if parts.len() >= 4 {
                let total: i64 = parts[1].parse().unwrap_or(0);
                let used: i64 = parts[2].parse().unwrap_or(0);
                let free: i64 = parts[3].parse().unwrap_or(0);

                return Ok(DiskUsage {
                    total,
                    used,
                    free,
                });
            }
        }

        Err("Failed to parse disk usage".to_string())
    }
}

#[tauri::command]
pub async fn open_path(path: String) -> Result<(), String> {
    let expanded = expand_path(&path);

    #[cfg(target_os = "macos")]
    {
        std::process::Command::new("open")
            .arg(&expanded)
            .spawn()
            .map_err(|e| e.to_string())?;
    }

    #[cfg(target_os = "windows")]
    {
        std::process::Command::new("explorer")
            .arg(&expanded)
            .spawn()
            .map_err(|e| e.to_string())?;
    }

    #[cfg(target_os = "linux")]
    {
        std::process::Command::new("xdg-open")
            .arg(&expanded)
            .spawn()
            .map_err(|e| e.to_string())?;
    }

    Ok(())
}

#[tauri::command]
pub async fn get_version() -> Result<String, String> {
    Ok(env!("CARGO_PKG_VERSION").to_string())
}

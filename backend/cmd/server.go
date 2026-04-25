package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/devcleaner/backend/provider"
	"github.com/devcleaner/backend/scanner"
	"github.com/devcleaner/backend/tools"
)

// Server HTTP 服务器
type Server struct {
	port      string
	scanner   *scanner.Scanner
	whitelist []string
}

// NewServer 创建服务器
func NewServer(port string) *Server {
	return &Server{
		port:      port,
		scanner:   scanner.NewScanner([]string{}),
		whitelist: []string{},
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := NewServer(port)
	
	// 路由
	http.HandleFunc("/api/tools", server.handleTools)
	http.HandleFunc("/api/tools/", server.handleToolByID)
	http.HandleFunc("/api/scan", server.handleScan)
	http.HandleFunc("/api/clean", server.handleClean)
	http.HandleFunc("/api/settings", server.handleSettings)
	http.HandleFunc("/api/system/disk", server.handleDiskUsage)
	http.HandleFunc("/api/system/version", server.handleVersion)
	
	// 静态文件（生产环境）
	// http.Handle("/", http.FileServer(http.Dir("./dist")))

	log.Printf("DevCleaner Backend Server starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// ToolResponse 工具响应
type ToolResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Paths       []string `json:"paths"`
	Enabled     bool     `json:"enabled"`
	Description string   `json:"description,omitempty"`
}

func (s *Server) handleTools(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		result := make([]ToolResponse, 0)
		for _, tool := range tools.AllTools {
			paths := make([]string, len(tool.Paths))
			for i, p := range tool.Paths {
				paths[i] = p.Path
			}
			result = append(result, ToolResponse{
				ID:      tool.ID,
				Name:    tool.Name,
				Paths:   paths,
				Enabled: true,
			})
		}
		jsonResponse(w, result)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleToolByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/api/tools/"):]
	tool := tools.GetToolByID(id)
	if tool == nil {
		http.Error(w, "Tool not found", http.StatusNotFound)
		return
	}
	jsonResponse(w, tool)
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ToolID string `json:"tool_id"`
		All    bool   `json:"all"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	start := time.Now()
	scannerEngine := scanner.NewScanner(s.whitelist)
	
	if req.All {
		// 扫描所有工具
		for _, tool := range tools.AllTools {
			paths := make([]scanner.PathInfo, len(tool.Paths))
			for i, p := range tool.Paths {
				paths[i] = scanner.PathInfo{Path: p.Path, Description: p.Description}
			}
			scannerEngine.ScanTool(tool.ID, paths)
		}
	} else if req.ToolID != "" {
		// 扫描指定工具
		tool := tools.GetToolByID(req.ToolID)
		if tool == nil {
			http.Error(w, "Tool not found", http.StatusNotFound)
			return
		}
		paths := make([]scanner.PathInfo, len(tool.Paths))
		for i, p := range tool.Paths {
			paths[i] = scanner.PathInfo{Path: p.Path, Description: p.Description}
		}
		scannerEngine.ScanTool(tool.ID, paths)
	}

	results := scannerEngine.CollectResults()
	stats := scanner.ScanStats{
		ScanDuration: time.Since(start),
		ScannedPaths: len(results),
	}
	for _, r := range results {
		stats.TotalSize += r.Size
		stats.TotalFiles += r.FileNum
	}

	jsonResponse(w, map[string]interface{}{
		"results": results,
		"stats":   stats,
	})
}

func (s *Server) handleClean(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ToolID string   `json:"tool_id"`
		Paths  []string `json:"paths"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 获取provider并清理
	p := provider.GetProvider(req.ToolID)
	if p == nil {
		http.Error(w, "Tool provider not found", http.StatusNotFound)
		return
	}

	cleanResult, err := p.Clean(req.Paths)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{
		"tool_id":  req.ToolID,
		"cleaned":  cleanResult.Cleaned,
		"failed":   cleanResult.Failed,
		"file_num": cleanResult.FileNum,
	}
	jsonResponse(w, result)
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		settings := map[string]interface{}{
			"threshold":    100,
			"whitelist":    []string{},
			"auto_scan":    false,
			"scan_interval": 7,
			"theme":        "auto",
		}
		jsonResponse(w, settings)
	case http.MethodPut:
		var settings map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		// TODO: 保存设置
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleDiskUsage(w http.ResponseWriter, r *http.Request) {
	// 简化实现，实际应使用 syscall 获取真实磁盘使用情况
	usage := map[string]int64{
		"total": 500 * 1024 * 1024 * 1024, // 500 GB
		"used":  250 * 1024 * 1024 * 1024, // 250 GB
		"free":  250 * 1024 * 1024 * 1024, // 250 GB
	}
	jsonResponse(w, usage)
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	version := map[string]string{
		"version": "0.1.0",
		"build":   "alpha",
	}
	jsonResponse(w, version)
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func init() {
	fmt.Println("DevCleaner Backend v0.1.0-alpha")
}

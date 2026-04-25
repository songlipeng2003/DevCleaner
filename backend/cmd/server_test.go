package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleTools(t *testing.T) {
	server := NewServer("8080")
	req := httptest.NewRequest("GET", "/api/tools", nil)
	w := httptest.NewRecorder()
	server.handleTools(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var tools []ToolResponse
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &tools); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if len(tools) == 0 {
		t.Error("Expected at least one tool")
	}

	// 检查工具字段
	for _, tool := range tools {
		if tool.ID == "" {
			t.Error("Tool ID is empty")
		}
		if tool.Name == "" {
			t.Error("Tool Name is empty")
		}
		if len(tool.Paths) == 0 {
			t.Errorf("Tool %s has no paths", tool.ID)
		}
	}
}

func TestHandleToolByID(t *testing.T) {
	server := NewServer("8080")

	// 测试存在的工具
	req := httptest.NewRequest("GET", "/api/tools/npm", nil)
	w := httptest.NewRecorder()
	server.handleToolByID(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for existing tool, got %d", resp.StatusCode)
	}

	var tool map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &tool); err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if tool["id"] != "npm" {
		t.Errorf("Expected tool ID 'npm', got %v", tool["id"])
	}

	// 测试不存在的工具
	req2 := httptest.NewRequest("GET", "/api/tools/nonexistent", nil)
	w2 := httptest.NewRecorder()
	server.handleToolByID(w2, req2)

	resp2 := w2.Result()
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404 for non-existent tool, got %d", resp2.StatusCode)
	}
}

func TestHandleScan(t *testing.T) {
	server := NewServer("8080")

	// 测试扫描单个工具
	scanReq := map[string]interface{}{
		"tool_id": "npm",
		"all":     false,
	}
	bodyBytes, _ := json.Marshal(scanReq)
	req := httptest.NewRequest("POST", "/api/scan", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.handleScan(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for scan, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Errorf("Failed to unmarshal scan response: %v", err)
	}

	if _, ok := result["results"]; !ok {
		t.Error("Scan response missing 'results' field")
	}
	if _, ok := result["stats"]; !ok {
		t.Error("Scan response missing 'stats' field")
	}

	// 测试扫描所有工具
	scanAllReq := map[string]interface{}{
		"all": true,
	}
	bodyBytesAll, _ := json.Marshal(scanAllReq)
	reqAll := httptest.NewRequest("POST", "/api/scan", bytes.NewReader(bodyBytesAll))
	reqAll.Header.Set("Content-Type", "application/json")
	wAll := httptest.NewRecorder()
	server.handleScan(wAll, reqAll)

	respAll := wAll.Result()
	defer respAll.Body.Close()

	if respAll.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for scan all, got %d", respAll.StatusCode)
	}

	// 测试无效请求
	invalidReq := httptest.NewRequest("GET", "/api/scan", nil)
	wInvalid := httptest.NewRecorder()
	server.handleScan(wInvalid, invalidReq)

	respInvalid := wInvalid.Result()
	defer respInvalid.Body.Close()

	if respInvalid.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 for GET scan, got %d", respInvalid.StatusCode)
	}
}

func TestHandleClean(t *testing.T) {
	server := NewServer("8080")

	// 测试清理（使用不存在的路径，预期清理0个文件）
	cleanReq := map[string]interface{}{
		"tool_id": "npm",
		"paths":   []string{"/tmp/nonexistent/path"},
	}
	bodyBytes, _ := json.Marshal(cleanReq)
	req := httptest.NewRequest("POST", "/api/clean", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.handleClean(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for clean, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	respBody, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Errorf("Failed to unmarshal clean response: %v", err)
	}

	if result["tool_id"] != "npm" {
		t.Errorf("Expected tool_id 'npm', got %v", result["tool_id"])
	}

	// 测试无效的工具ID
	invalidToolReq := map[string]interface{}{
		"tool_id": "nonexistent",
		"paths":   []string{"/tmp"},
	}
	bodyBytesInvalid, _ := json.Marshal(invalidToolReq)
	reqInvalid := httptest.NewRequest("POST", "/api/clean", bytes.NewReader(bodyBytesInvalid))
	reqInvalid.Header.Set("Content-Type", "application/json")
	wInvalid := httptest.NewRecorder()
	server.handleClean(wInvalid, reqInvalid)

	respInvalid := wInvalid.Result()
	defer respInvalid.Body.Close()

	if respInvalid.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404 for non-existent tool, got %d", respInvalid.StatusCode)
	}

	// 测试无效请求方法
	getReq := httptest.NewRequest("GET", "/api/clean", nil)
	wGet := httptest.NewRecorder()
	server.handleClean(wGet, getReq)

	respGet := wGet.Result()
	defer respGet.Body.Close()

	if respGet.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 for GET clean, got %d", respGet.StatusCode)
	}
}

func TestHandleSettings(t *testing.T) {
	server := NewServer("8080")

	// 测试获取设置
	req := httptest.NewRequest("GET", "/api/settings", nil)
	w := httptest.NewRecorder()
	server.handleSettings(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for GET settings, got %d", resp.StatusCode)
	}

	var settings map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &settings); err != nil {
		t.Errorf("Failed to unmarshal settings response: %v", err)
	}

	if _, ok := settings["threshold"]; !ok {
		t.Error("Settings missing 'threshold' field")
	}
	if _, ok := settings["whitelist"]; !ok {
		t.Error("Settings missing 'whitelist' field")
	}
	if _, ok := settings["auto_scan"]; !ok {
		t.Error("Settings missing 'auto_scan' field")
	}

	// 测试更新设置
	updateReq := map[string]interface{}{
		"threshold": 200,
	}
	bodyBytes, _ := json.Marshal(updateReq)
	reqPut := httptest.NewRequest("PUT", "/api/settings", bytes.NewReader(bodyBytes))
	reqPut.Header.Set("Content-Type", "application/json")
	wPut := httptest.NewRecorder()
	server.handleSettings(wPut, reqPut)

	respPut := wPut.Result()
	defer respPut.Body.Close()

	if respPut.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for PUT settings, got %d", respPut.StatusCode)
	}

	// 测试无效请求方法
	postReq := httptest.NewRequest("POST", "/api/settings", nil)
	wPost := httptest.NewRecorder()
	server.handleSettings(wPost, postReq)

	respPost := wPost.Result()
	defer respPost.Body.Close()

	if respPost.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 for POST settings, got %d", respPost.StatusCode)
	}
}

func TestHandleDiskUsage(t *testing.T) {
	server := NewServer("8080")

	req := httptest.NewRequest("GET", "/api/system/disk", nil)
	w := httptest.NewRecorder()
	server.handleDiskUsage(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for disk usage, got %d", resp.StatusCode)
	}

	var usage map[string]int64
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &usage); err != nil {
		t.Errorf("Failed to unmarshal disk usage response: %v", err)
	}

	if _, ok := usage["total"]; !ok {
		t.Error("Disk usage missing 'total' field")
	}
	if _, ok := usage["used"]; !ok {
		t.Error("Disk usage missing 'used' field")
	}
	if _, ok := usage["free"]; !ok {
		t.Error("Disk usage missing 'free' field")
	}
}

func TestHandleVersion(t *testing.T) {
	server := NewServer("8080")

	req := httptest.NewRequest("GET", "/api/system/version", nil)
	w := httptest.NewRecorder()
	server.handleVersion(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for version, got %d", resp.StatusCode)
	}

	var version map[string]string
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &version); err != nil {
		t.Errorf("Failed to unmarshal version response: %v", err)
	}

	if _, ok := version["version"]; !ok {
		t.Error("Version missing 'version' field")
	}
	if _, ok := version["build"]; !ok {
		t.Error("Version missing 'build' field")
	}
}

func TestMain(m *testing.M) {
	fmt.Println("Running API tests...")
	m.Run()
}
package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// These tests require a running instance of the scanner service
// You can start it with: go run ./cmd/main/main.go
// or with docker: docker compose -f ./deployments/docker/docker-compose.yml up

var (
	serverURL = "http://localhost:8081"
)

func init() {
	// Allow override of server URL from environment
	if url := os.Getenv("TEST_SERVER_URL"); url != "" {
		serverURL = url
	}
}

func TestHealthCheck(t *testing.T) {
	// Skip this test unless explicitly enabled
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set RUN_INTEGRATION_TESTS=true to run")
	}

	resp, err := http.Get(fmt.Sprintf("%s/health", serverURL))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)

	assert.Equal(t, "healthy", result["status"])
	assert.NotEmpty(t, result["nmap_version"])
}

func TestScanWorkflow(t *testing.T) {
	// Skip this test unless explicitly enabled
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set RUN_INTEGRATION_TESTS=true to run")
	}

	// 1. Start a scan
	scanID := startScan(t)
	assert.NotEmpty(t, scanID)

	// 2. Check scan status (may need to wait for completion)
	var scanStatus string
	var resultID string
	for i := 0; i < 30; i++ { // Wait up to 30 seconds
		scan := getScan(t, scanID)
		scanStatus = scan["status"].(string)
		if resultID, _ = scan["result_id"].(string); resultID != "" {
			break
		}

		if scanStatus == "COMPLETED" || scanStatus == "FAILED" {
			break
		}

		// Wait a bit before checking again
		time.Sleep(1 * time.Second)
	}

	// 3. Get scan result if available
	if resultID != "" {
		result := getScanResult(t, resultID)
		assert.NotNil(t, result)
		assert.Equal(t, scanID, result["scan_id"])
	}
}

func startScan(t *testing.T) string {
	// Create scan request targeting localhost (safe for testing)
	reqBody := map[string]interface{}{
		"target":          "127.0.0.1", // scan localhost only
		"ports":           "22,80,443", // common ports
		"scan_type":       "CONNECT",   // TCP connect is the safest
		"timing_template": 1,           // slow scan to be less intrusive
		"timeout_seconds": 5,           // quick timeout
	}

	jsonData, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	resp, err := http.Post(fmt.Sprintf("%s/api/v1/scans", serverURL), "application/json", bytes.NewBuffer(jsonData))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)

	return result["scan_id"].(string)
}

func getScan(t *testing.T, scanID string) map[string]interface{} {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/scans/%s", serverURL, scanID))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	assert.NoError(t, err)

	return result
}

func getScanResult(t *testing.T, resultID string) map[string]interface{} {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/results/%s", serverURL, resultID))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	assert.NoError(t, err)

	return result
}

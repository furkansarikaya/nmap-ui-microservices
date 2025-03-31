package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ScanRequest represents the request body for starting a scan
type ScanRequest struct {
	Target           string   `json:"target"`
	Ports            string   `json:"ports,omitempty"`
	ScanType         string   `json:"scan_type,omitempty"`
	TimingTemplate   int      `json:"timing_template,omitempty"`
	ServiceDetection bool     `json:"service_detection,omitempty"`
	OSDetection      bool     `json:"os_detection,omitempty"`
	ScriptScan       bool     `json:"script_scan,omitempty"`
	ExtraOptions     []string `json:"extra_options,omitempty"`
	TimeoutSeconds   int      `json:"timeout_seconds,omitempty"`
}

func main() {
	// Define command-line flags
	serverURL := flag.String("server", "http://localhost:8081", "Scanner service URL")
	target := flag.String("target", "", "Target to scan (required)")
	ports := flag.String("ports", "1-1000", "Ports to scan")
	scanType := flag.String("type", "SYN", "Scan type (SYN, CONNECT, UDP, VERSION, SCRIPT, ALL)")
	timing := flag.Int("timing", 4, "Timing template (0-5)")
	service := flag.Bool("service", false, "Enable service detection")
	osDetection := flag.Bool("os", false, "Enable OS detection")
	script := flag.Bool("script", false, "Enable script scanning")
	timeout := flag.Int("timeout", 300, "Timeout in seconds")
	wait := flag.Bool("wait", false, "Wait for scan to complete")
	format := flag.String("format", "json", "Output format (json, text)")

	// Parse command-line flags
	flag.Parse()

	// Validate required flags
	if *target == "" {
		fmt.Println("Error: target is required")
		flag.Usage()
		os.Exit(1)
	}

	// Create scan request
	req := ScanRequest{
		Target:           *target,
		Ports:            *ports,
		ScanType:         *scanType,
		TimingTemplate:   *timing,
		ServiceDetection: *service,
		OSDetection:      *osDetection,
		ScriptScan:       *script,
		TimeoutSeconds:   *timeout,
	}

	// Start scan
	scanID, err := startScan(*serverURL, req)
	if err != nil {
		fmt.Printf("Error starting scan: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Scan started with ID: %s\n", scanID)

	// Wait for scan to complete if requested
	if *wait {
		fmt.Println("Waiting for scan to complete...")
		for {
			scan, err := getScan(*serverURL, scanID)
			if err != nil {
				fmt.Printf("Error getting scan status: %v\n", err)
				os.Exit(1)
			}

			status := scan["status"].(string)
			fmt.Printf("Scan status: %s\n", status)

			if status == "COMPLETED" || status == "FAILED" || status == "CANCELLED" {
				break
			}

			time.Sleep(5 * time.Second)
		}

		// Get and print scan result
		if *format == "json" {
			printScanResultJSON(*serverURL, scanID)
		} else {
			printScanResultText(*serverURL, scanID)
		}
	}
}

// startScan starts a scan and returns the scan ID
func startScan(serverURL string, req ScanRequest) (string, error) {
	// Marshal request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request to server
	resp, err := http.Post(serverURL+"/api/v1/scans", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Get scan ID
	scanID, ok := result["scan_id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response format, scan_id not found")
	}

	return scanID, nil
}

// getScan gets a scan by ID
func getScan(serverURL string, scanID string) (map[string]interface{}, error) {
	// Send request to server
	resp, err := http.Get(serverURL + "/api/v1/scans/" + scanID)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}

// printScanResultJSON prints the scan result in JSON format
func printScanResultJSON(serverURL string, scanID string) {
	scan, err := getScan(serverURL, scanID)
	if err != nil {
		fmt.Printf("Error getting scan: %v\n", err)
		return
	}

	resultID, ok := scan["result_id"].(string)
	if !ok || resultID == "" {
		fmt.Println("No result available for this scan")
		return
	}

	// Get scan result
	resp, err := http.Get(serverURL + "/api/v1/results/" + resultID)
	if err != nil {
		fmt.Printf("Error getting scan result: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected status code: %d, body: %s\n", resp.StatusCode, string(body))
		return
	}

	// Parse and pretty-print JSON
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	prettyJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Printf("Error formatting JSON: %v\n", err)
		return
	}

	fmt.Println(string(prettyJSON))
}

// printScanResultText prints the scan result in a human-readable format
func printScanResultText(serverURL string, scanID string) {
	scan, err := getScan(serverURL, scanID)
	if err != nil {
		fmt.Printf("Error getting scan: %v\n", err)
		return
	}

	resultID, ok := scan["result_id"].(string)
	if !ok || resultID == "" {
		fmt.Println("No result available for this scan")
		return
	}

	// Get scan result
	resp, err := http.Get(serverURL + "/api/v1/results/" + resultID)
	if err != nil {
		fmt.Printf("Error getting scan result: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected status code: %d, body: %s\n", resp.StatusCode, string(body))
		return
	}

	// Parse result
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		return
	}

	// Print scan summary
	fmt.Println("=== Scan Summary ===")
	fmt.Printf("Scan ID: %s\n", scanID)
	fmt.Printf("Target: %s\n", scan["options"].(map[string]interface{})["target"])
	fmt.Printf("Start Time: %s\n", result["start_time"])
	fmt.Printf("End Time: %s\n", result["end_time"])
	fmt.Printf("Duration: %.2f seconds\n", result["duration"])
	fmt.Printf("Total Hosts: %d\n", int(result["total_hosts"].(float64)))
	fmt.Printf("Up Hosts: %d\n", int(result["up_hosts"].(float64)))
	fmt.Println()

	// Print hosts
	hosts, ok := result["hosts"].([]interface{})
	if !ok {
		fmt.Println("No hosts found")
		return
	}

	fmt.Printf("=== Hosts (%d) ===\n", len(hosts))
	for i, hostInterface := range hosts {
		host := hostInterface.(map[string]interface{})
		fmt.Printf("Host %d: %s\n", i+1, host["ip"])

		// Print hostnames
		hostnames, ok := host["hostnames"].([]interface{})
		if ok && len(hostnames) > 0 {
			fmt.Printf("  Hostnames: ")
			for j, hostname := range hostnames {
				if j > 0 {
					fmt.Print(", ")
				}
				fmt.Print(hostname)
			}
			fmt.Println()
		}

		// Print OS
		if os, ok := host["os"].(string); ok && os != "" {
			fmt.Printf("  OS: %s\n", os)
		}

		// Print ports
		ports, ok := host["ports"].([]interface{})
		if ok && len(ports) > 0 {
			fmt.Printf("  Open Ports (%d):\n", len(ports))
			for _, portInterface := range ports {
				port := portInterface.(map[string]interface{})
				fmt.Printf("    %s/%d: %s", port["protocol"], int(port["port"].(float64)), port["service"])

				if product, ok := port["product"].(string); ok && product != "" {
					fmt.Printf(" (%s", product)
					if version, ok := port["version"].(string); ok && version != "" {
						fmt.Printf(" %s", version)
					}
					fmt.Print(")")
				}
				fmt.Println()
			}
		} else {
			fmt.Println("  No open ports found")
		}

		fmt.Println()
	}
}

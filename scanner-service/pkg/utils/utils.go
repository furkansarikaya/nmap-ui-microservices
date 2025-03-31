package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os/exec"
	"strings"
)

// GenerateID generates a random ID
func GenerateID(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CheckPortStatus checks if a port is open on a host
func CheckPortStatus(host string, port int) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, 5*1000*1000*1000) // 5 seconds
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// IsNmapInstalled checks if nmap is installed
func IsNmapInstalled(nmapPath string) bool {
	path := nmapPath
	if path == "" {
		path = "nmap"
	}

	cmd := exec.Command(path, "--version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// GetNmapVersion returns the nmap version
func GetNmapVersion(nmapPath string) (string, error) {
	path := nmapPath
	if path == "" {
		path = "nmap"
	}

	output, err := exec.Command(path, "--version").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get nmap version: %w", err)
	}

	versionLine := strings.Split(string(output), "\n")[0]
	return versionLine, nil
}

// SanitizeInput sanitizes user input to prevent command injection
func SanitizeInput(input string) string {
	// Replace characters that could be used for command injection
	replacer := strings.NewReplacer(
		"`", "",
		"$", "",
		"&", "",
		"|", "",
		";", "",
		"<", "",
		">", "",
		"(", "",
		")", "",
		"\"", "",
		"'", "",
		"\n", "",
		"\r", "",
	)
	return replacer.Replace(input)
}

// ValidateIPAddress validates an IP address or CIDR range
func ValidateIPAddress(ip string) bool {
	// Check if it's a CIDR range
	if strings.Contains(ip, "/") {
		_, ipNet, err := net.ParseCIDR(ip)
		return err == nil && ipNet != nil
	}

	// Check if it's a valid IP address
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil
}

// ValidateHostname validates a hostname
func ValidateHostname(hostname string) bool {
	// Simple validation, check for invalid characters
	if strings.ContainsAny(hostname, " \t\n\r") {
		return false
	}

	// Try to resolve it
	_, err := net.ResolveIPAddr("ip", hostname)
	return err == nil
}

// PortRangeToSlice converts a port range string to a slice of integers
// Example: "1-100,200,300-400" -> [1,2,...,100,200,300,...,400]
func PortRangeToSlice(portRange string) ([]int, error) {
	var ports []int

	// Split by comma
	ranges := strings.Split(portRange, ",")

	for _, r := range ranges {
		// Check if it's a range or a single port
		if strings.Contains(r, "-") {
			// It's a range
			parts := strings.Split(r, "-")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid port range: %s", r)
			}

			// Parse start and end
			var start, end int
			if _, err := fmt.Sscanf(parts[0], "%d", &start); err != nil {
				return nil, fmt.Errorf("invalid start port: %s", parts[0])
			}
			if _, err := fmt.Sscanf(parts[1], "%d", &end); err != nil {
				return nil, fmt.Errorf("invalid end port: %s", parts[1])
			}

			// Validate range
			if start > end {
				return nil, fmt.Errorf("start port greater than end port: %d > %d", start, end)
			}
			if start < 1 || end > 65535 {
				return nil, fmt.Errorf("port range out of bounds (1-65535): %d-%d", start, end)
			}

			// Add ports to slice
			for port := start; port <= end; port++ {
				ports = append(ports, port)
			}
		} else {
			// It's a single port
			var port int
			if _, err := fmt.Sscanf(r, "%d", &port); err != nil {
				return nil, fmt.Errorf("invalid port: %s", r)
			}

			// Validate port
			if port < 1 || port > 65535 {
				return nil, fmt.Errorf("port out of bounds (1-65535): %d", port)
			}

			// Add port to slice
			ports = append(ports, port)
		}
	}

	return ports, nil
}

package domain

import (
	"time"
)

// ScanStatus represents the status of a scan
type ScanStatus string

// Scan status constants
const (
	ScanStatusPending   ScanStatus = "PENDING"
	ScanStatusRunning   ScanStatus = "RUNNING"
	ScanStatusCompleted ScanStatus = "COMPLETED"
	ScanStatusFailed    ScanStatus = "FAILED"
	ScanStatusCancelled ScanStatus = "CANCELLED"
)

// ScanType represents the type of a scan
type ScanType string

// Scan type constants
const (
	ScanTypeSYN     ScanType = "SYN"     // -sS: TCP SYN scan
	ScanTypeConnect ScanType = "CONNECT" // -sT: TCP connect scan
	ScanTypeUDP     ScanType = "UDP"     // -sU: UDP scan
	ScanTypeVersion ScanType = "VERSION" // -sV: Version detection
	ScanTypeScript  ScanType = "SCRIPT"  // -sC: Script scan
	ScanTypeAll     ScanType = "ALL"     // -A: Aggressive scan (-sV -sC -O)
)

// TimingTemplate represents the timing template for a scan
type TimingTemplate int

// Timing template constants
const (
	TimingParanoid   TimingTemplate = 0 // -T0: Paranoid timing
	TimingSneaky     TimingTemplate = 1 // -T1: Sneaky timing
	TimingPolite     TimingTemplate = 2 // -T2: Polite timing
	TimingNormal     TimingTemplate = 3 // -T3: Normal timing
	TimingAggressive TimingTemplate = 4 // -T4: Aggressive timing
	TimingInsane     TimingTemplate = 5 // -T5: Insane timing
)

// ScanOptions represents the options for a scan
type ScanOptions struct {
	Target           string         `json:"target"`            // Target host(s) or network
	Ports            string         `json:"ports"`             // Port specification (e.g., "22,80,443" or "1-1000")
	ScanType         ScanType       `json:"scan_type"`         // Type of scan
	TimingTemplate   TimingTemplate `json:"timing_template"`   // Timing template
	ServiceDetection bool           `json:"service_detection"` // Enable service/version detection
	OSDetection      bool           `json:"os_detection"`      // Enable OS detection
	ScriptScan       bool           `json:"script_scan"`       // Enable script scanning
	ExtraOptions     []string       `json:"extra_options"`     // Extra command-line options
	Timeout          time.Duration  `json:"timeout"`           // Scan timeout
}

// Scan represents a scan job
type Scan struct {
	ID          string      `json:"id"`           // Unique identifier
	UserID      string      `json:"user_id"`      // User who initiated the scan
	Options     ScanOptions `json:"options"`      // Scan options
	Status      ScanStatus  `json:"status"`       // Current status
	Progress    float64     `json:"progress"`     // Progress percentage (0-100)
	CreatedAt   time.Time   `json:"created_at"`   // When the scan was created
	StartedAt   *time.Time  `json:"started_at"`   // When the scan started
	CompletedAt *time.Time  `json:"completed_at"` // When the scan completed
	Error       string      `json:"error"`        // Error message if failed
	ResultID    string      `json:"result_id"`    // Reference to scan result
}

// Host represents a host from a scan result
type Host struct {
	IP        string       `json:"ip"`        // IP address
	Hostnames []string     `json:"hostnames"` // Hostnames
	Status    string       `json:"status"`    // Host status (up/down)
	OS        string       `json:"os"`        // Operating system
	Ports     []Port       `json:"ports"`     // Open ports
	Scripts   []Script     `json:"scripts"`   // Script results
	Metadata  HostMetadata `json:"metadata"`  // Additional metadata
}

// Port represents a port from a scan result
type Port struct {
	Port      int    `json:"port"`       // Port number
	Protocol  string `json:"protocol"`   // Protocol (tcp/udp)
	State     string `json:"state"`      // Port state (open/closed/filtered)
	Service   string `json:"service"`    // Service name
	Product   string `json:"product"`    // Product name
	Version   string `json:"version"`    // Version information
	ExtraInfo string `json:"extra_info"` // Extra information
}

// Script represents a script result from a scan
type Script struct {
	ID     string            `json:"id"`     // Script ID
	Output string            `json:"output"` // Script output
	Data   map[string]string `json:"data"`   // Structured data
}

// HostMetadata contains additional information about a host
type HostMetadata struct {
	Distance     int       `json:"distance"`       // Network distance (TTL)
	UpTime       float64   `json:"uptime"`         // System uptime in seconds
	LastBoot     time.Time `json:"last_boot"`      // Last boot time
	TCPSequence  string    `json:"tcp_sequence"`   // TCP sequence prediction
	IPIDSequence string    `json:"ip_id_sequence"` // IP ID sequence generation
}

// ScanResult represents the result of a scan
type ScanResult struct {
	ID         string    `json:"id"`          // Unique identifier
	ScanID     string    `json:"scan_id"`     // Reference to scan
	UserID     string    `json:"user_id"`     // User who initiated the scan
	StartTime  time.Time `json:"start_time"`  // When the scan started
	EndTime    time.Time `json:"end_time"`    // When the scan ended
	Duration   float64   `json:"duration"`    // Duration in seconds
	Command    string    `json:"command"`     // Command that was run
	Summary    string    `json:"summary"`     // Scan summary
	TotalHosts int       `json:"total_hosts"` // Total hosts scanned
	UpHosts    int       `json:"up_hosts"`    // Hosts that were up
	Hosts      []Host    `json:"hosts"`       // Host results
}

// ScanSummary represents a summary of a scan
type ScanSummary struct {
	ID         string     `json:"id"`          // Unique identifier
	UserID     string     `json:"user_id"`     // User who initiated the scan
	Target     string     `json:"target"`      // Target that was scanned
	Status     ScanStatus `json:"status"`      // Current status
	StartTime  *time.Time `json:"start_time"`  // When the scan started
	EndTime    *time.Time `json:"end_time"`    // When the scan ended
	Duration   float64    `json:"duration"`    // Duration in seconds
	TotalHosts int        `json:"total_hosts"` // Total hosts scanned
	UpHosts    int        `json:"up_hosts"`    // Hosts that were up
	OpenPorts  int        `json:"open_ports"`  // Total open ports found
	VulnCount  int        `json:"vuln_count"`  // Number of vulnerabilities found
	HasResults bool       `json:"has_results"` // Whether the scan has results
}

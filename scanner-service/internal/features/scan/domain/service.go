package domain

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/furkansarikaya/nmap-ui-microservices/scanner-service/pkg/errors"
	"github.com/furkansarikaya/nmap-ui-microservices/scanner-service/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ScanAdapter defines the interface for nmap adapter
type ScanAdapter interface {
	ExecuteScan(ctx context.Context, options ScanOptions) (*ScanResult, error)
	GetVersion() (string, error)
	IsAvailable() bool
}

// ScanRepository defines the interface for scan repository
type ScanRepository interface {
	SaveScan(scan *Scan) error
	UpdateScan(scan *Scan) error
	GetScanByID(id string) (*Scan, error)
	ListScans(userID string, limit, offset int) ([]*Scan, error)
	DeleteScan(id string) error
	SaveScanResult(result *ScanResult) error
	GetScanResultByID(id string) (*ScanResult, error)
	DeleteScanResult(id string) error
}

// ScanService handles scan operations
type ScanService struct {
	adapter            ScanAdapter
	repository         ScanRepository
	logger             *logger.Logger
	maxConcurrentScans int
	activeScans        map[string]*Scan
	mu                 sync.Mutex
}

// NewScanService creates a new ScanService
func NewScanService(adapter ScanAdapter, repository ScanRepository, logger *logger.Logger, maxConcurrentScans int) *ScanService {
	return &ScanService{
		adapter:            adapter,
		repository:         repository,
		logger:             logger,
		maxConcurrentScans: maxConcurrentScans,
		activeScans:        make(map[string]*Scan),
	}
}

// StartScan starts a new scan
func (s *ScanService) StartScan(ctx context.Context, userID string, options ScanOptions) (*Scan, error) {
	// Validate options
	if err := s.validateScanOptions(options); err != nil {
		return nil, err
	}

	// Check if we can run more scans
	s.mu.Lock()
	if len(s.activeScans) >= s.maxConcurrentScans {
		s.mu.Unlock()
		return nil, errors.NewUnavailable("maximum concurrent scans reached", nil)
	}

	// Create scan
	now := time.Now()
	scan := &Scan{
		ID:        uuid.New().String(),
		UserID:    userID,
		Options:   options,
		Status:    ScanStatusPending,
		Progress:  0,
		CreatedAt: now,
	}

	// Add to active scans
	s.activeScans[scan.ID] = scan
	s.mu.Unlock()

	// Save to repository
	if err := s.repository.SaveScan(scan); err != nil {
		s.mu.Lock()
		delete(s.activeScans, scan.ID)
		s.mu.Unlock()
		return nil, errors.NewInternal("failed to save scan", err)
	}

	// Start scan in a goroutine
	go s.executeScan(ctx, scan)

	return scan, nil
}

// GetScan gets a scan by ID
func (s *ScanService) GetScan(id string) (*Scan, error) {
	// Check active scans first
	s.mu.Lock()
	if scan, ok := s.activeScans[id]; ok {
		s.mu.Unlock()
		return scan, nil
	}
	s.mu.Unlock()

	// Check repository
	scan, err := s.repository.GetScanByID(id)
	if err != nil {
		return nil, errors.NewNotFound("scan not found", err)
	}

	return scan, nil
}

// ListScans lists scans for a user
func (s *ScanService) ListScans(userID string, limit, offset int) ([]*Scan, error) {
	scans, err := s.repository.ListScans(userID, limit, offset)
	if err != nil {
		return nil, errors.NewInternal("failed to list scans", err)
	}

	return scans, nil
}

// CancelScan cancels a running scan
func (s *ScanService) CancelScan(id string) error {
	// Get scan
	scan, err := s.GetScan(id)
	if err != nil {
		return err
	}

	// Check if scan is running
	if scan.Status != ScanStatusRunning && scan.Status != ScanStatusPending {
		return errors.NewInvalidInput("scan is not running or pending", nil)
	}

	// Update scan status
	scan.Status = ScanStatusCancelled
	now := time.Now()
	scan.CompletedAt = &now

	// Update in repository
	if err := s.repository.UpdateScan(scan); err != nil {
		return errors.NewInternal("failed to update scan", err)
	}

	// Remove from active scans
	s.mu.Lock()
	delete(s.activeScans, id)
	s.mu.Unlock()

	return nil
}

// GetScanResult gets a scan result by ID
func (s *ScanService) GetScanResult(id string) (*ScanResult, error) {
	result, err := s.repository.GetScanResultByID(id)
	if err != nil {
		return nil, errors.NewNotFound("scan result not found", err)
	}

	return result, nil
}

// ValidateNmap validates nmap installation
func (s *ScanService) ValidateNmap() error {
	if !s.adapter.IsAvailable() {
		return errors.NewUnavailable("nmap is not available", nil)
	}

	return nil
}

// GetNmapVersion gets nmap version
func (s *ScanService) GetNmapVersion() (string, error) {
	version, err := s.adapter.GetVersion()
	if err != nil {
		return "", errors.NewUnavailable("failed to get nmap version", err)
	}

	return version, nil
}

// executeScan executes a scan
func (s *ScanService) executeScan(ctx context.Context, scan *Scan) {
	// Create a cancellable context
	ctx, cancel := context.WithTimeout(ctx, scan.Options.Timeout)
	defer cancel()

	// Update scan status
	now := time.Now()
	scan.Status = ScanStatusRunning
	scan.StartedAt = &now
	scan.Progress = 0

	// Update in repository
	if err := s.repository.UpdateScan(scan); err != nil {
		s.logger.Error("Failed to update scan status",
			zap.String("scan_id", scan.ID),
			zap.Error(err),
		)
	}

	// Execute scan
	s.logger.Info("Starting scan",
		zap.String("scan_id", scan.ID),
		zap.String("target", scan.Options.Target),
	)

	result, err := s.adapter.ExecuteScan(ctx, scan.Options)

	// Update scan status and result
	if err != nil {
		s.logger.Error("Scan failed",
			zap.String("scan_id", scan.ID),
			zap.Error(err),
		)

		scan.Status = ScanStatusFailed
		scan.Error = err.Error()
	} else {
		s.logger.Info("Scan completed",
			zap.String("scan_id", scan.ID),
			zap.Int("total_hosts", result.TotalHosts),
			zap.Int("up_hosts", result.UpHosts),
		)

		scan.Status = ScanStatusCompleted
		scan.Progress = 100
		scan.ResultID = result.ID

		// Set scan ID in result
		result.ScanID = scan.ID
		result.UserID = scan.UserID

		// Save scan result
		if err := s.repository.SaveScanResult(result); err != nil {
			s.logger.Error("Failed to save scan result",
				zap.String("scan_id", scan.ID),
				zap.Error(err),
			)
		}
	}

	// Set completion time
	completedAt := time.Now()
	scan.CompletedAt = &completedAt

	// Update in repository
	if err := s.repository.UpdateScan(scan); err != nil {
		s.logger.Error("Failed to update scan status",
			zap.String("scan_id", scan.ID),
			zap.Error(err),
		)
	}

	// Remove from active scans
	s.mu.Lock()
	delete(s.activeScans, scan.ID)
	s.mu.Unlock()
}

// validateScanOptions validates scan options
func (s *ScanService) validateScanOptions(options ScanOptions) error {
	// Validate target
	if options.Target == "" {
		return errors.NewInvalidInput("target is required", nil)
	}

	// Validate timeout
	if options.Timeout == 0 {
		options.Timeout = 5 * time.Minute // Default timeout
	}

	// Validate ports
	if options.Ports == "" {
		options.Ports = "1-1000" // Default ports
	}

	// Validate timing template
	if options.TimingTemplate < TimingParanoid || options.TimingTemplate > TimingInsane {
		options.TimingTemplate = TimingNormal // Default timing template
	}

	return nil
}

// CreateScanSummary creates a scan summary from a scan and its result
func (s *ScanService) CreateScanSummary(scan *Scan, result *ScanResult) *ScanSummary {
	summary := &ScanSummary{
		ID:         scan.ID,
		UserID:     scan.UserID,
		Target:     scan.Options.Target,
		Status:     scan.Status,
		StartTime:  scan.StartedAt,
		EndTime:    scan.CompletedAt,
		HasResults: result != nil,
	}

	if scan.StartedAt != nil && scan.CompletedAt != nil {
		summary.Duration = scan.CompletedAt.Sub(*scan.StartedAt).Seconds()
	}

	if result != nil {
		summary.TotalHosts = result.TotalHosts
		summary.UpHosts = result.UpHosts

		// Count open ports
		for _, host := range result.Hosts {
			for _, port := range host.Ports {
				if port.State == "open" {
					summary.OpenPorts++
				}
			}
		}

		// Count vulnerabilities (example: count script results that contain "VULNERABLE")
		for _, host := range result.Hosts {
			for _, script := range host.Scripts {
				if strings.Contains(script.Output, "VULNERABLE") {
					summary.VulnCount++
				}
			}
		}
	}

	return summary
}

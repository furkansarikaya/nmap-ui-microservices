package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/furkansarikaya/nmap-ui-microservices/scanner-service/internal/features/scan/domain"
	"github.com/furkansarikaya/nmap-ui-microservices/scanner-service/pkg/errors"
	"github.com/furkansarikaya/nmap-ui-microservices/scanner-service/pkg/logger"
	"go.uber.org/zap"
)

// MemoryScanRepository is an in-memory implementation of the ScanRepository interface
type MemoryScanRepository struct {
	logger          *logger.Logger
	scans           map[string]*domain.Scan
	scanResults     map[string]*domain.ScanResult
	mu              sync.RWMutex
	retentionPeriod time.Duration
}

// NewMemoryScanRepository creates a new MemoryScanRepository
func NewMemoryScanRepository(logger *logger.Logger, retentionPeriod time.Duration) *MemoryScanRepository {
	repo := &MemoryScanRepository{
		logger:          logger,
		scans:           make(map[string]*domain.Scan),
		scanResults:     make(map[string]*domain.ScanResult),
		retentionPeriod: retentionPeriod,
	}

	// Start cleanup goroutine
	go repo.cleanupOldScans()

	return repo
}

// SaveScan saves a scan to the repository
func (r *MemoryScanRepository) SaveScan(scan *domain.Scan) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Make a deep copy to avoid modifying the original
	scanCopy := *scan
	r.scans[scan.ID] = &scanCopy

	r.logger.Debug("Saved scan",
		zap.String("scan_id", scan.ID),
		zap.String("user_id", scan.UserID),
	)

	return nil
}

// UpdateScan updates a scan in the repository
func (r *MemoryScanRepository) UpdateScan(scan *domain.Scan) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.scans[scan.ID]; !ok {
		return errors.NewNotFound(fmt.Sprintf("scan with ID %s not found", scan.ID), nil)
	}

	// Make a deep copy to avoid modifying the original
	scanCopy := *scan
	r.scans[scan.ID] = &scanCopy

	r.logger.Debug("Updated scan",
		zap.String("scan_id", scan.ID),
		zap.String("status", string(scan.Status)),
	)

	return nil
}

// GetScanByID gets a scan by ID from the repository
func (r *MemoryScanRepository) GetScanByID(id string) (*domain.Scan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	scan, ok := r.scans[id]
	if !ok {
		return nil, errors.NewNotFound(fmt.Sprintf("scan with ID %s not found", id), nil)
	}

	// Return a copy to avoid modifying the original
	scanCopy := *scan
	return &scanCopy, nil
}

// ListScans lists scans from the repository
func (r *MemoryScanRepository) ListScans(userID string, limit, offset int) ([]*domain.Scan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var scans []*domain.Scan

	// Filter by user ID if provided
	for _, scan := range r.scans {
		if userID == "" || scan.UserID == userID {
			// Make a copy to avoid modifying the original
			scanCopy := *scan
			scans = append(scans, &scanCopy)
		}
	}

	// Sort by created at (newest first)
	// In a real implementation, you would use a database query with ORDER BY
	// This is just a simple implementation for the in-memory repository
	for i := 0; i < len(scans)-1; i++ {
		for j := i + 1; j < len(scans); j++ {
			if scans[i].CreatedAt.Before(scans[j].CreatedAt) {
				scans[i], scans[j] = scans[j], scans[i]
			}
		}
	}

	// Apply pagination
	if offset >= len(scans) {
		return []*domain.Scan{}, nil
	}

	end := offset + limit
	if end > len(scans) {
		end = len(scans)
	}

	return scans[offset:end], nil
}

// DeleteScan deletes a scan from the repository
func (r *MemoryScanRepository) DeleteScan(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.scans[id]; !ok {
		return errors.NewNotFound(fmt.Sprintf("scan with ID %s not found", id), nil)
	}

	delete(r.scans, id)

	r.logger.Debug("Deleted scan", zap.String("scan_id", id))

	return nil
}

// SaveScanResult saves a scan result to the repository
func (r *MemoryScanRepository) SaveScanResult(result *domain.ScanResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Make a deep copy to avoid modifying the original
	resultCopy := *result
	r.scanResults[result.ID] = &resultCopy

	r.logger.Debug("Saved scan result",
		zap.String("result_id", result.ID),
		zap.String("scan_id", result.ScanID),
	)

	return nil
}

// GetScanResultByID gets a scan result by ID from the repository
func (r *MemoryScanRepository) GetScanResultByID(id string) (*domain.ScanResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result, ok := r.scanResults[id]
	if !ok {
		return nil, errors.NewNotFound(fmt.Sprintf("scan result with ID %s not found", id), nil)
	}

	// Return a copy to avoid modifying the original
	resultCopy := *result
	return &resultCopy, nil
}

// DeleteScanResult deletes a scan result from the repository
func (r *MemoryScanRepository) DeleteScanResult(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.scanResults[id]; !ok {
		return errors.NewNotFound(fmt.Sprintf("scan result with ID %s not found", id), nil)
	}

	delete(r.scanResults, id)

	r.logger.Debug("Deleted scan result", zap.String("result_id", id))

	return nil
}

// cleanupOldScans periodically removes old scans and results
func (r *MemoryScanRepository) cleanupOldScans() {
	ticker := time.NewTicker(6 * time.Hour) // Run cleanup every 6 hours
	defer ticker.Stop()

	for range ticker.C {
		r.mu.Lock()

		cutoffTime := time.Now().Add(-r.retentionPeriod)

		// Clean up old scans
		for id, scan := range r.scans {
			if scan.CreatedAt.Before(cutoffTime) {
				// Delete scan
				delete(r.scans, id)

				// Delete associated result if exists
				if scan.ResultID != "" {
					delete(r.scanResults, scan.ResultID)
				}

				r.logger.Debug("Cleaned up old scan",
					zap.String("scan_id", id),
					zap.Time("created_at", scan.CreatedAt),
				)
			}
		}

		// Clean up orphaned results (results without a scan)
		for resultID, result := range r.scanResults {
			if result.ScanID != "" {
				if _, ok := r.scans[result.ScanID]; !ok {
					delete(r.scanResults, resultID)

					r.logger.Debug("Cleaned up orphaned scan result",
						zap.String("result_id", resultID),
						zap.String("scan_id", result.ScanID),
					)
				}
			}
		}

		r.mu.Unlock()
	}
}

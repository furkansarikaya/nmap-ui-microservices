package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/furkansarikaya/nmap-ui-microservices/scanner-service/internal/features/scan/domain"
	"github.com/furkansarikaya/nmap-ui-microservices/scanner-service/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ScanHandler handles HTTP requests for scans
type ScanHandler struct {
	scanService *domain.ScanService
	logger      *logger.Logger
}

// NewScanHandler creates a new ScanHandler
func NewScanHandler(scanService *domain.ScanService, logger *logger.Logger) *ScanHandler {
	return &ScanHandler{
		scanService: scanService,
		logger:      logger,
	}
}

// StartScanRequest represents the request body for starting a scan
type StartScanRequest struct {
	Target           string                `json:"target" binding:"required"`
	Ports            string                `json:"ports,omitempty"`
	ScanType         domain.ScanType       `json:"scan_type,omitempty"`
	TimingTemplate   domain.TimingTemplate `json:"timing_template,omitempty"`
	ServiceDetection bool                  `json:"service_detection,omitempty"`
	OSDetection      bool                  `json:"os_detection,omitempty"`
	ScriptScan       bool                  `json:"script_scan,omitempty"`
	ExtraOptions     []string              `json:"extra_options,omitempty"`
	TimeoutSeconds   int                   `json:"timeout_seconds,omitempty"`
}

// StartScan handles the request to start a scan
func (h *ScanHandler) StartScan(c *gin.Context) {
	var req StartScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	// For now, use a default user ID
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "default-user" // Will be replaced with actual auth
	}

	// Create scan options from request
	options := domain.ScanOptions{
		Target:           req.Target,
		Ports:            req.Ports,
		ScanType:         req.ScanType,
		TimingTemplate:   req.TimingTemplate,
		ServiceDetection: req.ServiceDetection,
		OSDetection:      req.OSDetection,
		ScriptScan:       req.ScriptScan,
		ExtraOptions:     req.ExtraOptions,
	}

	// Set timeout
	if req.TimeoutSeconds > 0 {
		options.Timeout = time.Duration(req.TimeoutSeconds) * time.Second
	} else {
		options.Timeout = 5 * time.Minute // Default timeout
	}

	// Start scan
	scan, err := h.scanService.StartScan(c.Request.Context(), userID, options)
	if err != nil {
		h.logger.Error("Failed to start scan",
			zap.Error(err),
			zap.String("target", req.Target),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start scan: " + err.Error(),
		})
		return
	}

	h.logger.Info("Scan started",
		zap.String("scan_id", scan.ID),
		zap.String("target", req.Target),
	)

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Scan started",
		"scan_id": scan.ID,
	})
}

// GetScan handles the request to get a scan
func (h *ScanHandler) GetScan(c *gin.Context) {
	scanID := c.Param("id")
	if scanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Scan ID is required",
		})
		return
	}

	scan, err := h.scanService.GetScan(scanID)
	if err != nil {
		h.logger.Error("Failed to get scan",
			zap.Error(err),
			zap.String("scan_id", scanID),
		)

		c.JSON(http.StatusNotFound, gin.H{
			"error": "Failed to get scan: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, scan)
}

// ListScans handles the request to list scans
func (h *ScanHandler) ListScans(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	// For now, use a default user ID
	userID := c.GetString("user_id")
	if userID == "" {
		userID = "default-user" // Will be replaced with actual auth
	}

	// Parse pagination parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Validate pagination parameters
	if limit < 1 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	scans, err := h.scanService.ListScans(userID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list scans",
			zap.Error(err),
			zap.String("user_id", userID),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list scans: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"scans":  scans,
		"limit":  limit,
		"offset": offset,
		"count":  len(scans),
	})
}

// CancelScan handles the request to cancel a scan
func (h *ScanHandler) CancelScan(c *gin.Context) {
	scanID := c.Param("id")
	if scanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Scan ID is required",
		})
		return
	}

	err := h.scanService.CancelScan(scanID)
	if err != nil {
		h.logger.Error("Failed to cancel scan",
			zap.Error(err),
			zap.String("scan_id", scanID),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to cancel scan: " + err.Error(),
		})
		return
	}

	h.logger.Info("Scan cancelled", zap.String("scan_id", scanID))

	c.JSON(http.StatusOK, gin.H{
		"message": "Scan cancelled",
		"scan_id": scanID,
	})
}

// GetScanResult handles the request to get a scan result
func (h *ScanHandler) GetScanResult(c *gin.Context) {
	resultID := c.Param("id")
	if resultID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Result ID is required",
		})
		return
	}

	result, err := h.scanService.GetScanResult(resultID)
	if err != nil {
		h.logger.Error("Failed to get scan result",
			zap.Error(err),
			zap.String("result_id", resultID),
		)

		c.JSON(http.StatusNotFound, gin.H{
			"error": "Failed to get scan result: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetHealth handles the health check endpoint
func (h *ScanHandler) GetHealth(c *gin.Context) {
	// Check nmap installation
	err := h.scanService.ValidateNmap()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "Nmap is not available: " + err.Error(),
		})
		return
	}

	// Get nmap version
	version, err := h.scanService.GetNmapVersion()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "Failed to get nmap version: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "healthy",
		"nmap_version": version,
		"timestamp":    time.Now().Format(time.RFC3339),
	})
}

// RegisterRoutes registers the scan handler routes to the router
func (h *ScanHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")

	// Scan endpoints
	api.POST("/scans", h.StartScan)
	api.GET("/scans/:id", h.GetScan)
	api.GET("/scans", h.ListScans)
	api.DELETE("/scans/:id", h.CancelScan)

	// Scan result endpoints
	api.GET("/results/:id", h.GetScanResult)

	// Health check endpoint
	router.GET("/health", h.GetHealth)
}

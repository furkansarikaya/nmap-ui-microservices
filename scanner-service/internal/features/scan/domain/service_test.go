package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/furkansarikaya/nmap-ui-microservices/scanner-service/internal/features/scan/domain"
	"github.com/furkansarikaya/nmap-ui-microservices/scanner-service/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockScanAdapter is a mock implementation of ScanAdapter
type MockScanAdapter struct {
	mock.Mock
}

func (m *MockScanAdapter) ExecuteScan(ctx context.Context, options domain.ScanOptions) (*domain.ScanResult, error) {
	args := m.Called(ctx, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ScanResult), args.Error(1)
}

func (m *MockScanAdapter) GetVersion() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockScanAdapter) IsAvailable() bool {
	args := m.Called()
	return args.Bool(0)
}

// MockScanRepository is a mock implementation of ScanRepository
type MockScanRepository struct {
	mock.Mock
}

func (m *MockScanRepository) SaveScan(scan *domain.Scan) error {
	args := m.Called(scan)
	return args.Error(0)
}

func (m *MockScanRepository) UpdateScan(scan *domain.Scan) error {
	args := m.Called(scan)
	return args.Error(0)
}

func (m *MockScanRepository) GetScanByID(id string) (*domain.Scan, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Scan), args.Error(1)
}

func (m *MockScanRepository) ListScans(userID string, limit, offset int) ([]*domain.Scan, error) {
	args := m.Called(userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Scan), args.Error(1)
}

func (m *MockScanRepository) DeleteScan(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockScanRepository) SaveScanResult(result *domain.ScanResult) error {
	args := m.Called(result)
	return args.Error(0)
}

func (m *MockScanRepository) GetScanResultByID(id string) (*domain.ScanResult, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ScanResult), args.Error(1)
}

func (m *MockScanRepository) DeleteScanResult(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestStartScan(t *testing.T) {
	// Create mocks
	mockAdapter := new(MockScanAdapter)
	mockRepository := new(MockScanRepository)

	// Create logger
	zapLogger, _ := zap.NewDevelopment()
	log := &logger.Logger{Logger: zapLogger}

	// Create service
	service := domain.NewScanService(mockAdapter, mockRepository, log, 10)

	// Test data
	userID := "test-user"
	options := domain.ScanOptions{
		Target:  "192.168.1.1",
		Ports:   "1-1000",
		Timeout: 5 * time.Minute,
	}

	// Set up expectations
	mockRepository.On("SaveScan", mock.AnythingOfType("*domain.Scan")).Return(nil)

	// Execute test
	scan, err := service.StartScan(context.Background(), userID, options)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, scan)
	assert.Equal(t, userID, scan.UserID)
	assert.Equal(t, options.Target, scan.Options.Target)
	assert.Equal(t, domain.ScanStatusPending, scan.Status)

	// Verify expectations
	mockRepository.AssertExpectations(t)
}

func TestGetScan(t *testing.T) {
	// Create mocks
	mockAdapter := new(MockScanAdapter)
	mockRepository := new(MockScanRepository)

	// Create logger
	zapLogger, _ := zap.NewDevelopment()
	log := &logger.Logger{Logger: zapLogger}

	// Create service
	service := domain.NewScanService(mockAdapter, mockRepository, log, 10)

	// Test data
	scanID := "test-scan-id"
	expectedScan := &domain.Scan{
		ID:     scanID,
		UserID: "test-user",
		Status: domain.ScanStatusCompleted,
	}

	// Set up expectations
	mockRepository.On("GetScanByID", scanID).Return(expectedScan, nil)

	// Execute test
	scan, err := service.GetScan(scanID)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedScan.ID, scan.ID)
	assert.Equal(t, expectedScan.UserID, scan.UserID)
	assert.Equal(t, expectedScan.Status, scan.Status)

	// Verify expectations
	mockRepository.AssertExpectations(t)
}

func TestGetScanNotFound(t *testing.T) {
	// Create mocks
	mockAdapter := new(MockScanAdapter)
	mockRepository := new(MockScanRepository)

	// Create logger
	zapLogger, _ := zap.NewDevelopment()
	log := &logger.Logger{Logger: zapLogger}

	// Create service
	service := domain.NewScanService(mockAdapter, mockRepository, log, 10)

	// Test data
	scanID := "non-existent-scan-id"

	// Set up expectations
	mockRepository.On("GetScanByID", scanID).Return(nil, errors.New("scan not found"))

	// Execute test
	scan, err := service.GetScan(scanID)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, scan)

	// Verify expectations
	mockRepository.AssertExpectations(t)
}

func TestCancelScan(t *testing.T) {
	// Create mocks
	mockAdapter := new(MockScanAdapter)
	mockRepository := new(MockScanRepository)

	// Create logger
	zapLogger, _ := zap.NewDevelopment()
	log := &logger.Logger{Logger: zapLogger}

	// Create service
	service := domain.NewScanService(mockAdapter, mockRepository, log, 10)

	// Test data
	scanID := "test-scan-id"
	scan := &domain.Scan{
		ID:     scanID,
		UserID: "test-user",
		Status: domain.ScanStatusRunning,
	}

	// Set up expectations
	mockRepository.On("GetScanByID", scanID).Return(scan, nil)
	mockRepository.On("UpdateScan", mock.AnythingOfType("*domain.Scan")).Return(nil)

	// Execute test
	err := service.CancelScan(scanID)

	// Assertions
	assert.NoError(t, err)

	// Verify expectations
	mockRepository.AssertExpectations(t)

	// Verify scan was updated
	assert.Equal(t, domain.ScanStatusCancelled, scan.Status)
}

func TestValidateNmap(t *testing.T) {
	// Create mocks
	mockAdapter := new(MockScanAdapter)
	mockRepository := new(MockScanRepository)

	// Create logger
	zapLogger, _ := zap.NewDevelopment()
	log := &logger.Logger{Logger: zapLogger}

	// Create service
	service := domain.NewScanService(mockAdapter, mockRepository, log, 10)

	// Test with nmap available
	mockAdapter.On("IsAvailable").Return(true).Once()
	err := service.ValidateNmap()
	assert.NoError(t, err)

	// Test with nmap unavailable
	mockAdapter.On("IsAvailable").Return(false).Once()
	err = service.ValidateNmap()
	assert.Error(t, err)

	// Verify expectations
	mockAdapter.AssertExpectations(t)
}

func TestGetNmapVersion(t *testing.T) {
	// Create mocks
	mockAdapter := new(MockScanAdapter)
	mockRepository := new(MockScanRepository)

	// Create logger
	zapLogger, _ := zap.NewDevelopment()
	log := &logger.Logger{Logger: zapLogger}

	// Create service
	service := domain.NewScanService(mockAdapter, mockRepository, log, 10)

	// Test data
	expectedVersion := "Nmap version 7.92"

	// Set up expectations
	mockAdapter.On("GetVersion").Return(expectedVersion, nil)

	// Execute test
	version, err := service.GetNmapVersion()

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedVersion, version)

	// Verify expectations
	mockAdapter.AssertExpectations(t)
}

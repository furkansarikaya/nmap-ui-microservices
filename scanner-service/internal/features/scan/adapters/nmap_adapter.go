package adapters

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/furkansarikaya/nmap-ui-microservices/scanner-service/internal/features/scan/domain"
	"github.com/furkansarikaya/nmap-ui-microservices/scanner-service/pkg/errors"
	"github.com/furkansarikaya/nmap-ui-microservices/scanner-service/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// NmapXML represents the nmap XML output structure
type NmapXML struct {
	XMLName xml.Name `xml:"nmaprun"`
	Args    string   `xml:"args,attr"`
	Start   int64    `xml:"start,attr"`
	Version string   `xml:"version,attr"`
	Hosts   []struct {
		StartTime int64 `xml:"starttime,attr"`
		EndTime   int64 `xml:"endtime,attr"`
		Status    struct {
			State string `xml:"state,attr"`
		} `xml:"status"`
		Addresses []struct {
			Addr     string `xml:"addr,attr"`
			AddrType string `xml:"addrtype,attr"`
			Vendor   string `xml:"vendor,attr,omitempty"`
		} `xml:"address"`
		Hostnames struct {
			Hostnames []struct {
				Name string `xml:"name,attr"`
				Type string `xml:"type,attr"`
			} `xml:"hostname"`
		} `xml:"hostnames"`
		Ports struct {
			Ports []struct {
				Protocol string `xml:"protocol,attr"`
				PortID   int    `xml:"portid,attr"`
				State    struct {
					State  string `xml:"state,attr"`
					Reason string `xml:"reason,attr"`
				} `xml:"state"`
				Service struct {
					Name       string `xml:"name,attr"`
					Product    string `xml:"product,attr,omitempty"`
					Version    string `xml:"version,attr,omitempty"`
					ExtraInfo  string `xml:"extrainfo,attr,omitempty"`
					Method     string `xml:"method,attr"`
					Conf       string `xml:"conf,attr"`
					DeviceType string `xml:"devicetype,attr,omitempty"`
				} `xml:"service"`
				Scripts []struct {
					ID     string `xml:"id,attr"`
					Output string `xml:"output,attr"`
				} `xml:"script"`
			} `xml:"port"`
		} `xml:"ports"`
		OS struct {
			Matches []struct {
				Name     string `xml:"name,attr"`
				Accuracy string `xml:"accuracy,attr"`
			} `xml:"osmatch"`
		} `xml:"os"`
		Uptime struct {
			Seconds  string `xml:"seconds,attr"`
			LastBoot string `xml:"lastboot,attr,omitempty"`
		} `xml:"uptime"`
		Distance struct {
			Value string `xml:"value,attr"`
		} `xml:"distance"`
		TCPSequence struct {
			Index      string `xml:"index,attr"`
			Difficulty string `xml:"difficulty,attr"`
		} `xml:"tcpsequence"`
		IPIDSequence struct {
			Class string `xml:"class,attr"`
		} `xml:"ipidsequence"`
	} `xml:"host"`
	RunStats struct {
		Finished struct {
			Time    int64   `xml:"time,attr"`
			TimeStr string  `xml:"timestr,attr"`
			Elapsed float64 `xml:"elapsed,attr"`
			Summary string  `xml:"summary,attr"`
			Exit    string  `xml:"exit,attr"`
		} `xml:"finished"`
		Hosts struct {
			Up    int `xml:"up,attr"`
			Down  int `xml:"down,attr"`
			Total int `xml:"total,attr"`
		} `xml:"hosts"`
	} `xml:"runstats"`
}

// NmapAdapter is an adapter for nmap
type NmapAdapter struct {
	nmapPath string
	logger   *logger.Logger
}

// NewNmapAdapter creates a new NmapAdapter
func NewNmapAdapter(nmapPath string, logger *logger.Logger) *NmapAdapter {
	if nmapPath == "" {
		nmapPath = "nmap" // Use PATH by default
	}

	return &NmapAdapter{
		nmapPath: nmapPath,
		logger:   logger,
	}
}

// ExecuteScan executes an nmap scan with the given options
func (a *NmapAdapter) ExecuteScan(ctx context.Context, scanOptions domain.ScanOptions) (*domain.ScanResult, error) {
	startTime := time.Now()

	// Build nmap command
	args := a.buildCommandArgs(scanOptions)

	a.logger.Info("Executing nmap scan",
		zap.String("target", scanOptions.Target),
		zap.Strings("args", args),
	)

	// Create a temporary file for XML output
	tmpFile, err := os.CreateTemp("", "nmap-scan-*.xml")
	if err != nil {
		return nil, errors.NewInternal("failed to create temporary file", err)
	}
	tmpFileName := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpFileName)

	// Add XML output to args
	args = append(args, "-oX", tmpFileName)

	// Create command
	cmd := exec.CommandContext(ctx, a.nmapPath, args...)

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command
	if err := cmd.Run(); err != nil {
		// Check for context cancellation
		if ctx.Err() == context.Canceled {
			return nil, errors.NewTimeout("scan was cancelled", ctx.Err())
		}

		// Check for context timeout
		if ctx.Err() == context.DeadlineExceeded {
			return nil, errors.NewTimeout("scan timed out", ctx.Err())
		}

		a.logger.Error("Nmap scan failed",
			zap.Error(err),
			zap.String("stderr", stderr.String()),
		)

		return nil, errors.NewInternal("nmap scan failed", err)
	}

	// Read XML output
	xmlData, err := os.ReadFile(tmpFileName)
	if err != nil {
		return nil, errors.NewInternal("failed to read nmap output", err)
	}

	// Parse XML
	var nmapXML NmapXML
	if err := xml.Unmarshal(xmlData, &nmapXML); err != nil {
		return nil, errors.NewInternal("failed to parse nmap output", err)
	}

	// Convert to domain model
	result := a.convertToDomainModel(nmapXML, startTime)

	// Set scan ID and command
	result.ID = uuid.New().String()
	result.Command = a.nmapPath + " " + strings.Join(args, " ")

	a.logger.Info("Nmap scan completed",
		zap.String("target", scanOptions.Target),
		zap.Int("total_hosts", result.TotalHosts),
		zap.Int("up_hosts", result.UpHosts),
		zap.Int("host_count", len(result.Hosts)),
		zap.Float64("duration", result.Duration),
	)

	return result, nil
}

// buildCommandArgs builds nmap command arguments from scan options
func (a *NmapAdapter) buildCommandArgs(options domain.ScanOptions) []string {
	var args []string

	// Add target
	args = append(args, options.Target)

	// Add ports
	if options.Ports != "" {
		args = append(args, "-p", options.Ports)
	}

	// Add scan type
	switch options.ScanType {
	case domain.ScanTypeSYN:
		args = append(args, "-sS")
	case domain.ScanTypeConnect:
		args = append(args, "-sT")
	case domain.ScanTypeUDP:
		args = append(args, "-sU")
	case domain.ScanTypeVersion:
		args = append(args, "-sV")
	case domain.ScanTypeScript:
		args = append(args, "-sC")
	case domain.ScanTypeAll:
		args = append(args, "-A")
	}

	// Add timing template
	if options.TimingTemplate >= domain.TimingParanoid && options.TimingTemplate <= domain.TimingInsane {
		args = append(args, fmt.Sprintf("-T%d", options.TimingTemplate))
	}

	// Add service detection
	if options.ServiceDetection {
		args = append(args, "-sV")
	}

	// Add OS detection
	if options.OSDetection {
		args = append(args, "-O")
	}

	// Add script scan
	if options.ScriptScan {
		args = append(args, "-sC")
	}

	// Add extra options
	args = append(args, options.ExtraOptions...)

	return args
}

// convertToDomainModel converts NmapXML to domain.ScanResult
func (a *NmapAdapter) convertToDomainModel(nmapXML NmapXML, startTime time.Time) *domain.ScanResult {
	endTime := time.Unix(nmapXML.RunStats.Finished.Time, 0)

	result := &domain.ScanResult{
		StartTime:  startTime,
		EndTime:    endTime,
		Duration:   nmapXML.RunStats.Finished.Elapsed,
		Summary:    nmapXML.RunStats.Finished.Summary,
		TotalHosts: nmapXML.RunStats.Hosts.Total,
		UpHosts:    nmapXML.RunStats.Hosts.Up,
		Hosts:      make([]domain.Host, 0),
	}

	// Process hosts
	for _, xmlHost := range nmapXML.Hosts {
		// Skip hosts that are down
		if xmlHost.Status.State != "up" {
			continue
		}

		host := domain.Host{
			Status:    xmlHost.Status.State,
			Hostnames: make([]string, 0),
			Ports:     make([]domain.Port, 0),
			Scripts:   make([]domain.Script, 0),
			Metadata:  domain.HostMetadata{},
		}

		// Get IP address
		for _, addr := range xmlHost.Addresses {
			if addr.AddrType == "ipv4" {
				host.IP = addr.Addr
				break
			}
		}

		// Get hostnames
		for _, hostname := range xmlHost.Hostnames.Hostnames {
			host.Hostnames = append(host.Hostnames, hostname.Name)
		}

		// Get OS
		if len(xmlHost.OS.Matches) > 0 {
			host.OS = xmlHost.OS.Matches[0].Name
		}

		// Get ports
		for _, xmlPort := range xmlHost.Ports.Ports {
			port := domain.Port{
				Port:      xmlPort.PortID,
				Protocol:  xmlPort.Protocol,
				State:     xmlPort.State.State,
				Service:   xmlPort.Service.Name,
				Product:   xmlPort.Service.Product,
				Version:   xmlPort.Service.Version,
				ExtraInfo: xmlPort.Service.ExtraInfo,
			}

			// Get script results
			for _, xmlScript := range xmlPort.Scripts {
				script := domain.Script{
					ID:     xmlScript.ID,
					Output: xmlScript.Output,
					Data:   make(map[string]string),
				}

				host.Scripts = append(host.Scripts, script)
			}

			host.Ports = append(host.Ports, port)
		}

		// Get metadata
		if xmlHost.Distance.Value != "" {
			distance, _ := strconv.Atoi(xmlHost.Distance.Value)
			host.Metadata.Distance = distance
		}

		if xmlHost.Uptime.Seconds != "" {
			uptime, _ := strconv.ParseFloat(xmlHost.Uptime.Seconds, 64)
			host.Metadata.UpTime = uptime
		}

		if xmlHost.Uptime.LastBoot != "" {
			// Parse last boot time if available
			host.Metadata.LastBoot, _ = time.Parse("2006-01-02 15:04:05", xmlHost.Uptime.LastBoot)
		}

		if xmlHost.TCPSequence.Difficulty != "" {
			host.Metadata.TCPSequence = xmlHost.TCPSequence.Difficulty
		}

		if xmlHost.IPIDSequence.Class != "" {
			host.Metadata.IPIDSequence = xmlHost.IPIDSequence.Class
		}

		result.Hosts = append(result.Hosts, host)
	}

	return result
}

// GetVersion returns the nmap version
func (a *NmapAdapter) GetVersion() (string, error) {
	cmd := exec.Command(a.nmapPath, "--version")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", errors.NewUnavailable("failed to get nmap version", err)
	}

	version := strings.Split(out.String(), "\n")[0]
	return version, nil
}

// IsAvailable checks if nmap is available
func (a *NmapAdapter) IsAvailable() bool {
	_, err := a.GetVersion()
	return err == nil
}

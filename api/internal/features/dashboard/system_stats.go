package dashboard

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

const (
	bytesInMB = 1024 * 1024
	bytesInGB = 1024 * 1024 * 1024
)

// CommandExecutor is a function type for executing shell commands
type CommandExecutor func(cmd string) (string, error)

// GetSystemStatsOptions contains options for getting system stats
type GetSystemStatsOptions struct {
	CommandExecutor CommandExecutor // Optional: if nil, uses local exec.Command
	SystemReader    SystemReader    // Optional: if nil, creates appropriate reader based on CommandExecutor
}

// CollectSystemStats retrieves system statistics. Can be used by DashboardMonitor or MCP tools.
func CollectSystemStats(
	l logger.Logger,
	opts GetSystemStatsOptions,
) (SystemStats, error) {
	// Create command executor for basic system info commands
	cmdExecutor := opts.CommandExecutor
	if cmdExecutor == nil {
		cmdExecutor = func(cmd string) (string, error) {
			output, err := exec.Command("sh", "-c", cmd).Output()
			if err != nil {
				return "", fmt.Errorf("command failed: %w", err)
			}
			return strings.TrimSpace(string(output)), nil
		}
	}

	// Always use remote system reader (uses command executor)
	var systemReader SystemReader
	if opts.SystemReader != nil {
		systemReader = opts.SystemReader
	} else {
		systemReader = NewRemoteSystemReader(cmdExecutor, l)
	}

	osType, err := cmdExecutor("uname -s")
	if err != nil {
		l.Log(logger.Error, err.Error(), "")
		return SystemStats{}, err
	}
	osType = strings.TrimSpace(osType)

	stats := SystemStats{
		OSType:    osType,
		Timestamp: time.Now(),
		CPU:       CPUStats{PerCore: []CPUCore{}},
		Memory:    MemoryStats{},
		Load:      LoadStats{},
		Disk:      DiskStats{AllMounts: []DiskMount{}},
		Network:   NetworkStats{Interfaces: []NetworkInterface{}},
	}

	if hostname, err := cmdExecutor("hostname"); err == nil {
		stats.Hostname = strings.TrimSpace(hostname)
	}

	if kernelVersion, err := cmdExecutor("uname -r"); err == nil {
		stats.KernelVersion = strings.TrimSpace(kernelVersion)
	}

	if architecture, err := cmdExecutor("uname -m"); err == nil {
		stats.Architecture = strings.TrimSpace(architecture)
	}

	var uptime string
	if hostInfo, err := systemReader.HostInfo(); err == nil {
		uptime = time.Duration(hostInfo.Uptime * uint64(time.Second)).String()
	}

	if loadAvg, err := cmdExecutor("uptime"); err == nil {
		loadAvgStr := strings.TrimSpace(loadAvg)
		stats.Load = parseLoadAverage(loadAvgStr)
	}

	stats.Load.Uptime = uptime

	if cpuInfo, err := systemReader.CPUInfo(); err == nil && len(cpuInfo) > 0 {
		stats.CPUInfo = cpuInfo[0].ModelName
	}

	if stats.CPUCores == 0 {
		if coreCount, err := systemReader.CPUCounts(true); err == nil {
			stats.CPUCores = coreCount
		}
	}

	stats.CPU = getCPUStats(systemReader)

	if memInfo, err := systemReader.VirtualMemory(); err == nil {
		stats.Memory = MemoryStats{
			Total:      float64(memInfo.Total) / bytesInGB,
			Used:       float64(memInfo.Used) / bytesInGB,
			Percentage: memInfo.UsedPercent,
			RawInfo: fmt.Sprintf("Total: %s, Used: %s, Free: %s",
				formatBytes(memInfo.Total, "GB"),
				formatBytes(memInfo.Used, "GB"),
				formatBytes(memInfo.Free, "GB")),
		}
	}

	diskStats := DiskStats{
		AllMounts: []DiskMount{},
	}

	if diskInfo, err := systemReader.DiskPartitions(false); err == nil && len(diskInfo) > 0 {
		diskStats.AllMounts = make([]DiskMount, 0, len(diskInfo))

		for _, partition := range diskInfo {
			if usage, err := systemReader.DiskUsage(partition.Mountpoint); err == nil {
				mount := DiskMount{
					Filesystem: partition.Fstype,
					Size:       formatBytes(usage.Total, "GB"),
					Used:       formatBytes(usage.Used, "GB"),
					Avail:      formatBytes(usage.Free, "GB"),
					Capacity:   fmt.Sprintf("%.1f%%", usage.UsedPercent),
					MountPoint: partition.Mountpoint,
				}

				diskStats.AllMounts = append(diskStats.AllMounts, mount)

				if mount.MountPoint == "/" || (diskStats.MountPoint != "/" && diskStats.Total == 0) {
					diskStats.MountPoint = mount.MountPoint
					diskStats.Total = float64(usage.Total) / bytesInGB
					diskStats.Used = float64(usage.Used) / bytesInGB
					diskStats.Available = float64(usage.Free) / bytesInGB
					diskStats.Percentage = usage.UsedPercent
				}
			}
		}
	}
	// Ensure AllMounts is never nil (keep empty array if no partitions found)
	if diskStats.AllMounts == nil {
		diskStats.AllMounts = []DiskMount{}
	}

	stats.Disk = diskStats

	stats.Network = getNetworkStats(systemReader)

	return stats, nil
}

func formatBytes(bytes uint64, unit string) string {
	switch unit {
	case "MB":
		return fmt.Sprintf("%.2f MB", float64(bytes)/bytesInMB)
	case "GB":
		return fmt.Sprintf("%.2f GB", float64(bytes)/bytesInGB)
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

func parseLoadAverage(loadStr string) LoadStats {
	loadStats := LoadStats{}

	loadRe := regexp.MustCompile(`load averages?: ([\d.]+),? ([\d.]+),? ([\d.]+)`)
	matches := loadRe.FindStringSubmatch(loadStr)
	if len(matches) >= 4 {
		if one, err := strconv.ParseFloat(matches[1], 64); err == nil {
			loadStats.OneMin = one
		}
		if five, err := strconv.ParseFloat(matches[2], 64); err == nil {
			loadStats.FiveMin = five
		}
		if fifteen, err := strconv.ParseFloat(matches[3], 64); err == nil {
			loadStats.FifteenMin = fifteen
		}
	}

	return loadStats
}

func getCPUStats(reader SystemReader) CPUStats {
	cpuStats := CPUStats{
		Overall: 0.0,
		PerCore: []CPUCore{},
	}

	perCorePercent, err := reader.CPUPercent(time.Second, true)
	if err == nil && len(perCorePercent) > 0 {
		cpuStats.PerCore = make([]CPUCore, len(perCorePercent))
		var totalUsage float64 = 0

		for i, usage := range perCorePercent {
			cpuStats.PerCore[i] = CPUCore{
				CoreID: i,
				Usage:  usage,
			}
			totalUsage += usage
		}

		cpuStats.Overall = totalUsage / float64(len(perCorePercent))
	} else {
		if overallPercent, err := reader.CPUPercent(time.Second, false); err == nil && len(overallPercent) > 0 {
			cpuStats.Overall = overallPercent[0]
		}
	}

	return cpuStats
}

func getNetworkStats(reader SystemReader) NetworkStats {
	networkStats := NetworkStats{
		Interfaces: []NetworkInterface{},
	}

	if ioCounters, err := reader.IOCounters(true); err == nil {
		var totalSent, totalRecv, totalPacketsSent, totalPacketsRecv uint64

		for _, counter := range ioCounters {
			interfaces := NetworkInterface{
				Name:        counter.Name,
				BytesSent:   counter.BytesSent,
				BytesRecv:   counter.BytesRecv,
				PacketsSent: counter.PacketsSent,
				PacketsRecv: counter.PacketsRecv,
				ErrorIn:     counter.Errin,
				ErrorOut:    counter.Errout,
				DropIn:      counter.Dropin,
				DropOut:     counter.Dropout,
			}

			networkStats.Interfaces = append(networkStats.Interfaces, interfaces)

			totalSent += counter.BytesSent
			totalRecv += counter.BytesRecv
			totalPacketsSent += counter.PacketsSent
			totalPacketsRecv += counter.PacketsRecv
		}

		networkStats.TotalBytesSent = totalSent
		networkStats.TotalBytesRecv = totalRecv
		networkStats.TotalPacketsSent = totalPacketsSent
		networkStats.TotalPacketsRecv = totalPacketsRecv
	}

	return networkStats
}

// GetSystemStats retrieves system statistics using the service function with SSH command executor
func (m *DashboardMonitor) GetSystemStats() {
	// Check if context is cancelled before proceeding
	select {
	case <-m.ctx.Done():
		return
	default:
	}

	// Use SSH-based command executor
	cmdExecutor := func(cmd string) (string, error) {
		return m.getCommandOutput(cmd)
	}

	stats, err := CollectSystemStats(m.log, GetSystemStatsOptions{
		CommandExecutor: cmdExecutor,
	})
	if err != nil {
		m.BroadcastError(err.Error(), GetSystemStats)
		return
	}

	m.Broadcast(string(GetSystemStats), stats)
}

func (m *DashboardMonitor) getCommandOutput(cmd string) (string, error) {
	// Use centralized session creation with retry logic
	session, err := m.sshManager.NewSessionWithRetry("")
	if err != nil {
		m.log.Log(logger.Error, "Failed to create new session", err.Error())
		return "", err
	}
	defer session.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	err = session.Run(cmd)
	if err != nil {
		errMsg := fmt.Sprintf("Command failed: %s, stderr: %s", err.Error(), stderrBuf.String())
		m.log.Log(logger.Error, errMsg, "")
		return "", errors.New(errMsg)
	}

	return stdoutBuf.String(), nil
}

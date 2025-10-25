package dashboard

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

const (
	bytesInMB = 1024 * 1024
	bytesInGB = 1024 * 1024 * 1024
)

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

func (m *DashboardMonitor) GetSystemStats() {
	osType, err := m.getCommandOutput("uname -s")
	if err != nil {
		m.BroadcastError(err.Error(), GetSystemStats)
		return
	}
	osType = strings.TrimSpace(osType)

	stats := SystemStats{
		OSType:    osType,
		Timestamp: time.Now(),
		Memory:    MemoryStats{},
		Load:      LoadStats{},
		Disk:      DiskStats{AllMounts: []DiskMount{}},
	}

	// Get hostname
	if hostname, err := m.getCommandOutput("hostname"); err == nil {
		stats.Hostname = strings.TrimSpace(hostname)
	}

	// Get kernel version
	if kernelVersion, err := m.getCommandOutput("uname -r"); err == nil {
		stats.KernelVersion = strings.TrimSpace(kernelVersion)
	}

	// Get architecture
	if architecture, err := m.getCommandOutput("uname -m"); err == nil {
		stats.Architecture = strings.TrimSpace(architecture)
	}

	// Get uptime 
	var uptime string
	if hostInfo, err := host.Info(); err == nil {
		uptime = time.Duration(hostInfo.Uptime * uint64(time.Second)).String()
	}

	// Parse load averages
	if loadAvg, err := m.getCommandOutput("uptime"); err == nil {
		loadAvgStr := strings.TrimSpace(loadAvg)
		stats.Load = parseLoadAverage(loadAvgStr)
	}

	// Set uptime after parsing load averages
	stats.Load.Uptime = uptime

	// Get CPU info
	if cpuInfo, err := cpu.Info(); err == nil && len(cpuInfo) > 0 {
		stats.CPUInfo = cpuInfo[0].ModelName
	}

	// Get CPU cores count
	if cores, err := cpu.Counts(true); err == nil {
		stats.CPUCores = cores
	} else {
		// Fallback: use nproc command
		if coresStr, err := m.getCommandOutput("nproc"); err == nil {
			if coresInt, err := strconv.Atoi(strings.TrimSpace(coresStr)); err == nil {
				stats.CPUCores = coresInt
			}
		}
	}

	if memInfo, err := mem.VirtualMemory(); err == nil {
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

	if diskInfo, err := disk.Partitions(false); err == nil {
		diskStats := DiskStats{
			AllMounts: make([]DiskMount, 0, len(diskInfo)),
		}

		for _, partition := range diskInfo {
			if usage, err := disk.Usage(partition.Mountpoint); err == nil {
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

		stats.Disk = diskStats
	}

	m.Broadcast(string(GetSystemStats), stats)
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

func (m *DashboardMonitor) getCommandOutput(cmd string) (string, error) {
	session, err := m.client.NewSession()
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
		return "", fmt.Errorf(errMsg)
	}

	return stdoutBuf.String(), nil
}

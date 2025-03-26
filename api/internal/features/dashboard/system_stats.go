package dashboard

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

func parseSize(sizeStr string) float64 {
	re := regexp.MustCompile(`^([\d.]+)([KMGT])?`)
	matches := re.FindStringSubmatch(sizeStr)
	if len(matches) < 2 {
		return 0
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0
	}

	if len(matches) < 3 {
		return value
	}

	unit := strings.ToUpper(matches[2])
	switch unit {
	case "K":
		return value / 1024
	case "M":
		return value
	case "G":
		return value * 1024
	case "T":
		return value * 1024 * 1024
	default:
		return value
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

	if loadAvg, err := m.getCommandOutput("uptime"); err == nil {
		loadAvgStr := strings.TrimSpace(loadAvg)
		stats.Load = parseLoadAverage(loadAvgStr)
		uptimeRe := regexp.MustCompile(`up\s+(.*?),\s+\d+\s+users?`)
		uptimeMatches := uptimeRe.FindStringSubmatch(loadAvgStr)
		if len(uptimeMatches) > 1 {
			stats.Load.Uptime = uptimeMatches[1]
		}
	}

	switch osType {
	case "Linux":
		if cpuInfo, err := m.getCommandOutput("grep 'model name' /proc/cpuinfo | head -1"); err == nil {
			cpuInfoStr := strings.TrimSpace(cpuInfo)
			stats.CPUInfo = strings.Replace(cpuInfoStr, "model name\t: ", "", 1)
		}

		if memInfo, err := m.getCommandOutput("free -m | head -2"); err == nil {
			memInfoStr := strings.TrimSpace(memInfo)
			stats.Memory = parseLinuxMemory(memInfoStr)
		}

		if diskInfo, err := m.getCommandOutput("df -h | grep -v 'tmpfs\\|udev'"); err == nil {
			diskInfoStr := strings.TrimSpace(diskInfo)
			stats.Disk = parseLinuxDisk(diskInfoStr)
		}

	case "Darwin":
		if cpuInfo, err := m.getCommandOutput("sysctl -n machdep.cpu.brand_string"); err == nil {
			stats.CPUInfo = strings.TrimSpace(cpuInfo)
		}

		if memInfo, err := m.getCommandOutput("top -l 1 | grep PhysMem"); err == nil {
			memInfoStr := strings.TrimSpace(memInfo)
			stats.Memory = parseDarwinMemory(memInfoStr)
		}

		if diskInfo, err := m.getCommandOutput("df -h"); err == nil {
			diskInfoStr := strings.TrimSpace(diskInfo)
			stats.Disk = parseDarwinDisk(diskInfoStr)
		}
	}

	m.Broadcast(string(GetSystemStats), stats)
}

// parseLoadAverage extracts load average from uptime command output
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

// parseLinuxMemory extracts memory information from Linux free command
func parseLinuxMemory(memStr string) MemoryStats {
	memStats := MemoryStats{
		RawInfo: memStr,
	}

	lines := strings.Split(memStr, "\n")
	if len(lines) < 2 {
		return memStats
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 3 {
		return memStats
	}

	if total, err := strconv.ParseFloat(fields[1], 64); err == nil {
		memStats.Total = total
	}

	if used, err := strconv.ParseFloat(fields[2], 64); err == nil {
		memStats.Used = used
	}

	if memStats.Total > 0 {
		memStats.Percentage = (memStats.Used / memStats.Total) * 100
	}

	return memStats
}

// parseDarwinMemory extracts memory information from macOS top command
func parseDarwinMemory(memStr string) MemoryStats {
	memStats := MemoryStats{
		RawInfo: memStr,
	}

	re := regexp.MustCompile(`PhysMem: (\d+)([KMGT]) used .* (\d+)([KMGT]) unused`)
	matches := re.FindStringSubmatch(memStr)
	if len(matches) >= 5 {
		usedVal, _ := strconv.ParseFloat(matches[1], 64)
		usedUnit := matches[2]

		unusedVal, _ := strconv.ParseFloat(matches[3], 64)
		unusedUnit := matches[4]

		used := convertToMB(usedVal, usedUnit)
		unused := convertToMB(unusedVal, unusedUnit)

		memStats.Used = used
		memStats.Total = used + unused

		if memStats.Total > 0 {
			memStats.Percentage = (memStats.Used / memStats.Total) * 100
		}
	}

	return memStats
}

// parseLinuxDisk extracts disk information from Linux df command
func parseLinuxDisk(diskStr string) DiskStats {
	diskStats := DiskStats{
		AllMounts: []DiskMount{},
	}

	lines := strings.Split(diskStr, "\n")
	if len(lines) < 2 {
		return diskStats
	}

	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		mount := DiskMount{
			Filesystem: fields[0],
			Size:       fields[1],
			Used:       fields[2],
			Avail:      fields[3],
			Capacity:   fields[4],
			MountPoint: fields[5],
		}

		diskStats.AllMounts = append(diskStats.AllMounts, mount)

		if mount.MountPoint == "/" || (diskStats.MountPoint != "/" && diskStats.Total == 0) {
			diskStats.MountPoint = mount.MountPoint
			diskStats.Total = parseSize(mount.Size)
			diskStats.Used = parseSize(mount.Used)
			diskStats.Available = parseSize(mount.Avail)
			percentStr := strings.TrimSuffix(mount.Capacity, "%")
			if percent, err := strconv.ParseFloat(percentStr, 64); err == nil {
				diskStats.Percentage = percent
			}
		}
	}

	return diskStats
}

// parseDarwinDisk extracts disk information from macOS df command
func parseDarwinDisk(diskStr string) DiskStats {
	return parseLinuxDisk(diskStr)
}

func convertToMB(value float64, unit string) float64 {
	switch unit {
	case "K":
		return value / 1024
	case "M":
		return value
	case "G":
		return value * 1024
	case "T":
		return value * 1024 * 1024
	default:
		return value
	}
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

package dashboard

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// Custom types to replace gopsutil types
type HostInfoStat struct {
	Uptime uint64
}

type CPUInfoStat struct {
	ModelName string
}

type VirtualMemoryStat struct {
	Total       uint64
	Used        uint64
	Free        uint64
	Available   uint64
	Buffers     uint64
	Cached      uint64
	UsedPercent float64
}

type PartitionStat struct {
	Device     string
	Mountpoint string
	Fstype     string
}

type UsageStat struct {
	Path        string
	Total       uint64
	Used        uint64
	Free        uint64
	InodesTotal uint64
	InodesUsed  uint64
	InodesFree  uint64
	UsedPercent float64
}

type IOCountersStat struct {
	Name        string
	BytesSent   uint64
	BytesRecv   uint64
	PacketsSent uint64
	PacketsRecv uint64
	Errin       uint64
	Errout      uint64
	Dropin      uint64
	Dropout     uint64
}

// SystemReader abstracts system information reading operations.
// Similar to Docker tunnel pattern, this allows both local and remote system access.
type SystemReader interface {
	// HostInfo returns host information including uptime
	HostInfo() (*HostInfoStat, error)
	// CPUInfo returns CPU information
	CPUInfo() ([]CPUInfoStat, error)
	// CPUCounts returns the number of CPU cores
	CPUCounts(logical bool) (int, error)
	// CPUPercent returns CPU usage percentages
	CPUPercent(interval time.Duration, percpu bool) ([]float64, error)
	// VirtualMemory returns memory statistics
	VirtualMemory() (*VirtualMemoryStat, error)
	// DiskPartitions returns disk partition information
	DiskPartitions(all bool) ([]PartitionStat, error)
	// DiskUsage returns disk usage for a mountpoint
	DiskUsage(path string) (*UsageStat, error)
	// IOCounters returns network I/O statistics
	IOCounters(pernic bool) ([]IOCountersStat, error)
}

// LocalSystemReader uses standard library and system commands for local system access
type LocalSystemReader struct{}

func NewLocalSystemReader() *LocalSystemReader {
	return &LocalSystemReader{}
}

func (r *LocalSystemReader) HostInfo() (*HostInfoStat, error) {
	// Read /proc/uptime
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/uptime: %w", err)
	}

	parts := strings.Fields(string(data))
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid uptime format")
	}

	uptimeSec, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse uptime: %w", err)
	}

	return &HostInfoStat{
		Uptime: uint64(uptimeSec),
	}, nil
}

func (r *LocalSystemReader) CPUInfo() ([]CPUInfoStat, error) {
	// Read /proc/cpuinfo
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/cpuinfo: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "model name") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				modelName := strings.TrimSpace(parts[1])
				return []CPUInfoStat{{ModelName: modelName}}, nil
			}
		}
	}

	return []CPUInfoStat{{ModelName: "Unknown"}}, nil
}

func (r *LocalSystemReader) CPUCounts(logical bool) (int, error) {
	// Read /proc/cpuinfo and count processors
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return 0, fmt.Errorf("failed to read /proc/cpuinfo: %w", err)
	}

	count := 0
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "processor") {
			count++
		}
	}

	if count == 0 {
		return 0, fmt.Errorf("no processors found in /proc/cpuinfo")
	}

	return count, nil
}

func (r *LocalSystemReader) CPUPercent(interval time.Duration, percpu bool) ([]float64, error) {
	cmd := exec.Command("sh", "-c", "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%%* *id.*/\\1/' | awk '{print 100 - $1}'")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU percentage: %w", err)
	}

	val, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CPU percentage: %w", err)
	}

	if percpu {
		coreCount, err := r.CPUCounts(true)
		if err != nil {
			return nil, fmt.Errorf("failed to get CPU count: %w", err)
		}
		percentages := make([]float64, coreCount)
		for i := range percentages {
			percentages[i] = val
		}
		return percentages, nil
	}

	return []float64{val}, nil
}

func (r *LocalSystemReader) VirtualMemory() (*VirtualMemoryStat, error) {
	// Read /proc/meminfo
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/meminfo: %w", err)
	}

	memInfo := &VirtualMemoryStat{}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := parts[0]
		value, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			continue
		}
		value *= 1024 // Convert from KB to bytes

		switch key {
		case "MemTotal:":
			memInfo.Total = value
		case "MemFree:":
			memInfo.Free = value
		case "MemAvailable:":
			memInfo.Available = value
		case "Buffers:":
			memInfo.Buffers = value
		case "Cached:":
			memInfo.Cached = value
		}
	}

	memInfo.Used = memInfo.Total - memInfo.Free
	if memInfo.Total > 0 {
		memInfo.UsedPercent = float64(memInfo.Used) / float64(memInfo.Total) * 100
	}

	return memInfo, nil
}

func (r *LocalSystemReader) DiskPartitions(all bool) ([]PartitionStat, error) {
	// Use df command
	cmd := exec.Command("df", "-T")
	if all {
		cmd = exec.Command("df", "-aT")
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk partitions: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	partitions := make([]PartitionStat, 0)

	// Skip header line
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 7 {
			continue
		}

		partitions = append(partitions, PartitionStat{
			Device:     parts[0],
			Mountpoint: parts[6],
			Fstype:     parts[1],
		})
	}

	return partitions, nil
}

func (r *LocalSystemReader) DiskUsage(path string) (*UsageStat, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return nil, fmt.Errorf("failed to statfs %s: %w", path, err)
	}

	total := uint64(stat.Blocks) * uint64(stat.Bsize)
	free := uint64(stat.Bavail) * uint64(stat.Bsize)
	used := total - free

	usage := &UsageStat{
		Path:        path,
		Total:       total,
		Used:        used,
		Free:        free,
		InodesTotal: stat.Files,
		InodesFree:  stat.Ffree,
		InodesUsed:  stat.Files - stat.Ffree,
	}

	if total > 0 {
		usage.UsedPercent = float64(used) / float64(total) * 100
	}

	return usage, nil
}

func (r *LocalSystemReader) IOCounters(pernic bool) ([]IOCountersStat, error) {
	// Read /proc/net/dev
	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/net/dev: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	counters := make([]IOCountersStat, 0)

	// Skip header lines (first 2 lines)
	for i := 2; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 10 {
			continue
		}

		name := strings.TrimSuffix(parts[0], ":")
		bytesRecv, _ := strconv.ParseUint(parts[1], 10, 64)
		packetsRecv, _ := strconv.ParseUint(parts[2], 10, 64)
		errin, _ := strconv.ParseUint(parts[3], 10, 64)
		dropin, _ := strconv.ParseUint(parts[4], 10, 64)
		bytesSent, _ := strconv.ParseUint(parts[9], 10, 64)
		packetsSent, _ := strconv.ParseUint(parts[10], 10, 64)
		errout, _ := strconv.ParseUint(parts[11], 10, 64)
		dropout, _ := strconv.ParseUint(parts[12], 10, 64)

		counters = append(counters, IOCountersStat{
			Name:        name,
			BytesRecv:   bytesRecv,
			PacketsRecv: packetsRecv,
			Errin:       errin,
			Dropin:      dropin,
			BytesSent:   bytesSent,
			PacketsSent: packetsSent,
			Errout:      errout,
			Dropout:     dropout,
		})
	}

	return counters, nil
}

// RemoteSystemReader uses SSH commands to read remote system information
type RemoteSystemReader struct {
	cmdExecutor CommandExecutor
	logger      logger.Logger
}

func NewRemoteSystemReader(cmdExecutor CommandExecutor, logger logger.Logger) *RemoteSystemReader {
	return &RemoteSystemReader{
		cmdExecutor: cmdExecutor,
		logger:      logger,
	}
}

func (r *RemoteSystemReader) HostInfo() (*HostInfoStat, error) {
	uptimeStr, err := r.cmdExecutor("cat /proc/uptime")
	if err != nil {
		return nil, fmt.Errorf("failed to get uptime: %w", err)
	}

	parts := strings.Fields(strings.TrimSpace(uptimeStr))
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid uptime format")
	}

	uptimeSec, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse uptime: %w", err)
	}

	return &HostInfoStat{
		Uptime: uint64(uptimeSec),
	}, nil
}

func (r *RemoteSystemReader) CPUInfo() ([]CPUInfoStat, error) {
	cpuInfoStr, err := r.cmdExecutor("cat /proc/cpuinfo | grep 'model name' | head -1 | cut -d ':' -f 2 | sed 's/^[[:space:]]*//'")
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU info: %w", err)
	}

	modelName := strings.TrimSpace(cpuInfoStr)
	return []CPUInfoStat{
		{
			ModelName: modelName,
		},
	}, nil
}

func (r *RemoteSystemReader) CPUCounts(logical bool) (int, error) {
	var cmd string
	if logical {
		cmd = "nproc"
	} else {
		cmd = "nproc --all"
	}

	output, err := r.cmdExecutor(cmd)
	if err != nil {
		return 0, fmt.Errorf("failed to get CPU count: %w", err)
	}

	count, err := strconv.Atoi(strings.TrimSpace(output))
	if err != nil {
		return 0, fmt.Errorf("failed to parse CPU count: %w", err)
	}

	return count, nil
}

func (r *RemoteSystemReader) CPUPercent(interval time.Duration, percpu bool) ([]float64, error) {
	// Use a simple command-based approach for CPU percentage
	// This uses 'top' or 'vmstat' to get current CPU usage
	var cmd string
	if percpu {
		// For per-core, we'll get overall and divide (simplified approach)
		// A more accurate implementation would parse /proc/stat with two samples
		cmd = "grep -c processor /proc/cpuinfo"
		coreCountStr, err := r.cmdExecutor(cmd)
		if err != nil {
			return nil, fmt.Errorf("failed to get CPU count: %w", err)
		}
		coreCount, err := strconv.Atoi(strings.TrimSpace(coreCountStr))
		if err != nil {
			return nil, fmt.Errorf("failed to parse CPU count: %w", err)
		}

		// Get overall CPU usage and replicate for each core
		overallSlice, err := r.getOverallCPUPercent()
		if err != nil {
			return nil, err
		}

		var overallValue float64
		if len(overallSlice) > 0 {
			overallValue = overallSlice[0]
		}

		percentages := make([]float64, coreCount)
		for i := range percentages {
			percentages[i] = overallValue
		}
		return percentages, nil
	}

	// Get overall CPU usage
	return r.getOverallCPUPercent()
}

func (r *RemoteSystemReader) getOverallCPUPercent() ([]float64, error) {
	cmd := "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%%* *id.*/\\1/' | awk '{print 100 - $1}'"
	output, err := r.cmdExecutor(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU percentage: %w", err)
	}

	if output == "" {
		return nil, fmt.Errorf("empty output from CPU percentage command")
	}

	val, err := strconv.ParseFloat(strings.TrimSpace(output), 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CPU percentage: %w", err)
	}

	return []float64{val}, nil
}

func (r *RemoteSystemReader) VirtualMemory() (*VirtualMemoryStat, error) {
	memInfoStr, err := r.cmdExecutor("cat /proc/meminfo")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/meminfo: %w", err)
	}

	memInfo := &VirtualMemoryStat{}
	lines := strings.Split(memInfoStr, "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := parts[0]
		value, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			continue
		}
		value *= 1024 // Convert from KB to bytes

		switch key {
		case "MemTotal:":
			memInfo.Total = value
		case "MemFree:":
			memInfo.Free = value
		case "MemAvailable:":
			memInfo.Available = value
		case "Buffers:":
			memInfo.Buffers = value
		case "Cached:":
			memInfo.Cached = value
		}
	}

	memInfo.Used = memInfo.Total - memInfo.Free
	if memInfo.Total > 0 {
		memInfo.UsedPercent = float64(memInfo.Used) / float64(memInfo.Total) * 100
	}

	return memInfo, nil
}

func (r *RemoteSystemReader) DiskPartitions(all bool) ([]PartitionStat, error) {
	cmd := "df -T"
	if all {
		cmd = "df -aT"
	}

	output, err := r.cmdExecutor(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk partitions: %w", err)
	}

	lines := strings.Split(output, "\n")
	partitions := make([]PartitionStat, 0)

	// Skip header line
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 7 {
			continue
		}

		partitions = append(partitions, PartitionStat{
			Device:     parts[0],
			Mountpoint: parts[6],
			Fstype:     parts[1],
		})
	}

	return partitions, nil
}

func (r *RemoteSystemReader) DiskUsage(path string) (*UsageStat, error) {
	output, err := r.cmdExecutor(fmt.Sprintf("df -B1 %s | tail -1", path))
	if err != nil {
		return nil, fmt.Errorf("failed to get disk usage: %w", err)
	}

	parts := strings.Fields(strings.TrimSpace(output))
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid df output format")
	}

	total, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse total: %w", err)
	}

	used, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse used: %w", err)
	}

	free, err := strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse free: %w", err)
	}

	usage := &UsageStat{
		Path:        path,
		Total:       total,
		Used:        used,
		Free:        free,
		InodesTotal: 0,
		InodesUsed:  0,
		InodesFree:  0,
	}

	if total > 0 {
		usage.UsedPercent = float64(used) / float64(total) * 100
	}

	return usage, nil
}

func (r *RemoteSystemReader) IOCounters(pernic bool) ([]IOCountersStat, error) {
	netDevStr, err := r.cmdExecutor("cat /proc/net/dev")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/net/dev: %w", err)
	}

	lines := strings.Split(netDevStr, "\n")
	counters := make([]IOCountersStat, 0)

	// Skip header lines (first 2 lines)
	for i := 2; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 10 {
			continue
		}

		name := strings.TrimSuffix(parts[0], ":")
		bytesRecv, _ := strconv.ParseUint(parts[1], 10, 64)
		packetsRecv, _ := strconv.ParseUint(parts[2], 10, 64)
		errin, _ := strconv.ParseUint(parts[3], 10, 64)
		dropin, _ := strconv.ParseUint(parts[4], 10, 64)
		bytesSent, _ := strconv.ParseUint(parts[9], 10, 64)
		packetsSent, _ := strconv.ParseUint(parts[10], 10, 64)
		errout, _ := strconv.ParseUint(parts[11], 10, 64)
		dropout, _ := strconv.ParseUint(parts[12], 10, 64)

		counters = append(counters, IOCountersStat{
			Name:        name,
			BytesRecv:   bytesRecv,
			PacketsRecv: packetsRecv,
			Errin:       errin,
			Dropin:      dropin,
			BytesSent:   bytesSent,
			PacketsSent: packetsSent,
			Errout:      errout,
			Dropout:     dropout,
		})
	}

	return counters, nil
}

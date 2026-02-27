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

// procStatSample represents a snapshot of /proc/stat CPU times
type procStatSample struct {
	cpuTimes map[string][]uint64 // core name -> [user, nice, system, idle, iowait, irq, softirq, steal, guest, guest_nice]
}

// readProcStat reads /proc/stat and parses CPU times
func (r *LocalSystemReader) readProcStat() (*procStatSample, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return nil, err
	}

	sample := &procStatSample{
		cpuTimes: make(map[string][]uint64),
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "cpu") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		cpuName := fields[0]
		times := make([]uint64, 10)
		for i := 1; i < 11 && i < len(fields); i++ {
			val, err := strconv.ParseUint(fields[i], 10, 64)
			if err != nil {
				continue
			}
			times[i-1] = val
		}
		sample.cpuTimes[cpuName] = times
	}

	return sample, nil
}

// calculatePerCoreCPUPercent calculates CPU usage percentage for each core
func (r *LocalSystemReader) calculatePerCoreCPUPercent(first, second *procStatSample) ([]float64, error) {
	var percentages []float64

	// Get core count to determine how many cores we should have
	coreCount, err := r.CPUCounts(true)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU count: %w", err)
	}

	// Process each CPU core (cpu0, cpu1, cpu2, etc.)
	for i := 0; i < coreCount; i++ {
		cpuName := fmt.Sprintf("cpu%d", i)
		firstTimes, ok1 := first.cpuTimes[cpuName]
		secondTimes, ok2 := second.cpuTimes[cpuName]

		if !ok1 || !ok2 {
			// If we can't find this core, use 0.0
			percentages = append(percentages, 0.0)
			continue
		}

		percent := r.calculateCPUPercentFromTimes(firstTimes, secondTimes)
		percentages = append(percentages, percent)
	}

	if len(percentages) == 0 {
		return nil, fmt.Errorf("no CPU cores found in /proc/stat")
	}

	return percentages, nil
}

// calculateOverallCPUPercent calculates overall CPU usage percentage
func (r *LocalSystemReader) calculateOverallCPUPercent(first, second *procStatSample) ([]float64, error) {
	firstTimes, ok1 := first.cpuTimes["cpu"]
	secondTimes, ok2 := second.cpuTimes["cpu"]

	if !ok1 || !ok2 {
		return nil, fmt.Errorf("failed to find overall CPU stats in /proc/stat")
	}

	percent := r.calculateCPUPercentFromTimes(firstTimes, secondTimes)
	return []float64{percent}, nil
}

// calculateCPUPercentFromTimes calculates CPU percentage from two time samples
func (r *LocalSystemReader) calculateCPUPercentFromTimes(firstTimes, secondTimes []uint64) float64 {
	if len(firstTimes) < 4 || len(secondTimes) < 4 {
		return 0.0
	}

	// Calculate total time (sum of all CPU time fields)
	var firstTotal, secondTotal uint64
	var firstIdle, secondIdle uint64

	for i := 0; i < len(firstTimes) && i < len(secondTimes); i++ {
		firstTotal += firstTimes[i]
		secondTotal += secondTimes[i]
	}

	// Idle time is the sum of idle and iowait (indices 3 and 4)
	firstIdle = firstTimes[3]
	if len(firstTimes) > 4 {
		firstIdle += firstTimes[4]
	}

	secondIdle = secondTimes[3]
	if len(secondTimes) > 4 {
		secondIdle += secondTimes[4]
	}

	// Calculate differences
	totalDiff := secondTotal - firstTotal
	idleDiff := secondIdle - firstIdle

	if totalDiff == 0 {
		return 0.0
	}

	// CPU usage = 100 - (idle percentage)
	idlePercent := float64(idleDiff) / float64(totalDiff) * 100.0
	cpuUsage := 100.0 - idlePercent

	// Clamp to valid range
	if cpuUsage < 0.0 {
		cpuUsage = 0.0
	}
	if cpuUsage > 100.0 {
		cpuUsage = 100.0
	}

	return cpuUsage
}

func (r *LocalSystemReader) CPUPercent(interval time.Duration, percpu bool) ([]float64, error) {
	// Use /proc/stat for accurate CPU usage calculation
	// Read first sample
	firstSample, err := r.readProcStat()
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/stat: %w", err)
	}

	// Wait for the interval
	time.Sleep(interval)

	// Read second sample
	secondSample, err := r.readProcStat()
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/stat: %w", err)
	}

	if percpu {
		return r.calculatePerCoreCPUPercent(firstSample, secondSample)
	}

	return r.calculateOverallCPUPercent(firstSample, secondSample)
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
	// Use /proc/stat for accurate CPU usage calculation
	// Read first sample
	firstSample, err := r.readProcStatRemote()
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/stat: %w", err)
	}

	// Wait for the interval
	time.Sleep(interval)

	// Read second sample
	secondSample, err := r.readProcStatRemote()
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/stat: %w", err)
	}

	if percpu {
		return r.calculatePerCoreCPUPercentRemote(firstSample, secondSample)
	}

	return r.calculateOverallCPUPercentRemote(firstSample, secondSample)
}

// readProcStatRemote reads /proc/stat via command executor and parses CPU times
func (r *RemoteSystemReader) readProcStatRemote() (*procStatSample, error) {
	output, err := r.cmdExecutor("cat /proc/stat")
	if err != nil {
		return nil, err
	}

	sample := &procStatSample{
		cpuTimes: make(map[string][]uint64),
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "cpu") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		cpuName := fields[0]
		times := make([]uint64, 10)
		for i := 1; i < 11 && i < len(fields); i++ {
			val, err := strconv.ParseUint(fields[i], 10, 64)
			if err != nil {
				continue
			}
			times[i-1] = val
		}
		sample.cpuTimes[cpuName] = times
	}

	return sample, nil
}

// calculatePerCoreCPUPercentRemote calculates CPU usage percentage for each core
func (r *RemoteSystemReader) calculatePerCoreCPUPercentRemote(first, second *procStatSample) ([]float64, error) {
	var percentages []float64

	// Get core count to determine how many cores we should have
	coreCount, err := r.CPUCounts(true)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU count: %w", err)
	}

	// Process each CPU core (cpu0, cpu1, cpu2, etc.)
	for i := 0; i < coreCount; i++ {
		cpuName := fmt.Sprintf("cpu%d", i)
		firstTimes, ok1 := first.cpuTimes[cpuName]
		secondTimes, ok2 := second.cpuTimes[cpuName]

		if !ok1 || !ok2 {
			// If we can't find this core, use 0.0
			percentages = append(percentages, 0.0)
			continue
		}

		percent := r.calculateCPUPercentFromTimesRemote(firstTimes, secondTimes)
		percentages = append(percentages, percent)
	}

	if len(percentages) == 0 {
		return nil, fmt.Errorf("no CPU cores found in /proc/stat")
	}

	return percentages, nil
}

// calculateOverallCPUPercentRemote calculates overall CPU usage percentage
func (r *RemoteSystemReader) calculateOverallCPUPercentRemote(first, second *procStatSample) ([]float64, error) {
	firstTimes, ok1 := first.cpuTimes["cpu"]
	secondTimes, ok2 := second.cpuTimes["cpu"]

	if !ok1 || !ok2 {
		return nil, fmt.Errorf("failed to find overall CPU stats in /proc/stat")
	}

	percent := r.calculateCPUPercentFromTimesRemote(firstTimes, secondTimes)
	return []float64{percent}, nil
}

// calculateCPUPercentFromTimesRemote calculates CPU percentage from two time samples
func (r *RemoteSystemReader) calculateCPUPercentFromTimesRemote(firstTimes, secondTimes []uint64) float64 {
	if len(firstTimes) < 4 || len(secondTimes) < 4 {
		return 0.0
	}

	// Calculate total time (sum of all CPU time fields)
	var firstTotal, secondTotal uint64
	var firstIdle, secondIdle uint64

	for i := 0; i < len(firstTimes) && i < len(secondTimes); i++ {
		firstTotal += firstTimes[i]
		secondTotal += secondTimes[i]
	}

	// Idle time is the sum of idle and iowait (indices 3 and 4)
	firstIdle = firstTimes[3]
	if len(firstTimes) > 4 {
		firstIdle += firstTimes[4]
	}

	secondIdle = secondTimes[3]
	if len(secondTimes) > 4 {
		secondIdle += secondTimes[4]
	}

	// Calculate differences
	totalDiff := secondTotal - firstTotal
	idleDiff := secondIdle - firstIdle

	if totalDiff == 0 {
		return 0.0
	}

	// CPU usage = 100 - (idle percentage)
	idlePercent := float64(idleDiff) / float64(totalDiff) * 100.0
	cpuUsage := 100.0 - idlePercent

	// Clamp to valid range
	if cpuUsage < 0.0 {
		cpuUsage = 0.0
	}
	if cpuUsage > 100.0 {
		cpuUsage = 100.0
	}

	return cpuUsage
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

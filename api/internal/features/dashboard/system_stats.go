package dashboard

import (
	"bytes"
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

// systemStatsScript gathers all system metrics in a single shell invocation,
// separating outputs with unique markers so we can parse them in one pass.
// This replaces 15+ individual SSH sessions with a single session.
const systemStatsScript = `echo '===HOSTNAME==='
hostname
echo '===UNAME_S==='
uname -s
echo '===UNAME_R==='
uname -r
echo '===UNAME_M==='
uname -m
echo '===UPTIME_RAW==='
cat /proc/uptime
echo '===UPTIME==='
uptime
echo '===CPUINFO==='
cat /proc/cpuinfo | grep 'model name' | head -1 | cut -d ':' -f 2 | sed 's/^[[:space:]]*//'
echo '===NPROC==='
nproc
echo '===PROC_STAT_1==='
cat /proc/stat
sleep 1
echo '===PROC_STAT_2==='
cat /proc/stat
echo '===MEMINFO==='
cat /proc/meminfo
echo '===DF_T==='
df -T
echo '===DF_B1==='
df -B1
echo '===NET_DEV==='
cat /proc/net/dev
echo '===DONE==='
`

// sectionMarkers defines the ordered markers in systemStatsScript output.
var sectionMarkers = []string{
	"HOSTNAME", "UNAME_S", "UNAME_R", "UNAME_M",
	"UPTIME_RAW", "UPTIME", "CPUINFO", "NPROC",
	"PROC_STAT_1", "PROC_STAT_2",
	"MEMINFO", "DF_T", "DF_B1", "NET_DEV", "DONE",
}

// CollectSystemStats retrieves system statistics. Can be used by DashboardMonitor or MCP tools.
func CollectSystemStats(
	l logger.Logger,
	opts GetSystemStatsOptions,
) (SystemStats, error) {
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

// getSystemStats collects all system metrics in a single SSH session and
// broadcasts the result to every subscribed monitor.
func (p *OrgPoller) getSystemStats() {
	select {
	case <-p.ctx.Done():
		return
	default:
	}

	stats, err := p.collectSystemStatsBatched()
	if err != nil {
		p.log.Log(logger.Error, "Failed to collect system stats", err.Error())
		p.broadcastError(err.Error(), GetSystemStats)
		return
	}

	p.broadcast(string(GetSystemStats), stats)
}

// collectSystemStatsBatched runs systemStatsScript in one SSH session and
// parses the delimited output, reducing ~16 sessions to 1.
func (p *OrgPoller) collectSystemStatsBatched() (SystemStats, error) {
	session, err := p.sshManager.NewSessionWithRetry("")
	if err != nil {
		return SystemStats{}, fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	if err := session.Run(systemStatsScript); err != nil {
		return SystemStats{}, fmt.Errorf("stats script failed: %w, stderr: %s", err, stderrBuf.String())
	}

	return parseBatchedStatsOutput(stdoutBuf.String())
}

// parseSections splits the batched script output by ===MARKER=== delimiters.
func parseSections(output string) map[string]string {
	sections := make(map[string]string, len(sectionMarkers))
	for i, marker := range sectionMarkers {
		startTag := "===" + marker + "==="
		startIdx := strings.Index(output, startTag)
		if startIdx == -1 {
			continue
		}
		contentStart := startIdx + len(startTag)

		var contentEnd int
		if i+1 < len(sectionMarkers) {
			nextTag := "===" + sectionMarkers[i+1] + "==="
			nextIdx := strings.Index(output[contentStart:], nextTag)
			if nextIdx == -1 {
				contentEnd = len(output)
			} else {
				contentEnd = contentStart + nextIdx
			}
		} else {
			contentEnd = len(output)
		}

		sections[marker] = strings.TrimSpace(output[contentStart:contentEnd])
	}
	return sections
}

// parseBatchedStatsOutput constructs SystemStats from the batched script output.
func parseBatchedStatsOutput(output string) (SystemStats, error) {
	sec := parseSections(output)

	osType := sec["UNAME_S"]
	if osType == "" {
		return SystemStats{}, fmt.Errorf("failed to parse OS type from batched output")
	}

	stats := SystemStats{
		OSType:        osType,
		Hostname:      sec["HOSTNAME"],
		KernelVersion: sec["UNAME_R"],
		Architecture:  sec["UNAME_M"],
		Timestamp:     time.Now(),
		CPU:           CPUStats{PerCore: []CPUCore{}},
		Memory:        MemoryStats{},
		Load:          LoadStats{},
		Disk:          DiskStats{AllMounts: []DiskMount{}},
		Network:       NetworkStats{Interfaces: []NetworkInterface{}},
	}

	// Uptime from /proc/uptime
	var uptimeStr string
	if raw := sec["UPTIME_RAW"]; raw != "" {
		parts := strings.Fields(raw)
		if len(parts) > 0 {
			if uptimeSec, err := strconv.ParseFloat(parts[0], 64); err == nil {
				uptimeStr = time.Duration(uint64(uptimeSec) * uint64(time.Second)).String()
			}
		}
	}

	// Load averages from uptime output
	if uptimeOutput := sec["UPTIME"]; uptimeOutput != "" {
		stats.Load = parseLoadAverage(uptimeOutput)
	}
	stats.Load.Uptime = uptimeStr

	// CPU model
	stats.CPUInfo = sec["CPUINFO"]

	// CPU core count
	if nprocStr := sec["NPROC"]; nprocStr != "" {
		if count, err := strconv.Atoi(strings.TrimSpace(nprocStr)); err == nil {
			stats.CPUCores = count
		}
	}

	// CPU usage from two /proc/stat samples (1s apart, captured in the script)
	if procStat1, procStat2 := sec["PROC_STAT_1"], sec["PROC_STAT_2"]; procStat1 != "" && procStat2 != "" {
		first := parseProcStatText(procStat1)
		second := parseProcStatText(procStat2)
		stats.CPU = cpuStatsFromSamples(first, second, stats.CPUCores)
	}

	// Memory
	if memRaw := sec["MEMINFO"]; memRaw != "" {
		stats.Memory = parseMemInfoText(memRaw)
	}

	// Disk: merge df -T (for fstypes) with df -B1 (for byte-level usage)
	stats.Disk = parseDiskSections(sec["DF_T"], sec["DF_B1"])

	// Network
	if netDev := sec["NET_DEV"]; netDev != "" {
		stats.Network = parseNetDevText(netDev)
	}

	return stats, nil
}

// parseProcStatText parses /proc/stat text into a sample map.
func parseProcStatText(text string) map[string][]uint64 {
	cpuTimes := make(map[string][]uint64)
	for _, line := range strings.Split(text, "\n") {
		if !strings.HasPrefix(line, "cpu") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		cpuName := fields[0]
		times := make([]uint64, 0, 10)
		for _, f := range fields[1:] {
			val, err := strconv.ParseUint(f, 10, 64)
			if err != nil {
				break
			}
			times = append(times, val)
		}
		cpuTimes[cpuName] = times
	}
	return cpuTimes
}

// cpuPercentFromTimes calculates CPU usage % between two time slices.
func cpuPercentFromTimes(first, second []uint64) float64 {
	if len(first) < 4 || len(second) < 4 {
		return 0
	}
	var firstTotal, secondTotal uint64
	for i := 0; i < len(first) && i < len(second); i++ {
		firstTotal += first[i]
		secondTotal += second[i]
	}
	firstIdle := first[3]
	if len(first) > 4 {
		firstIdle += first[4]
	}
	secondIdle := second[3]
	if len(second) > 4 {
		secondIdle += second[4]
	}
	totalDiff := secondTotal - firstTotal
	if totalDiff == 0 {
		return 0
	}
	idleDiff := secondIdle - firstIdle
	usage := 100.0 - float64(idleDiff)/float64(totalDiff)*100.0
	if usage < 0 {
		return 0
	}
	if usage > 100 {
		return 100
	}
	return usage
}

// cpuStatsFromSamples builds CPUStats from two /proc/stat snapshots.
func cpuStatsFromSamples(first, second map[string][]uint64, coreCount int) CPUStats {
	stats := CPUStats{PerCore: make([]CPUCore, 0, coreCount)}

	// Per-core
	for i := 0; i < coreCount; i++ {
		name := fmt.Sprintf("cpu%d", i)
		f, ok1 := first[name]
		s, ok2 := second[name]
		usage := 0.0
		if ok1 && ok2 {
			usage = cpuPercentFromTimes(f, s)
		}
		stats.PerCore = append(stats.PerCore, CPUCore{CoreID: i, Usage: usage})
	}

	// Overall
	if f, ok := first["cpu"]; ok {
		if s, ok := second["cpu"]; ok {
			stats.Overall = cpuPercentFromTimes(f, s)
		}
	}
	return stats
}

// parseMemInfoText parses /proc/meminfo text into MemoryStats.
func parseMemInfoText(text string) MemoryStats {
	mem := &VirtualMemoryStat{}
	for _, line := range strings.Split(text, "\n") {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		value, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			continue
		}
		value *= 1024 // KB → bytes
		switch parts[0] {
		case "MemTotal:":
			mem.Total = value
		case "MemFree:":
			mem.Free = value
		case "MemAvailable:":
			mem.Available = value
		case "Buffers:":
			mem.Buffers = value
		case "Cached:":
			mem.Cached = value
		}
	}
	mem.Used = mem.Total - mem.Free
	if mem.Total > 0 {
		mem.UsedPercent = float64(mem.Used) / float64(mem.Total) * 100
	}
	return MemoryStats{
		Total:      float64(mem.Total) / bytesInGB,
		Used:       float64(mem.Used) / bytesInGB,
		Percentage: mem.UsedPercent,
		RawInfo: fmt.Sprintf("Total: %s, Used: %s, Free: %s",
			formatBytes(mem.Total, "GB"),
			formatBytes(mem.Used, "GB"),
			formatBytes(mem.Free, "GB")),
	}
}

// parseDiskSections builds DiskStats from df -T and df -B1 outputs.
func parseDiskSections(dfT, dfB1 string) DiskStats {
	ds := DiskStats{AllMounts: []DiskMount{}}

	// Build mountpoint → fstype map from df -T
	fsTypes := make(map[string]string)
	if dfT != "" {
		for i, line := range strings.Split(dfT, "\n") {
			if i == 0 {
				continue // header
			}
			parts := strings.Fields(strings.TrimSpace(line))
			if len(parts) >= 7 {
				fsTypes[parts[6]] = parts[1]
			}
		}
	}

	if dfB1 == "" {
		return ds
	}

	for i, line := range strings.Split(dfB1, "\n") {
		if i == 0 {
			continue // header
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 6 {
			continue
		}
		mountpoint := parts[5]
		total, _ := strconv.ParseUint(parts[1], 10, 64)
		used, _ := strconv.ParseUint(parts[2], 10, 64)
		free, _ := strconv.ParseUint(parts[3], 10, 64)

		var usedPct float64
		if total > 0 {
			usedPct = float64(used) / float64(total) * 100
		}

		mount := DiskMount{
			Filesystem: fsTypes[mountpoint],
			Size:       formatBytes(total, "GB"),
			Used:       formatBytes(used, "GB"),
			Avail:      formatBytes(free, "GB"),
			Capacity:   fmt.Sprintf("%.1f%%", usedPct),
			MountPoint: mountpoint,
		}
		ds.AllMounts = append(ds.AllMounts, mount)

		if mountpoint == "/" || (ds.MountPoint != "/" && ds.Total == 0) {
			ds.MountPoint = mountpoint
			ds.Total = float64(total) / bytesInGB
			ds.Used = float64(used) / bytesInGB
			ds.Available = float64(free) / bytesInGB
			ds.Percentage = usedPct
		}
	}
	return ds
}

// parseNetDevText parses /proc/net/dev text into NetworkStats.
func parseNetDevText(text string) NetworkStats {
	ns := NetworkStats{Interfaces: []NetworkInterface{}}
	lines := strings.Split(text, "\n")
	for i := 2; i < len(lines); i++ { // skip 2 header lines
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 13 {
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

		iface := NetworkInterface{
			Name: name, BytesSent: bytesSent, BytesRecv: bytesRecv,
			PacketsSent: packetsSent, PacketsRecv: packetsRecv,
			ErrorIn: errin, ErrorOut: errout, DropIn: dropin, DropOut: dropout,
		}
		ns.Interfaces = append(ns.Interfaces, iface)
		ns.TotalBytesSent += bytesSent
		ns.TotalBytesRecv += bytesRecv
		ns.TotalPacketsSent += packetsSent
		ns.TotalPacketsRecv += packetsRecv
	}
	return ns
}

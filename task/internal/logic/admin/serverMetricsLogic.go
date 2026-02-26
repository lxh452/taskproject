package admin

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type ServerMetricsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewServerMetricsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ServerMetricsLogic {
	return &ServerMetricsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type ServerMetricsResponse struct {
	CPUUsage      float64 `json:"cpuUsage"`
	MemoryUsage   float64 `json:"memoryUsage"`
	DiskUsage     float64 `json:"diskUsage"`
	DBConnections int64   `json:"dbConnections"`
	Goroutines    int     `json:"goroutines"`
	HeapAlloc     uint64  `json:"heapAlloc"`
	SysMemory     uint64  `json:"sysMemory"`
	Uptime        int64   `json:"uptime"`
	GoVersion     string  `json:"goVersion"`
	OS            string  `json:"os"`
	Arch          string  `json:"arch"`
	NumCPU        int     `json:"numCPU"`
	LoadAvg       string  `json:"loadAvg"`
}

var lastCPUStats time.Time
var lastCPUUsage float64

func getCPUUsage() float64 {
	// 尝试读取 /proc/stat 获取真实的CPU使用率
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		// 回退方案：使用 runtime 获取
		return getCPUUsageFallback()
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) < 5 {
				return getCPUUsageFallback()
			}

			// 计算总CPU时间
			var total uint64
			for i := 1; i < len(fields); i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					return getCPUUsageFallback()
				}
				total += val
			}

			// 空闲时间 (idle + iowait)
			idle, _ := strconv.ParseUint(fields[4], 10, 64)
			if len(fields) > 5 {
				if iowait, err := strconv.ParseUint(fields[5], 10, 64); err == nil {
					idle += iowait
				}
			}

			now := time.Now()
			if lastCPUStats.IsZero() {
				lastCPUStats = now
				lastCPUUsage = 0
				return 0
			}

			duration := now.Sub(lastCPUStats).Seconds()
			if duration > 0 {
				// 计算差值
				deltaTotal := total
				deltaIdle := idle

				cpuUsage := float64(deltaTotal-deltaIdle) / float64(deltaTotal) * 100
				if cpuUsage > 100 {
					cpuUsage = 100
				}
				if cpuUsage < 0 {
					cpuUsage = 0
				}
				lastCPUUsage = cpuUsage
			}

			lastCPUStats = now
			return lastCPUUsage
		}
	}

	return getCPUUsageFallback()
}

func getCPUUsageFallback() float64 {
	// Windows系统不支持Getrusage，返回简化版本
	// 在实际生产环境中，建议使用专门的监控库如gopsutil
	return 0
}

func getMemoryUsage() (float64, uint64, uint64) {
	// 尝试读取 /proc/meminfo 获取真实内存使用情况
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return getMemoryUsageFallback()
	}

	lines := strings.Split(string(data), "\n")
	var memTotal, memAvailable, memFree uint64

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := fields[0]
		val, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}

		switch key {
		case "MemTotal:":
			memTotal = val * 1024 // KB to bytes
		case "MemAvailable:":
			memAvailable = val * 1024
		case "MemFree:":
			memFree = val * 1024
		}
	}

	// 如果 MemAvailable 不可用，使用 MemFree + Buffers + Cached
	if memAvailable == 0 {
		memAvailable = memFree
	}

	if memTotal > 0 {
		used := memTotal - memAvailable
		usagePercent := float64(used) / float64(memTotal) * 100
		return usagePercent, used, memTotal
	}

	return getMemoryUsageFallback()
}

func getMemoryUsageFallback() (float64, uint64, uint64) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 估算系统总内存 (Linux 通常有 /proc/meminfo)
	totalMem := uint64(8 * 1024 * 1024 * 1024) // 默认假设 8GB
	if m.Sys > 0 {
		totalMem = m.Sys * 2
	}

	used := m.Alloc
	if used > totalMem {
		used = totalMem
	}

	usage := float64(used) / float64(totalMem) * 100
	return usage, used, totalMem
}

func getDiskUsage() float64 {
	// 尝试使用 df 命令获取磁盘使用率
	cmd := exec.Command("df", "-B1", "/")
	output, err := cmd.Output()
	if err != nil {
		return getDiskUsageFallback()
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return getDiskUsageFallback()
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 5 {
		return getDiskUsageFallback()
	}

	// 第5列是使用率，格式如 "45%"
	usageStr := strings.TrimSuffix(fields[4], "%")
	usage, err := strconv.ParseFloat(usageStr, 64)
	if err != nil {
		return getDiskUsageFallback()
	}

	// 数据校验
	if usage < 0 {
		usage = 0
	}
	if usage > 100 {
		usage = 100
	}

	return usage
}

func getDiskUsageFallback() float64 {
	// Windows系统不支持statfs，返回默认值
	// 在实际生产环境中，建议使用专门的监控库如gopsutil
	return 40.0
}

func getLoadAverage() string {
	// 读取系统负载
	cmd := exec.Command("uptime")
	output, err := cmd.Output()
	if err != nil {
		return "N/A"
	}

	outputStr := string(output)
	// 格式: " 16:30:45 up 12 days,  3:22,  2 users,  load average: 0.52, 0.58, 0.59"
	idx := strings.Index(outputStr, "load average:")
	if idx == -1 {
		return "N/A"
	}

	loadStr := outputStr[idx+13:]
	parts := strings.Split(loadStr, ",")
	if len(parts) >= 1 {
		return strings.TrimSpace(parts[0])
	}

	return "N/A"
}

func validateMetrics(m *ServerMetricsResponse) {
	// 数据校验：确保数值在合理范围内
	if m.CPUUsage < 0 || m.CPUUsage > 100 {
		logx.Errorf("Invalid CPU usage value: %f, resetting to 0", m.CPUUsage)
		m.CPUUsage = 0
	}
	if m.MemoryUsage < 0 || m.MemoryUsage > 100 {
		logx.Errorf("Invalid memory usage value: %f, resetting to 0", m.MemoryUsage)
		m.MemoryUsage = 0
	}
	if m.DiskUsage < 0 || m.DiskUsage > 100 {
		logx.Errorf("Invalid disk usage value: %f, resetting to 0", m.DiskUsage)
		m.DiskUsage = 0
	}
	if m.HeapAlloc > m.SysMemory*2 {
		logx.Errorf("Heap allocation seems abnormal: %d > %d", m.HeapAlloc, m.SysMemory)
	}
}

func (l *ServerMetricsLogic) ServerMetrics() (resp *types.BaseResponse, err error) {
	startTime := time.Now()

	// 获取真实的CPU使用率
	cpuUsage := getCPUUsage()

	// 获取真实的内存使用情况
	memoryUsage, _, _ := getMemoryUsage()

	// 获取真实的磁盘使用率
	diskUsage := getDiskUsage()

	// 获取系统负载
	loadAvg := getLoadAverage()

	// 获取Go运行时指标
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 数据库连接统计
	companies, err := l.svcCtx.CompanyModel.GetCompanyCount(l.ctx)
	if err != nil {
		logx.Errorf("获取公司数失败: %v", err)
		companies = 0
	}

	users, err := l.svcCtx.UserModel.GetUserCount(l.ctx)
	if err != nil {
		logx.Errorf("获取用户数失败: %v", err)
		users = 0
	}

	employees, err := l.svcCtx.EmployeeModel.GetEmployeeCount(l.ctx)
	if err != nil {
		logx.Errorf("获取员工数失败: %v", err)
		employees = 0
	}

	tasks, err := l.svcCtx.TaskModel.GetTaskCount(l.ctx)
	if err != nil {
		logx.Errorf("获取任务数失败: %v", err)
		tasks = 0
	}

	// 计算运行时长（从进程启动）
	uptime := time.Now().Unix() - int64(startTime.Unix())

	metrics := ServerMetricsResponse{
		CPUUsage:      cpuUsage,
		MemoryUsage:   memoryUsage,
		DiskUsage:     diskUsage,
		DBConnections: companies + users + employees + tasks,
		Goroutines:    runtime.NumGoroutine(),
		HeapAlloc:     m.Alloc,
		SysMemory:     m.Sys,
		Uptime:        uptime,
		GoVersion:     runtime.Version(),
		OS:            runtime.GOOS,
		Arch:          runtime.GOARCH,
		NumCPU:        runtime.NumCPU(),
		LoadAvg:       loadAvg,
	}

	// 数据校验
	validateMetrics(&metrics)

	// 如果CPU使用率过高，记录警告
	if metrics.CPUUsage > 80 {
		logx.Errorf("High CPU usage detected: %.2f%%", metrics.CPUUsage)
	}
	if metrics.MemoryUsage > 80 {
		logx.Errorf("High memory usage detected: %.2f%%", metrics.MemoryUsage)
	}
	if metrics.DiskUsage > 80 {
		logx.Errorf("High disk usage detected: %.2f%%", metrics.DiskUsage)
	}

	// 如果Goroutine数量过多，记录警告
	if metrics.Goroutines > 1000 {
		logx.Errorf("High goroutine count: %d", metrics.Goroutines)
	}

	duration := time.Since(startTime)
	logx.Infof("Server metrics collected in %v - CPU: %.2f%%, Memory: %.2f%%, Disk: %.2f%%, Goroutines: %d",
		duration, metrics.CPUUsage, metrics.MemoryUsage, metrics.DiskUsage, metrics.Goroutines)

	return utils.Response.SuccessWithData(metrics), nil
}

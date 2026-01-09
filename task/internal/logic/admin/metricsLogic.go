package admin

import (
	"context"
	"runtime"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/zeromicro/go-zero/core/logx"
)

type MetricsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取服务器性能指标
func NewMetricsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MetricsLogic {
	return &MetricsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MetricsLogic) GetMetrics() (resp *types.BaseResponse, err error) {
	// 1. 获取CPU使用率
	cpuPercent, err := cpu.Percent(0, false)
	cpuUsage := 0.0
	if err == nil && len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	} else {
		logx.Errorf("获取CPU使用率失败: %v", err)
	}

	// 2. 获取内存使用率
	memInfo, err := mem.VirtualMemory()
	memoryUsage := 0.0
	var memoryTotal int64 = 0
	var memoryUsed int64 = 0
	if err == nil {
		memoryUsage = memInfo.UsedPercent
		memoryTotal = int64(memInfo.Total)
		memoryUsed = int64(memInfo.Used)
	} else {
		logx.Errorf("获取内存使用率失败: %v", err)
	}

	// 3. 获取磁盘使用率
	diskInfo, err := disk.Usage("/")
	diskUsage := 0.0
	var diskTotal int64 = 0
	var diskUsed int64 = 0
	if err == nil {
		diskUsage = diskInfo.UsedPercent
		diskTotal = int64(diskInfo.Total)
		diskUsed = int64(diskInfo.Used)
	} else {
		logx.Errorf("获取磁盘使用率失败: %v", err)
	}

	// 4. 获取Go协程数
	goroutineCount := runtime.NumGoroutine()

	metricsResp := types.ServerMetricsResponse{
		CPUUsage:    cpuUsage,
		MemoryUsage: memoryUsage,
		MemoryTotal: memoryTotal,
		MemoryUsed:  memoryUsed,
		DiskUsage:   diskUsage,
		DiskTotal:   diskTotal,
		DiskUsed:    diskUsed,
		GoRoutines:  goroutineCount,
		ActiveConns: 0,
		MySQLConns:  0,
		RedisConns:  0,
		Uptime:      0,
	}

	return utils.Response.SuccessWithData(metricsResp), nil
}

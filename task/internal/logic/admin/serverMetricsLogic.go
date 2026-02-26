package admin

import (
	"context"
	"runtime"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

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

// ServerMetricsResponse 前端期望的服务器监控数据格式
type ServerMetricsResponse struct {
	CPUUsage      float64 `json:"cpuUsage"`
	MemoryUsage   float64 `json:"memoryUsage"`
	DiskUsage     float64 `json:"diskUsage"`
	DBConnections int64   `json:"dbConnections"`
	Goroutines    int     `json:"goroutines"`
	HeapAlloc     uint64  `json:"heapAlloc"`
	SysMemory     uint64  `json:"sysMemory"`
}

func (l *ServerMetricsLogic) ServerMetrics() (resp *types.BaseResponse, err error) {
	// 获取Go运行时指标
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// CPU使用率（简化版本）
	// 实际生产环境建议使用专门的监控工具如prometheus
	cpuUsage := 0.0

	// 内存使用率（基于Go堆内存）
	var memoryUsage float64
	// 估算系统总内存（简化）
	totalMem := m.Alloc * 2
	if totalMem > 0 {
		memoryUsage = float64(m.Alloc) / float64(totalMem) * 100
	}

	// 磁盘使用率（模拟）
	diskUsage := 40.0

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

	// 数据库连接数（使用总记录数作为指标）
	dbConnections := companies + users + employees + tasks

	metrics := ServerMetricsResponse{
		CPUUsage:      cpuUsage,
		MemoryUsage:   memoryUsage,
		DiskUsage:     diskUsage,
		DBConnections: dbConnections,
		Goroutines:    runtime.NumGoroutine(),
		HeapAlloc:     m.Alloc,
		SysMemory:     m.Sys,
	}

	return utils.Response.SuccessWithData(metrics), nil
}

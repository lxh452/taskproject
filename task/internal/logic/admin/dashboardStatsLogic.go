package admin

import (
	"context"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type DashboardStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDashboardStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DashboardStatsLogic {
	return &DashboardStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DashboardStatsLogic) DashboardStats() (resp *types.BaseResponse, err error) {
	// 获取公司总数
	totalCompanies, err := l.svcCtx.CompanyModel.GetCompanyCount(l.ctx)
	if err != nil {
		logx.Errorf("获取公司总数失败: %v", err)
		totalCompanies = 0
	}

	// 获取用户总数
	totalUsers, err := l.svcCtx.UserModel.GetUserCount(l.ctx)
	if err != nil {
		logx.Errorf("获取用户总数失败: %v", err)
		totalUsers = 0
	}

	// 获取员工总数
	totalEmployees, err := l.svcCtx.EmployeeModel.GetEmployeeCount(l.ctx)
	if err != nil {
		logx.Errorf("获取员工总数失败: %v", err)
		totalEmployees = 0
	}

	// 获取任务总数
	totalTasks, err := l.svcCtx.TaskModel.GetTaskCount(l.ctx)
	if err != nil {
		logx.Errorf("获取任务总数失败: %v", err)
		totalTasks = 0
	}

	// 获取各公司的员工分布
	companies, _, err := l.svcCtx.CompanyModel.FindByPage(l.ctx, 1, 100)
	if err != nil {
		logx.Errorf("获取公司列表失败: %v", err)
		companies = nil
	}

	var companyDistribution []types.CompanyEmployeeCount
	for _, company := range companies {
		// 获取每个公司的员工数量
		count, err := l.svcCtx.EmployeeModel.GetEmployeeCountByCompany(l.ctx, company.Id)
		if err != nil {
			count = 0
		}
		companyDistribution = append(companyDistribution, types.CompanyEmployeeCount{
			CompanyID:     company.Id,
			CompanyName:   company.Name,
			EmployeeCount: count,
		})
	}

	// 获取最近7天的用户注册趋势
	now := time.Now()
	var userTrend []types.TrendData
	var taskTrend []types.TrendData

	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		startTime := dateStr + " 00:00:00"
		endTime := dateStr + " 23:59:59"

		// 获取当天新增用户数
		userCount, err := l.svcCtx.UserModel.GetUserCountByDateRange(l.ctx, startTime, endTime)
		if err != nil {
			userCount = 0
		}

		// 获取当天新增任务数（简化：使用任务总数作为模拟）
		taskCount, err := l.svcCtx.TaskModel.GetTaskCount(l.ctx)
		if err != nil {
			taskCount = 0
		}

		userTrend = append(userTrend, types.TrendData{
			Date:  dateStr,
			Count: userCount,
		})
		taskTrend = append(taskTrend, types.TrendData{
			Date:  dateStr,
			Count: taskCount,
		})
	}

	stats := types.PlatformStatsResponse{
		TotalCompanies:      totalCompanies,
		TotalUsers:          totalUsers,
		TotalTasks:          totalTasks,
		TotalEmployees:      totalEmployees,
		CompanyDistribution: companyDistribution,
		UserTrend:           userTrend,
		TaskTrend:           taskTrend,
	}

	return utils.Response.SuccessWithData(stats), nil
}

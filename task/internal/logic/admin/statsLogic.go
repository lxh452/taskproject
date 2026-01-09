package admin

import (
	"context"
	"fmt"
	"time"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPlatformStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取平台统计概览
func NewGetPlatformStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPlatformStatsLogic {
	return &GetPlatformStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPlatformStatsLogic) GetPlatformStats() (resp *types.BaseResponse, err error) {
	// 1. 获取公司总数
	totalCompanies, err := l.svcCtx.CompanyModel.GetCompanyCount(l.ctx)
	if err != nil {
		logx.Errorf("获取公司总数失败: %v", err)
		totalCompanies = 0
	}

	// 2. 获取用户总数
	totalUsers, err := l.svcCtx.UserModel.GetUserCount(l.ctx)
	if err != nil {
		logx.Errorf("获取用户总数失败: %v", err)
		totalUsers = 0
	}

	// 3. 获取任务总数
	totalTasks, err := l.svcCtx.TaskModel.GetTaskCount(l.ctx)
	if err != nil {
		logx.Errorf("获取任务总数失败: %v", err)
		totalTasks = 0
	}

	// 4. 获取员工总数
	totalEmployees, err := l.svcCtx.EmployeeModel.GetEmployeeCount(l.ctx)
	if err != nil {
		logx.Errorf("获取员工总数失败: %v", err)
		totalEmployees = 0
	}

	// 5. 获取公司员工分布
	companyDistribution, err := l.getCompanyEmployeeDistribution()
	if err != nil {
		logx.Errorf("获取公司员工分布失败: %v", err)
		companyDistribution = []types.CompanyEmployeeCount{}
	}

	// 6. 获取用户注册趋势（最近30天）
	userTrend, err := l.getUserRegistrationTrend(30)
	if err != nil {
		logx.Errorf("获取用户注册趋势失败: %v", err)
		userTrend = []types.TrendData{}
	}

	// 7. 获取任务创建趋势（最近30天）
	taskTrend, err := l.getTaskCreationTrend(30)
	if err != nil {
		logx.Errorf("获取任务创建趋势失败: %v", err)
		taskTrend = []types.TrendData{}
	}

	statsResp := types.PlatformStatsResponse{
		TotalCompanies:      totalCompanies,
		TotalUsers:          totalUsers,
		TotalTasks:          totalTasks,
		TotalEmployees:      totalEmployees,
		CompanyDistribution: companyDistribution,
		UserTrend:           userTrend,
		TaskTrend:           taskTrend,
	}

	return utils.Response.SuccessWithData(statsResp), nil
}

// getCompanyEmployeeDistribution 获取公司员工分布
func (l *GetPlatformStatsLogic) getCompanyEmployeeDistribution() ([]types.CompanyEmployeeCount, error) {
	// 获取所有公司
	companies, _, err := l.svcCtx.CompanyModel.FindByPage(l.ctx, 1, 100)
	if err != nil {
		return nil, err
	}

	var distribution []types.CompanyEmployeeCount
	for _, company := range companies {
		// 获取每个公司的员工数量
		count, err := l.svcCtx.EmployeeModel.GetEmployeeCountByCompany(l.ctx, company.Id)
		if err != nil {
			logx.Errorf("获取公司 %s 员工数量失败: %v", company.Id, err)
			count = 0
		}

		distribution = append(distribution, types.CompanyEmployeeCount{
			CompanyID:     company.Id,
			CompanyName:   company.Name,
			EmployeeCount: count,
		})
	}

	return distribution, nil
}

// getUserRegistrationTrend 获取用户注册趋势
func (l *GetPlatformStatsLogic) getUserRegistrationTrend(days int) ([]types.TrendData, error) {
	var trend []types.TrendData
	now := time.Now()

	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		startTime := fmt.Sprintf("%s 00:00:00", dateStr)
		endTime := fmt.Sprintf("%s 23:59:59", dateStr)

		count, err := l.svcCtx.UserModel.GetUserCountByDateRange(l.ctx, startTime, endTime)
		if err != nil {
			logx.Errorf("获取 %s 用户注册数量失败: %v", dateStr, err)
			count = 0
		}

		trend = append(trend, types.TrendData{
			Date:  dateStr,
			Count: count,
		})
	}

	return trend, nil
}

// getTaskCreationTrend 获取任务创建趋势
func (l *GetPlatformStatsLogic) getTaskCreationTrend(days int) ([]types.TrendData, error) {
	var trend []types.TrendData
	now := time.Now()

	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		startTime := fmt.Sprintf("%s 00:00:00", dateStr)
		endTime := fmt.Sprintf("%s 23:59:59", dateStr)

		count, err := l.svcCtx.TaskModel.GetTaskCountByDateRange(l.ctx, startTime, endTime)
		if err != nil {
			logx.Errorf("获取 %s 任务创建数量失败: %v", dateStr, err)
			count = 0
		}

		trend = append(trend, types.TrendData{
			Date:  dateStr,
			Count: count,
		})
	}

	return trend, nil
}

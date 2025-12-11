package task

import (
	"context"
	"errors"

	"task_Project/model/task"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTaskListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTaskListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTaskListLogic {
	return &GetTaskListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTaskListLogic) GetTaskList(req *types.TaskListRequest) (resp *types.BaseResponse, err error) {
	// 1. 参数验证
	validator := utils.NewValidator()
	page, pageSize, errs := validator.ValidatePageParams(req.Page, req.PageSize)
	if len(errs) > 0 {
		return utils.Response.ValidationError(errs[0]), nil
	}

	employeeId, ok := utils.Common.GetCurrentEmployeeID(l.ctx)
	if !ok || employeeId == "" {
		return nil, errors.New("获取员工信息失败，请重新登录后再试")
	}

	// 3. 根据用户角色获取任务列表
	var tasks []*task.Task
	var total int64

	// 根据查询条件获取任务
	if req.CompanyID != "" {
		// 查询指定公司的任务
		tasks, total, err = l.svcCtx.TaskModel.FindByCompany(l.ctx, req.CompanyID, page, pageSize)
		if err != nil {
			return nil, err
		}
	} else if req.DepartmentID != "" {
		// 查询指定部门的任务
		tasks, total, err = l.svcCtx.TaskModel.FindByDepartment(l.ctx, req.DepartmentID, page, pageSize)
		if err != nil {
			return nil, err
		}
	} else {
		// 查询用户参与的任务：创建者/负责人/节点执行人（支持多员工字段）
		tasks, total, err = l.svcCtx.TaskModel.FindByInvolved(l.ctx, employeeId, page, pageSize)
		if err != nil {
			return nil, err
		}

	}

	// 4. 过滤和分页
	var filteredTasks []*task.Task
	for _, t := range tasks {
		// 状态过滤
		if req.Status > 0 && t.TaskStatus != int64(req.Status) {
			continue
		}
		// 优先级过滤
		if req.Priority > 0 && t.TaskPriority != int64(req.Priority) {
			continue
		}
		filteredTasks = append(filteredTasks, t)
	}

	total = int64(len(filteredTasks))

	// 分页
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= int(total) {
		filteredTasks = []*task.Task{}
	} else {
		if end > int(total) {
			end = int(total)
		}
		filteredTasks = filteredTasks[start:end]
	}

	// 5. 转换为响应格式
	converter := utils.NewConverter()
	taskInfos := converter.ToTaskInfoList(filteredTasks)

	// 6. 构建分页响应
	pageResponse := converter.ToPageResponse(taskInfos, int(total), page, pageSize)

	return utils.Response.Success(pageResponse), nil
}

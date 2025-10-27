// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package position

import (
	"context"
	"database/sql"
	"time"

	"task_Project/model/company"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreatePositionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建职位
func NewCreatePositionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePositionLogic {
	return &CreatePositionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreatePositionLogic) CreatePosition(req *types.CreatePositionRequest) (resp *types.BaseResponse, err error) {
	// 参数验证
	validator := utils.NewValidator()
	if validator.IsEmpty(req.DepartmentID) {
		return utils.Response.ValidationError("部门ID不能为空"), nil
	}
	if validator.IsEmpty(req.PositionName) {
		return utils.Response.ValidationError("职位名称不能为空"), nil
	}
	if req.PositionLevel <= 0 {
		return utils.Response.ValidationError("职位级别必须大于0"), nil
	}

	// 检查部门是否存在
	_, err = l.svcCtx.DepartmentModel.FindOne(l.ctx, req.DepartmentID)
	if err != nil {
		logx.Errorf("查询部门失败: %v", err)
		return utils.Response.ErrorWithKey("department_not_found"), nil
	}

	// 生成职位ID
	common := utils.NewCommon()
	positionID := common.GenerateID()

	// 创建职位
	position := &company.Position{
		Id:               positionID,
		DepartmentId:     req.DepartmentID,
		PositionName:     req.PositionName,
		PositionLevel:    int64(req.PositionLevel),
		JobDescription:   utils.Common.ToSqlNullString(req.JobDescription),
		SalaryRangeMin:   utils.Common.ToSqlNullFloat64(req.SalaryRangeMin),
		SalaryRangeMax:   utils.Common.ToSqlNullFloat64(req.SalaryRangeMax),
		IsManagement:     int64(req.IsManagement),
		MaxEmployees:     int64(req.MaxEmployees),
		CurrentEmployees: 0,
		Status:           1, // 正常状态
		CreateTime:       time.Now(),
		UpdateTime:       time.Now(),
		DeleteTime:       sql.NullTime{Time: time.Time{}, Valid: false},
	}

	_, err = l.svcCtx.PositionModel.Insert(l.ctx, position)
	if err != nil {
		logx.Errorf("创建职位失败: %v", err)
		return utils.Response.InternalError("创建职位失败"), err
	}

	return utils.Response.SuccessWithKey("positionId", positionID), nil
}

// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package position

import (
	"context"
	"database/sql"
	"time"

	"task_Project/model/company"
	"task_Project/model/role"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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

	// 使用事务创建职位和角色关联
	err = l.svcCtx.TransactionService.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 创建职位
		positionModelWithSession := l.svcCtx.TransactionHelper.GetPositionModelWithSession(session)
		_, err := positionModelWithSession.Insert(ctx, position)
		if err != nil {
			return err
		}

		// 如果提供了角色ID列表，批量创建职位角色关联
		if len(req.RoleIds) > 0 {
			positionRoleModelWithSession := l.svcCtx.TransactionHelper.GetPositionRoleModelWithSession(session)
			userId, _ := utils.Common.GetCurrentUserID(ctx)

			for _, roleId := range req.RoleIds {
				// 验证角色是否存在
				_, err := l.svcCtx.RoleModel.FindOne(ctx, roleId)
				if err != nil {
					logx.Errorf("角色不存在 roleId=%s: %v", roleId, err)
					continue // 跳过不存在的角色，继续处理其他角色
				}

				// 检查是否已经存在关联
				_, err = positionRoleModelWithSession.FindOneByPositionIdRoleId(ctx, positionID, roleId)
				if err == nil {
					logx.Infof("职位角色关联已存在 positionId=%s roleId=%s", positionID, roleId)
					continue // 已存在，跳过
				}

				// 创建职位角色关联
				positionRole := &role.PositionRole{
					Id:         utils.Common.GenId("pr"),
					PositionId: positionID,
					RoleId:     roleId,
					GrantBy:    sql.NullString{String: userId, Valid: userId != ""},
					GrantTime:  time.Now(),
					ExpireTime: sql.NullTime{Valid: false},
					Status:     1,
				}
				_, err = positionRoleModelWithSession.Insert(ctx, positionRole)
				if err != nil {
					logx.Errorf("创建职位角色关联失败 positionId=%s roleId=%s: %v", positionID, roleId, err)
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		logx.Errorf("创建职位失败: %v", err)
		return utils.Response.InternalError("创建职位失败"), err
	}

	return utils.Response.SuccessWithKey("positionId", positionID), nil
}

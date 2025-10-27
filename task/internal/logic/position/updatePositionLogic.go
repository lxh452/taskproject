// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package position

import (
	"context"
	"task_Project/model/company"
	"time"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatePositionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新职位信息
func NewUpdatePositionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePositionLogic {
	return &UpdatePositionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdatePositionLogic) UpdatePosition(req *types.UpdatePositionRequest) (resp *types.BaseResponse, err error) {
	// 参数验证
	validator := utils.NewValidator()
	if validator.IsEmpty(req.ID) {
		return utils.Response.ValidationError("职位ID不能为空"), nil
	}

	// 检查职位是否存在
	if _, err = l.svcCtx.PositionModel.FindOne(l.ctx, req.ID); err != nil {
		logx.Errorf("查询职位失败: %v", err)
		return utils.Response.ErrorWithKey("position_not_found"), nil
	}

	// 更新职位信息
	updateData := l.updateData(req)

	if len(updateData) == 0 {
		return utils.Response.ValidationError("没有需要更新的字段"), nil
	}

	updateData["update_time"] = time.Now()
	var position company.Position
	err = utils.Common.MapToStructWithMapstructure(updateData, &position)
	if err != nil {
		logx.Errorf("转换结构体失败: %v", err)
		return utils.Response.ErrorWithKey("position_not_found"), nil
	}
	position.Id = req.ID
	err = l.svcCtx.PositionModel.Update(l.ctx, &position)
	if err != nil {
		logx.Errorf("更新职位信息失败: %v", err)
		return utils.Response.InternalError("更新职位信息失败"), err
	}

	return utils.Response.Success("更新职位信息成功"), nil
}

func (l *UpdatePositionLogic) updateData(req *types.UpdatePositionRequest) map[string]interface{} {
	updateData := make(map[string]interface{})
	if !utils.Validator.IsEmpty(req.PositionName) {
		updateData["position_name"] = req.PositionName
	}
	if !utils.Validator.IsEmpty(req.PositionCode) {
		updateData["position_code"] = req.PositionCode
	}
	if req.PositionLevel > 0 {
		updateData["position_level"] = req.PositionLevel
	}
	if !utils.Validator.IsEmpty(req.RequiredSkills) {
		updateData["required_skills"] = req.RequiredSkills
	}
	if !utils.Validator.IsEmpty(req.JobDescription) {
		updateData["job_description"] = req.JobDescription
	}
	if !utils.Validator.IsEmpty(req.Responsibilities) {
		updateData["responsibilities"] = req.Responsibilities
	}
	if !utils.Validator.IsEmpty(req.Requirements) {
		updateData["requirements"] = req.Requirements
	}
	if req.SalaryRangeMin > 0 {
		updateData["salary_range_min"] = req.SalaryRangeMin
	}
	if req.SalaryRangeMax > 0 {
		updateData["salary_range_max"] = req.SalaryRangeMax
	}
	if req.IsManagement >= 0 {
		updateData["is_management"] = req.IsManagement
	}
	if req.MaxEmployees > 0 {
		updateData["max_employees"] = req.MaxEmployees
	}
	return updateData
}

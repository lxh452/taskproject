package company

import (
	"context"
	"time"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetInviteCodeListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取邀请码列表
func NewGetInviteCodeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInviteCodeListLogic {
	return &GetInviteCodeListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetInviteCodeListLogic) GetInviteCodeList(req *types.GetInviteCodeListRequest) (resp *types.BaseResponse, err error) {
	// 获取当前用户ID
	userID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 获取当前员工信息
	employee, err := l.svcCtx.EmployeeModel.FindByUserID(l.ctx, userID)
	if err != nil || employee == nil {
		return utils.Response.BusinessError("您尚未加入任何公司"), nil
	}

	// 获取公司信息
	company, err := l.svcCtx.CompanyModel.FindOne(l.ctx, employee.CompanyId)
	if err != nil {
		logx.Errorf("查询公司失败: %v", err)
		return utils.Response.InternalError("查询公司失败"), nil
	}

	// 检查权限：只有公司创始人、人事部门或管理人员可以查看邀请码列表
	isFounder := company.Owner == userID
	isHR := false
	isManager := false

	if employee.DepartmentId.Valid {
		dept, _ := l.svcCtx.DepartmentModel.FindOne(l.ctx, employee.DepartmentId.String)
		if dept != nil && dept.DepartmentCode.Valid && dept.DepartmentCode.String == "HR" {
			isHR = true
		}
	}

	if employee.PositionId.Valid {
		pos, _ := l.svcCtx.PositionModel.FindOne(l.ctx, employee.PositionId.String)
		if pos != nil && pos.IsManagement == 1 {
			isManager = true
		}
	}

	if !isFounder && !isHR && !isManager {
		return utils.Response.BusinessError("只有公司创始人、人事部门或管理人员可以查看邀请码列表"), nil
	}

	// 获取邀请码列表
	inviteCodes, err := l.svcCtx.InviteCodeService.ListInviteCodesByCompany(l.ctx, employee.CompanyId)
	if err != nil {
		logx.Errorf("获取邀请码列表失败: %v", err)
		return utils.Response.InternalError("获取邀请码列表失败"), nil
	}

	logx.Infof("[GetInviteCodeList] 获取到邀请码数量: %d, companyId=%s", len(inviteCodes), employee.CompanyId)

	// 构建响应数据
	var list []types.InviteCodeInfo
	now := time.Now().Unix()

	for _, code := range inviteCodes {
		// 确定状态
		status := "active"
		if now > code.ExpireAt {
			status = "expired"
		} else if code.MaxUses > 0 && code.UsedCount >= code.MaxUses {
			status = "exhausted"
		}

		// 获取创建者名称
		creatorName := ""
		if code.CreatedBy != "" {
			creator, _ := l.svcCtx.EmployeeModel.FindOne(l.ctx, code.CreatedBy)
			if creator != nil {
				creatorName = creator.RealName
			}
		}

		// 计算过期天数
		expireDays := int((code.ExpireAt - code.CreatedAt) / (24 * 60 * 60))

		list = append(list, types.InviteCodeInfo{
			InviteCode:  code.Code,
			CompanyID:   code.CompanyID,
			CompanyName: code.CompanyName,
			CreatedBy:   code.CreatedBy,
			CreatorName: creatorName,
			ExpireDays:  expireDays,
			ExpireAt:    time.Unix(code.ExpireAt, 0).Format("2006-01-02 15:04:05"),
			MaxUses:     code.MaxUses,
			UsedCount:   code.UsedCount,
			Status:      status,
			CreateTime:  time.Unix(code.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		})
	}

	// 分页处理
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	total := int64(len(list))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start > len(list) {
		start = len(list)
	}
	if end > len(list) {
		end = len(list)
	}

	pagedList := list[start:end]

	logx.Infof("[GetInviteCodeList] 返回数据: total=%d, page=%d, pageSize=%d, listLen=%d", total, page, pageSize, len(pagedList))

	return utils.Response.Success(map[string]interface{}{
		"list":  pagedList,
		"total": total,
	}), nil
}

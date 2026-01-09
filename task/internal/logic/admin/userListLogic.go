package admin

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 管理员获取用户列表
func NewUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserListLogic {
	return &UserListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserListLogic) UserList(req *types.AdminUserListRequest) (resp *types.BaseResponse, err error) {
	// 设置默认分页参数
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 查询用户列表
	var users []types.AdminUserInfo
	var total int64

	if req.Username != "" {
		// 搜索用户
		userList, count, err := l.svcCtx.UserModel.SearchUsers(l.ctx, req.Username, page, pageSize)
		if err != nil {
			logx.Errorf("搜索用户失败: %v", err)
			return utils.Response.Error(500, "搜索用户失败"), nil
		}
		total = count

		for _, u := range userList {
			userInfo := types.AdminUserInfo{
				ID:         u.Id,
				Username:   u.Username,
				Email:      u.Email.String,
				Phone:      u.Phone.String,
				RealName:   u.RealName.String,
				Avatar:     u.Avatar.String,
				Status:     int(u.Status),
				CreateTime: u.CreateTime.Format("2006-01-02 15:04:05"),
			}

			// 获取用户关联的员工信息
			employee, err := l.svcCtx.EmployeeModel.FindByUserID(l.ctx, u.Id)
			if err == nil && employee != nil {
				userInfo.CompanyID = employee.CompanyId
				// 获取公司名称
				company, err := l.svcCtx.CompanyModel.FindOne(l.ctx, employee.CompanyId)
				if err == nil && company != nil {
					userInfo.CompanyName = company.Name
				}
			}

			users = append(users, userInfo)
		}
	} else {
		// 分页查询所有用户
		userList, count, err := l.svcCtx.UserModel.FindByPage(l.ctx, page, pageSize)
		if err != nil {
			logx.Errorf("查询用户列表失败: %v", err)
			return utils.Response.Error(500, "查询用户列表失败"), nil
		}
		total = count

		for _, u := range userList {
			userInfo := types.AdminUserInfo{
				ID:         u.Id,
				Username:   u.Username,
				Email:      u.Email.String,
				Phone:      u.Phone.String,
				RealName:   u.RealName.String,
				Avatar:     u.Avatar.String,
				Status:     int(u.Status),
				CreateTime: u.CreateTime.Format("2006-01-02 15:04:05"),
			}

			// 获取用户关联的员工信息
			employee, err := l.svcCtx.EmployeeModel.FindByUserID(l.ctx, u.Id)
			if err == nil && employee != nil {
				userInfo.CompanyID = employee.CompanyId
				// 获取公司名称
				company, err := l.svcCtx.CompanyModel.FindOne(l.ctx, employee.CompanyId)
				if err == nil && company != nil {
					userInfo.CompanyName = company.Name
				}
			}

			users = append(users, userInfo)
		}
	}

	// 按状态筛选
	if req.Status != 0 {
		var filtered []types.AdminUserInfo
		for _, u := range users {
			if u.Status == req.Status {
				filtered = append(filtered, u)
			}
		}
		users = filtered
	}

	// 按公司筛选
	if req.CompanyID != "" {
		var filtered []types.AdminUserInfo
		for _, u := range users {
			if u.CompanyID == req.CompanyID {
				filtered = append(filtered, u)
			}
		}
		users = filtered
	}

	return utils.Response.SuccessWithData(map[string]interface{}{
		"list":     users,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}), nil
}

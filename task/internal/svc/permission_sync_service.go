package svc

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"task_Project/model/user_auth"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

// PermissionSyncService 权限同步服务
// 用于将职位关联的角色权限同步到 user_permission 表
type PermissionSyncService struct {
	svcCtx *ServiceContext
}

// NewPermissionSyncService 创建权限同步服务
func NewPermissionSyncService(svcCtx *ServiceContext) *PermissionSyncService {
	return &PermissionSyncService{
		svcCtx: svcCtx,
	}
}

// SyncEmployeePermissions 同步员工权限（通过职位->角色->权限）
// userId: 用户ID
// employeeId: 员工ID
// positionId: 职位ID（如果为空，则清除该员工的所有权限）
func (s *PermissionSyncService) SyncEmployeePermissions(ctx context.Context, userId, employeeId, positionId string) error {
	// 如果职位ID为空，清除该员工的所有角色授权权限
	if positionId == "" {
		return s.clearRoleGrantedPermissions(ctx, userId)
	}

	// 通过职位查询角色
	roles, err := s.svcCtx.PositionRoleModel.ListRolesByPositionId(ctx, positionId)
	if err != nil {
		logx.Errorf("查询职位角色失败: %v", err)
		return err
	}

	// 先清除该员工的所有角色授权权限（grant_type=1）
	if err := s.clearRoleGrantedPermissions(ctx, userId); err != nil {
		logx.Errorf("清除旧权限失败: %v", err)
		return err
	}

	// 为每个角色的权限创建 user_permission 记录
	for _, role := range roles {
		if role.Status != 1 || !role.Permissions.Valid || role.Permissions.String == "" {
			continue
		}

		// 解析权限（JSON数组格式，如 [1,2,3]）
		var permCodes []int
		if err := json.Unmarshal([]byte(role.Permissions.String), &permCodes); err != nil {
			logx.Errorf("解析角色权限失败 roleId=%s: %v", role.Id, err)
			continue
		}

		// 为每个权限码创建 user_permission 记录
		for _, permCode := range permCodes {
			permissionId := utils.Common.GenerateID()
			permission := &user_auth.UserPermission{
				Id:             permissionId,
				UserId:         userId,
				PermissionCode: strconv.Itoa(permCode),
				PermissionName: s.getPermissionName(permCode),
				ResourceType:   2,                                            // 2-接口
				ResourceId:     sql.NullString{String: role.Id, Valid: true}, // 关联角色ID
				GrantType:      1,                                            // 1-角色授权
				GrantBy:        sql.NullString{Valid: false},                 // 系统自动授权
				ExpireTime:     sql.NullTime{Valid: false},
				Status:         1,
				CreateTime:     time.Now(),
				UpdateTime:     time.Now(),
			}

			_, err := s.svcCtx.UserPermissionModel.Insert(ctx, permission)
			if err != nil {
				logx.Errorf("插入用户权限失败 userId=%s permissionCode=%d: %v", userId, permCode, err)
				// 继续处理其他权限，不中断
				continue
			}
		}
	}

	logx.Infof("同步员工权限成功 userId=%s employeeId=%s positionId=%s roles=%d", userId, employeeId, positionId, len(roles))
	return nil
}

// clearRoleGrantedPermissions 清除用户的所有角色授权权限（grant_type=1）
func (s *PermissionSyncService) clearRoleGrantedPermissions(ctx context.Context, userId string) error {
	// 直接删除该用户的所有角色授权权限
	// UserPermissionModel 接口已经包含了 DeleteByUserIdAndGrantType 方法
	return s.svcCtx.UserPermissionModel.DeleteByUserIdAndGrantType(ctx, userId, 1)
}

// getPermissionName 根据权限码获取权限名称
func (s *PermissionSyncService) getPermissionName(permCode int) string {
	// 这里可以根据权限码映射到权限名称
	// 暂时返回通用名称，后续可以扩展
	permNames := map[int]string{
		1:  "任务创建",
		2:  "任务更新",
		3:  "任务删除",
		4:  "任务审批",
		5:  "任务节点创建",
		6:  "任务节点更新",
		7:  "任务节点删除",
		8:  "交接创建",
		9:  "交接审批",
		10: "交接拒绝",
		11: "公司创建",
		12: "公司更新",
		13: "公司删除",
		14: "部门创建",
		15: "部门更新",
		16: "部门删除",
		17: "职位创建",
		18: "职位更新",
		19: "职位删除",
		20: "角色创建",
		21: "角色更新",
		22: "角色删除",
		23: "角色分配",
		24: "角色撤销",
		25: "员工创建",
		26: "员工更新",
		27: "员工删除",
		28: "员工离职",
		29: "通知创建",
		30: "通知删除",
	}
	if name, ok := permNames[permCode]; ok {
		return name
	}
	return "权限" + strconv.Itoa(permCode)
}

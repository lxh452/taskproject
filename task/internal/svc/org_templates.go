package svc

import (
	"context"
	"database/sql"
	"time"

	"task_Project/model/company"
	roleModel "task_Project/model/role"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ApplyDefaultOrgStructureWithSession 在指定事务中初始化组织结构模板
func (s *ServiceContext) ApplyDefaultOrgStructureWithSession(ctx context.Context, session sqlx.Session, companyID string) error {
	type posTpl struct {
		Name    string
		Code    string
		Level   int64
		JobType int64
		IsMgmt  int64
	}

	type deptTpl struct {
		Name      string
		Code      string
		Positions []posTpl // 每个部门可以有自己特定的职位
		Children  []deptTpl
	}

	// 定义每个部门的专属职位（避免重复）
	roots := []deptTpl{
		{
			Name: "人力资源部", Code: "HR",
			Positions: []posTpl{
				{Name: "人力资源经理", Code: "HR_MGR", Level: 4, JobType: 2, IsMgmt: 1},
				{Name: "招聘专员", Code: "HR_REC", Level: 2, JobType: 1, IsMgmt: 0},
				{Name: "培训专员", Code: "HR_TRN", Level: 2, JobType: 1, IsMgmt: 0},
			},
		},
		{
			Name: "研发部", Code: "RND",
			Positions: []posTpl{
				{Name: "研发总监", Code: "RND_DIR", Level: 5, JobType: 2, IsMgmt: 1},
				{Name: "技术经理", Code: "RND_MGR", Level: 4, JobType: 2, IsMgmt: 1},
			},
			Children: []deptTpl{
				{
					Name: "后端组", Code: "BE",
					Positions: []posTpl{
						{Name: "后端组长", Code: "BE_LEAD", Level: 3, JobType: 2, IsMgmt: 1},
						{Name: "高级后端工程师", Code: "BE_SEN", Level: 3, JobType: 0, IsMgmt: 0},
						{Name: "后端工程师", Code: "BE_ENG", Level: 2, JobType: 0, IsMgmt: 0},
					},
				},
				{
					Name: "前端组", Code: "FE",
					Positions: []posTpl{
						{Name: "前端组长", Code: "FE_LEAD", Level: 3, JobType: 2, IsMgmt: 1},
						{Name: "高级前端工程师", Code: "FE_SEN", Level: 3, JobType: 0, IsMgmt: 0},
						{Name: "前端工程师", Code: "FE_ENG", Level: 2, JobType: 0, IsMgmt: 0},
					},
				},
			},
		},
		{
			Name: "市场部", Code: "MKT",
			Positions: []posTpl{
				{Name: "市场经理", Code: "MKT_MGR", Level: 4, JobType: 2, IsMgmt: 1},
				{Name: "市场专员", Code: "MKT_SPE", Level: 2, JobType: 1, IsMgmt: 0},
			},
		},
		{
			Name: "销售部", Code: "SLS",
			Positions: []posTpl{
				{Name: "销售经理", Code: "SLS_MGR", Level: 4, JobType: 2, IsMgmt: 1},
				{Name: "销售代表", Code: "SLS_REP", Level: 2, JobType: 1, IsMgmt: 0},
			},
		},
		{
			Name: "财务部", Code: "FIN",
			Positions: []posTpl{
				{Name: "财务经理", Code: "FIN_MGR", Level: 4, JobType: 2, IsMgmt: 1},
				{Name: "会计", Code: "FIN_ACC", Level: 2, JobType: 1, IsMgmt: 0},
				{Name: "出纳", Code: "FIN_CSH", Level: 2, JobType: 1, IsMgmt: 0},
			},
		},
		{
			Name: "行政部", Code: "ADM",
			Positions: []posTpl{
				{Name: "行政经理", Code: "ADM_MGR", Level: 4, JobType: 2, IsMgmt: 1},
				{Name: "行政专员", Code: "ADM_SPE", Level: 2, JobType: 1, IsMgmt: 0},
			},
		},
		{
			Name: "运维部", Code: "OPS",
			Positions: []posTpl{
				{Name: "运维经理", Code: "OPS_MGR", Level: 4, JobType: 2, IsMgmt: 1},
				{Name: "高级运维工程师", Code: "OPS_SEN", Level: 3, JobType: 0, IsMgmt: 0},
				{Name: "运维工程师", Code: "OPS_ENG", Level: 2, JobType: 0, IsMgmt: 0},
			},
		},
	}

	depModel := s.TransactionHelper.GetDepartmentModelWithSession(session)
	posModel := s.TransactionHelper.GetPositionModelWithSession(session)
	roleModelWithSession := s.TransactionHelper.GetRoleModelWithSession(session)
	positionRoleModel := s.TransactionHelper.GetPositionRoleModelWithSession(session)

	// ========== 1. 创建默认角色 ==========
	// 部门经理角色 - 拥有部门管理、员工管理、任务管理等权限
	managerRoleID := utils.Common.GenId("role")
	// 权限码: 任务(1-5), 任务节点(10-13), 部门(45-48), 员工(70-74), 通知(30-32)
	managerPermissions := "[1,2,3,4,5,10,11,12,13,30,31,32,45,46,47,48,70,71,72,73,74]"
	managerRole := &roleModel.Role{
		Id:              managerRoleID,
		CompanyId:       companyID,
		RoleName:        "部门经理",
		RoleCode:        "DEPT_MANAGER",
		RoleDescription: utils.Common.ToSqlNullString("部门经理角色，拥有部门内管理权限"),
		IsSystem:        1,
		Permissions:     utils.Common.ToSqlNullString(managerPermissions),
		Status:          1,
		CreateTime:      time.Now(),
		UpdateTime:      time.Now(),
	}
	if _, err := roleModelWithSession.Insert(ctx, managerRole); err != nil {
		logx.Errorf("创建部门经理角色失败: %v", err)
		return err
	}

	// 普通员工角色 - 基础权限
	employeeRoleID := utils.Common.GenId("role")
	// 权限码: 任务查看(1), 任务节点查看(10), 通知查看(30)
	employeePermissions := "[1,10,30]"
	employeeRole := &roleModel.Role{
		Id:              employeeRoleID,
		CompanyId:       companyID,
		RoleName:        "普通员工",
		RoleCode:        "EMPLOYEE",
		RoleDescription: utils.Common.ToSqlNullString("普通员工角色，拥有基础操作权限"),
		IsSystem:        1,
		Permissions:     utils.Common.ToSqlNullString(employeePermissions),
		Status:          1,
		CreateTime:      time.Now(),
		UpdateTime:      time.Now(),
	}
	if _, err := roleModelWithSession.Insert(ctx, employeeRole); err != nil {
		logx.Errorf("创建普通员工角色失败: %v", err)
		return err
	}

	// 高级员工角色 - 比普通员工多一些权限
	seniorRoleID := utils.Common.GenId("role")
	// 权限码: 任务(1-5), 任务节点(10-13), 通知(30-32)
	seniorPermissions := "[1,2,3,4,5,10,11,12,13,30,31,32]"
	seniorRole := &roleModel.Role{
		Id:              seniorRoleID,
		CompanyId:       companyID,
		RoleName:        "高级员工",
		RoleCode:        "SENIOR_EMPLOYEE",
		RoleDescription: utils.Common.ToSqlNullString("高级员工角色，拥有更多任务操作权限"),
		IsSystem:        1,
		Permissions:     utils.Common.ToSqlNullString(seniorPermissions),
		Status:          1,
		CreateTime:      time.Now(),
		UpdateTime:      time.Now(),
	}
	if _, err := roleModelWithSession.Insert(ctx, seniorRole); err != nil {
		logx.Errorf("创建高级员工角色失败: %v", err)
		return err
	}

	logx.Infof("已为公司 %s 创建默认角色: 部门经理(%s), 高级员工(%s), 普通员工(%s)",
		companyID, managerRoleID, seniorRoleID, employeeRoleID)

	// ========== 2. 创建部门和职位，并绑定角色 ==========
	var createDept func(parentID sql.NullString, tpl deptTpl) (string, error)
	createDept = func(parentID sql.NullString, tpl deptTpl) (string, error) {
		deptID := utils.Common.GenId("dept")
		d := &company.Department{
			Id:                 deptID,
			CompanyId:          companyID,
			ParentId:           parentID,
			DepartmentName:     tpl.Name,
			DepartmentCode:     utils.Common.ToSqlNullString(tpl.Code),
			DepartmentPriority: 1,
			ManagerId:          sql.NullString{},
			Description:        sql.NullString{String: "系统默认模板", Valid: true},
			Status:             1,
			CreateTime:         time.Now(),
			UpdateTime:         time.Now(),
		}
		if _, err := depModel.Insert(ctx, d); err != nil {
			return "", err
		}

		// 为该部门创建专属职位
		for _, p := range tpl.Positions {
			posID := utils.Common.GenId("pos")
			pos := &company.Position{
				Id:               posID,
				DepartmentId:     deptID,
				PositionName:     p.Name,
				PositionCode:     utils.Common.ToSqlNullString(p.Code),
				JobType:          p.JobType,
				PositionLevel:    p.Level,
				RequiredSkills:   sql.NullString{},
				JobDescription:   sql.NullString{String: "系统默认模板", Valid: true},
				Responsibilities: sql.NullString{},
				Requirements:     sql.NullString{},
				SalaryRangeMin:   sql.NullFloat64{},
				SalaryRangeMax:   sql.NullFloat64{},
				IsManagement:     p.IsMgmt,
				MaxEmployees:     0,
				CurrentEmployees: 0,
				Status:           1,
				CreateTime:       time.Now(),
				UpdateTime:       time.Now(),
			}
			if _, err := posModel.Insert(ctx, pos); err != nil {
				return "", err
			}

			// ========== 3. 根据职位类型绑定对应角色 ==========
			var roleID string
			if p.IsMgmt == 1 {
				// 管理岗位绑定"部门经理"角色
				roleID = managerRoleID
			} else if p.Level >= 3 {
				// 高级岗位绑定"高级员工"角色
				roleID = seniorRoleID
			} else {
				// 普通岗位绑定"普通员工"角色
				roleID = employeeRoleID
			}

			posRoleID := utils.Common.GenId("pr")
			posRole := &roleModel.PositionRole{
				Id:         posRoleID,
				PositionId: posID,
				RoleId:     roleID,
				GrantBy:    sql.NullString{String: "SYSTEM", Valid: true},
				GrantTime:  time.Now(),
				ExpireTime: sql.NullTime{}, // 不过期
				Status:     1,
				CreateTime: time.Now(),
			}
			if _, err := positionRoleModel.Insert(ctx, posRole); err != nil {
				logx.Errorf("绑定职位角色失败: posID=%s, roleID=%s, err=%v", posID, roleID, err)
				return "", err
			}
		}

		// 递归创建子部门
		for _, child := range tpl.Children {
			if _, err := createDept(sql.NullString{String: deptID, Valid: true}, child); err != nil {
				return "", err
			}
		}
		return deptID, nil
	}

	for _, root := range roots {
		if _, err := createDept(sql.NullString{}, root); err != nil {
			logx.Errorf("clone default org failed: %v", err)
			return err
		}
	}

	logx.Infof("已为公司 %s 完成组织架构模板初始化", companyID)
	return nil
}

// ApplyDefaultOrgStructure clones system default departments and positions to a new company
// 这个方法会创建新的事务，用于独立调用
func (s *ServiceContext) ApplyDefaultOrgStructure(ctx context.Context, companyID string) error {
	return s.TransactionService.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		return s.ApplyDefaultOrgStructureWithSession(ctx, session, companyID)
	})
}

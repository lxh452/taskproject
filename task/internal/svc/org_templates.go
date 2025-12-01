package svc

import (
	"context"
	"database/sql"
	"time"

	"task_Project/model/company"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ApplyDefaultOrgStructure clones system default departments and positions to a new company
func (s *ServiceContext) ApplyDefaultOrgStructure(ctx context.Context, companyID string) error {
	type deptTpl struct {
		Name     string
		Code     string
		Children []deptTpl
	}
	// minimal default structure
	roots := []deptTpl{
		{Name: "人力资源部", Code: "HR"},
		{Name: "研发部", Code: "RND", Children: []deptTpl{
			{Name: "后端组", Code: "BE"},
			{Name: "前端组", Code: "FE"},
		}},
		{Name: "市场部", Code: "MKT"},
		{Name: "销售部", Code: "SLS"},
		{Name: "财务部", Code: "FIN"},
		{Name: "行政部", Code: "ADM"},
		{Name: "运维部", Code: "OPS"},
	}

	// default positions by dept code (fallback to common set)
	type posTpl struct {
		Name    string
		Code    string
		Level   int64
		JobType int64
		IsMgmt  int64
	}

	commonPositions := []posTpl{
		{Name: "经理", Code: "MGR", Level: 3, JobType: 2, IsMgmt: 1},
		{Name: "高级", Code: "SEN", Level: 3, JobType: 0, IsMgmt: 0},
		{Name: "工程师", Code: "ENG", Level: 2, JobType: 0, IsMgmt: 0},
		{Name: "助理", Code: "AST", Level: 1, JobType: 1, IsMgmt: 0},
	}

	return s.TransactionService.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		depModel := s.TransactionHelper.GetDepartmentModelWithSession(session)
		posModel := s.TransactionHelper.GetPositionModelWithSession(session)

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
			// create default positions under this department
			for _, p := range commonPositions {
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
			}
			// create children
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
		return nil
	})
}

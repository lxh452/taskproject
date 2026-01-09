package svc

import (
	"context"
	"errors"

	"task_Project/model/company"
	"task_Project/model/user"

	"github.com/zeromicro/go-zero/core/logx"
)

// StatusCheckerService 用户和公司状态检查服务
type StatusCheckerService struct {
	userModel    user.UserModel
	companyModel company.CompanyModel
}

// NewStatusCheckerService 创建状态检查服务
func NewStatusCheckerService(userModel user.UserModel, companyModel company.CompanyModel) *StatusCheckerService {
	return &StatusCheckerService{
		userModel:    userModel,
		companyModel: companyModel,
	}
}

// CheckUserStatus 检查用户状态
// 返回 nil 表示用户状态正常，返回 error 表示用户被封禁或禁用
func (s *StatusCheckerService) CheckUserStatus(ctx context.Context, userID string) error {
	if userID == "" {
		return nil
	}

	userInfo, err := s.userModel.FindOne(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return errors.New("用户不存在")
		}
		logx.Errorf("查询用户状态失败: %v, userId=%s", err, userID)
		// 查询失败时不阻止请求，避免数据库问题影响正常使用
		return nil
	}

	// status = 2 表示封禁
	if userInfo.Status == 2 {
		return errors.New("您的账号已被封禁，请联系管理员")
	}

	// status != 1 表示非正常状态
	if userInfo.Status != 1 {
		return errors.New("您的账号已被禁用，请联系管理员")
	}

	return nil
}

// CheckCompanyStatus 检查公司状态
// 返回 nil 表示公司状态正常，返回 error 表示公司被禁用
func (s *StatusCheckerService) CheckCompanyStatus(ctx context.Context, companyID string) error {
	if companyID == "" {
		return nil
	}

	companyInfo, err := s.companyModel.FindOne(ctx, companyID)
	if err != nil {
		if errors.Is(err, company.ErrNotFound) {
			return errors.New("公司不存在")
		}
		logx.Errorf("查询公司状态失败: %v, companyId=%s", err, companyID)
		// 查询失败时不阻止请求，避免数据库问题影响正常使用
		return nil
	}

	// status = 0 表示禁用
	if companyInfo.Status == 0 {
		return errors.New("您所在的公司已被禁用，请联系管理员")
	}

	return nil
}

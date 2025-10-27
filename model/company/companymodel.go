package company

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CompanyModel = (*customCompanyModel)(nil)

type (
	// CompanyModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCompanyModel.
	CompanyModel interface {
		companyModel
		withSession(session sqlx.Session) CompanyModel

		// 公司CRUD操作
		FindByOwner(ctx context.Context, owner string) ([]*Company, error)
		FindByStatus(ctx context.Context, status int) ([]*Company, error)
		FindByAttributes(ctx context.Context, attributes int) ([]*Company, error)
		FindByBusiness(ctx context.Context, business int) ([]*Company, error)
		FindByPage(ctx context.Context, page, pageSize int) ([]*Company, int64, error)
		SearchCompanies(ctx context.Context, keyword string, page, pageSize int) ([]*Company, int64, error)
		UpdateStatus(ctx context.Context, id string, status int) error
		UpdateBasicInfo(ctx context.Context, id, name, description, address, phone, email string) error
		UpdateAttributes(ctx context.Context, id string, attributes, business int) error
		SoftDelete(ctx context.Context, id string) error
		BatchUpdateStatus(ctx context.Context, ids []string, status int) error
		GetCompanyCount(ctx context.Context) (int64, error)
		GetCompanyCountByStatus(ctx context.Context, status int) (int64, error)
		GetCompanyCountByOwner(ctx context.Context, owner string) (int64, error)
		GetCompanyCountByAttributes(ctx context.Context, attributes int) (int64, error)
		GetCompanyCountByBusiness(ctx context.Context, business int) (int64, error)
	}

	customCompanyModel struct {
		*defaultCompanyModel
	}
)

// NewCompanyModel returns a model for the database table.
func NewCompanyModel(conn sqlx.SqlConn) CompanyModel {
	return &customCompanyModel{
		defaultCompanyModel: newCompanyModel(conn),
	}
}

func (m *customCompanyModel) withSession(session sqlx.Session) CompanyModel {
	return NewCompanyModel(sqlx.NewSqlConnFromSession(session))
}

// FindByOwner 根据拥有者查找公司
func (m *customCompanyModel) FindByOwner(ctx context.Context, owner string) ([]*Company, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `owner` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", companyRows, m.table)
	var resp []*Company
	err := m.conn.QueryRowsCtx(ctx, &resp, query, owner)
	return resp, err
}

// FindByStatus 根据状态查找公司
func (m *customCompanyModel) FindByStatus(ctx context.Context, status int) ([]*Company, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `status` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", companyRows, m.table)
	var resp []*Company
	err := m.conn.QueryRowsCtx(ctx, &resp, query, status)
	return resp, err
}

// FindByAttributes 根据企业属性查找公司
func (m *customCompanyModel) FindByAttributes(ctx context.Context, attributes int) ([]*Company, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `company_attributes` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", companyRows, m.table)
	var resp []*Company
	err := m.conn.QueryRowsCtx(ctx, &resp, query, attributes)
	return resp, err
}

// FindByBusiness 根据公司业务查找公司
func (m *customCompanyModel) FindByBusiness(ctx context.Context, business int) ([]*Company, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `company_business` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", companyRows, m.table)
	var resp []*Company
	err := m.conn.QueryRowsCtx(ctx, &resp, query, business)
	return resp, err
}

// FindByPage 分页查找公司
func (m *customCompanyModel) FindByPage(ctx context.Context, page, pageSize int) ([]*Company, int64, error) {
	offset := (page - 1) * pageSize

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `delete_time` IS NULL", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `delete_time` IS NULL ORDER BY `create_time` DESC LIMIT ? OFFSET ?", companyRows, m.table)
	var resp []*Company
	err = m.conn.QueryRowsCtx(ctx, &resp, query, pageSize, offset)
	return resp, total, err
}

// SearchCompanies 搜索公司
func (m *customCompanyModel) SearchCompanies(ctx context.Context, keyword string, page, pageSize int) ([]*Company, int64, error) {
	offset := (page - 1) * pageSize
	keyword = "%" + keyword + "%"

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE (`name` LIKE ? OR `description` LIKE ? OR `address` LIKE ?) AND `delete_time` IS NULL", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, keyword, keyword, keyword)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE (`name` LIKE ? OR `description` LIKE ? OR `address` LIKE ?) AND `delete_time` IS NULL ORDER BY `create_time` DESC LIMIT ? OFFSET ?", companyRows, m.table)
	var resp []*Company
	err = m.conn.QueryRowsCtx(ctx, &resp, query, keyword, keyword, keyword, pageSize, offset)
	return resp, total, err
}

// UpdateStatus 更新公司状态
func (m *customCompanyModel) UpdateStatus(ctx context.Context, id string, status int) error {
	query := fmt.Sprintf("UPDATE %s SET `status` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, id)
	return err
}

// UpdateBasicInfo 更新公司基本信息
func (m *customCompanyModel) UpdateBasicInfo(ctx context.Context, id, name, description, address, phone, email string) error {
	query := fmt.Sprintf("UPDATE %s SET `name` = ?, `description` = ?, `address` = ?, `phone` = ?, `email` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, name, description, address, phone, email, id)
	return err
}

// UpdateAttributes 更新公司属性
func (m *customCompanyModel) UpdateAttributes(ctx context.Context, id string, attributes, business int) error {
	query := fmt.Sprintf("UPDATE %s SET `company_attributes` = ?, `company_business` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, attributes, business, id)
	return err
}

// SoftDelete 软删除公司
func (m *customCompanyModel) SoftDelete(ctx context.Context, id string) error {
	query := fmt.Sprintf("UPDATE %s SET `delete_time` = NOW(), `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

// BatchUpdateStatus 批量更新公司状态
func (m *customCompanyModel) BatchUpdateStatus(ctx context.Context, ids []string, status int) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := strings.Repeat("?,", len(ids)-1) + "?"
	query := fmt.Sprintf("UPDATE %s SET `status` = ?, `update_time` = NOW() WHERE `id` IN (%s)", m.table, placeholders)

	args := make([]interface{}, len(ids)+1)
	args[0] = status
	for i, id := range ids {
		args[i+1] = id
	}

	_, err := m.conn.ExecCtx(ctx, query, args...)
	return err
}

// GetCompanyCount 获取公司总数
func (m *customCompanyModel) GetCompanyCount(ctx context.Context) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `delete_time` IS NULL", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query)
	return count, err
}

// GetCompanyCountByStatus 根据状态获取公司数量
func (m *customCompanyModel) GetCompanyCountByStatus(ctx context.Context, status int) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `status` = ? AND `delete_time` IS NULL", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, status)
	return count, err
}

// GetCompanyCountByOwner 根据拥有者获取公司数量
func (m *customCompanyModel) GetCompanyCountByOwner(ctx context.Context, owner string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `owner` = ? AND `delete_time` IS NULL", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, owner)
	return count, err
}

// GetCompanyCountByAttributes 根据企业属性获取公司数量
func (m *customCompanyModel) GetCompanyCountByAttributes(ctx context.Context, attributes int) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `company_attributes` = ? AND `delete_time` IS NULL", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, attributes)
	return count, err
}

// GetCompanyCountByBusiness 根据公司业务获取公司数量
func (m *customCompanyModel) GetCompanyCountByBusiness(ctx context.Context, business int) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `company_business` = ? AND `delete_time` IS NULL", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, business)
	return count, err
}

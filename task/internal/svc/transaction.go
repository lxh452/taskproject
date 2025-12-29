package svc

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// TransactionService 事务服务
type TransactionService struct {
	conn sqlx.SqlConn
}

// NewTransactionService 创建事务服务
func NewTransactionService(conn sqlx.SqlConn) *TransactionService {
	return &TransactionService{
		conn: conn,
	}
}

// TransactCtx 执行事务
func (t *TransactionService) TransactCtx(ctx context.Context, fn func(ctx context.Context, session sqlx.Session) error) error {
	return t.conn.TransactCtx(ctx, fn)
}

// WithSession 创建带会话的上下文
func (t *TransactionService) WithSession(session sqlx.Session) *TransactionService {
	return &TransactionService{
		conn: sqlx.NewSqlConnFromSession(session),
	}
}

// GetConn 获取连接
func (t *TransactionService) GetConn() sqlx.SqlConn {
	return t.conn
}

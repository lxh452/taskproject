package svc

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// SQLExecutorService SQL脚本执行服务
type SQLExecutorService struct {
	conn   sqlx.SqlConn
	sqlDir string
}

// NewSQLExecutorService 创建SQL执行服务
func NewSQLExecutorService(conn sqlx.SqlConn, sqlDir string) *SQLExecutorService {
	return &SQLExecutorService{
		conn:   conn,
		sqlDir: sqlDir,
	}
}

// AutoMigrate 启动时自动执行数据库迁移
// 按照依赖顺序执行所有SQL脚本
func (s *SQLExecutorService) AutoMigrate(ctx context.Context) error {
	logx.Info("========== 开始自动执行数据库迁移 ==========")
	startTime := time.Now()

	// 定义执行顺序（按依赖关系排序）
	scriptOrder := []string{
		"user.sql",
		"user_auth.sql",
		"company.sql",
		"role.sql",
		"task.sql",
		"task_checklist.sql",
		"handover_approval.sql",
		"join_application.sql",
		"add_has_joine_company.sql", // 注意：实际文件名是 add_has_joine_company.sql（少了一个d）
	}

	successCount := 0
	skipCount := 0
	failCount := 0

	for _, scriptName := range scriptOrder {
		scriptPath := filepath.Join(s.sqlDir, scriptName)

		// 检查文件是否存在
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			logx.Infof("[SQL迁移] 跳过不存在的脚本: %s", scriptName)
			skipCount++
			continue
		}

		// 执行脚本
		success, err := s.executeScriptFile(ctx, scriptPath, scriptName)
		if err != nil {
			logx.Errorf("[SQL迁移] 脚本执行出错: %s, error=%v", scriptName, err)
			failCount++
		} else if success {
			successCount++
		} else {
			// 部分失败但继续执行
			failCount++
		}
	}

	// 执行目录中其他未在顺序列表中的脚本
	entries, err := os.ReadDir(s.sqlDir)
	if err == nil {
		executedSet := make(map[string]bool)
		for _, name := range scriptOrder {
			executedSet[name] = true
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".sql") {
				continue
			}
			if executedSet[entry.Name()] {
				continue
			}

			scriptPath := filepath.Join(s.sqlDir, entry.Name())
			success, err := s.executeScriptFile(ctx, scriptPath, entry.Name())
			if err != nil {
				logx.Errorf("[SQL迁移] 脚本执行出错: %s, error=%v", entry.Name(), err)
				failCount++
			} else if success {
				successCount++
			} else {
				failCount++
			}
		}
	}

	duration := time.Since(startTime)
	logx.Infof("========== 数据库迁移完成 ==========")
	logx.Infof("[SQL迁移] 统计: 成功=%d, 跳过=%d, 失败=%d, 耗时=%v", successCount, skipCount, failCount, duration)

	return nil
}

// executeScriptFile 执行单个SQL脚本文件
func (s *SQLExecutorService) executeScriptFile(ctx context.Context, filePath, fileName string) (bool, error) {
	logx.Infof("[SQL迁移] 开始执行: %s", fileName)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	statements := s.parseSQLStatements(string(content))
	if len(statements) == 0 {
		logx.Infof("[SQL迁移] 脚本为空: %s", fileName)
		return true, nil
	}

	successCount := 0
	skipCount := 0
	failCount := 0

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		_, err := s.conn.ExecCtx(ctx, stmt)
		if err != nil {
			errStr := err.Error()
			// 忽略常见的"已存在"错误
			if strings.Contains(errStr, "already exists") ||
				strings.Contains(errStr, "Duplicate column") ||
				strings.Contains(errStr, "Duplicate key") ||
				strings.Contains(errStr, "Duplicate entry") ||
				strings.Contains(errStr, "1060") || // Duplicate column name
				strings.Contains(errStr, "1061") || // Duplicate key name
				strings.Contains(errStr, "1050") { // Table already exists
				skipCount++
				continue
			}
			logx.Errorf("[SQL迁移] 语句执行失败: %s, error=%v", truncateSQL(stmt, 80), err)
			failCount++
			continue
		}
		successCount++
	}

	logx.Infof("[SQL迁移] %s 完成: 成功=%d, 跳过=%d, 失败=%d",
		fileName, successCount, skipCount, failCount)

	return failCount == 0, nil
}

// parseSQLStatements 解析SQL语句（按分号分割，处理注释和动态SQL）
func (s *SQLExecutorService) parseSQLStatements(content string) []string {
	var statements []string
	var currentStmt strings.Builder
	inMultiLineComment := false
	inPrepareBlock := false // 是否在 PREPARE 块中
	prepareDepth := 0       // PREPARE 语句的嵌套深度

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		upperLine := strings.ToUpper(trimmedLine)

		// 处理多行注释
		if inMultiLineComment {
			if strings.Contains(trimmedLine, "*/") {
				inMultiLineComment = false
				// 取 */ 之后的内容
				idx := strings.Index(trimmedLine, "*/")
				trimmedLine = strings.TrimSpace(trimmedLine[idx+2:])
				if trimmedLine == "" {
					continue
				}
				upperLine = strings.ToUpper(trimmedLine)
			} else {
				// 如果在 PREPARE 块中，需要保留注释内容
				if inPrepareBlock {
					currentStmt.WriteString(line)
					currentStmt.WriteString("\n")
				}
				continue
			}
		}

		// 检查多行注释开始
		if strings.HasPrefix(trimmedLine, "/*") {
			if strings.Contains(trimmedLine, "*/") {
				// 同一行结束的注释，跳过
				continue
			}
			inMultiLineComment = true
			// 如果在 PREPARE 块中，需要保留注释
			if inPrepareBlock {
				currentStmt.WriteString(line)
				currentStmt.WriteString("\n")
			}
			continue
		}

		// 跳过空行和单行注释（除非在 PREPARE 块中）
		if !inPrepareBlock {
			if trimmedLine == "" || strings.HasPrefix(trimmedLine, "--") || strings.HasPrefix(trimmedLine, "#") {
				continue
			}
		}

		// 检测动态 SQL 块：从 SET @ 开始，到 DEALLOCATE PREPARE 结束
		// 如果遇到 SET @，可能是动态 SQL 的开始，需要等待 PREPARE 确认
		// 如果已经遇到 PREPARE，则确认是动态 SQL 块
		if strings.HasPrefix(upperLine, "SET @") {
			// 可能是动态 SQL 的开始，但需要等待 PREPARE 确认
			// 暂时不设置 inPrepareBlock，继续累积语句
		}

		// 检测 PREPARE 语句（确认是动态 SQL 块）
		// 如果当前语句包含 SET @，则从 SET @ 开始的所有内容都是动态 SQL 块的一部分
		if strings.HasPrefix(upperLine, "PREPARE ") {
			// 检查 currentStmt 中是否有 SET @，如果有，则确认是动态 SQL 块
			currentContent := strings.ToUpper(currentStmt.String())
			if strings.Contains(currentContent, "SET @") {
				inPrepareBlock = true
				prepareDepth = 1
			}
		}

		currentStmt.WriteString(line)
		currentStmt.WriteString("\n")

		// 检测 DEALLOCATE PREPARE 语句（结束动态 SQL 块）
		if strings.HasPrefix(upperLine, "DEALLOCATE PREPARE ") {
			// 标记块即将结束，等待分号
			if inPrepareBlock {
				prepareDepth = 0
			}
		}

		// 检查语句是否结束（以分号结尾）
		if strings.HasSuffix(trimmedLine, ";") {
			// 如果在动态 SQL 块中，且已经遇到 DEALLOCATE PREPARE，则结束块
			if inPrepareBlock && prepareDepth == 0 {
				stmt := strings.TrimSpace(currentStmt.String())
				if stmt != "" && stmt != ";" {
					statements = append(statements, stmt)
				}
				currentStmt.Reset()
				inPrepareBlock = false
				prepareDepth = 0
			} else if !inPrepareBlock {
				// 检查当前语句是否包含 SET @ 但还没有 PREPARE
				// 如果是，可能是独立的 SET @ 语句，应该提交
				currentContent := strings.ToUpper(strings.TrimSpace(currentStmt.String()))
				if strings.HasPrefix(currentContent, "SET @") && !strings.Contains(currentContent, "PREPARE ") {
					// 独立的 SET @ 语句，可以提交
					stmt := strings.TrimSpace(currentStmt.String())
					if stmt != "" && stmt != ";" {
						statements = append(statements, stmt)
					}
					currentStmt.Reset()
				} else {
					// 普通语句，直接提交
					stmt := strings.TrimSpace(currentStmt.String())
					if stmt != "" && stmt != ";" {
						// 跳过 SELECT 语句（迁移脚本中不应该有查询语句）
						upperStmt := strings.ToUpper(strings.TrimSpace(stmt))
						if !strings.HasPrefix(upperStmt, "SELECT ") {
							statements = append(statements, stmt)
						} else {
							logx.Infof("[SQL迁移] 跳过 SELECT 语句: %s", truncateSQL(stmt, 50))
						}
					}
					currentStmt.Reset()
				}
			}
			// 如果在动态 SQL 块中但还没遇到 DEALLOCATE，继续累积
		}
	}

	// 处理最后一条没有分号的语句
	if currentStmt.Len() > 0 {
		stmt := strings.TrimSpace(currentStmt.String())
		if stmt != "" && stmt != ";" {
			upperStmt := strings.ToUpper(strings.TrimSpace(stmt))
			if !strings.HasPrefix(upperStmt, "SELECT ") {
				statements = append(statements, stmt)
			} else {
				logx.Infof("[SQL迁移] 跳过 SELECT 语句: %s", truncateSQL(stmt, 50))
			}
		}
	}

	return statements
}

// truncateSQL 截断SQL语句用于日志
func truncateSQL(sql string, maxLen int) string {
	sql = strings.ReplaceAll(sql, "\n", " ")
	sql = strings.ReplaceAll(sql, "\r", "")
	sql = strings.TrimSpace(sql)
	if len(sql) > maxLen {
		return sql[:maxLen] + "..."
	}
	return sql
}

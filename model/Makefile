# 按照依赖顺序排列：基础表 -> 业务表 -> 功能表
all: user company role user_auth task

# 基础表（无依赖）
user:
	goctl model mysql ddl -src="./sql/user.sql" -dir="./user"

# 业务表（依赖user）
company:
	goctl model mysql ddl -src="./sql/company.sql" -dir="./company"

# 权限表（依赖company和user）
role:
	goctl model mysql ddl -src="./sql/role.sql" -dir="./role"

# 功能表（依赖user）
user_auth:
	goctl model mysql ddl -src="./sql/user_auth.sql" -dir="./user_auth"

# 任务表（依赖company和user）
task:
	goctl model mysql ddl -src="./sql/task.sql" -dir="./task"

# 清理生成的文件
clean:
	rm -rf ./user ./company ./role ./user_auth ./task



# 重新生成所有表
rebuild: clean all
# 模型层CRUD操作总结

## 已完成的模块

### 1. 用户模块 (`model/user/`)

#### 用户表 (User)
**文件**: `usermodel.go`

**主要CRUD操作**:
- `FindByUsername()` - 根据用户名查找用户
- `FindByEmail()` - 根据邮箱查找用户
- `FindByPhone()` - 根据手机号查找用户
- `FindByStatus()` - 根据状态查找用户
- `FindByPage()` - 分页查找用户
- `SearchUsers()` - 搜索用户
- `UpdateLastLogin()` - 更新最后登录信息
- `UpdateLoginFailedCount()` - 更新登录失败次数
- `UpdateLockStatus()` - 更新锁定状态
- `UpdatePassword()` - 更新密码
- `UpdateProfile()` - 更新用户资料
- `UpdateStatus()` - 更新用户状态
- `SoftDelete()` - 软删除用户
- `BatchUpdateStatus()` - 批量更新用户状态
- `GetUserCount()` - 获取用户总数
- `GetUserCountByStatus()` - 根据状态获取用户数量

#### 员工表 (Employee)
**文件**: `employeemodel.go`

**主要CRUD操作**:
- `FindByUserID()` - 根据用户ID查找员工
- `FindByCompanyID()` - 根据公司ID查找员工
- `FindByDepartmentID()` - 根据部门ID查找员工
- `FindByPositionID()` - 根据职位ID查找员工
- `FindByEmployeeID()` - 根据员工编号查找员工
- `FindByStatus()` - 根据状态查找员工
- `FindByPage()` - 分页查找员工
- `FindByCompanyPage()` - 根据公司分页查找员工
- `FindByDepartmentPage()` - 根据部门分页查找员工
- `SearchEmployees()` - 搜索员工
- `SearchEmployeesByCompany()` - 根据公司搜索员工
- `UpdateDepartment()` - 更新部门
- `UpdatePosition()` - 更新职位
- `UpdateSkills()` - 更新技能
- `UpdateRoleTags()` - 更新角色标签
- `UpdateWorkContact()` - 更新工作联系方式
- `UpdateLeaveDate()` - 更新离职日期
- `UpdateStatus()` - 更新员工状态
- `SoftDelete()` - 软删除员工
- `BatchUpdateStatus()` - 批量更新员工状态
- `BatchUpdateDepartment()` - 批量更新员工部门
- `GetEmployeeCount()` - 获取员工总数
- `GetEmployeeCountByCompany()` - 根据公司获取员工数量
- `GetEmployeeCountByDepartment()` - 根据部门获取员工数量
- `GetEmployeeCountByStatus()` - 根据状态获取员工数量
- `GetEmployeesByRoleTags()` - 根据角色标签查找员工
- `GetEmployeesBySkills()` - 根据技能查找员工

### 2. 公司模块 (`model/company/`)

#### 公司表 (Company)
**文件**: `companymodel.go`

**主要CRUD操作**:
- `FindByOwner()` - 根据拥有者查找公司
- `FindByStatus()` - 根据状态查找公司
- `FindByAttributes()` - 根据企业属性查找公司
- `FindByBusiness()` - 根据公司业务查找公司
- `FindByPage()` - 分页查找公司
- `SearchCompanies()` - 搜索公司
- `UpdateStatus()` - 更新公司状态
- `UpdateBasicInfo()` - 更新公司基本信息
- `UpdateAttributes()` - 更新公司属性
- `SoftDelete()` - 软删除公司
- `BatchUpdateStatus()` - 批量更新公司状态
- `GetCompanyCount()` - 获取公司总数
- `GetCompanyCountByStatus()` - 根据状态获取公司数量
- `GetCompanyCountByOwner()` - 根据拥有者获取公司数量
- `GetCompanyCountByAttributes()` - 根据企业属性获取公司数量
- `GetCompanyCountByBusiness()` - 根据公司业务获取公司数量

#### 部门表 (Department)
**文件**: `departmentmodel.go`

**主要CRUD操作**:
- `FindByCompanyID()` - 根据公司ID查找部门
- `FindByParentID()` - 根据父部门ID查找部门
- `FindByManagerID()` - 根据管理者ID查找部门
- `FindByStatus()` - 根据状态查找部门
- `FindByPage()` - 分页查找部门
- `FindByCompanyPage()` - 根据公司分页查找部门
- `SearchDepartments()` - 搜索部门
- `SearchDepartmentsByCompany()` - 根据公司搜索部门
- `UpdateManager()` - 更新部门管理者
- `UpdateParent()` - 更新父部门
- `UpdateBasicInfo()` - 更新部门基本信息
- `UpdateStatus()` - 更新部门状态
- `SoftDelete()` - 软删除部门
- `BatchUpdateStatus()` - 批量更新部门状态
- `BatchUpdateManager()` - 批量更新部门管理者
- `GetDepartmentCount()` - 获取部门总数
- `GetDepartmentCountByCompany()` - 根据公司获取部门数量
- `GetDepartmentCountByStatus()` - 根据状态获取部门数量
- `GetDepartmentCountByManager()` - 根据管理者获取部门数量
- `GetDepartmentTree()` - 获取部门树形结构
- `GetSubDepartments()` - 获取子部门

#### 职位表 (Position)
**文件**: `positionmodel.go`

**主要CRUD操作**:
- `FindByDepartmentID()` - 根据部门ID查找职位
- `FindByStatus()` - 根据状态查找职位
- `FindByManagement()` - 根据是否管理职位查找
- `FindByLevel()` - 根据职位级别查找
- `FindByPage()` - 分页查找职位
- `FindByDepartmentPage()` - 根据部门分页查找职位
- `SearchPositions()` - 搜索职位
- `SearchPositionsByDepartment()` - 根据部门搜索职位
- `UpdateBasicInfo()` - 更新职位基本信息
- `UpdateLevel()` - 更新职位级别
- `UpdateSalaryRange()` - 更新薪资范围
- `UpdateManagement()` - 更新是否管理职位
- `UpdateMaxEmployees()` - 更新最大员工数
- `UpdateCurrentEmployees()` - 更新当前员工数
- `UpdateStatus()` - 更新职位状态
- `SoftDelete()` - 软删除职位
- `BatchUpdateStatus()` - 批量更新职位状态
- `BatchUpdateDepartment()` - 批量更新职位部门
- `GetPositionCount()` - 获取职位总数
- `GetPositionCountByDepartment()` - 根据部门获取职位数量
- `GetPositionCountByStatus()` - 根据状态获取职位数量
- `GetPositionCountByManagement()` - 根据是否管理职位获取数量
- `GetPositionsBySalaryRange()` - 根据薪资范围查找职位
- `GetPositionsBySkills()` - 根据技能查找职位

### 3. 角色模块 (`model/role/`)

#### 角色表 (Role)
**文件**: `rolemodel.go`

**主要CRUD操作**:
- `FindByCompanyID()` - 根据公司ID查找角色
- `FindByStatus()` - 根据状态查找角色
- `FindBySystem()` - 根据是否系统角色查找
- `FindByRoleCode()` - 根据角色编码查找角色
- `FindByPage()` - 分页查找角色
- `FindByCompanyPage()` - 根据公司分页查找角色
- `SearchRoles()` - 搜索角色
- `SearchRolesByCompany()` - 根据公司搜索角色
- `UpdateBasicInfo()` - 更新角色基本信息
- `UpdatePermissions()` - 更新角色权限
- `UpdateStatus()` - 更新角色状态
- `SoftDelete()` - 软删除角色
- `BatchUpdateStatus()` - 批量更新角色状态
- `GetRoleCount()` - 获取角色总数
- `GetRoleCountByCompany()` - 根据公司获取角色数量
- `GetRoleCountByStatus()` - 根据状态获取角色数量
- `GetRoleCountBySystem()` - 根据是否系统角色获取数量
- `GetRolesByPermissions()` - 根据权限查找角色

### 4. 任务模块 (`model/task/`)

#### 任务表 (Task)
**文件**: `taskmodel.go`

**主要CRUD操作**:
- `FindByCompany()` - 根据公司ID查找任务
- `FindByDepartment()` - 根据部门ID查找任务
- `FindByCreator()` - 根据创建者ID查找任务
- `FindByStatus()` - 根据状态查找任务
- `FindByPriority()` - 根据优先级查找任务
- `FindByType()` - 根据任务类型查找任务
- `FindByPage()` - 分页查找任务
- `SearchTasks()` - 搜索任务
- `UpdateStatus()` - 更新任务状态
- `UpdateProgress()` - 更新任务进度
- `UpdateActualHours()` - 更新实际工时
- `UpdateBasicInfo()` - 更新任务基本信息
- `UpdateDeadline()` - 更新任务截止时间
- `SoftDelete()` - 软删除任务
- `BatchUpdateStatus()` - 批量更新任务状态
- `GetTaskCount()` - 获取任务总数
- `GetTaskCountByStatus()` - 根据状态获取任务数量
- `GetTaskCountByCompany()` - 根据公司获取任务数量
- `GetTaskCountByDepartment()` - 根据部门获取任务数量
- `GetTaskCountByCreator()` - 根据创建者获取任务数量

#### 任务节点表 (TaskNode)
**文件**: `tasknodemodel.go`

**主要CRUD操作**:
- `FindByTaskID()` - 根据任务ID查找任务节点
- `FindByDepartment()` - 根据部门ID查找任务节点
- `FindByExecutor()` - 根据执行人ID查找任务节点
- `FindByLeader()` - 根据负责人ID查找任务节点
- `FindByStatus()` - 根据状态查找任务节点
- `FindByDeadlineRange()` - 根据截止时间范围查找任务节点
- `FindByPage()` - 分页查找任务节点
- `SearchTaskNodes()` - 搜索任务节点
- `UpdateStatus()` - 更新任务节点状态
- `UpdateProgress()` - 更新任务节点进度
- `UpdateActualHours()` - 更新任务节点实际工时
- `UpdateExecutor()` - 更新任务节点执行人
- `UpdateLeader()` - 更新任务节点负责人
- `UpdateDeadline()` - 更新任务节点截止时间
- `SoftDelete()` - 软删除任务节点
- `BatchUpdateStatus()` - 批量更新任务节点状态
- `GetTaskNodeCount()` - 获取任务节点总数
- `GetTaskNodeCountByStatus()` - 根据状态获取任务节点数量
- `GetTaskNodeCountByTask()` - 根据任务获取任务节点数量
- `GetTaskNodeCountByDepartment()` - 根据部门获取任务节点数量
- `GetTaskNodeCountByExecutor()` - 根据执行人获取任务节点数量

#### 任务日志表 (TaskLog)
**文件**: `tasklogmodel.go`

**主要CRUD操作**:
- `FindByTaskID()` - 根据任务ID查找任务日志
- `FindByTaskNodeID()` - 根据任务节点ID查找任务日志
- `FindByOperator()` - 根据操作人ID查找任务日志
- `FindByLogType()` - 根据日志类型查找任务日志
- `FindByPage()` - 分页查找任务日志
- `SearchTaskLogs()` - 搜索任务日志
- `GetTaskLogCount()` - 获取任务日志总数
- `GetTaskLogCountByTask()` - 根据任务获取任务日志数量
- `GetTaskLogCountByTaskNode()` - 根据任务节点获取任务日志数量
- `GetTaskLogCountByOperator()` - 根据操作人获取任务日志数量
- `GetTaskLogCountByLogType()` - 根据日志类型获取任务日志数量

#### 任务交接表 (TaskHandover)
**文件**: `taskhandovermodel.go`

**主要CRUD操作**:
- `FindByTaskID()` - 根据任务ID查找任务交接
- `FindByTaskNodeID()` - 根据任务节点ID查找任务交接
- `FindByFromEmployee()` - 根据交接人ID查找任务交接
- `FindByToEmployee()` - 根据接收人ID查找任务交接
- `FindByStatus()` - 根据状态查找任务交接
- `FindByPage()` - 分页查找任务交接
- `SearchTaskHandovers()` - 搜索任务交接
- `UpdateStatus()` - 更新任务交接状态
- `UpdateApproval()` - 更新任务交接审批
- `SoftDelete()` - 软删除任务交接
- `BatchUpdateStatus()` - 批量更新任务交接状态
- `GetTaskHandoverCount()` - 获取任务交接总数
- `GetTaskHandoverCountByStatus()` - 根据状态获取任务交接数量
- `GetTaskHandoverCountByTask()` - 根据任务获取任务交接数量
- `GetTaskHandoverCountByTaskNode()` - 根据任务节点获取任务交接数量
- `GetTaskHandoverCountByFromEmployee()` - 根据交接人获取任务交接数量
- `GetTaskHandoverCountByToEmployee()` - 根据接收人获取任务交接数量

## 通用功能特性

### 1. 查询功能
- **单条查询**: 根据唯一标识查找单条记录
- **列表查询**: 根据条件查找多条记录
- **分页查询**: 支持分页的列表查询
- **搜索功能**: 支持关键词模糊搜索
- **条件查询**: 根据各种条件组合查询

### 2. 更新功能
- **单字段更新**: 更新特定字段
- **批量更新**: 批量更新多条记录
- **软删除**: 使用delete_time字段标记删除
- **状态管理**: 统一的状态更新机制

### 3. 统计功能
- **计数查询**: 获取各种条件下的记录数量
- **分组统计**: 根据不同维度统计数据

### 4. 业务特性
- **关联查询**: 支持跨表关联查询
- **树形结构**: 支持部门等树形结构数据
- **权限控制**: 支持基于权限的数据过滤
- **搜索优化**: 支持多字段模糊搜索

## 使用示例

### 基本查询
```go
// 根据ID查找用户
user, err := userModel.FindOne(ctx, "user123")

// 根据用户名查找用户
user, err := userModel.FindByUsername(ctx, "zhangsan")

// 分页查找用户
users, total, err := userModel.FindByPage(ctx, 1, 10)
```

### 条件查询
```go
// 根据公司查找员工
employees, err := employeeModel.FindByCompanyID(ctx, "company123")

// 根据部门查找职位
positions, err := positionModel.FindByDepartmentID(ctx, "dept123")

// 搜索功能
users, total, err := userModel.SearchUsers(ctx, "张三", 1, 10)
```

### 更新操作
```go
// 更新用户状态
err := userModel.UpdateStatus(ctx, "user123", 1)

// 批量更新员工部门
err := employeeModel.BatchUpdateDepartment(ctx, []string{"emp1", "emp2"}, "dept123")

// 软删除
err := userModel.SoftDelete(ctx, "user123")
```

### 统计查询
```go
// 获取用户总数
count, err := userModel.GetUserCount(ctx)

// 根据状态获取员工数量
count, err := employeeModel.GetEmployeeCountByStatus(ctx, 1)
```

## 注意事项

1. **SQL注入防护**: 所有查询都使用参数化查询，防止SQL注入
2. **软删除**: 所有删除操作都是软删除，通过delete_time字段标记
3. **分页优化**: 分页查询包含总数统计，避免N+1查询问题
4. **索引优化**: 查询条件都基于数据库索引字段
5. **事务支持**: 所有操作都支持事务上下文
6. **错误处理**: 统一的错误处理机制，包括NotFound等特殊情况

## 待完成模块

- 通知模块 (Notification)
- 用户权限模块 (UserPermission)
- 员工角色关联模块 (EmployeeRole)
- 系统配置模块 (SystemConfig)
- 操作日志模块 (OperationLog)

所有CRUD操作都已经实现并经过测试，可以直接在业务逻辑中使用。

# 工具类使用指南

## 概述

本工具类集合提供了统一的响应处理、数据转换、参数验证和通用功能，大大减少了重复代码，提高了代码的可维护性和一致性。

## 工具类结构

```
task/internal/utils/
├── response.go      # 响应处理工具
├── converter.go     # 数据转换工具
├── validator.go     # 参数验证工具
├── common.go        # 通用工具
└── README.md        # 使用指南
```

## 1. 响应处理工具 (Response)

### 功能
- 统一的成功和错误响应格式
- 预定义的错误消息枚举
- 业务错误消息管理

### 使用示例

```go
import "task_Project/task/internal/utils"

// 成功响应
resp := utils.Response.Success(data)
resp := utils.Response.SuccessWithMessage("操作成功", data)
resp := utils.Response.SuccessWithKey("login", data)

// 错误响应
resp := utils.Response.Error(400, "参数错误")
resp := utils.Response.BusinessError("user_not_found")
resp := utils.Response.ValidationError("用户名不能为空")
resp := utils.Response.UnauthorizedError()
resp := utils.Response.NotFoundError("company_not_found")
resp := utils.Response.ConflictError("username_exists")
resp := utils.Response.InternalError("系统错误")
```

### 预定义错误消息

```go
// 业务错误消息
"user_not_found"           // 用户不存在
"username_exists"          // 用户名已存在
"email_exists"             // 邮箱已被注册
"company_not_found"        // 公司不存在
"company_name_exists"      // 公司名称已存在
"employee_already_exists"  // 该用户已经是该公司的员工
// ... 更多错误消息
```

## 2. 数据转换工具 (Converter)

### 功能
- Model层到Types层的自动转换
- 时间格式化
- 分页响应构建

### 使用示例

```go
import "task_Project/task/internal/utils"

// 单个对象转换
userInfo := utils.Converter.ToUserInfo(user)
companyInfo := utils.Converter.ToCompanyInfo(company)
employeeInfo := utils.Converter.ToEmployeeInfo(employee)

// 列表转换
userList := utils.Converter.ToUserInfoList(users)
companyList := utils.Converter.ToCompanyInfoList(companies)
employeeList := utils.Converter.ToEmployeeInfoList(employees)

// 分页响应
pageResp := utils.Converter.ToPageResponse(list, total, page, pageSize)
```

## 3. 参数验证工具 (Validator)

### 功能
- 字符串验证（空值、邮箱、手机号、密码等）
- 数值验证（范围、正数、非负数等）
- 分页参数验证
- 必填字段验证

### 使用示例

```go
import "task_Project/task/internal/utils"

// 基础验证
if utils.Validator.IsEmpty(str) {
    return utils.Response.ValidationError("字段不能为空")
}

// 格式验证
if emailErr := utils.Validator.ValidateEmail(email); emailErr != "" {
    return utils.Response.ValidationError(emailErr)
}

if phoneErr := utils.Validator.ValidatePhone(phone); phoneErr != "" {
    return utils.Response.ValidationError(phoneErr)
}

// 必填字段验证
requiredFields := map[string]string{
    "用户名": username,
    "密码":  password,
}
if errors := utils.Validator.ValidateRequired(requiredFields); len(errors) > 0 {
    return utils.Response.ValidationError(errors[0])
}

// 分页参数验证
page, pageSize, errors := utils.Validator.ValidatePageParams(req.Page, req.PageSize)
```

## 4. 通用工具 (Common)

### 功能
- 用户信息获取
- ID生成
- 时间处理
- 字符串处理
- 数值处理

### 使用示例

```go
import "task_Project/task/internal/utils"

// 获取当前用户信息
userID, ok := utils.Common.GetCurrentUserID(ctx)
username, _ := utils.Common.GetCurrentUsername(ctx)
realName, _ := utils.Common.GetCurrentRealName(ctx)

// ID生成
id := utils.Common.GenerateID()
idWithPrefix := utils.Common.GenerateIDWithPrefix("EMP")

// 时间处理
now := utils.Common.GetCurrentTime()
timeStr := utils.Common.FormatTime(now)
dateStr := utils.Common.FormatDate(now)

// 字符串处理
if utils.Common.IsEmptyString(str) {
    // 处理空字符串
}
trimmed := utils.Common.TrimString(str)
defaultStr := utils.Common.DefaultString(str, "默认值")

// 数值处理
maxVal := utils.Common.MaxInt(a, b)
minVal := utils.Common.MinInt(a, b)
defaultInt := utils.Common.DefaultInt(value, 0)
```

## 5. 完整使用示例

### 登录逻辑示例

```go
func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.BaseResponse, err error) {
    // 参数验证
    if utils.Validator.IsEmpty(req.Username) || utils.Validator.IsEmpty(req.Password) {
        return utils.Response.ValidationError("用户名和密码不能为空"), nil
    }

    // 查找用户
    userModel := user.NewUserModel(l.svcCtx.Config.MySQL.DataSource)
    userInfo, err := userModel.FindByUsername(l.ctx, req.Username)
    if err != nil {
        if errors.Is(err, user.ErrNotFound) {
            return utils.Response.BusinessError("login_failed"), nil
        }
        logx.Errorf("查找用户失败: %v", err)
        return utils.Response.InternalError("查找用户失败"), nil
    }

    // 检查用户状态
    if userInfo.Status != 1 {
        return utils.Response.BusinessError("user_disabled"), nil
    }

    // 验证密码
    err = bcrypt.CompareHashAndPassword([]byte(userInfo.PasswordHash), []byte(req.Password))
    if err != nil {
        return utils.Response.BusinessError("login_failed"), nil
    }

    // 生成JWT令牌
    token, err := l.svcCtx.JWTMiddleware.GenerateToken(userInfo.Id, userInfo.Username, userInfo.RealName, "user")
    if err != nil {
        logx.Errorf("生成JWT令牌失败: %v", err)
        return utils.Response.InternalError("生成JWT令牌失败"), nil
    }

    // 返回成功响应
    loginResp := types.LoginResponse{
        Token:    token,
        UserID:   userInfo.Id,
        Username: userInfo.Username,
        RealName: userInfo.RealName,
    }

    return utils.Response.SuccessWithKey("login", loginResp), nil
}
```

### 公司列表查询示例

```go
func (l *GetCompanyListLogic) GetCompanyList(req *types.CompanyListRequest) (resp *types.BaseResponse, err error) {
    // 参数验证和默认值设置
    req.Page, req.PageSize, _ = utils.Validator.ValidatePageParams(req.Page, req.PageSize)

    // 获取当前用户ID
    userID, ok := utils.Common.GetCurrentUserID(l.ctx)
    if !ok {
        return utils.Response.UnauthorizedError(), nil
    }

    // 查询公司列表
    companyModel := company.NewCompanyModel(l.svcCtx.Config.MySQL.DataSource)
    companies, total, err := companyModel.FindByPage(l.ctx, req.Page, req.PageSize)
    if err != nil {
        logx.Errorf("查询公司列表失败: %v", err)
        return utils.Response.InternalError("查询失败"), nil
    }

    // 过滤用户自己的公司
    filteredCompanies := []*company.Company{}
    for _, comp := range companies {
        if comp.Owner == userID {
            filteredCompanies = append(filteredCompanies, comp)
        }
    }

    // 转换为响应格式
    companyList := utils.Converter.ToCompanyInfoList(filteredCompanies)

    // 构建响应
    response := types.CompanyListResponse{
        List:     companyList,
        Total:    int64(len(filteredCompanies)),
        Page:     req.Page,
        PageSize: req.PageSize,
    }

    return utils.Response.SuccessWithKey("query", response), nil
}
```

## 6. 最佳实践

### 1. 统一错误处理
```go
// 推荐：使用预定义的错误消息
return utils.Response.BusinessError("user_not_found"), nil

// 不推荐：硬编码错误消息
return &types.BaseResponse{Code: 400, Msg: "用户不存在"}, nil
```

### 2. 参数验证
```go
// 推荐：使用验证工具
if emailErr := utils.Validator.ValidateEmail(req.Email); emailErr != "" {
    return utils.Response.ValidationError(emailErr), nil
}

// 不推荐：手动验证
if !strings.Contains(req.Email, "@") {
    return utils.Response.ValidationError("邮箱格式不正确"), nil
}
```

### 3. 数据转换
```go
// 推荐：使用转换工具
companyList := utils.Converter.ToCompanyInfoList(companies)

// 不推荐：手动转换
companyList := make([]types.CompanyInfo, 0, len(companies))
for _, comp := range companies {
    companyInfo := types.CompanyInfo{
        ID:   comp.Id,
        Name: comp.Name,
        // ... 更多字段
    }
    companyList = append(companyList, companyInfo)
}
```

### 4. 响应构建
```go
// 推荐：使用响应工具
return utils.Response.SuccessWithKey("create", data), nil

// 不推荐：手动构建响应
return &types.BaseResponse{
    Code: 200,
    Msg:  "创建成功",
    Data: data,
}, nil
```

## 7. 扩展指南

### 添加新的错误消息
在 `response.go` 的 `BusinessErrorMessages` 中添加：
```go
"new_error_key": "新的错误消息",
```

### 添加新的验证方法
在 `validator.go` 中添加：
```go
func (v *Validator) ValidateNewField(value string) string {
    // 验证逻辑
    return ""
}
```

### 添加新的转换方法
在 `converter.go` 中添加：
```go
func (c *Converter) ToNewInfo(model *NewModel) types.NewInfo {
    return types.NewInfo{
        // 转换逻辑
    }
}
```

## 8. 注意事项

1. **错误消息一致性**: 使用预定义的错误消息键，确保错误消息的一致性
2. **参数验证**: 在业务逻辑开始前进行参数验证
3. **数据转换**: 使用转换工具避免手动转换的重复代码
4. **响应格式**: 统一使用响应工具构建API响应
5. **日志记录**: 在关键操作点添加适当的日志记录
6. **错误处理**: 区分业务错误和系统错误，返回适当的错误码

通过使用这些工具类，可以大大减少重复代码，提高代码的可维护性和一致性。

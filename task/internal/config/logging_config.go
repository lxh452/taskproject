package config

// LoggingConfig 日志记录配置
type LoggingConfig struct {
	// Enabled 是否启用自动日志记录
	Enabled bool `json:"enabled"`

	// LogLevel 日志级别: debug, info, warn, error
	LogLevel string `json:"logLevel"`

	// AutoLogRequests 是否自动记录HTTP请求
	AutoLogRequests bool `json:"autoLogRequests"`

	// LogQueryOperations 是否记录查询操作
	LogQueryOperations bool `json:"logQueryOperations"`

	// LogResponseBody 是否记录响应体（谨慎使用，可能包含敏感信息）
	LogResponseBody bool `json:"logResponseBody"`

	// MaxLogSize 单条日志最大长度（字节）
	MaxLogSize int `json:"maxLogSize"`

	// ExcludePaths 排除的路径（不记录日志）
	ExcludePaths []string `json:"excludePaths"`

	// IncludePaths 包含的路径（只记录这些路径的日志）
	IncludePaths []string `json:"includePaths"`

	// SensitiveFields 敏感字段（在日志中脱敏）
	SensitiveFields []string `json:"sensitiveFields"`
}

// DefaultLoggingConfig 默认日志记录配置
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Enabled:            true,
		LogLevel:           "info",
		AutoLogRequests:    true,
		LogQueryOperations: false,
		LogResponseBody:    false,
		MaxLogSize:         1024,
		ExcludePaths: []string{
			"/api/health",
			"/api/metrics",
			"/api/static",
			"/api/assets",
		},
		SensitiveFields: []string{
			"password",
			"token",
			"secret",
			"key",
			"email",
			"phone",
			"id_card",
		},
	}
}

// LoggingPolicy 日志记录策略
type LoggingPolicy struct {
	// ModulePolicies 模块级别的日志策略
	ModulePolicies map[string]ModulePolicy `json:"modulePolicies"`

	// ActionPolicies 操作级别的日志策略
	ActionPolicies map[string]ActionPolicy `json:"actionPolicies"`
}

// ModulePolicy 模块级别的日志策略
type ModulePolicy struct {
	// Enabled 是否启用该模块的日志记录
	Enabled bool `json:"enabled"`

	// LogLevel 该模块的日志级别
	LogLevel string `json:"logLevel"`

	// LogActions 需要记录的操作列表（空表示记录所有操作）
	LogActions []string `json:"logActions"`

	// ExcludeActions 排除的操作列表
	ExcludeActions []string `json:"excludeActions"`
}

// ActionPolicy 操作级别的日志策略
type ActionPolicy struct {
	// Enabled 是否启用该操作的日志记录
	Enabled bool `json:"enabled"`

	// LogLevel 该操作的日志级别
	LogLevel string `json:"logLevel"`

	// LogRequestBody 是否记录请求体
	LogRequestBody bool `json:"logRequestBody"`

	// LogResponseBody 是否记录响应体
	LogResponseBody bool `json:"logResponseBody"`

	// Sensitive 是否包含敏感信息
	Sensitive bool `json:"sensitive"`
}

// DefaultLoggingPolicy 默认日志记录策略
func DefaultLoggingPolicy() LoggingPolicy {
	return LoggingPolicy{
		ModulePolicies: map[string]ModulePolicy{
			"auth": {
				Enabled:        true,
				LogLevel:       "info",
				LogActions:     []string{"login", "logout", "register", "reset_password"},
				ExcludeActions: []string{"verify"},
			},
			"task": {
				Enabled:    true,
				LogLevel:   "info",
				LogActions: []string{"create", "update", "delete", "complete"},
			},
			"user": {
				Enabled:    true,
				LogLevel:   "info",
				LogActions: []string{"create", "update", "delete", "ban", "unban"},
			},
			"company": {
				Enabled:    true,
				LogLevel:   "info",
				LogActions: []string{"create", "update", "delete", "enable", "disable"},
			},
			"employee": {
				Enabled:    true,
				LogLevel:   "info",
				LogActions: []string{"create", "update", "delete", "join", "leave"},
			},
			"file": {
				Enabled:    true,
				LogLevel:   "info",
				LogActions: []string{"upload", "delete", "download"},
			},
			"role": {
				Enabled:    true,
				LogLevel:   "info",
				LogActions: []string{"create", "assign", "revoke", "update"},
			},
		},
		ActionPolicies: map[string]ActionPolicy{
			"login": {
				Enabled:        true,
				LogLevel:       "info",
				LogRequestBody: false, // 不记录密码等敏感信息
				Sensitive:      true,
			},
			"register": {
				Enabled:        true,
				LogLevel:       "info",
				LogRequestBody: false,
				Sensitive:      true,
			},
			"create": {
				Enabled:         true,
				LogLevel:        "info",
				LogRequestBody:  true,
				LogResponseBody: false,
			},
			"delete": {
				Enabled:         true,
				LogLevel:        "warn",
				LogRequestBody:  true,
				LogResponseBody: false,
			},
			"update": {
				Enabled:         true,
				LogLevel:        "info",
				LogRequestBody:  true,
				LogResponseBody: false,
			},
		},
	}
}

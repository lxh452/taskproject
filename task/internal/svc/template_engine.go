package svc

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/collection"
	"github.com/zeromicro/go-zero/core/logx"
)

// TemplateType 模板类型
type TemplateType string

const (
	TemplateTypeEmail        TemplateType = "email"
	TemplateTypeNotification TemplateType = "notification"
	TemplateTypeDocument     TemplateType = "document"
)

// TemplateData 统一模板数据模型
type TemplateData struct {
	// 基础信息
	Type     TemplateType `json:"type"`
	Category string       `json:"category"`

	// 内容信息
	Title   string                 `json:"title"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`

	// 样式配置
	Style TemplateStyle `json:"style"`

	// 行为配置
	Actions []TemplateAction `json:"actions"`

	// 元数据
	Metadata TemplateMetadata `json:"metadata"`
}

// TemplateStyle 模板样式配置
type TemplateStyle struct {
	Theme      string `json:"theme"`      // 主题色：primary, success, warning, danger, info
	Background string `json:"background"` // 背景色
	Border     string `json:"border"`     // 边框色
}

// TemplateAction 模板行为配置
type TemplateAction struct {
	Text string `json:"text"` // 按钮文本
	URL  string `json:"url"`  // 跳转链接
	Type string `json:"type"` // 按钮类型：primary, secondary, danger
}

// TemplateMetadata 模板元数据
type TemplateMetadata struct {
	Timestamp time.Time `json:"timestamp"` // 时间戳
	Source    string    `json:"source"`    // 来源
	Priority  int       `json:"priority"`  // 优先级
}

// TemplateDefinition 模板定义
type TemplateDefinition struct {
	Name     string       `json:"name"`     // 模板名称
	Type     TemplateType `json:"type"`     // 模板类型
	Category string       `json:"category"` // 分类
	File     string       `json:"file"`     // 模板文件路径
	Priority int          `json:"priority"` // 优先级
}

// TemplateEngine 通用模板引擎
type TemplateEngine struct {
	mu          sync.RWMutex
	templates   map[string]*template.Template // 已编译的模板
	definitions map[string]TemplateDefinition // 模板定义
	cache       *collection.Cache             // 模板缓存
	baseDir     string                        // 模板基础目录
}

// NewTemplateEngine 创建新的模板引擎
func NewTemplateEngine(baseDir string) (*TemplateEngine, error) {
	if baseDir == "" {
		baseDir = "./templates"
	}

	// 创建缓存，TTL为1小时
	cache, err := collection.NewCache(time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	engine := &TemplateEngine{
		templates:   make(map[string]*template.Template),
		definitions: make(map[string]TemplateDefinition),
		cache:       cache,
		baseDir:     baseDir,
	}

	// 自动加载模板
	if err := engine.loadTemplates(); err != nil {
		logx.Errorf("Failed to load templates: %v", err)
		// 不返回错误，允许后续手动加载
	}

	return engine, nil
}

// loadTemplates 加载所有模板
func (e *TemplateEngine) loadTemplates() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 清空现有模板
	e.templates = make(map[string]*template.Template)
	e.definitions = make(map[string]TemplateDefinition)

	// 遍历模板目录
	return filepath.WalkDir(e.baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".tpl") {
			return nil
		}

		// 计算相对路径作为模板名称
		relPath, err := filepath.Rel(e.baseDir, path)
		if err != nil {
			return err
		}

		templateName := strings.TrimSuffix(relPath, ".tpl")
		templateName = strings.ReplaceAll(templateName, string(filepath.Separator), "/")

		// 解析模板类型和分类
		parts := strings.Split(templateName, "/")
		if len(parts) < 2 {
			logx.Infof("Invalid template path format: %s", templateName)
			return nil
		}

		templateType := TemplateType(parts[0])
		category := parts[1]
		name := templateName
		if len(parts) > 2 {
			name = strings.Join(parts[2:], "/")
		}

		// 加载模板
		tmpl, err := template.ParseFiles(path)
		if err != nil {
			logx.Errorf("Failed to parse template %s: %v", path, err)
			return nil
		}

		// 注册模板
		e.templates[templateName] = tmpl
		e.definitions[templateName] = TemplateDefinition{
			Name:     name,
			Type:     templateType,
			Category: category,
			File:     path,
			Priority: 1,
		}

		logx.Infof("Loaded template: %s", templateName)
		return nil
	})
}

// Render 渲染模板
func (e *TemplateEngine) Render(templateName string, data TemplateData) (string, error) {
	// 检查缓存
	if cached, ok := e.cache.Get(templateName); ok {
		if result, ok := cached.(string); ok {
			return result, nil
		}
	}

	e.mu.RLock()
	tmpl, exists := e.templates[templateName]
	e.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("template %s not found", templateName)
	}

	// 执行模板渲染
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	result := buf.String()

	// 缓存结果
	e.cache.Set(templateName, result)

	return result, nil
}

// RenderWithFallback 渲染模板，支持回退
func (e *TemplateEngine) RenderWithFallback(primaryTemplate string, fallbackTemplate string, data TemplateData) (string, error) {
	result, err := e.Render(primaryTemplate, data)
	if err != nil && fallbackTemplate != "" {
		logx.Infof("Failed to render primary template %s, using fallback: %v", primaryTemplate, err)
		return e.Render(fallbackTemplate, data)
	}
	return result, err
}

// RegisterTemplate 手动注册模板
func (e *TemplateEngine) RegisterTemplate(def TemplateDefinition, content string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	tmpl, err := template.New(def.Name).Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse template content: %w", err)
	}

	templateName := fmt.Sprintf("%s/%s/%s", def.Type, def.Category, def.Name)
	e.templates[templateName] = tmpl
	e.definitions[templateName] = def

	logx.Infof("Registered template: %s", templateName)
	return nil
}

// ListTemplates 列出所有模板
func (e *TemplateEngine) ListTemplates(filterType TemplateType, category string) []TemplateDefinition {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var result []TemplateDefinition
	for _, def := range e.definitions {
		if (filterType == "" || def.Type == filterType) &&
			(category == "" || def.Category == category) {
			result = append(result, def)
		}
	}
	return result
}

// ClearCache 清理缓存（go-zero v1.9.3 的 Cache 没有 Clear 方法，改为重新创建）
func (e *TemplateEngine) ClearCache() {
	// go-zero v1.9.3 的 Cache 没有 Clear 方法，重新创建缓存实例
	cache, err := collection.NewCache(time.Hour)
	if err != nil {
		logx.Errorf("Failed to recreate cache: %v", err)
		return
	}
	e.cache = cache
	logx.Info("Template cache cleared")
}

// Reload 重新加载模板
func (e *TemplateEngine) Reload() error {
	return e.loadTemplates()
}

// GetTemplateInfo 获取模板信息
func (e *TemplateEngine) GetTemplateInfo(templateName string) (TemplateDefinition, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	def, exists := e.definitions[templateName]
	return def, exists
}

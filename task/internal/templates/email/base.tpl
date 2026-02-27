<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        /* 基础样式 */
        body {
            margin: 0;
            padding: 0;
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
            background-color: #f4f5f7;
            line-height: 1.6;
        }
        
        .email-container {
            max-width: 600px;
            margin: 32px auto;
            background: #ffffff;
            border-radius: 12px;
            box-shadow: 0 4px 20px rgba(0,0,0,0.08);
            overflow: hidden;
        }
        
        .email-header {
            height: 4px;
            background: {{if eq .Style.Theme "success"}}linear-gradient(90deg, #059669, #10b981){{else if eq .Style.Theme "warning"}}linear-gradient(90deg, #d97706, #f59e0b){{else if eq .Style.Theme "danger"}}linear-gradient(90deg, #dc2626, #ef4444){{else if eq .Style.Theme "info"}}linear-gradient(90deg, #0284c7, #0ea5e9){{else}}linear-gradient(90deg, #ea580c, #f97316){{end}};
        }
        
        .email-badge {
            display: inline-block;
            padding: 6px 14px;
            font-size: 12px;
            font-weight: 600;
            border-radius: 20px;
            letter-spacing: 0.5px;
            margin: 32px 40px 24px;
            {{if eq .Style.Theme "success"}}background: #d1fae5; color: #065f46;{{else if eq .Style.Theme "warning"}}background: #fef3c7; color: #92400e;{{else if eq .Style.Theme "danger"}}background: #fee2e2; color: #991b1b;{{else if eq .Style.Theme "info"}}background: #dbeafe; color: #1e40af;{{else}}background: #ffedd5; color: #9a3412;{{end}}
        }
        
        .email-content {
            padding: 0 40px 32px;
        }
        
        .email-title {
            margin: 0 0 8px;
            font-size: 24px;
            font-weight: 700;
            color: #111827;
        }
        
        .email-message {
            margin: 0 0 24px;
            font-size: 15px;
            color: #6b7280;
        }
        
        .info-card {
            background: {{if eq .Style.Theme "success"}}#f0fdf4{{else if eq .Style.Theme "warning"}}#fffbeb{{else if eq .Style.Theme "danger"}}#fef2f2{{else if eq .Style.Theme "info"}}#f0f9ff{{else}}#fff7ed{{end}};
            border: 1px solid {{if eq .Style.Theme "success"}}#bbf7d0{{else if eq .Style.Theme "warning"}}#fde68a{{else if eq .Style.Theme "danger"}}#fecaca{{else if eq .Style.Theme "info"}}#bae6fd{{else}}#fed7aa{{end}};
            border-radius: 10px;
            margin-bottom: 24px;
        }
        
        .info-row {
            padding: 18px 24px;
            border-bottom: 1px solid {{if eq .Style.Theme "success"}}#bbf7d0{{else if eq .Style.Theme "warning"}}#fde68a{{else if eq .Style.Theme "danger"}}#fecaca{{else if eq .Style.Theme "info"}}#bae6fd{{else}}#fed7aa{{end}};
        }
        
        .info-row:last-child {
            border-bottom: none;
        }
        
        .info-label {
            margin: 0 0 4px;
            font-size: 12px;
            color: #9ca3af;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        
        .info-value {
            margin: 0;
            font-size: 15px;
            font-weight: 600;
            color: #111827;
        }
        
        .actions {
            display: flex;
            gap: 12px;
            margin-top: 24px;
        }
        
        .btn {
            display: inline-block;
            padding: 12px 24px;
            text-decoration: none;
            border-radius: 8px;
            font-weight: 600;
            font-size: 14px;
            border: none;
            cursor: pointer;
        }
        
        .btn-primary {
            background: #ea580c;
            color: white;
        }
        
        .btn-secondary {
            background: #6b7280;
            color: white;
        }
        
        .btn-danger {
            background: #dc2626;
            color: white;
        }
        
        .email-footer {
            padding: 24px 40px;
            background: #f9fafb;
            border-top: 1px solid #e5e7eb;
            text-align: center;
            color: #6b7280;
            font-size: 12px;
        }
    </style>
</head>
<body>
    <div class="email-container">
        <!-- 顶部品牌条 -->
        <div class="email-header"></div>
        
        <!-- 头部 -->
        <div class="email-badge">{{.Title}}</div>
        
        <!-- 内容区域 -->
        <div class="email-content">
            <h1 class="email-title">{{.Title}}</h1>
            <p class="email-message">{{.Message}}</p>
            
            {{template "content" .}}
            
            <!-- 操作按钮 -->
            {{if .Actions}}
            <div class="actions">
                {{range .Actions}}
                <a href="{{.URL}}" class="btn btn-{{.Type}}">{{.Text}}</a>
                {{end}}
            </div>
            {{end}}
        </div>
        
        <!-- 底部 -->
        <div class="email-footer">
            <p>此邮件由任务管理系统自动发送，请勿回复</p>
            <p>&copy; {{.Metadata.Timestamp.Year}} 任务管理系统</p>
        </div>
    </div>
</body>
</html>
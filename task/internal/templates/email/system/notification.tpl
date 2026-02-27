{{define "system_notification"}}
{{template "base" .}}
{{end}}

{{define "content"}}
{{if .Data}}
<div class="info-card">
    {{if .Data.eventType}}
    <div class="info-row">
        <p class="info-label">事件类型</p>
        <p class="info-value">{{.Data.eventType}}</p>
    </div>
    {{end}}
    
    {{if .Data.systemMessage}}
    <div class="info-row">
        <p class="info-label">系统消息</p>
        <p class="info-value">{{.Data.systemMessage}}</p>
    </div>
    {{end}}
    
    {{if .Data.timestamp}}
    <div class="info-row">
        <p class="info-label">发生时间</p>
        <p class="info-value">{{.Data.timestamp}}</p>
    </div>
    {{end}}
    
    {{if .Data.severity}}
    <div class="info-row">
        <p class="info-label">严重程度</p>
        <p class="info-value" style="color: {{if eq .Data.severity "high"}}#dc2626{{else if eq .Data.severity "medium"}}#d97706{{else}}#059669{{end}};">
            {{.Data.severity}}
        </p>
    </div>
    {{end}}
</div>
{{end}}

{{if .Data.details}}
<div style="background: #f9fafb; border: 1px solid #e5e7eb; border-radius: 8px; padding: 16px; margin-top: 16px;">
    <p style="margin: 0 0 8px; font-size: 14px; font-weight: 600; color: #374151;">详细信息：</p>
    <pre style="margin: 0; font-size: 12px; color: #6b7280; white-space: pre-wrap;">{{.Data.details}}</pre>
</div>
{{end}}
{{end}}
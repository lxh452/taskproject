{{define "task_notification"}}
{{template "base" .}}
{{end}}

{{define "content"}}
{{if .Data}}
<div class="info-card">
    {{if .Data.taskTitle}}
    <div class="info-row">
        <p class="info-label">任务名称</p>
        <p class="info-value">{{.Data.taskTitle}}</p>
    </div>
    {{end}}
    
    {{if .Data.nodeName}}
    <div class="info-row">
        <p class="info-label">节点名称</p>
        <p class="info-value">{{.Data.nodeName}}</p>
    </div>
    {{end}}
    
    {{if .Data.deadline}}
    <div class="info-row">
        <p class="info-label">截止时间</p>
        <p class="info-value" style="color: #dc2626;">{{.Data.deadline}}</p>
    </div>
    {{end}}
    
    {{if .Data.progress}}
    <div class="info-row">
        <p class="info-label">当前进度</p>
        <p class="info-value" style="color: #ea580c;">{{.Data.progress}}%</p>
    </div>
    {{end}}
    
    {{if .Data.completeTime}}
    <div class="info-row">
        <p class="info-label">完成时间</p>
        <p class="info-value" style="color: #059669;">{{.Data.completeTime}}</p>
    </div>
    {{end}}
    
    {{if .Data.employeeName}}
    <div class="info-row">
        <p class="info-label">相关人员</p>
        <p class="info-value">{{.Data.employeeName}}</p>
    </div>
    {{end}}
    
    {{if .Data.message}}
    <div class="info-row">
        <p class="info-label">附加信息</p>
        <p class="info-value">{{.Data.message}}</p>
    </div>
    {{end}}
</div>
{{end}}

{{if .Data.taskId}}
<p style="color: #6b7280; font-size: 14px;">
    任务ID: {{.Data.taskId}}
</p>
{{end}}
{{end}}
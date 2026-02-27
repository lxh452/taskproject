{{define "auth_verification_code"}}
{{template "base" .}}
{{end}}

{{define "content"}}
<div class="info-card">
    <div class="info-row" style="text-align: center; padding: 32px 24px;">
        <p class="info-label" style="font-size: 14px; margin-bottom: 16px;">您的验证码</p>
        <p class="info-value" style="font-size: 42px; font-weight: bold; letter-spacing: 12px; color: {{if eq .Style.Theme "danger"}}#dc2626{{else}}#ea580c{{end}};">
            {{.Data.code}}
        </p>
    </div>
</div>

<div style="background: #f9fafb; border-radius: 8px; padding: 16px; margin-top: 20px;">
    <p style="margin: 0 0 8px; font-size: 13px; color: #6b7280;">
        <span style="color: #9ca3af;">⏰</span> 验证码有效期为 <strong>5 分钟</strong>，请尽快使用
    </p>
    <p style="margin: 0; font-size: 13px; color: #6b7280;">
        <span style="color: #9ca3af;">🔒</span> 如非本人操作，请忽略此邮件
    </p>
</div>

{{if .Data.purpose}}
<div style="margin-top: 20px; padding: 12px 16px; background: #fff7ed; border-left: 4px solid #f97316; border-radius: 4px;">
    <p style="margin: 0; font-size: 13px; color: #9a3412;">
        用途：{{.Data.purpose}}
    </p>
</div>
{{end}}
{{end}}

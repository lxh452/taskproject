<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>任务更新通知</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif; background-color: #f4f5f7; line-height: 1.6;">
    <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="padding: 32px 16px;">
        <tr>
            <td align="center">
                <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="max-width: 600px; background: #ffffff; border-radius: 4px; box-shadow: 0 1px 3px rgba(0,0,0,0.08);">
                    <!-- 顶部品牌条 -->
                    <tr>
                        <td style="height: 4px; background: #0891b2;"></td>
                    </tr>
                    <!-- 头部 -->
                    <tr>
                        <td style="padding: 32px 40px 24px;">
                            <table role="presentation" width="100%" cellspacing="0" cellpadding="0">
                                <tr>
                                    <td>
                                        <span style="display: inline-block; padding: 6px 12px; background: #cffafe; color: #155e75; font-size: 12px; font-weight: 600; border-radius: 4px; letter-spacing: 0.5px;">内容更新</span>
                                    </td>
                                </tr>
                            </table>
                        </td>
                    </tr>
                    <!-- 主体内容 -->
                    <tr>
                        <td style="padding: 0 40px;">
                            <h1 style="margin: 0 0 8px; font-size: 22px; font-weight: 600; color: #111827;">任务更新通知</h1>
                            <p style="margin: 0 0 24px; font-size: 14px; color: #6b7280;">{{.EmployeeName}}，您参与的任务有新的更新</p>
                        </td>
                    </tr>
                    <!-- 任务信息 -->
                    <tr>
                        <td style="padding: 0 40px 24px;">
                            <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background: #f9fafb; border: 1px solid #e5e7eb; border-radius: 6px;">
                                <tr>
                                    <td style="padding: 20px 24px; border-bottom: 1px solid #e5e7eb;">
                                        <p style="margin: 0 0 4px; font-size: 12px; color: #6b7280; text-transform: uppercase; letter-spacing: 0.5px;">任务名称</p>
                                        <p style="margin: 0; font-size: 16px; font-weight: 600; color: #111827;">{{.TaskTitle}}</p>
                                    </td>
                                </tr>
                                <tr>
                                    <td style="padding: 16px 24px; border-bottom: 1px solid #e5e7eb;">
                                        <p style="margin: 0 0 4px; font-size: 12px; color: #6b7280; text-transform: uppercase; letter-spacing: 0.5px;">更新内容</p>
                                        <p style="margin: 0; font-size: 14px; color: #374151;">{{.UpdateContent}}</p>
                                    </td>
                                </tr>
                                <tr>
                                    <td style="padding: 16px 24px;">
                                        <p style="margin: 0 0 4px; font-size: 12px; color: #6b7280; text-transform: uppercase; letter-spacing: 0.5px;">更新时间</p>
                                        <p style="margin: 0; font-size: 14px; color: #374151;">{{.UpdateTime}}</p>
                                    </td>
                                </tr>
                            </table>
                        </td>
                    </tr>
                    <!-- 操作按钮 -->
                    <tr>
                        <td style="padding: 0 40px 32px;">
                            {{if .BaseURL}}
                            <a href="{{.BaseURL}}/#/tasks/detail/{{.TaskId}}" style="display: inline-block; padding: 12px 24px; background: #0891b2; color: #ffffff; text-decoration: none; font-size: 14px; font-weight: 500; border-radius: 4px;">查看任务详情</a>
                            {{else}}
                            <a href="/#/tasks/detail/{{.TaskId}}" style="display: inline-block; padding: 12px 24px; background: #0891b2; color: #ffffff; text-decoration: none; font-size: 14px; font-weight: 500; border-radius: 4px;">查看任务详情</a>
                            {{end}}
                        </td>
                    </tr>
                    <!-- 页脚 -->
                    <tr>
                        <td style="padding: 24px 40px; background: #f9fafb; border-top: 1px solid #e5e7eb;">
                            <p style="margin: 0 0 4px; font-size: 12px; color: #9ca3af;">此邮件由系统自动发送，请勿直接回复。</p>
                            <p style="margin: 0; font-size: 12px; color: #9ca3af;">© {{.Year}} 企业任务管理平台</p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>

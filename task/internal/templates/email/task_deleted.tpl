<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>任务删除通知</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif; background-color: #f4f5f7; line-height: 1.6;">
    <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="padding: 32px 16px;">
        <tr>
            <td align="center">
                <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="max-width: 600px; background: #ffffff; border-radius: 4px; box-shadow: 0 1px 3px rgba(0,0,0,0.08);">
                    <!-- 顶部品牌条 -->
                    <tr>
                        <td style="height: 4px; background: #dc2626;"></td>
                    </tr>
                    <!-- 头部 -->
                    <tr>
                        <td style="padding: 32px 40px 24px;">
                            <table role="presentation" width="100%" cellspacing="0" cellpadding="0">
                                <tr>
                                    <td>
                                        <span style="display: inline-block; padding: 6px 12px; background: #fee2e2; color: #991b1b; font-size: 12px; font-weight: 600; border-radius: 4px; letter-spacing: 0.5px;">已删除</span>
                                    </td>
                                </tr>
                            </table>
                        </td>
                    </tr>
                    <!-- 主体内容 -->
                    <tr>
                        <td style="padding: 0 40px;">
                            <h1 style="margin: 0 0 8px; font-size: 22px; font-weight: 600; color: #111827;">任务删除通知</h1>
                            <p style="margin: 0 0 24px; font-size: 14px; color: #6b7280;">{{.EmployeeName}}，您参与的任务已被删除</p>
                        </td>
                    </tr>
                    <!-- 任务信息 -->
                    <tr>
                        <td style="padding: 0 40px 24px;">
                            <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background: #fef2f2; border: 1px solid #fecaca; border-radius: 6px;">
                                <tr>
                                    <td style="padding: 20px 24px; border-bottom: 1px solid #fecaca;">
                                        <p style="margin: 0 0 4px; font-size: 12px; color: #6b7280; text-transform: uppercase; letter-spacing: 0.5px;">任务名称</p>
                                        <p style="margin: 0; font-size: 16px; font-weight: 600; color: #111827;">{{.TaskTitle}}</p>
                                    </td>
                                </tr>
                                <tr>
                                    <td style="padding: 16px 24px; border-bottom: 1px solid #fecaca;">
                                        <p style="margin: 0 0 4px; font-size: 12px; color: #6b7280; text-transform: uppercase; letter-spacing: 0.5px;">删除时间</p>
                                        <p style="margin: 0; font-size: 14px; color: #374151;">{{.DeleteTime}}</p>
                                    </td>
                                </tr>
                                <tr>
                                    <td style="padding: 16px 24px;">
                                        <p style="margin: 0 0 4px; font-size: 12px; color: #6b7280; text-transform: uppercase; letter-spacing: 0.5px;">操作人</p>
                                        <p style="margin: 0; font-size: 14px; color: #374151;">{{.DeletedBy}}</p>
                                    </td>
                                </tr>
                            </table>
                        </td>
                    </tr>
                    <!-- 提示信息 -->
                    <tr>
                        <td style="padding: 0 40px 32px;">
                            <p style="margin: 0; font-size: 13px; color: #6b7280;">如有疑问，请联系任务管理员或相关负责人。</p>
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

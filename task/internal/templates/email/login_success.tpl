<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>登录通知</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif; background-color: #f4f5f7; line-height: 1.6;">
    <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="padding: 32px 16px;">
        <tr>
            <td align="center">
                <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="max-width: 600px; background: #ffffff; border-radius: 4px; box-shadow: 0 1px 3px rgba(0,0,0,0.08);">
                    <!-- 顶部品牌条 -->
                    <tr>
                        <td style="height: 4px; background: #059669;"></td>
                    </tr>
                    <!-- 头部 -->
                    <tr>
                        <td style="padding: 32px 40px 24px;">
                            <table role="presentation" width="100%" cellspacing="0" cellpadding="0">
                                <tr>
                                    <td>
                                        <span style="display: inline-block; padding: 6px 12px; background: #d1fae5; color: #065f46; font-size: 12px; font-weight: 600; border-radius: 4px; letter-spacing: 0.5px;">安全通知</span>
                                    </td>
                                </tr>
                            </table>
                        </td>
                    </tr>
                    <!-- 主体内容 -->
                    <tr>
                        <td style="padding: 0 40px;">
                            <h1 style="margin: 0 0 8px; font-size: 22px; font-weight: 600; color: #111827;">登录成功通知</h1>
                            <p style="margin: 0 0 24px; font-size: 14px; color: #6b7280;">{{.Username}}，您的账户刚刚完成了一次登录</p>
                        </td>
                    </tr>
                    <!-- 登录信息 -->
                    <tr>
                        <td style="padding: 0 40px 24px;">
                            <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background: #f9fafb; border: 1px solid #e5e7eb; border-radius: 6px;">
                                <tr>
                                    <td style="padding: 16px 24px; border-bottom: 1px solid #e5e7eb;">
                                        <p style="margin: 0 0 4px; font-size: 12px; color: #6b7280; text-transform: uppercase; letter-spacing: 0.5px;">登录时间</p>
                                        <p style="margin: 0; font-size: 14px; color: #374151;">{{.LoginTime}}</p>
                                    </td>
                                </tr>
                                <tr>
                                    <td style="padding: 16px 24px; border-bottom: 1px solid #e5e7eb;">
                                        <p style="margin: 0 0 4px; font-size: 12px; color: #6b7280; text-transform: uppercase; letter-spacing: 0.5px;">登录IP</p>
                                        <p style="margin: 0; font-size: 14px; color: #374151; font-family: 'SF Mono', Monaco, Consolas, monospace;">{{.LoginIP}}</p>
                                    </td>
                                </tr>
                                <tr>
                                    <td style="padding: 16px 24px;">
                                        <p style="margin: 0 0 4px; font-size: 12px; color: #6b7280; text-transform: uppercase; letter-spacing: 0.5px;">设备信息</p>
                                        <p style="margin: 0; font-size: 14px; color: #374151;">{{.DeviceInfo}}</p>
                                    </td>
                                </tr>
                            </table>
                        </td>
                    </tr>
                    <!-- 提示信息 -->
                    <tr>
                        <td style="padding: 0 40px 32px;">
                            <p style="margin: 0; font-size: 13px; color: #6b7280;">如非本人操作，请立即修改密码并联系管理员。</p>
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

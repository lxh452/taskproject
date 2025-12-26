<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>任务交接通知</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif; background-color: #f4f5f7; line-height: 1.6;">
    <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="padding: 32px 16px;">
        <tr>
            <td align="center">
                <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="max-width: 600px; background: #ffffff; border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.08);">
                    <!-- 顶部品牌条 -->
                    <tr>
                        <td style="height: 4px; background: linear-gradient(90deg, #7c3aed, #8b5cf6); border-radius: 12px 12px 0 0;"></td>
                    </tr>
                    <!-- 头部 -->
                    <tr>
                        <td style="padding: 32px 40px 24px;">
                            <span style="display: inline-block; padding: 6px 14px; background: #ede9fe; color: #5b21b6; font-size: 12px; font-weight: 600; border-radius: 20px; letter-spacing: 0.5px;">任务交接</span>
                        </td>
                    </tr>
                    <!-- 主体内容 -->
                    <tr>
                        <td style="padding: 0 40px;">
                            <h1 style="margin: 0 0 8px; font-size: 24px; font-weight: 700; color: #111827;">任务交接通知</h1>
                            <p style="margin: 0 0 24px; font-size: 15px; color: #6b7280;">{{.EmployeeName}}，您有一项任务交接需要处理</p>
                        </td>
                    </tr>
                    <!-- 交接信息 -->
                    <tr>
                        <td style="padding: 0 40px 24px;">
                            <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background: #faf5ff; border: 1px solid #e9d5ff; border-radius: 10px;">
                                <tr>
                                    <td style="padding: 20px 24px;">
                                        <p style="margin: 0; font-size: 15px; color: #374151; line-height: 1.7;">{{.Message}}</p>
                                    </td>
                                </tr>
                            </table>
                        </td>
                    </tr>
                    <!-- 详细信息 -->
                    <tr>
                        <td style="padding: 0 40px 24px;">
                            <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background: #f9fafb; border: 1px solid #e5e7eb; border-radius: 10px;">
                                <tr>
                                    <td style="padding: 18px 24px; border-bottom: 1px solid #e5e7eb;">
                                        <p style="margin: 0 0 4px; font-size: 12px; color: #9ca3af; text-transform: uppercase; letter-spacing: 0.5px;">交接编号</p>
                                        <p style="margin: 0; font-size: 13px; color: #6b7280; font-family: 'SF Mono', Monaco, Consolas, monospace;">{{.HandoverID}}</p>
                                    </td>
                                </tr>
                                {{if .TaskTitle}}
                                <tr>
                                    <td style="padding: 18px 24px;">
                                        <p style="margin: 0 0 4px; font-size: 12px; color: #9ca3af; text-transform: uppercase; letter-spacing: 0.5px;">相关任务</p>
                                        <p style="margin: 0; font-size: 15px; font-weight: 500; color: #111827;">{{.TaskTitle}}</p>
                                    </td>
                                </tr>
                                {{end}}
                            </table>
                        </td>
                    </tr>
                    <!-- 提示信息 -->
                    <tr>
                        <td style="padding: 0 40px 32px;">
                            <p style="margin: 0; font-size: 13px; color: #9ca3af;">请及时登录系统处理相关事务，如有疑问请联系相关负责人。</p>
                        </td>
                    </tr>
                    <!-- 页脚 -->
                    <tr>
                        <td style="padding: 24px 40px; background: #f9fafb; border-top: 1px solid #e5e7eb; border-radius: 0 0 12px 12px;">
                            <p style="margin: 0 0 4px; font-size: 12px; color: #9ca3af;">此邮件由系统自动发送，请勿直接回复。</p>
                            <p style="margin: 0; font-size: 12px; color: #9ca3af;">© {{.Year}} Task Helper</p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>

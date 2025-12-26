<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>任务进度预警</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif; background-color: #f4f5f7; line-height: 1.6;">
    <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="padding: 32px 16px;">
        <tr>
            <td align="center">
                <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="max-width: 600px; background: #ffffff; border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.08);">
                    <!-- 顶部品牌条 -->
                    <tr>
                        <td style="height: 4px; background: linear-gradient(90deg, #d97706, #f59e0b); border-radius: 12px 12px 0 0;"></td>
                    </tr>
                    <!-- 头部 -->
                    <tr>
                        <td style="padding: 32px 40px 24px;">
                            <span style="display: inline-block; padding: 6px 14px; background: #fef3c7; color: #92400e; font-size: 12px; font-weight: 600; border-radius: 20px; letter-spacing: 0.5px;">进度预警</span>
                        </td>
                    </tr>
                    <!-- 主体内容 -->
                    <tr>
                        <td style="padding: 0 40px;">
                            <h1 style="margin: 0 0 8px; font-size: 24px; font-weight: 700; color: #111827;">任务进度预警</h1>
                            <p style="margin: 0 0 24px; font-size: 15px; color: #6b7280;">以下任务节点进度滞后，请关注</p>
                        </td>
                    </tr>
                    <!-- 任务信息 -->
                    <tr>
                        <td style="padding: 0 40px 24px;">
                            <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background: #fffbeb; border: 1px solid #fde68a; border-radius: 10px;">
                                <tr>
                                    <td style="padding: 18px 24px; border-bottom: 1px solid #fde68a;">
                                        <p style="margin: 0 0 4px; font-size: 12px; color: #9ca3af; text-transform: uppercase; letter-spacing: 0.5px;">节点名称</p>
                                        <p style="margin: 0; font-size: 16px; font-weight: 600; color: #111827;">{{.NodeName}}</p>
                                    </td>
                                </tr>
                                <tr>
                                    <td style="padding: 18px 24px; border-bottom: 1px solid #fde68a;">
                                        <p style="margin: 0 0 4px; font-size: 12px; color: #9ca3af; text-transform: uppercase; letter-spacing: 0.5px;">当前进度</p>
                                        <p style="margin: 0; font-size: 15px; font-weight: 600; color: #d97706;">{{.Progress}}%</p>
                                    </td>
                                </tr>
                                <tr>
                                    <td style="padding: 18px 24px;">
                                        <p style="margin: 0 0 4px; font-size: 12px; color: #9ca3af; text-transform: uppercase; letter-spacing: 0.5px;">截止日期</p>
                                        <p style="margin: 0; font-size: 15px; color: #374151;">{{.Deadline}}</p>
                                    </td>
                                </tr>
                            </table>
                        </td>
                    </tr>
                    <!-- 提示信息 -->
                    <tr>
                        <td style="padding: 0 40px 32px;">
                            <p style="margin: 0; font-size: 13px; color: #9ca3af;">请登录系统查看任务详情并加快进度。</p>
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

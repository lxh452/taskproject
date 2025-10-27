package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
)

// ExampleUsage 中间件使用示例
func ExampleUsage() {
	// 这个文件展示了如何在业务逻辑中使用中间件
	// 实际使用时请删除此文件
}

// ExampleJWTUsage JWT中间件使用示例
func ExampleJWTUsage(w http.ResponseWriter, r *http.Request) {
	// 从上下文中获取用户信息
	userID, ok := GetUserID(r.Context())
	if !ok {
		http.Error(w, "用户ID获取失败", http.StatusUnauthorized)
		return
	}

	username, _ := GetUsername(r.Context())
	realName, _ := GetRealName(r.Context())
	role, _ := GetRole(r.Context())

	// 返回用户信息
	response := map[string]interface{}{
		"userId":   userID,
		"username": username,
		"realName": realName,
		"role":     role,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ExampleEmailUsage 邮件中间件使用示例
func ExampleEmailUsage(ctx context.Context, emailMiddleware *EmailMiddleware) {
	// 发送简单邮件
	msg := EmailMessage{
		To:      []string{"user@example.com"},
		Subject: "任务提醒",
		Body:    "您有一个新任务需要处理",
		IsHTML:  false,
	}

	if err := emailMiddleware.SendEmail(ctx, msg); err != nil {
		logx.Errorf("发送邮件失败: %v", err)
	}

	// 发送HTML邮件
	htmlMsg := EmailMessage{
		To:      []string{"user@example.com"},
		Subject: "任务详情",
		Body:    "<h1>任务详情</h1><p>您有一个新任务需要处理</p>",
		IsHTML:  true,
	}

	if err := emailMiddleware.SendEmail(ctx, htmlMsg); err != nil {
		logx.Errorf("发送HTML邮件失败: %v", err)
	}

	// 发送模板邮件
	template := "您好 %s，您有一个新任务：%s"
	data := map[string]interface{}{
		"name": "张三",
		"task": "完成项目文档",
	}

	if err := emailMiddleware.SendTemplateEmail(ctx, []string{"user@example.com"}, "任务提醒", template, data); err != nil {
		logx.Errorf("发送模板邮件失败: %v", err)
	}
}

// ExampleSMSUsage 短信中间件使用示例
func ExampleSMSUsage(ctx context.Context, smsMiddleware *SMSMiddleware) {
	// 发送验证码短信
	if err := smsMiddleware.SendVerificationCode(ctx, "13800138000", "123456"); err != nil {
		logx.Errorf("发送验证码失败: %v", err)
	}

	// 发送通知短信
	if err := smsMiddleware.SendNotificationSMS(ctx, "13800138000", "您有一个新任务需要处理"); err != nil {
		logx.Errorf("发送通知短信失败: %v", err)
	}

	// 发送自定义短信
	msg := SMSMessage{
		PhoneNumbers: []string{"13800138000"},
		TemplateCode: "SMS_123456789",
		TemplateParam: map[string]string{
			"name": "张三",
			"task": "完成项目文档",
		},
		SignName: "企业任务系统",
	}

	if err := smsMiddleware.SendSMS(ctx, msg); err != nil {
		logx.Errorf("发送自定义短信失败: %v", err)
	}
}

// ExampleJWTTokenGeneration JWT令牌生成示例
func ExampleJWTTokenGeneration(jwtMiddleware *JWTMiddleware) {
	// 生成令牌
	token, err := jwtMiddleware.GenerateToken("user123", "zhangsan", "张三", "admin")
	if err != nil {
		logx.Errorf("生成JWT令牌失败: %v", err)
		return
	}

	logx.Infof("生成的JWT令牌: %s", token)

	// 解析令牌
	claims, err := jwtMiddleware.ParseToken(token)
	if err != nil {
		logx.Errorf("解析JWT令牌失败: %v", err)
		return
	}

	logx.Infof("解析的JWT声明: %+v", claims)

	// 刷新令牌
	newToken, err := jwtMiddleware.RefreshToken(token)
	if err != nil {
		logx.Errorf("刷新JWT令牌失败: %v", err)
		return
	}

	logx.Infof("刷新后的JWT令牌: %s", newToken)
}

// ExampleMiddlewareIntegration 中间件集成示例
func ExampleMiddlewareIntegration(jwtMiddleware *JWTMiddleware, emailMiddleware *EmailMiddleware, smsMiddleware *SMSMiddleware) {
	ctx := context.Background()

	// 使用JWT中间件
	token, err := jwtMiddleware.GenerateToken("user123", "zhangsan", "张三", "admin")
	if err != nil {
		logx.Errorf("生成JWT令牌失败: %v", err)
		return
	}

	// 使用邮件中间件
	emailMsg := EmailMessage{
		To:      []string{"user@example.com"},
		Subject: "系统通知",
		Body:    "您的JWT令牌已生成: " + token,
		IsHTML:  false,
	}

	if err := emailMiddleware.SendEmail(ctx, emailMsg); err != nil {
		logx.Errorf("发送邮件失败: %v", err)
	}

	// 使用短信中间件
	if err := smsMiddleware.SendNotificationSMS(ctx, "13800138000", "您的JWT令牌已生成"); err != nil {
		logx.Errorf("发送短信失败: %v", err)
	}

	logx.Info("中间件集成示例完成")
}

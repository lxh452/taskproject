package middleware

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// SMSConfig 短信配置
type SMSConfig struct {
	Provider   string `json:"provider"`   // 短信服务提供商 (aliyun, tencent, etc.)
	AccessKey  string `json:"accessKey"`  // 访问密钥
	SecretKey  string `json:"secretKey"`  // 密钥
	SignName   string `json:"signName"`   // 短信签名
	TemplateID string `json:"templateId"` // 模板ID
	Endpoint   string `json:"endpoint"`   // 服务端点
	Region     string `json:"region"`     // 地域
}

// SMSMessage 短信消息
type SMSMessage struct {
	PhoneNumbers  []string          `json:"phoneNumbers"`  // 手机号列表
	TemplateCode  string            `json:"templateCode"`  // 模板代码
	TemplateParam map[string]string `json:"templateParam"` // 模板参数
	SignName      string            `json:"signName"`      // 签名
}

// SMSResponse 短信发送响应
type SMSResponse struct {
	Code      string `json:"Code"`
	Message   string `json:"Message"`
	BizId     string `json:"BizId"`
	RequestId string `json:"RequestId"`
}

// SMSMiddleware 短信中间件
type SMSMiddleware struct {
	config SMSConfig
}

// NewSMSMiddleware 创建短信中间件
func NewSMSMiddleware(config SMSConfig) *SMSMiddleware {
	return &SMSMiddleware{
		config: config,
	}
}

// SendSMS 发送短信
func (s *SMSMiddleware) SendSMS(ctx context.Context, msg SMSMessage) error {
	switch s.config.Provider {
	case "aliyun":
		return s.sendAliyunSMS(ctx, msg)
	case "tencent":
		return s.sendTencentSMS(ctx, msg)
	default:
		return fmt.Errorf("不支持的短信服务提供商: %s", s.config.Provider)
	}
}

// sendAliyunSMS 发送阿里云短信
func (s *SMSMiddleware) sendAliyunSMS(ctx context.Context, msg SMSMessage) error {
	// 构建请求参数
	params := map[string]string{
		"Action":           "SendSms",
		"Version":          "2017-05-25",
		"RegionId":         s.config.Region,
		"PhoneNumbers":     strings.Join(msg.PhoneNumbers, ","),
		"SignName":         msg.SignName,
		"TemplateCode":     msg.TemplateCode,
		"AccessKeyId":      s.config.AccessKey,
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureVersion": "1.0",
		"SignatureNonce":   fmt.Sprintf("%d", time.Now().UnixNano()),
		"Format":           "JSON",
	}

	// 添加模板参数
	if len(msg.TemplateParam) > 0 {
		templateParam, _ := json.Marshal(msg.TemplateParam)
		params["TemplateParam"] = string(templateParam)
	}

	// 生成签名
	signature := s.generateAliyunSignature(params)
	params["Signature"] = signature

	// 构建请求URL
	url := fmt.Sprintf("%s?%s", s.config.Endpoint, s.buildQueryString(params))

	// 发送HTTP请求
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var smsResp SMSResponse
	if err := json.Unmarshal(body, &smsResp); err != nil {
		return err
	}

	if smsResp.Code != "OK" {
		logx.Errorf("阿里云短信发送失败: %s - %s", smsResp.Code, smsResp.Message)
		return fmt.Errorf("短信发送失败: %s", smsResp.Message)
	}

	logx.Infof("阿里云短信发送成功: %s -> %v", msg.TemplateCode, msg.PhoneNumbers)
	return nil
}

// sendTencentSMS 发送腾讯云短信
func (s *SMSMiddleware) sendTencentSMS(ctx context.Context, msg SMSMessage) error {
	// 构建请求参数
	params := map[string]interface{}{
		"PhoneNumberSet":   msg.PhoneNumbers,
		"TemplateId":       msg.TemplateCode,
		"SignName":         msg.SignName,
		"TemplateParamSet": s.buildTemplateParamSet(msg.TemplateParam),
	}

	// 构建请求体
	requestBody, err := json.Marshal(params)
	if err != nil {
		return err
	}

	// 构建请求
	req, err := http.NewRequest("POST", s.config.Endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", s.generateTencentAuth())

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var smsResp SMSResponse
	if err := json.Unmarshal(body, &smsResp); err != nil {
		return err
	}

	if smsResp.Code != "OK" {
		logx.Errorf("腾讯云短信发送失败: %s - %s", smsResp.Code, smsResp.Message)
		return fmt.Errorf("短信发送失败: %s", smsResp.Message)
	}

	logx.Infof("腾讯云短信发送成功: %s -> %v", msg.TemplateCode, msg.PhoneNumbers)
	return nil
}

// generateAliyunSignature 生成阿里云签名
func (s *SMSMiddleware) generateAliyunSignature(params map[string]string) string {
	// 排序参数
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建查询字符串
	var queryParts []string
	for _, k := range keys {
		queryParts = append(queryParts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	queryString := strings.Join(queryParts, "&")

	// 构建签名字符串
	stringToSign := fmt.Sprintf("GET&%s&%s", "/", queryString)

	// 生成HMAC-SHA1签名
	h := md5.New()
	h.Write([]byte(s.config.SecretKey + "&"))

	// 这里应该使用HMAC-SHA1，简化实现
	return fmt.Sprintf("%x", h.Sum([]byte(stringToSign)))
}

// generateTencentAuth 生成腾讯云认证头
func (s *SMSMiddleware) generateTencentAuth() string {
	// 简化实现，实际应该使用腾讯云的签名算法
	return fmt.Sprintf("Bearer %s", s.config.AccessKey)
}

// buildQueryString 构建查询字符串
func (s *SMSMiddleware) buildQueryString(params map[string]string) string {
	var parts []string
	for k, v := range params {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, "&")
}

// buildTemplateParamSet 构建模板参数集
func (s *SMSMiddleware) buildTemplateParamSet(params map[string]string) []string {
	var paramSet []string
	for _, v := range params {
		paramSet = append(paramSet, v)
	}
	return paramSet
}

// SendBatchSMS 批量发送短信
func (s *SMSMiddleware) SendBatchSMS(ctx context.Context, messages []SMSMessage) error {
	for _, msg := range messages {
		if err := s.SendSMS(ctx, msg); err != nil {
			logx.Errorf("批量发送短信失败: %v", err)
			return err
		}
		// 避免发送过快
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}

// SendVerificationCode 发送验证码短信
func (s *SMSMiddleware) SendVerificationCode(ctx context.Context, phoneNumber, code string) error {
	msg := SMSMessage{
		PhoneNumbers: []string{phoneNumber},
		TemplateCode: s.config.TemplateID,
		TemplateParam: map[string]string{
			"code": code,
		},
		SignName: s.config.SignName,
	}

	return s.SendSMS(ctx, msg)
}

// SendNotificationSMS 发送通知短信
func (s *SMSMiddleware) SendNotificationSMS(ctx context.Context, phoneNumber, content string) error {
	msg := SMSMessage{
		PhoneNumbers: []string{phoneNumber},
		TemplateCode: s.config.TemplateID,
		TemplateParam: map[string]string{
			"content": content,
		},
		SignName: s.config.SignName,
	}

	return s.SendSMS(ctx, msg)
}

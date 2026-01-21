package svc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	// 邀请码在Redis中的key前缀
	InviteCodeKeyPrefix = "invite:code:"
	// 邀请码默认过期时间（7天）
	InviteCodeExpireSeconds = 7 * 24 * 60 * 60
)

// InviteCodeData 邀请码数据
type InviteCodeData struct {
	Code        string `json:"code"` // 邀请码
	CompanyID   string `json:"companyId"`
	CompanyName string `json:"companyName"`
	CreatedBy   string `json:"createdBy"` // 创建者员工ID
	CreatedAt   int64  `json:"createdAt"` // 创建时间戳
	ExpireAt    int64  `json:"expireAt"`  // 过期时间戳
	MaxUses     int    `json:"maxUses"`   // 最大使用次数，0表示不限制
	UsedCount   int    `json:"usedCount"` // 已使用次数
}

// InviteCodeService 邀请码服务
type InviteCodeService struct {
	redisClient *redis.Redis
}

// NewInviteCodeService 创建邀请码服务
func NewInviteCodeService(redisClient *redis.Redis) *InviteCodeService {
	return &InviteCodeService{
		redisClient: redisClient,
	}
}

// GenerateInviteCode 生成邀请码
// companyID: 公司ID
// companyName: 公司名称
// createdBy: 创建者员工ID
// expireDays: 过期天数（默认7天）
// maxUses: 最大使用次数（0表示不限制）
func (s *InviteCodeService) GenerateInviteCode(ctx context.Context, companyID, companyName, createdBy string, expireDays, maxUses int) (string, error) {
	if expireDays <= 0 {
		expireDays = 7
	}

	// 生成随机邀请码（8位字母数字）
	code, err := s.generateRandomCode(8)
	if err != nil {
		return "", fmt.Errorf("生成邀请码失败: %w", err)
	}

	// 构建邀请码数据
	now := time.Now()
	data := InviteCodeData{
		Code:        code,
		CompanyID:   companyID,
		CompanyName: companyName,
		CreatedBy:   createdBy,
		CreatedAt:   now.Unix(),
		ExpireAt:    now.Add(time.Duration(expireDays) * 24 * time.Hour).Unix(),
		MaxUses:     maxUses,
		UsedCount:   0,
	}

	// 序列化数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("序列化邀请码数据失败: %w", err)
	}

	// 存储到Redis
	key := InviteCodeKeyPrefix + code
	expireSeconds := expireDays * 24 * 60 * 60
	logx.Infof("[InviteCodeService] 存储邀请码到Redis: key=%s, expireSeconds=%d, data=%s", key, expireSeconds, string(jsonData))

	if err := s.redisClient.Setex(key, string(jsonData), expireSeconds); err != nil {
		logx.Errorf("[InviteCodeService] 存储邀请码到Redis失败: key=%s, err=%v", key, err)
		return "", fmt.Errorf("存储邀请码到Redis失败: %w", err)
	}

	logx.Infof("[InviteCodeService] 生成邀请码成功: code=%s, key=%s, companyID=%s, expireDays=%d", code, key, companyID, expireDays)
	return code, nil
}

// ParseInviteCode 解析邀请码
// 返回邀请码数据，如果邀请码无效或已过期则返回错误
func (s *InviteCodeService) ParseInviteCode(ctx context.Context, code string) (*InviteCodeData, error) {
	code = strings.TrimSpace(strings.ToUpper(code))
	if code == "" {
		return nil, errors.New("邀请码不能为空")
	}

	// 从Redis获取邀请码数据
	key := InviteCodeKeyPrefix + code
	logx.Infof("[InviteCodeService] 解析邀请码: code=%s, key=%s", code, key)

	// 先检查key是否存在
	exists, existsErr := s.redisClient.Exists(key)
	logx.Infof("[InviteCodeService] 检查key是否存在: key=%s, exists=%v, err=%v", key, exists, existsErr)

	jsonData, err := s.redisClient.Get(key)
	if err != nil {
		logx.Errorf("[InviteCodeService] 从Redis获取邀请码失败: code=%s, key=%s, err=%v", code, key, err)
		return nil, errors.New("邀请码无效或已过期")
	}

	logx.Infof("[InviteCodeService] Redis返回数据: code=%s, dataLen=%d, data=%s", code, len(jsonData), jsonData)

	if jsonData == "" {
		logx.Errorf("[InviteCodeService] 邀请码数据为空: code=%s, key=%s", code, key)
		return nil, errors.New("邀请码无效或已过期")
	}

	// 反序列化数据
	var data InviteCodeData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		logx.Errorf("反序列化邀请码数据失败: code=%s, err=%v", code, err)
		return nil, errors.New("邀请码数据损坏")
	}

	// 检查是否过期
	if time.Now().Unix() > data.ExpireAt {
		return nil, errors.New("邀请码已过期")
	}

	// 检查使用次数
	if data.MaxUses > 0 && data.UsedCount >= data.MaxUses {
		return nil, errors.New("邀请码已达到最大使用次数")
	}

	return &data, nil
}

// UseInviteCode 使用邀请码（增加使用计数）
func (s *InviteCodeService) UseInviteCode(ctx context.Context, code string) error {
	code = strings.TrimSpace(strings.ToUpper(code))
	key := InviteCodeKeyPrefix + code

	// 获取当前数据
	jsonData, err := s.redisClient.Get(key)
	if err != nil || jsonData == "" {
		return errors.New("邀请码无效或已过期")
	}

	var data InviteCodeData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return errors.New("邀请码数据损坏")
	}

	// 增加使用计数
	data.UsedCount++

	// 更新Redis（保持原有过期时间）
	ttl, err := s.redisClient.Ttl(key)
	if err != nil || ttl <= 0 {
		ttl = InviteCodeExpireSeconds
	}

	updatedJson, _ := json.Marshal(data)
	if err := s.redisClient.Setex(key, string(updatedJson), int(ttl)); err != nil {
		logx.Errorf("更新邀请码使用次数失败: code=%s, err=%v", code, err)
		// 不影响主流程
	}

	return nil
}

// RevokeInviteCode 撤销邀请码
func (s *InviteCodeService) RevokeInviteCode(ctx context.Context, code string) error {
	code = strings.TrimSpace(strings.ToUpper(code))
	key := InviteCodeKeyPrefix + code

	_, err := s.redisClient.Del(key)
	if err != nil {
		return fmt.Errorf("撤销邀请码失败: %w", err)
	}

	logx.Infof("撤销邀请码成功: code=%s", code)
	return nil
}

// ListInviteCodesByCompany 获取公司的所有邀请码列表
func (s *InviteCodeService) ListInviteCodesByCompany(ctx context.Context, companyID string) ([]InviteCodeData, error) {
	// 扫描所有邀请码key
	pattern := InviteCodeKeyPrefix + "*"
	keys, err := s.redisClient.Keys(pattern)
	if err != nil {
		logx.Errorf("扫描邀请码失败: pattern=%s, err=%v", pattern, err)
		return nil, fmt.Errorf("扫描邀请码失败: %w", err)
	}

	logx.Infof("[ListInviteCodesByCompany] 扫描到的key数量: %d, pattern=%s, companyID=%s", len(keys), pattern, companyID)

	var result []InviteCodeData

	for _, key := range keys {
		// 获取邀请码数据
		jsonData, err := s.redisClient.Get(key)
		if err != nil || jsonData == "" {
			logx.Errorf("[ListInviteCodesByCompany] 获取key失败或数据为空: key=%s, err=%v", key, err)
			continue
		}

		var data InviteCodeData
		if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
			logx.Errorf("解析邀请码数据失败: key=%s, err=%v", key, err)
			continue
		}

		logx.Infof("[ListInviteCodesByCompany] 解析邀请码: key=%s, companyID=%s, dataCompanyID=%s", key, companyID, data.CompanyID)

		// 只返回当前公司的邀请码
		if data.CompanyID == companyID {
			// 从key中提取邀请码
			code := strings.TrimPrefix(key, InviteCodeKeyPrefix)
			data.Code = code

			// 返回所有邀请码（包括已过期的），让前端处理过滤
			result = append(result, data)
			logx.Infof("[ListInviteCodesByCompany] 添加邀请码到结果: code=%s", code)
		}
	}

	logx.Infof("获取公司邀请码列表: companyID=%s, count=%d", companyID, len(result))
	return result, nil
}

// generateRandomCode 生成随机邀请码
func (s *InviteCodeService) generateRandomCode(length int) (string, error) {
	// 使用字母和数字（排除容易混淆的字符：0,O,I,1,L）
	const charset = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// 如果加密随机失败，使用base64编码
		randomBytes := make([]byte, length)
		rand.Read(randomBytes)
		encoded := base64.StdEncoding.EncodeToString(randomBytes)
		// 只取字母数字
		result := ""
		for _, c := range encoded {
			if (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
				result += string(c)
				if len(result) >= length {
					break
				}
			}
		}
		return result, nil
	}

	result := make([]byte, length)
	for i, b := range bytes {
		result[i] = charset[int(b)%len(charset)]
	}
	return string(result), nil
}

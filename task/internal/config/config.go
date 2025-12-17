// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"os"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf

	// 系统配置
	System struct {
		// BaseURL 用于邮件中的前端访问地址，如：https://task.yourcompany.com
		// 如果为空，则邮件中的链接会使用相对路径
		BaseURL string `json:"baseURL"`
	} `json:"system"`

	// 文件存储配置
	FileStorage struct {
		// StorageRoot 本地存储根目录（当使用本地存储时）
		StorageRoot string `json:"storageRoot"`
		// URLPrefix 访问URL前缀
		URLPrefix string `json:"urlPrefix"`
		// StorageType 存储类型: local(本地) 或 cos(腾讯云COS)
		StorageType string `json:"storageType"`
		// COS配置（当StorageType为cos时使用）
		COS struct {
			SecretId  string `json:"secretId"`
			SecretKey string `json:"secretKey"`
			Bucket    string `json:"bucket"`
			Region    string `json:"region"`
		} `json:"cos"`
	} `json:"fileStorage"`

	// 数据库配置
	MySQL struct {
		DataSource string `json:"dataSource"`
	} `json:"mysql"`

	// Redis配置
	Redis struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	} `json:"redis"`

	// JWT配置
	JWT struct {
		SecretKey   string        `json:"secretKey"`
		ExpireTime  time.Duration `json:"expireTime"`
		RefreshTime time.Duration `json:"refreshTime"`
		Issuer      string        `json:"issuer"`
		Audience    string        `json:"audience"`
	} `json:"jwt"`

	// 邮件配置
	Email struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		From     string `json:"from"`
		UseTLS   bool   `json:"useTLS"`
	} `json:"email"`

	// 短信配置
	SMS struct {
		Provider   string `json:"provider"`
		AccessKey  string `json:"accessKey"`
		SecretKey  string `json:"secretKey"`
		SignName   string `json:"signName"`
		TemplateID string `json:"templateId"`
		Endpoint   string `json:"endpoint"`
		Region     string `json:"region"`
	} `json:"sms"`

	// RabbitMQ配置
	RabbitMQ struct {
		URL      string `json:"url"`
		Exchange string `json:"exchange"`
		Queue    string `json:"queue"`
	} `json:"rabbitmq"`

	// MongoDB配置
	Mongo struct {
		Host            string        `json:"host"`
		Port            int           `json:"port"`
		Database        string        `json:"database"`
		Username        string        `json:"username"`
		Password        string        `json:"password"`
		AuthSource      string        `json:"authSource"`
		Timeout         time.Duration `json:"timeout"`
		MaxIdleConns    int           `json:"maxIdleConns"`
		MaxOpenConns    int           `json:"maxOpenConns"`
		MaxConnLifetime time.Duration `json:"maxConnLifetime"`
		MaxConnIdleTime time.Duration `json:"maxConnIdleTime"`
	} `json:"mongo"`
}

// ApplyEnvOverrides 从环境变量覆盖配置
// 环境变量命名规则: MYSQL_DATASOURCE, REDIS_HOST, MONGO_HOST 等
// 支持 Railway / Render / Fly.io 等云平台的环境变量注入
func (c *Config) ApplyEnvOverrides() {
	logx.Info("[Config] 检查环境变量覆盖...")

	overrideCount := 0

	// MySQL
	if v := os.Getenv("MYSQL_DATASOURCE"); v != "" {
		c.MySQL.DataSource = v
		overrideCount++
		logx.Info("[Config] MYSQL_DATASOURCE 已从环境变量覆盖")
	}
	// 支持 Railway 的 DATABASE_URL 格式
	if v := os.Getenv("DATABASE_URL"); v != "" && c.MySQL.DataSource == "" {
		c.MySQL.DataSource = v
		overrideCount++
		logx.Info("[Config] DATABASE_URL 已从环境变量覆盖 MySQL")
	}

	// Redis
	if v := os.Getenv("REDIS_HOST"); v != "" {
		c.Redis.Host = v
		overrideCount++
		logx.Info("[Config] REDIS_HOST 已从环境变量覆盖")
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.Redis.Port = port
			overrideCount++
		}
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		c.Redis.Password = v
		overrideCount++
		logx.Info("[Config] REDIS_PASSWORD 已从环境变量覆盖")
	}
	if v := os.Getenv("REDIS_DB"); v != "" {
		if db, err := strconv.Atoi(v); err == nil {
			c.Redis.DB = db
			overrideCount++
		}
	}
	// 支持 Railway 的 REDIS_URL 格式（需要解析）
	if v := os.Getenv("REDIS_URL"); v != "" {
		// Railway 格式: redis://default:password@host:port
		// 简单处理：如果有 REDIS_URL 但没有单独的 Host，则记录日志提示
		logx.Infof("[Config] 检测到 REDIS_URL: %s (如需使用请手动解析或设置 REDIS_HOST/PORT/PASSWORD)", maskPassword(v))
	}

	// MongoDB
	if v := os.Getenv("MONGO_HOST"); v != "" {
		c.Mongo.Host = v
		overrideCount++
		logx.Info("[Config] MONGO_HOST 已从环境变量覆盖")
	}
	if v := os.Getenv("MONGO_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.Mongo.Port = port
			overrideCount++
		}
	}
	if v := os.Getenv("MONGO_DATABASE"); v != "" {
		c.Mongo.Database = v
		overrideCount++
	}
	if v := os.Getenv("MONGO_USERNAME"); v != "" {
		c.Mongo.Username = v
		overrideCount++
	}
	if v := os.Getenv("MONGO_PASSWORD"); v != "" {
		c.Mongo.Password = v
		overrideCount++
		logx.Info("[Config] MONGO_PASSWORD 已从环境变量覆盖")
	}
	if v := os.Getenv("MONGO_AUTHSOURCE"); v != "" {
		c.Mongo.AuthSource = v
		overrideCount++
	}
	// 支持 Railway 的 MONGO_URL 格式
	if v := os.Getenv("MONGO_URL"); v != "" {
		logx.Infof("[Config] 检测到 MONGO_URL (如需使用请手动解析)")
	}

	// RabbitMQ
	if v := os.Getenv("RABBITMQ_URL"); v != "" {
		c.RabbitMQ.URL = v
		overrideCount++
		logx.Info("[Config] RABBITMQ_URL 已从环境变量覆盖")
	}
	if v := os.Getenv("RABBITMQ_EXCHANGE"); v != "" {
		c.RabbitMQ.Exchange = v
		overrideCount++
	}
	if v := os.Getenv("RABBITMQ_QUEUE"); v != "" {
		c.RabbitMQ.Queue = v
		overrideCount++
	}

	// JWT
	if v := os.Getenv("JWT_SECRET_KEY"); v != "" {
		c.JWT.SecretKey = v
		overrideCount++
		logx.Info("[Config] JWT_SECRET_KEY 已从环境变量覆盖")
	}

	// Email
	if v := os.Getenv("EMAIL_HOST"); v != "" {
		c.Email.Host = v
		overrideCount++
	}
	if v := os.Getenv("EMAIL_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.Email.Port = port
			overrideCount++
		}
	}
	if v := os.Getenv("EMAIL_USERNAME"); v != "" {
		c.Email.Username = v
		overrideCount++
	}
	if v := os.Getenv("EMAIL_PASSWORD"); v != "" {
		c.Email.Password = v
		overrideCount++
	}

	// System
	if v := os.Getenv("SYSTEM_BASE_URL"); v != "" {
		c.System.BaseURL = v
		overrideCount++
	}

	// FileStorage - COS配置
	if v := os.Getenv("TENCENT_CLOUD_SECRET_ID"); v != "" {
		c.FileStorage.COS.SecretId = v
		overrideCount++
		logx.Info("[Config] TENCENT_CLOUD_SECRET_ID 已从环境变量覆盖")
	}
	if v := os.Getenv("TENCENT_CLOUD_SECRET_KEY"); v != "" {
		c.FileStorage.COS.SecretKey = v
		overrideCount++
		logx.Info("[Config] TENCENT_CLOUD_SECRET_KEY 已从环境变量覆盖")
	}
	if v := os.Getenv("TENCENT_CLOUD_COS_BUCKET"); v != "" {
		c.FileStorage.COS.Bucket = v
		overrideCount++
		logx.Info("[Config] TENCENT_CLOUD_COS_BUCKET 已从环境变量覆盖")
	}
	if v := os.Getenv("TENCENT_CLOUD_COS_REGION"); v != "" {
		c.FileStorage.COS.Region = v
		overrideCount++
		logx.Info("[Config] TENCENT_CLOUD_COS_REGION 已从环境变量覆盖")
	}
	if v := os.Getenv("TENCENT_CLOUD_COS_URL_PREFIX"); v != "" {
		c.FileStorage.URLPrefix = v
		overrideCount++
		logx.Info("[Config] TENCENT_CLOUD_COS_URL_PREFIX 已从环境变量覆盖")
	}
	if v := os.Getenv("FILE_STORAGE_TYPE"); v != "" {
		c.FileStorage.StorageType = v
		overrideCount++
		logx.Infof("[Config] FILE_STORAGE_TYPE 已从环境变量覆盖: %s", v)
	}

	// Server (go-zero RestConf)
	if v := os.Getenv("PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.Port = port
			overrideCount++
			logx.Infof("[Config] PORT 已从环境变量覆盖: %d", port)
		}
	}
	if v := os.Getenv("HOST"); v != "" {
		c.Host = v
		overrideCount++
	}

	if overrideCount > 0 {
		logx.Infof("[Config] 共 %d 个配置项已从环境变量覆盖", overrideCount)
	} else {
		logx.Info("[Config] 未检测到环境变量覆盖，使用 YAML 配置")
	}
}

// maskPassword 隐藏密码用于日志
func maskPassword(url string) string {
	// 简单处理：如果URL中有@符号，隐藏密码部分
	if len(url) > 20 {
		return url[:10] + "****" + url[len(url)-10:]
	}
	return "****"
}

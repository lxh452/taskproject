// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"time"

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
		// StorageRoot 本地存储根目录
		StorageRoot string `json:"storageRoot"`
		// URLPrefix 访问URL前缀
		URLPrefix string `json:"urlPrefix"`
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

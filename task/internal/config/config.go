// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"time"

	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	
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
}

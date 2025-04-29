package api

import (
	"github.com/gogf/gf/v2/frame/g"
)

type Alert struct {
	Status       string `v:"required" dc:"告警状态"`
	Labels       Labels
	Annotations  Annotations
	StartsAt     string
	EndsAt       string
	GeneratorURL string
}

type Annotations struct {
	Description string
	Summary     string
}

type Labels struct {
	Alertname string `v:"required" dc:"prometheus告警名称"`
	Env       string
	Namespace string
	Severity  string
}

// prometheus的告警通过飞书发送
type PrometheusFSReq struct {
	g.Meta      `path:"/api/prometheus/fs" method:"post" tags:"prometheus告警推送" summary:"飞书推送告警"`
	ChatId      string `dc:"飞书的chat_id" in:"query"`
	Alerts      []Alert
	ExternalURL string
	Hook        string `dc:"飞书的消息机器人地址" json:"hook"`
}
type PrometheusFSRes struct{}

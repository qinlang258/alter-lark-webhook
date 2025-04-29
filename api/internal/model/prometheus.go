package model

import "github.com/gogf/gf/v2/os/gtime"

type PrometheusReportListOutput struct {
	Id         int
	Alertname  string `json:"alertname"`
	K8sCluster string `json:"k8s_cluster"`
	Level      string
	StartTime  gtime.Time        `json:"start_time"`
	Labels     map[string]string `json:"Labels"`
}

type PrometheusListAlertnameKeyOutput struct {
	Id        int    `json:"id"`
	Alertname string `json:"alertname"`
}

package api

import "github.com/gogf/gf/v2/frame/g"

type SendOomToFeishuReq struct {
	g.Meta   `path:"/api/feishu/oom" method:"post" tags:"prometheus告警推送" summary:"飞书推送告警"`
	ImageUrl string
}

type SendOomToFeishuRes struct {
	Status bool
}

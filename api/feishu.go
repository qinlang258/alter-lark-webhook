package api

import "github.com/gogf/gf/v2/frame/g"

type OomToLarkReq struct {
	g.Meta   `path:"/api/feishu/oom" method:"post" tags:"oom上传推送" summary:"lark推送告警"`
	S3Url    string
	ImageUrl string
}

type OomToLarkRes struct {
	Status bool
}

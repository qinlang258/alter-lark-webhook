package controller

import (
	"alter-lark-webhook/api"
	"alter-lark-webhook/internal/service"
	"context"
)

var Feishu = cFeishu{}

type cFeishu struct{}

func (c *cFeishu) OomToLark(ctx context.Context, req *api.OomToLarkReq) (res *api.OomToLarkRes, err error) {
	err = service.Feishu().OomToLark(ctx, req.ImageUrl, req.S3Url)
	return
}

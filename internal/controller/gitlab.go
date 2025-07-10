package controller

import (
	"alter-lark-webhook/api"
	"alter-lark-webhook/internal/service"
	"context"

	"github.com/gogf/gf/v2/os/glog"
)

var Gitlab = cGitlab{}

type cGitlab struct{}

func (c *cGitlab) SendOomToFeishu(ctx context.Context, req *api.SendOomToFeishuReq) (res *api.SendOomToFeishuRes, err error) {
	status, err := service.Gitlab().SendOomToFeishu(ctx, req.ImageUrl)
	if err != nil {
		glog.Error(ctx, err.Error())
		return &api.SendOomToFeishuRes{Status: status}, err
	}

	return &api.SendOomToFeishuRes{Status: status}, nil
}

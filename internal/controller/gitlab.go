package controller

import (
	"alter-lark-webhook/api"
	"alter-lark-webhook/internal/service"
	"context"
	"fmt"

	"github.com/gogf/gf/v2/os/glog"
)

var Gitlab = cGitlab{}

type cGitlab struct{}

func (c *cGitlab) SendOomToFeishu(ctx context.Context, req *api.SendOomToFeishuReq) (res *api.SendOomToFeishuRes, err error) {

	//return &api.SendOomToFeishuRes{Status: status}, nil
	userInfo, payload, err := service.Gitlab().GetByImageUrlSendOomToFeishu(ctx, req.ImageUrl)
	if err != nil {
		glog.Error(ctx, err.Error())
		return &api.SendOomToFeishuRes{Status: false}, err
	}

	fmt.Println("payload: ------------------------- ", payload)

	err = service.Feishu().SendPrometheusOomAlertToFeishu(ctx, payload, "", userInfo["user_id"])
	if err != nil {
		glog.Error(ctx, err)
		return &api.SendOomToFeishuRes{Status: false}, err
	}

	return &api.SendOomToFeishuRes{Status: true}, nil

}

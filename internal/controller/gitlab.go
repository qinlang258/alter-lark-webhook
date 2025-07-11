package controller

import (
	"alter-lark-webhook/api"
	"context"
)

var Gitlab = cGitlab{}

type cGitlab struct{}

func (c *cGitlab) SendOomToFeishu(ctx context.Context, req *api.SendOomToFeishuReq) (res *api.SendOomToFeishuRes, err error) {
	// status, err := service.Gitlab().send
	// if err != nil {
	// 	glog.Error(ctx, err.Error())
	// 	return &api.SendOomToFeishuRes{Status: status}, err
	// }

	//return &api.SendOomToFeishuRes{Status: status}, nil

	return nil, nil
}

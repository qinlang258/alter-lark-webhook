// ================================================================================
// Code generated by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"alter-lark-webhook/internal/model"
	"context"
)

type (
	IFeishu interface {
		Notify(ctx context.Context, in *model.FsMsgInput, status, itemName string) error
		GetUserIdByCommitItem(ctx context.Context, itemName string) (*string, error)
		SendPrometheusOomAlertToFeishu(ctx context.Context, payload map[string]interface{}, status, userId string) error
	}
)

var (
	localFeishu IFeishu
)

func Feishu() IFeishu {
	if localFeishu == nil {
		panic("implement not found for interface IFeishu, forgot register?")
	}
	return localFeishu
}

func RegisterFeishu(i IFeishu) {
	localFeishu = i
}

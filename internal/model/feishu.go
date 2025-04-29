package model

import (
	"github.com/gogf/gf/v2/frame/g"
)

// 通过api获取所有项目, 并批量向数据库中写入
type FsDepartmentInput struct {
	AppId               string `description:"对应config.toml中的feishu.departments.appId"`
	AppSecret           string `description:"对应config.toml中的feishu.departments.appSecret"`
	TopOpenDepartmentId string `description:"对应config.toml中的feishu.departments.topOpenDepartmentId"`
	TopDepartmentId     string `description:"对应config.toml中的feishu.departments.topDepartmentId"`
	TopDepartmentName   string `description:"对应config.toml中的feishu.departments.topDepartmentName"`
}

type FsDepartmentOutput struct {
	OpenDepartmentId   string `description:"用来在具体某个应用中标识一个部门，同一个部门 在不同应用中的 open_department_id 不相同"`
	DepartmentId       string `description:"用来标识租户内一个唯一的部门"`
	Name               string `description:"部门名称"`
	LeaderUserId       string `description:"用户id"`
	ParentDepartmentId string `description:""`
	HasListChild       bool   `description:"是否已获取过子部门"`
}

// 发送飞书消息
type FsMsgInput struct {
	Hook          string
	ReceiveIdType string      `description:"只支持open_id(单个用户)/chat_id(单个群组), 2选1"`
	ReceiveId     string      `description:"只支持open_id(单个用户)/chat_id(单个群组)这2种类型的值,是具体的id值, 不是名称"`
	MsgType       string      `description:"消息类型 包括: text、post、image、file、audio、media、sticker、interactive、share_chat、share_user等"`
	Content       g.MapStrAny `description:"消息内容,不同的消息类型, 内容格式不同"`
}

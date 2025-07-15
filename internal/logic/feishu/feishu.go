package feishu

import (
	"alter-lark-webhook/internal/dao"
	"alter-lark-webhook/internal/logic/tools"
	"alter-lark-webhook/internal/model"
	"alter-lark-webhook/internal/model/entity"
	"alter-lark-webhook/internal/service"

	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"net/http"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
	larkv3 "github.com/larksuite/oapi-sdk-go/v3"
	larkcontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type sFeishu struct {
	client *larkv3.Client
}

func init() {
	fs := New()
	ctx := context.Background()

	// 读取配置
	appId := g.Cfg().MustGet(ctx, "feishu.appId").String()
	appSecret := g.Cfg().MustGet(ctx, "feishu.appSecret").String()

	if appId == "" || appSecret == "" {
		glog.Errorf(ctx, "飞书配置缺失: appId=%s, appSecret=%s", appId, appSecret)
		panic("无法初始化飞书客户端：配置缺失")
	}

	fs.client = larkv3.NewClient(appId, appSecret, larkv3.WithOpenBaseUrl(larkv3.LarkBaseUrl))

	service.RegisterFeishu(fs)

}

func New() *sFeishu {
	return &sFeishu{}
}

func (s *sFeishu) formatTimeUtc8(timeStr string) string {
	// 1. 空值或默认零值检查
	if timeStr == "" || timeStr == "0001-01-01T00:00:00Z" {
		return "N/A" // 或返回空字符串 ""
	}

	// 2. 定义支持的格式（兼容原有布局和ISO 8601）
	layouts := []string{
		"2006-01-02 15:04:05", // 原有格式
		time.RFC3339,          // ISO 8601（如 "2025-07-03T08:02:40.243Z"）
	}

	// 3. 尝试按多种格式解析时间
	var parsedTime time.Time
	var err error
	for _, layout := range layouts {
		parsedTime, err = time.Parse(layout, timeStr)
		if err == nil {
			break // 解析成功则退出循环
		}
	}
	if err != nil {
		return "Invalid Time" // 所有格式均解析失败
	}

	// 4. 加载东八区时区（优先使用标准时区，失败则回退）
	cstLoc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		cstLoc = time.FixedZone("CST", 8*3600)
	}

	// 5. 转换为东八区并格式化输出
	return parsedTime.In(cstLoc).Format("2006-01-02 15:04:05")
}

// Notify 用于向飞书发送通知消息
func (s *sFeishu) Notify(ctx context.Context, in *model.FsMsgInput, status, itemName string) error {
	//将 content 转换为 JSON 字节流
	bytesData, err := json.Marshal(in.Content)
	if err != nil {
		glog.Error(ctx, "Failed to marshal content:", err)
		return err
	}

	// 初始化提取的字段变量
	var alertData map[string]interface{}
	err = json.Unmarshal(bytesData, &alertData)
	if err != nil {
		glog.Error(ctx, err)
		return err
	}

	// 安全地访问嵌套字段 alertData
	var alertname, severity, description, env, startsAt, generatorURL, summary string

	// 提取 template_variable 字段，进行格式检查
	data, ok := alertData["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("data field missing or not in expected format")
	}

	templateVariable, ok := data["template_variable"].(map[string]interface{})
	if !ok {
		glog.Error(ctx, "template_variable 字段缺失或格式不符合预期")
	}

	// 提取具体字段
	alertname = tools.ExtractField(templateVariable, "alertname")
	severity = tools.ExtractField(templateVariable, "severity")
	description = tools.ExtractField(templateVariable, "description")
	env = tools.ExtractField(templateVariable, "env")

	startsAt = tools.ExtractField(templateVariable, "startsAt")
	startsAt = s.formatTimeUtc8(startsAt) // 格式化时间为东八区

	generatorURL = tools.ExtractField(templateVariable, "generatorURL")
	summary = tools.ExtractField(templateVariable, "summary")
	itemName = tools.ExtractField(templateVariable, "itemName")

	dbPayload := make(map[string]interface{})

	if status == "resolved" {
		endsAt := tools.ExtractField(templateVariable, "endsAt")
		endsAt = s.formatTimeUtc8(endsAt) // 格式化时间为东八区

		dbPayload = map[string]interface{}{
			"alertname":   alertname,
			"env":         env,
			"k8s_cluster": "stx",
			"item_name":   itemName, // 添加 itemName 字段
			"level":       severity,
			"start_time":  startsAt,
			"end_time":    endsAt,
			"labels":      tools.ExtractOtherLabels(templateVariable, false), // 提取其他标签并格式化为飞书消息格式
			// 其他标签提取,
			"summary":     summary,
			"status":      status,
			"description": description,
			"generator":   generatorURL,
			"is_resolved": "1",
		}
	} else {
		dbPayload = map[string]interface{}{
			"alertname":   alertname,
			"env":         env,
			"k8s_cluster": "stx",
			"item_name":   itemName, // 添加 itemName 字段
			"level":       severity,
			"start_time":  startsAt,
			"labels":      tools.ExtractOtherLabels(templateVariable, false),
			"summary":     summary,
			"status":      status,
			"description": description,
			"generator":   generatorURL,
			"is_resolved": "0",
		}
	}

	//记录到数据库
	_, err = service.Prometheus().Record(ctx, dbPayload)
	if err != nil {
		glog.Error(ctx, "Prometheus告警记录添加失败: %v", err)
		return err
	}

	var payload map[string]interface{}

	if alertname == "Watchdog" && env == "prod" {
		payload = tools.BuildWatchDogrichTextMessage(alertname, severity, description, env, startsAt, tools.ExtractOtherLabels(templateVariable, true), status, summary)
	}

	payload = tools.BuildRichTextMessage(alertname, severity, description, env, startsAt, generatorURL, tools.ExtractOtherLabels(templateVariable, true), status, summary)

	// 修改调用条件，增加resolved状态判断
	if severity == "critical" || severity == "warning" || severity == "resolved" || severity == "watchdog" {
		s.sendToFeishu(ctx, payload, in.Hook)
	}

	//新增对异常容器的
	if alertname == "KubePodCrashLooping" {
		userId, err := s.GetUserIdByCommitItem(ctx, itemName)
		if err != nil {
			return err
		}
		service.Feishu().SendPrometheusOomAlertToFeishu(ctx, payload, status, *userId)

	}
	return nil
}

// gitlab的相关接口可以根据环境变量传递容器镜像地址。prometheus并没有这个参数
func (s *sFeishu) GetUserIdByCommitItem(ctx context.Context, itemName string) (*string, error) {

	workloadName := tools.ExtractDeploymentOrSTSName(itemName)

	var deployRecord entity.DeployHistory
	dao.DeployHistory.Ctx(ctx).
		Where("service_name = ?", workloadName).
		Where("type like ?", fmt.Sprintf("%%%s%%", "cd")).
		OrderDesc("deploy_time").
		Scan(&deployRecord)

	data, err := service.Gitlab().GetUserInfoByImageUrl(ctx, deployRecord.Image)
	fmt.Println(data)

	// 构建消息体
	req := larkcontact.NewBatchGetIdUserReqBuilder().
		UserIdType(`open_id`).
		Body(larkcontact.NewBatchGetIdUserReqBodyBuilder().
			Emails([]string{data["committerEmail"]}).
			Mobiles([]string{}).
			IncludeResigned(true).
			Build()).
		Build()

	// 发送消息
	resp, err := s.client.Contact.User.BatchGetId(ctx, req)
	if err != nil {
		glog.Error(ctx, "获取user_id失败: %v\n", err)
		return nil, err
	}

	// 服务端错误处理
	if !resp.Success() {
		glog.Error(ctx, "logId: %s, error response: code=%d, msg=%s\n", resp.RequestId(), resp.Code, resp.Msg)
		return nil, err
	}

	if len(resp.Data.UserList) == 1 {
		return resp.Data.UserList[0].UserId, nil
	} else {
		return nil, fmt.Errorf("没有找到或者找到多个用户")
	}

}

// 发送消息到飞书
func (s *sFeishu) sendToFeishu(ctx context.Context, payload map[string]interface{}, hook string) error {
	// 将消息体转换为 JSON 字节流
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		glog.Error(ctx, "消息体转换为JSON失败: %v", err)
		return err
	}

	fmt.Println("payloadBytes: [[[[[[[[[[[[[[[[[[[[[[[[[[]]]]]]]]]]]]]]]]]]]]]]]]]]  ", payloadBytes)

	// 创建 HTTP POST 请求
	hookurl := "https://open.larksuite.com/open-apis/bot/v2/hook/" + hook
	req, err := http.NewRequest("POST", hookurl, bytes.NewBuffer(payloadBytes))
	if err != nil {
		glog.Error(ctx, "创建请求失败: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// 发送 HTTP 请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		glog.Error(ctx, "请求飞书失败: %v", err)
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (s *sFeishu) SendPrometheusOomAlertToFeishu(ctx context.Context, payload map[string]interface{}, status, userId string) error {
	var alertname, severity, description, env, startsAt, summary string

	// 提取 template_variable 字段，进行格式检查
	data, ok := payload["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("data field missing or not in expected format")
	}

	templateVariable, ok := data["template_variable"].(map[string]interface{})
	if !ok {
		glog.Error(ctx, "template_variable 字段缺失或格式不符合预期")
	}

	// 提取具体字段
	alertname = tools.ExtractField(templateVariable, "alertname")
	severity = tools.ExtractField(templateVariable, "severity")
	description = tools.ExtractField(templateVariable, "description")
	env = tools.ExtractField(templateVariable, "env")

	startsAt = tools.ExtractField(templateVariable, "startsAt")
	startsAt = s.formatTimeUtc8(startsAt) // 格式化时间为东八区

	summary = tools.ExtractField(templateVariable, "summary")
	//itemName = tools.ExtractField(templateVariable, "itemName")

	applicationPayload := tools.BuildOOMRichTextMessage(alertname, severity, description, env, startsAt, tools.ExtractOtherLabels(templateVariable, true), status, summary)

	// 序列化内容
	contentJSON, err := json.Marshal(applicationPayload)
	if err != nil {
		glog.Error(ctx, "序列化富文本内容失败: %v\n", err)
		return err
	}

	// 构建消息体
	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType("open_id").
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(userId).
			MsgType("interactive").
			Content(string(contentJSON)).
			Build()).
		Build()

	// 发送消息
	resp, err := s.client.Im.V1.Message.Create(ctx, req)
	if err != nil {
		glog.Error(ctx, "发送消息失败: %v\n", err)
		return err
	}

	fmt.Println("resp--------------------------------------", resp)

	// 服务端错误处理
	if !resp.Success() {
		glog.Error(ctx, "logId: %s, error response: code=%d, msg=%s\n", resp.RequestId(), resp.Code, resp.Msg)
		glog.Error(ctx, resp.Msg)
		return err
	}

	glog.Info(ctx, "消息发送成功: %s\n", resp.Msg)
	return nil

}

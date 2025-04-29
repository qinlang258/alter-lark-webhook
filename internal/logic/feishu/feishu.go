package feishu

import (
	"alter-lark-webhook/internal/model"
	"alter-lark-webhook/internal/service"
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"net/http"

	"github.com/gogf/gf/v2/os/glog"
)

type sFeishu struct {
}

func init() {
	service.RegisterFeishu(New())
}

func New() *sFeishu {
	return &sFeishu{}
}

// Notify 用于向飞书发送通知消息
func (s *sFeishu) Notify(ctx context.Context, in *model.FsMsgInput) error {
	// 将 content 转换为 JSON 字节流
	bytesData, err := json.Marshal(in.Content)
	if err != nil {
		return err
	}

	// 初始化提取的字段变量
	var alertData map[string]interface{}
	err = json.Unmarshal(bytesData, &alertData)
	if err != nil {
		return err
	}

	// 安全地访问嵌套字段 alertData
	var alertname, severity, description, env, startsAt, generatorURL string
	var otherlabels map[string]interface{}
	var otherlabelsStr string

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
	alertname = extractField(templateVariable, "alertname")
	severity = extractField(templateVariable, "severity")
	description = extractField(templateVariable, "description")
	env = extractField(templateVariable, "env")
	startsAt = extractField(templateVariable, "startsAt")
	generatorURL = extractField(templateVariable, "generatorURL")

	// 提取其它标签
	otherlabelsStr = extractOtherLabels(templateVariable)

	// 根据 severity 来构建消息
	textMessage := buildTextMessage(alertname, severity, description, env, startsAt, generatorURL, otherlabelsStr)

	// 构建消息体，发送给飞书机器人
	payload := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]interface{}{
			"text": textMessage,
		},
	}

	// 如果严重性为 "critical"，发送普通消息
	if severity == "critical" || severity == "warning" {
		return s.sendToFeishu(ctx, payload, alertname, severity, env, startsAt, otherlabels, in.Hook)
	}
	return nil
}

// 提取字段
func extractField(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

// 提取其他标签并格式化
func extractOtherLabels(templateVariable map[string]interface{}) string {
	var otherlabels map[string]interface{}
	var otherlabelsStr string

	if labels, ok := templateVariable["otherlabels"].(map[string]interface{}); ok {
		otherlabels = labels
		for k, v := range otherlabels {
			otherlabelsStr += fmt.Sprintf("%s: %v\n", k, v)
		}
	} else {
		otherlabelsStr = "{}"
	}

	return otherlabelsStr
}

// 构建文本消息
func buildTextMessage(alertname, severity, description, env, startsAt, generatorURL, otherlabelsStr string) string {
	return fmt.Sprintf(
		"告警名称: %s\n"+
			"严重程度: %s\n"+
			"描述: \n\t%s\n"+
			"环境: %s\n"+
			"开始时间: %s\n"+
			"告警链接: %s\n"+
			"其它标签: \n%s",
		alertname,
		severity,
		description,
		env,
		startsAt,
		generatorURL,
		otherlabelsStr,
	)
}

// 发送消息到飞书
func (s *sFeishu) sendToFeishu(ctx context.Context, payload map[string]interface{}, alertname, severity, env, startsAt string, otherlabels map[string]interface{}, hook string) error {
	// 将消息体转换为 JSON 字节流
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// 创建 HTTP POST 请求
	hookurl := "https://open.larksuite.com/open-apis/bot/v2/hook/" + hook
	req, err := http.NewRequest("POST", hookurl, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// 发送 HTTP 请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

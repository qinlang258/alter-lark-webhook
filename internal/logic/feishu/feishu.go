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
	//textMessage := buildRichTextMessage(alertname, severity, description, env, startsAt, generatorURL, otherlabelsStr)

	payload := buildRichTextMessage(alertname, severity, description, env, startsAt, generatorURL, otherlabelsStr)

	// 修改调用条件，增加resolved状态判断
	if severity == "critical" || severity == "warning" || severity == "resolved" {
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

func buildRichTextMessage(alertname, severity, description, env, startsAt, generatorURL, otherlabelsStr string) map[string]interface{} {
	color := "green" // 默认设为绿色
	status := "告警通知"

	// 判断是否为恢复状态
	if severity == "resolved" {
		status = "告警恢复"
	} else {
		// 非恢复状态才按严重程度设置颜色
		if severity == "critical" {
			color = "red"
		} else if severity == "warning" {
			color = "orange"
		}
	}

	return map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"tag":     "plain_text",
					"content": fmt.Sprintf("【%s】%s", severity, status),
				},
				"template": color,
			},
			"elements": []map[string]interface{}{
				{
					"tag": "div",
					"fields": []map[string]interface{}{
						{
							"is_short": true,
							"text": map[string]interface{}{
								"tag":     "lark_md",
								"content": fmt.Sprintf("​**告警名称**:\n%s", alertname),
							},
						},
						{
							"is_short": true,
							"text": map[string]interface{}{
								"tag":     "lark_md",
								"content": fmt.Sprintf("​**严重程度**:\n<font color=\"%s\">%s</font>", color, severity),
							},
						},
					},
				},
				{
					"tag":     "markdown",
					"content": fmt.Sprintf("​**描述**:\n%s", description),
				},
				{
					"tag": "hr",
				},
				{
					"tag": "note",
					"elements": []map[string]interface{}{
						{
							"tag":     "plain_text",
							"content": fmt.Sprintf("环境: %s | 开始时间: %s", env, startsAt),
						},
					},
				},
				{
					"tag": "action",
					"actions": []map[string]interface{}{
						{
							"tag": "button",
							"text": map[string]interface{}{
								"tag":     "plain_text",
								"content": "查看详情",
							},
							"url":  generatorURL,
							"type": "primary",
						},
					},
				},
			},
		},
	}
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

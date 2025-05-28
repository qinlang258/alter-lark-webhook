package feishu

import (
	"alter-lark-webhook/internal/model"
	"alter-lark-webhook/internal/service"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

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

func (s *sFeishu) formatTimeUtc8(timeStr string) string {
	layout := "2006-01-02 15:04:05"

	// 解析为UTC时间
	utcTime, err := time.Parse(layout, timeStr)
	if err != nil {
		log.Fatal(err)
	}

	// 加载东八区时区
	cstLoc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		cstLoc = time.FixedZone("CST", 8*3600)
	}

	// 转换为东八区时间
	cstTime := utcTime.In(cstLoc)

	// 返回格式化后的时间字符串
	return cstTime.Format(layout)
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
	var alertname, severity, description, env, startsAt, generatorURL, status, summary string
	var otherlabels map[string]interface{}
	var otherlabelsStr string

	fmt.Println("alertData:           ", alertData)

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
	startsAt = s.formatTimeUtc8(startsAt) // 格式化时间为东八区

	generatorURL = extractField(templateVariable, "generatorURL")
	status = extractField(templateVariable, "status")
	summary = extractField(templateVariable, "summary")

	// 提取其它标签
	otherlabelsStr = extractOtherLabels(templateVariable)

	// 根据 severity 来构建消息
	//textMessage := buildRichTextMessage(alertname, severity, description, env, startsAt, generatorURL, otherlabelsStr)

	payload := buildRichTextMessage(alertname, severity, description, env, startsAt, generatorURL, otherlabelsStr, status, summary)

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
	var builder strings.Builder

	if labels, ok := templateVariable["otherlabels"].(map[string]interface{}); ok {
		builder.WriteString("{")
		first := true
		for k, v := range labels {
			if !first {
				builder.WriteString("\n")
			}
			builder.WriteString(fmt.Sprintf("%s: %v", k, v))
			first = false
		}
		builder.WriteString("}")
	} else {
		builder.WriteString("{}")
	}

	return builder.String()
}

func buildRichTextMessage(alertname, severity, description, env, startsAt, generatorURL, otherlabelsStr, status, summary string) map[string]interface{} {
	// 初始化变量
	var color, titlePrefix string
	isResolved := status == "resolved"

	// 设置状态和颜色
	if isResolved {
		status = "告警恢复"
		color = "green"
		titlePrefix = "✅"
	} else {
		status = "告警通知"
		titlePrefix = "⚠️"
		switch severity {
		case "critical":
			color = "red"
		case "warning":
			color = "orange"
		default:
			color = "blue"
		}
	}

	// 构建消息卡片
	return map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"tag":     "plain_text",
					"content": fmt.Sprintf("%s【%s】%s", titlePrefix, strings.ToUpper(severity), status),
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
								"content": fmt.Sprintf("​**状态**:\n<font color=\"%s\">%s</font>", color, status),
							},
						},
					},
				},
				{
					"tag": "div",
					"fields": []map[string]interface{}{
						{
							"is_short": true,
							"text": map[string]interface{}{
								"tag":     "lark_md",
								"content": fmt.Sprintf("​**环境**:\n%s", env),
							},
						},
						{
							"is_short": true,
							"text": map[string]interface{}{
								"tag":     "lark_md",
								"content": fmt.Sprintf("​**时间**:\n%s", startsAt),
							},
						},
					},
				},
				{
					"tag":     "markdown",
					"content": fmt.Sprintf("​**描述**:\n%s", description),
				},
				{
					"tag":     "markdown",
					"content": fmt.Sprintf("​**summary**:\n%s", summary),
				},
				{
					"tag":     "markdown",
					"content": fmt.Sprintf("​**其他标签**:\n```\n%s\n```", otherlabelsStr),
				},
				{
					"tag": "hr",
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

// 标签解析示例（需根据实际数据结构实现）
func parseLabels(labelsStr string) map[string]string {
	// 实现具体的标签解析逻辑
	return map[string]string{
		"severity":   "warning",
		"alertname":  "ai-high-cpu-used",
		"container":  "ubuntu-container",
		"env":        "prod",
		"namespace":  "monitoring",
		"pod":        "ubuntu-deployment-649b48f48c-w9h98",
		"prometheus": "monitoring/k8s",
	}
}

// 构建标签展示组件
func buildLabelComponents(labelsStr string) []map[string]interface{} {
	// 示例标签解析逻辑（需根据实际数据结构实现）
	labels := parseLabels(labelsStr)

	// 标签分类配置
	labelGroups := map[string][]string{
		"🖥️ 系统资源": {"pod", "namespace", "container"},
		"🚨 告警信息":  {"severity", "alertname"},
		"🌍 环境配置":  {"env", "cluster"},
	}

	var components []map[string]interface{}

	for groupName, keys := range labelGroups {
		var fields []map[string]interface{}
		for _, k := range keys {
			if v, ok := labels[k]; ok {
				fields = append(fields, map[string]interface{}{
					"tag": "div",
					"text": map[string]interface{}{
						"tag": "lark_md",
						"content": fmt.Sprintf("`%s:` <font color='%s'>%s</font>",
							k,
							getLabelColor(k),
							v),
					},
				})
			}
		}
		if len(fields) > 0 {
			components = append(components, map[string]interface{}{
				"tag":              "column_set",
				"flex_mode":        "flow",
				"background_style": "grey",
				"columns": []map[string]interface{}{
					{
						"tag":    "column",
						"width":  "weighted",
						"weight": 30,
						"elements": []map[string]interface{}{
							{
								"tag":     "markdown",
								"content": fmt.Sprintf("​**​%s**​", groupName),
							},
						},
					},
					{
						"tag":      "column",
						"width":    "weighted",
						"weight":   70,
						"elements": fields,
					},
				},
			})
		}
	}
	return components
}

// 获取标签颜色（示例实现）
func getLabelColor(key string) string {
	colorMap := map[string]string{
		"critical":  "#FF4D4D",
		"warning":   "#FF9A2E",
		"pod":       "#3370FF",
		"namespace": "#3370FF",
		"container": "#3370FF",
		"env":       "#00B567",
		"alertname": "#FF9A2E",
	}
	if color, ok := colorMap[key]; ok {
		return color
	}
	return "#666"
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
		glog.Error(ctx, "请求飞书失败: %v", err)
		return err
	}
	defer resp.Body.Close()

	return nil
}

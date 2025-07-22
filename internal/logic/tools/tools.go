package tools

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

func GetMapStr(data g.Map, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func GetMapByte(data g.Map, key string) []byte {
	if val, ok := data[key].([]byte); ok {
		return val
	}
	return nil
}

func GetMapInt64(data g.Map, key string) int64 {
	if val, ok := data[key].(int64); ok {
		return val
	}
	return 0
}

func GetMapInt(data g.Map, key string) int {
	if val, ok := data[key].(int); ok {
		return val
	}
	return 0
}

func GetMapInt32(data g.Map, key string) int32 {
	if val, ok := data[key].(int32); ok {
		return val
	}
	return 0
}

func GetMapTime(data g.Map, key string) *gtime.Time {
	if val, ok := data[key].(*gtime.Time); ok {
		return val
	}
	return nil
}

func GetGjsonjson(data g.Map, key string) *gjson.Json {
	if val, ok := data[key].(*gjson.Json); ok {
		return val
	}
	return nil
}

func BuildOOMRichTextMessage(alertname, severity, description, env, startsAt, otherlabelsStr, status, summary string) map[string]interface{} {
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
			},
		},
	}
}

func ParseJSONToMap(jsonStr string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func BuildWatchDogrichTextMessage(alertname, severity, description, env, startsAt, otherlabelsStr, status, summary string) map[string]interface{} {
	// 初始化变量
	var color, titlePrefix string
	//isResolved := status == "resolved"

	// 设置状态和颜色
	status = "告警恢复"
	color = "green"
	titlePrefix = "✅"

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
			},
		},
	}
}

func BuildRichTextMessage(alertname, severity, description, env, startsAt, generatorURL, otherlabelsStr, status, summary string) map[string]interface{} {
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

func ExtractDeploymentOrSTSName(podName string) string {
	// 正则表达式匹配Deployment Pod名称（带hash的部分）
	deploymentRegex := regexp.MustCompile(`^(.*)-[a-z0-9]{8,10}-[a-z0-9]{5}$`)
	// 正则表达式匹配StatefulSet Pod名称（带数字的部分）
	stsRegex := regexp.MustCompile(`^(.*)-\d+$`)

	if deploymentRegex.MatchString(podName) {
		matches := deploymentRegex.FindStringSubmatch(podName)
		if len(matches) >= 2 {
			return matches[1] // 返回Deployment名称
		}
	} else if stsRegex.MatchString(podName) {
		matches := stsRegex.FindStringSubmatch(podName)
		if len(matches) >= 2 {
			return matches[1] // 返回STS名称
		}
	}

	return podName // 如果都不匹配，返回原始名称
}

// 提取字段
func ExtractField(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func RemoveOuterLayer(jsonStr string) (string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", err
	}

	// 删除otherlabels键
	innerData := data["otherlabels"]
	delete(data, "otherlabels")

	// 将内部字段提升到顶层
	for k, v := range innerData.(map[string]interface{}) {
		data[k] = v
	}

	// 重新序列化
	result, _ := json.Marshal(data)
	return string(result), nil
}

func ExtractOtherLabels(templateVariable map[string]interface{}, forFeishu bool) string {
	// 1. 移除 "otherlabels" 外层（如果存在）
	if innerData, exists := templateVariable["otherlabels"].(map[string]interface{}); exists {
		for k, v := range innerData {
			templateVariable[k] = v // 将内部字段提升到顶层
		}
		delete(templateVariable, "otherlabels") // 移除外层键
	}

	// 2. 提取非保留字段
	reservedFields := map[string]bool{
		"alertname": true, "severity": true, "description": true,
		"env": true, "startsAt": true, "generatorURL": true,
		"status": true, "summary": true, "endsAt": true,
	}

	// 预估算容量
	otherLabels := make(map[string]interface{}, len(templateVariable)-len(reservedFields))

	for key, val := range templateVariable {
		if !reservedFields[key] && !isEmptyValue(val) {
			otherLabels[key] = val
		}
	}

	if len(otherLabels) == 0 {
		return "{}"
	}

	// 3. 根据输出格式处理
	if forFeishu {
		var sb strings.Builder
		// 预估算容量
		sb.Grow(len(otherLabels) * 16) // 估算平均每个键值对约16字符

		for k, v := range otherLabels {
			sb.WriteString(fmt.Sprintf("%s: %v\n", k, v))
		}
		return strings.TrimSpace(sb.String())
	}

	jsonData, _ := json.Marshal(otherLabels)
	return string(jsonData)
}

func isEmptyValue(v interface{}) bool {
	if v == nil {
		return true
	}

	switch val := v.(type) {
	case string:
		return len(val) == 0
	case map[string]interface{}:
		return len(val) == 0
	case []interface{}:
		return len(val) == 0
	default:
		return false
	}
}

package controller

import (
	"context"

	"alter-lark-webhook/api"
	"alter-lark-webhook/internal/model"
	"alter-lark-webhook/internal/service"

	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/v2/os/gtime"
)

var Prometheus = cPrometheus{}

type cPrometheus struct{}

// prometheus的告警通过飞书发送, 这里的响应信息是查看不到的, 所以主要以日志的形式记录
// 因为告警的标签等是动态的, 即不同的服务或告警信息不一样, 所以需要动态解析, 不能依赖于req参数
func (c *cPrometheus) PrometheusFS(ctx context.Context, req *api.PrometheusFSReq) (res *api.PrometheusFSRes, err error) {
	alters, err := service.Prometheus().GetRawAlertInfo(ctx)
	if err != nil {
		return nil, err
	}

	if err := gtime.SetTimeZone("Asia/Shanghai"); err != nil {
		glog.Fatal(ctx, "时区设置失败:", err)
	}
	for _, alert := range alters {
		status := alert.Get("status").String()
		env := alert.Get("labels.env").String()
		alertname := alert.Get("labels.alertname").String()
		generatorURL := alert.Get("generatorURL").String()
		severity := alert.Get("labels.severity").String()
		itemName := alert.Get("labels.pod").String()
		var startsAt, endsAt gtime.Time

		if alert.Get("startsAt") != nil {
			// 这里的时间是UTC时间, 需要转换为本地时间
			startsAt = *gtime.New(alert.Get("startsAt").String())
		} else {
			startsAt = *gtime.New()
		}

		if alert.Get("endsAt") != nil {
			// 这里的时间是UTC时间, 需要转换为本地时间
			endsAt = *gtime.New(alert.Get("endsAt").String())
		} else {
			endsAt = *gtime.New()
		}

		summary := alert.Get("annotations.summary").String()
		description := alert.Get("annotations.description").String()

		label := alert.Get("labels").Map()

		// 第1步: 先发送到群, 并at到人
		in := model.FsMsgInput{
			Hook:          req.Hook,
			ReceiveIdType: "chat_id",
			ReceiveId:     req.ChatId,
			MsgType:       "interactive",
			Content: map[string]interface{}{
				"type": "template",
				"data": map[string]interface{}{
					"template_variable": map[string]interface{}{
						"itemName":     itemName,
						"alertname":    alertname,
						"generatorURL": generatorURL,
						"severity":     severity,
						"startsAt":     startsAt,
						"endsAt":       endsAt,
						"summary":      summary,
						"description":  description,
						"otherlabels":  label,
						"env":          env,
						"status":       status,
					},
				},
			},
		}

		if err = service.Feishu().Notify(ctx, &in, status, itemName); err != nil {
			glog.Error("prometheus告警发送到群失败: %s", err.Error())
			return nil, err
		}

	}

	return nil, nil
}

func (c *cPrometheus) Test(ctx context.Context, req *api.PrometheusTestReq) (res *api.PrometheusTestRes, err error) {

	service.Prometheus().Test(ctx, make(map[string]interface{}))

	return nil, nil
}

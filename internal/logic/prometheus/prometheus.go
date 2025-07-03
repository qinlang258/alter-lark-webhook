package prometheus

import (
	"alter-lark-webhook/internal/dao"
	"alter-lark-webhook/internal/logic/tools"
	"alter-lark-webhook/internal/model/entity"
	"alter-lark-webhook/internal/service"

	"github.com/gogf/gf/v2/os/gtime"

	"context"
	"fmt"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
)

type sPrometheus struct{}

func init() {
	service.RegisterPrometheus(New())
}

func New() *sPrometheus {
	return &sPrometheus{}
}

func (s *sPrometheus) GetRawAlertInfo(ctx context.Context) (alerts []*gjson.Json, err error) {
	alerts = make([]*gjson.Json, 0)

	//jsonInfo, err := g.RequestFromCtx(ctx).GetJson()

	if err != nil {
		g.Log().Errorf(ctx, "prometheus告警信息解析失败: %s", err.Error())
		return nil, err
	}

	bodyStr := g.RequestFromCtx(ctx).GetBodyString()
	fmt.Println("bodyStr:        ", bodyStr)

	bodyJson, err := gjson.DecodeToJson(bodyStr)
	if err != nil {
		g.Log().Errorf(ctx, "prometheus告警信息解析失败: %s", err.Error())
		return nil, err
	}

	alertsI := bodyJson.Get("alerts").Slice()
	if len(alertsI) == 0 {
		g.Log().Errorf(ctx, "告警信息为空")
		return nil, nil
	}
	for _, alert := range alertsI {
		alerts = append(alerts, gjson.New(alert))
	}
	return alerts, nil
}

func (s *sPrometheus) Record(ctx context.Context, record g.Map) (bool, error) {
	data := entity.PrometheusReport{}

	// 检查是否存在相同条件的未解决记录
	oldRecord := entity.PrometheusReport{}
	oldCount, err := dao.PrometheusReport.Ctx(ctx).
		Where("k8s_cluster", tools.GetMapStr(record, "k8s_cluster")).
		Where("alertname", tools.GetMapStr(record, "alertname")).
		Where("env", tools.GetMapStr(record, "env")).
		Where("summary", tools.GetMapStr(record, "summary")).
		Where("level", tools.GetMapStr(record, "level")).
		Where("labels", tools.GetMapStr(record, "labels")).
		Where("is_resolved", 0).
		Order("start_time DESC").
		Limit(1).
		Count()

	if err != nil {
		g.Log().Errorf(ctx, "查询旧记录失败: %s", err.Error())
		return false, err
	}

	// 如果是已解决的告警，更新旧记录的结束时间
	if tools.GetMapInt(record, "is_resolved") == 1 {
		if oldRecord.Id > 0 {
			endTime := tools.GetMapStr(record, "end_time")
			utc8EndTime := gtime.NewFromStr(endTime).Add(8 * gtime.H)
			_, err := dao.PrometheusReport.Ctx(ctx).
				Where("id", oldRecord.Id).
				Data(g.Map{
					"end_time":    utc8EndTime,
					"is_resolved": 1,
				}).
				Update()
			return err == nil, err
		}
		return true, nil // 无匹配记录时直接返回
	}

	// 如果没有找到未解决的旧记录，则插入新记录
	if oldCount == 0 {
		data.Alertname = tools.GetMapStr(record, "alertname")
		data.K8SCluster = tools.GetMapStr(record, "k8s_cluster")
		data.Env = tools.GetMapStr(record, "env")
		data.Level = tools.GetMapStr(record, "level")
		startTime := tools.GetMapStr(record, "start_time")
		data.StartTime = gtime.NewFromStr(startTime).Add(8 * gtime.H)
		data.Labels = tools.GetMapStr(record, "labels")
		data.Description = tools.GetMapStr(record, "description")
		data.Summary = tools.GetMapStr(record, "summary")
		data.IsResolved = 0

		_, err := dao.PrometheusReport.Ctx(ctx).Insert(data)
		if err != nil {
			g.Log().Errorf(ctx, "插入新告警记录失败: %s", err.Error())
			return false, err
		}
		return true, nil
	}

	// 存在重复记录时跳过插入
	return true, nil
}

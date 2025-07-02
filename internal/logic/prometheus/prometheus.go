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

	if tools.GetMapInt(record, "is_resolved") == 1 {
		// 查询上一条同条件的告警记录（假设使用数据库ORM操作）
		oldRecord := entity.PrometheusReport{}
		err := dao.PrometheusReport.Ctx(ctx).
			Where("k8s_cluster", tools.GetMapStr(record, "k8s_cluster")).
			Where("alertname", tools.GetMapStr(record, "alertname")).
			Where("env", tools.GetMapStr(record, "env")).
			Where("summary", tools.GetMapStr(record, "summary")).
			Where("is_resolved", 0). // 只查询未解决的记录
			Order("start_time DESC").
			Limit(1).
			Scan(&oldRecord)

		if err != nil {
			return false, err
		}

		fmt.Println("存在该记录：", oldRecord.StartTime, oldRecord.EndTime)

		//startTime:    2025-06-30 16:00:13
		//endTime:      0001-01-01 16:05:43

		endTime := tools.GetMapStr(record, "end_time")
		utc8EndTime := gtime.NewFromStr(endTime).Add(8 * gtime.H) // 将结束时间转换为北京时间

		// 更新上一条记录的 end_time
		if oldRecord.Id > 0 { // 确保记录存在
			_, err := dao.PrometheusReport.Ctx(ctx).
				Where("id", oldRecord.Id).
				Data(g.Map{
					"end_time":    utc8EndTime,
					"is_resolved": 1, // 标记为已解决
				}).
				Update()
			if err != nil {
				return false, err
			}

		}
	} else {
		data.Alertname = tools.GetMapStr(record, "alertname")
		data.K8SCluster = tools.GetMapStr(record, "k8s_cluster")
		data.Env = tools.GetMapStr(record, "env")
		data.Level = tools.GetMapStr(record, "level")
		startTime := tools.GetMapStr(record, "start_time")
		data.StartTime = gtime.NewFromStr(startTime).Add(8 * gtime.H)

		//data.EndTime = tools.GetMapTime(record, "end_time")
		data.Labels = tools.GetMapStr(record, "labels")
		data.Description = tools.GetMapStr(record, "description")
		data.Summary = tools.GetMapStr(record, "summary")
		data.Labels = tools.GetMapStr(record, "labels")
		data.IsResolved = 0

		// 插入新记录（当前告警）
		_, err := dao.PrometheusReport.Ctx(ctx).Insert(data)
		return err == nil, err
	}

	return true, nil
}

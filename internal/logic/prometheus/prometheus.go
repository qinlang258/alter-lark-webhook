package prometheus

import (
	"alter-lark-webhook/internal/dao"
	"alter-lark-webhook/internal/logic/tools"
	"alter-lark-webhook/internal/model/entity"
	"alter-lark-webhook/internal/service"

	"github.com/gogf/gf/v2/os/glog"
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
	k8sCluster := tools.GetMapStr(record, "k8s_cluster")
	alertname := tools.GetMapStr(record, "alertname")
	env := tools.GetMapStr(record, "env")
	summary := tools.GetMapStr(record, "summary")
	level := tools.GetMapStr(record, "level")
	labels := tools.GetMapStr(record, "labels")
	itemName := tools.GetMapStr(record, "item_name")
	startTime := tools.GetMapStr(record, "start_time")

	// 检查是否存在相同条件的未解决记录
	oldRecord := &entity.PrometheusReport{}
	err := dao.PrometheusReport.Ctx(ctx).
		Where("k8s_cluster", k8sCluster).
		Where("alertname", alertname).
		Where("env", env).
		Where("item_name", itemName).
		Where("level", level).
		Where("is_resolved", 0).
		Order("start_time DESC").
		Limit(1).
		Scan(oldRecord)

	if err != nil {
		glog.Errorf(ctx, "没有找到老的记录: %s", err.Error())
		return false, err
	}

	// 如果没有找到未解决的旧记录，则插入新记录
	data := &entity.PrometheusReport{}

	if oldRecord.Id == 0 && oldRecord.IsResolved == 0 {
		data.Alertname = alertname
		data.K8SCluster = k8sCluster
		data.Env = env
		data.Level = level
		data.ItemName = itemName
		fmt.Println("startTime:::::::::::::: ", startTime)
		data.StartTime = gtime.NewFromStr(startTime).Add(8 * gtime.H)
		data.Labels = labels
		data.Description = tools.GetMapStr(record, "description")
		data.Summary = summary
		data.IsResolved = 0

		_, err := dao.PrometheusReport.Ctx(ctx).Insert(data)
		if err != nil {
			g.Log().Errorf(ctx, "插入新告警记录失败: %s", err.Error())
			return false, err
		}
		return true, nil
	} else {
		// 如果是已解决的告警，更新旧记录的结束时间
		if tools.GetMapStr(record, "is_resolved") == "1" {
			endTime := tools.GetMapStr(record, "end_time")
			var utc8EndTime *gtime.Time
			if endTime == "N/A" {
				endTime = ""
				utc8EndTime = nil
			} else {
				utc8EndTime = gtime.NewFromStr(endTime).Add(8 * gtime.H)
			}
			_, err := dao.PrometheusReport.Ctx(ctx).
				Where("id = ? and item_name = ?", oldRecord.Id, itemName).
				Data(g.Map{
					"end_time":    utc8EndTime,
					"is_resolved": 1,
				}).
				Update()
			if err != nil {
				g.Log().Errorf(ctx, "更新告警记录失败: %s", err.Error())
				return false, err
			}
			g.Log().Infof(ctx, "更新告警记录成功: %s", oldRecord.Id)
			return true, nil // 无匹配记录时直接返回
		}
	}
	// 存在重复记录时跳过插入
	return true, nil
}

func (s *sPrometheus) Test(ctx context.Context, query g.Map) {
	data := entity.PrometheusReport{}
	err := dao.PrometheusReport.Ctx(ctx).Where("id = ?", 555).Scan(&data)
	if err != nil {
		glog.Error(ctx, err.Error())
	}

	fmt.Println(data)
}

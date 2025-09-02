package prometheus

import (
	"alter-lark-webhook/internal/dao"
	"alter-lark-webhook/internal/logic/tools"
	"alter-lark-webhook/internal/model/entity"
	"alter-lark-webhook/internal/service"
	"time"

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

// 单独处理已解决告警
func (s *sPrometheus) handleResolvedAlert(ctx context.Context, record g.Map, oldRecord *entity.PrometheusReport) (bool, error) {
	fmt.Println("oldRecord: ", oldRecord)
	endTime := tools.GetMapStr(record, "end_time")
	var utc8EndTime *gtime.Time
	if endTime == "N/A" {
		utc8EndTime = nil
	} else {
		utc8EndTime = gtime.NewFromStr(endTime)
	}

	_, err := dao.PrometheusReport.Ctx(ctx).
		Where("id = ?", oldRecord.Id).
		Data(g.Map{
			"end_time":    utc8EndTime,
			"is_resolved": 1,
		}).
		Update()
	if err != nil {
		glog.Errorf(ctx, "更新告警记录失败: %s", err.Error())
		return false, err
	}
	return true, nil // 解决告警总是需要通知
}

func (s *sPrometheus) Record(ctx context.Context, record g.Map) (bool, error) {
	// 提取公共字段
	k8sCluster := tools.GetMapStr(record, "k8s_cluster")
	alertname := tools.GetMapStr(record, "alertname")
	env := tools.GetMapStr(record, "env")
	itemName := tools.GetMapStr(record, "item_name")
	level := tools.GetMapStr(record, "level")
	startTime := tools.GetMapStr(record, "start_time")
	// labels的map顺序是不固定的
	labels := tools.GetMapStr(record, "labels")
	summary := tools.GetMapStr(record, "summary")

	// 查询最近一条未解决的相同告警
	oldRecord := &entity.PrometheusReport{}

	IsResolved, err := dao.PrometheusReport.Ctx(ctx).
		Where("k8s_cluster", k8sCluster).
		Where("alertname", alertname).
		Where("env", env).
		Where("item_name", itemName).
		Where("level", level).
		Where("is_resolved", 0).
		Order("start_time DESC").
		Count()

	if err != nil {
		glog.Error(ctx, err.Error())
	}

	if IsResolved == 0 {
		glog.Info(ctx, "没有找到相关的告警记录")
	} else {
		err = dao.PrometheusReport.Ctx(ctx).
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
			glog.Errorf(ctx, "查询告警记录失败: %s", err.Error())
		}
	}

	gtime.SetTimeZone("UTC-8")

	now := gtime.Now()
	shouldResend := false

	// 检查是否需要重新发送告警（满足以下任一条件）：
	// 1. 没有找到未解决的旧记录（全新告警）
	// 2. 旧记录的start_time距离当前时间超过10分钟
	if IsResolved == 0 {
		shouldResend = true

		data := &entity.PrometheusReport{}
		data.Alertname = alertname
		data.K8SCluster = k8sCluster
		data.Env = env
		data.Level = level
		data.ItemName = itemName
		data.StartTime = gtime.NewFromStr(startTime)
		data.Labels = labels
		data.Description = tools.GetMapStr(record, "description")
		data.Summary = summary
		data.IsResolved = 0

		_, err := dao.PrometheusReport.Ctx(ctx).Insert(data)
		if err != nil {
			glog.Errorf(ctx, "插入新告警记录失败: %s", err.Error())
		}
	}

	//当一条记录重复告警的时候，任然发告警
	if now.Sub(gtime.NewFromStr(startTime)) > 10*time.Minute && oldRecord.IsResolved == 0 {
		glog.Infof(ctx, "告警记录超过10分钟未解决，重新发送告警: %s", record)
		shouldResend = true
	}

	// 处理已解决状态
	if tools.GetMapStr(record, "is_resolved") == "1" {
		return s.handleResolvedAlert(ctx, record, oldRecord)
	}

	return shouldResend, nil // 不触发告警
}

func (s *sPrometheus) Test(ctx context.Context, query g.Map) {
	data := entity.PrometheusReport{}
	err := dao.PrometheusReport.Ctx(ctx).Where("id = ?", 555).Scan(&data)
	if err != nil {
		glog.Error(ctx, err.Error())
	}

}

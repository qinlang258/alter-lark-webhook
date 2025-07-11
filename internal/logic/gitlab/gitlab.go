package gitlab

import (
	"alter-lark-webhook/internal/dao"
	"alter-lark-webhook/internal/logic/tools"
	"alter-lark-webhook/internal/model/entity"
	"alter-lark-webhook/internal/service"
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/xanzy/go-gitlab"
)

type sGitlab struct {
	GitClient *gitlab.Client
}

func init() {
	myGitlab := New()
	// 初始化gitlab客户端
	if myGitlab.GitClient == nil {
		ctx := context.Background()
		url := g.Cfg().MustGet(ctx, "gitlab.url").String()
		token := g.Cfg().MustGet(ctx, "gitlab.token").String()
		git, err := gitlab.NewClient(token, gitlab.WithBaseURL(url+"/api/v4"))
		if err != nil {
			g.Log().Fatalf(ctx, "创建gitlab客户端失败: %v", err)
		}
		myGitlab.GitClient = git
	}

	service.RegisterGitlab(myGitlab)
}

func New() *sGitlab {
	return &sGitlab{}
}

func (s *sGitlab) GetProjectIDByPath(ctx context.Context, projectPath string) (int, error) {

	project, _, err := s.GitClient.Projects.GetProject(projectPath, &gitlab.GetProjectOptions{})
	if err != nil {
		return 0, fmt.Errorf("查询项目失败: %v", err)
	}

	return project.ID, nil

}

func (s *sGitlab) GetUserInfoByImageUrl(ctx context.Context, imageUrl string) (map[string]string, error) {
	// 原始 imageUrl: 116981788283.dkr.ecr.ap-east-1.amazonaws.com/chief/user/chief-sso.git/sso-server:dev-dev-68c9e27e-20250710_103848

	message := make(map[string]string)

	fieldsList := strings.Split(imageUrl, ":")
	commitId := strings.Split(fieldsList[1], "-")[2]

	imageRe := regexp.MustCompile(`116981788283.dkr.ecr.ap-east-1.amazonaws.com/(.*?).git/([^:]+)`)

	match := imageRe.FindStringSubmatch(imageUrl)
	if len(match) < 3 {
		return nil, fmt.Errorf("未找到匹配的 Git 路径")
	}

	message["projectPath"] = match[1]
	message["serviceName"] = match[2]

	projectId, err := s.GetProjectIDByPath(ctx, match[1])
	if err != nil {
		glog.Error(ctx, err.Error())
		return nil, err
	}

	// 查询提交信息
	commit, _, err := s.GitClient.Commits.GetCommit(projectId, commitId, &gitlab.GetCommitOptions{})
	if err != nil {
		glog.Error(ctx, err.Error())
		//return nil, fmt.Errorf("查询提交信息失败: %v", err)
	}

	message["committerName"] = commit.CommitterName
	message["committerEmail"] = commit.CommitterEmail
	message["commitId"] = commitId

	//fmt.Println("============================ ", message)

	return message, nil

}

func (s *sGitlab) GetByImageUrlSendOomToFeishu(ctx context.Context, imageUrl string) (map[string]string, map[string]interface{}, error) {
	data, err := s.GetUserInfoByImageUrl(ctx, imageUrl)
	if err != nil {
		glog.Error(ctx, err)
		return nil, nil, err
	}

	//根据 serviceName 获取最近的告警日志
	var prometheusReport entity.PrometheusReport
	err = dao.PrometheusReport.Ctx(ctx).
		Where("item_name like ?", fmt.Sprintf("%%%s%%", data["serviceName"])).
		Where("is_resolved = 1").
		OrderDesc("start_time").
		Scan(&prometheusReport)

	if err != nil {
		glog.Error(ctx, err.Error())
		return nil, nil, err
	}

	oomPayload := tools.BuildOOMRichTextMessage(prometheusReport.Alertname, prometheusReport.Level, prometheusReport.Description, prometheusReport.Env, prometheusReport.StartTime.String(), prometheusReport.Labels, prometheusReport.Level, prometheusReport.Summary)

	return data, oomPayload, nil

}

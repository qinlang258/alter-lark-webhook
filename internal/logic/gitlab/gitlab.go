package gitlab

import (
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

func (s *sGitlab) SendOomToFeishu(ctx context.Context, imageUrl string) (bool, error) {
	// 原始 imageUrl: 116981788283.dkr.ecr.ap-east-1.amazonaws.com/chief/user/chief-sso.git/sso-server:dev-dev-68c9e27e-20250710_103848

	fieldsList := strings.Split(imageUrl, ":")
	commitId := strings.Split(fieldsList[1], "-")[3]

	re := regexp.MustCompile(`116981788283.dkr.ecr.ap-east-1.amazonaws.com/(.*?).git`)

	match := re.FindStringSubmatch(imageUrl)
	if len(match) < 2 {
		return false, fmt.Errorf("未找到匹配的 Git 路径")
	}

	projectId, err := s.GetProjectIDByPath(ctx, match[1])
	if err != nil {
		glog.Error(ctx, err.Error())
		return false, err
	}

	// 查询提交信息
	commit, _, err := s.GitClient.Commits.GetCommit(projectId, commitId, &gitlab.GetCommitOptions{})
	if err != nil {
		return false, fmt.Errorf("查询提交信息失败: %v", err)
	}

	committerName := commit.CommitterName
	committerEmail := commit.CommitterEmail

	fmt.Println(committerName, committerEmail, commit.AuthorEmail)

	return true, nil
}

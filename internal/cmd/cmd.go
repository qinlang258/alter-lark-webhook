package cmd

import (
	"alter-lark-webhook/internal/controller"
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Bind(
					controller.Prometheus,
					controller.Gitlab,
					controller.Feishu,
				)
			})

			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware()
				// 处理 WebSocket 连接
				group.Bind(controller.Skywalking) // 你为路由绑定具体的控制器方法
			})

			s.SetPort(8000)
			s.Run()
			return nil
		},
	}
)

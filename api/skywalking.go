package api

import (
	"fmt"

	"github.com/gogf/gf/v2/net/ghttp"
)

func AlarmHandler(r *ghttp.Request) {
	// 1. 获取原始 body
	body := r.GetBody()

	// 2. 打印或处理
	fmt.Println("收到 SkyWalking 告警：")
	fmt.Println(string(body))

	// 3. 返回 200（SkyWalking 只认 HTTP 成功）
	r.Response.WriteStatus(200)
}

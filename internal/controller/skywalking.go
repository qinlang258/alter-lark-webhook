package controller

import (
	"fmt"

	"github.com/gogf/gf/v2/net/ghttp"
)

type cSkywalking struct{}

var Skywalking = cSkywalking{}

// SkyWalking 告警接收
func (c *cSkywalking) Skywalking(r *ghttp.Request) {
	body := r.GetBody()

	fmt.Println(body)

	// 一定要返回 200
	r.Response.WriteStatus(200)
}

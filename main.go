package main

import (
	_ "alter-lark-webhook/internal/packed"

	_ "alter-lark-webhook/internal/logic"

	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/v2/os/gctx"

	"alter-lark-webhook/internal/cmd"
)

func main() {
	g.Log().SetFlags(glog.F_TIME_STD | glog.F_FILE_SHORT)

	gtime.SetTimeZone("UTC-8")

	cmd.Main.Run(gctx.New())
}

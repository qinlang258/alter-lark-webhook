package main

import (
	_ "alter-lark-webhook/internal/packed"

	_ "alter-lark-webhook/internal/logic"

	"github.com/gogf/gf/v2/os/gctx"

	"alter-lark-webhook/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.New())
}

package middlewares

import (
	"finally-gi/gi"
	"fmt"
	"time"
)

func TimeLog() gi.HTTPHandler {
	return func(ctx *gi.Context) {
		now := time.Now()
		ctx.Next()
		fmt.Println(fmt.Sprintf("[%v: %v] : take %8dms", ctx.Method, ctx.Path, time.Since(now)/time.Millisecond))
	}
}

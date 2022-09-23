package middlewares

import (
	"finally-gi/gi"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

func trace(msg string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:])
	str := strings.Builder{}

	str.WriteString(fmt.Sprintf("err : \n%s\n", msg))
	for _, pc := range pcs[:n] {
		funcInfo := runtime.FuncForPC(pc)
		file, line := funcInfo.FileLine(pc)
		str.WriteString(fmt.Sprintf("\tat file:[%s] [%d row] \n", file, line))
	}
	return str.String()
}

func Recover() gi.HTTPHandler {
	return func(ctx *gi.Context) {
		defer func() {
			if err := recover(); err != nil {
				msg := fmt.Sprintf("%s", err)
				log.Println(trace(msg))
				ctx.Fail(http.StatusBadGateway, "server err")
			}
		}()
		ctx.Next()
	}
}

package grunway

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/amattn/deeperror"
)

type MiddlewareProcessor interface {
	Process(routePtr *Route, ctx *Context) (terminateEarly bool, derr *deeperror.DeepError)
}

type PostProcessor interface {
	Process(ctx *Context) (terminateEarly bool, derr *deeperror.DeepError)
}

type CommonLogger struct {
}

func (logger *CommonLogger) Process(ctx *Context) (terminateEarly bool, derr *deeperror.DeepError) {
	commonLogFormat(ctx)
	return false, nil
}

func commonLogFormat(ctx *Context) {
	// http://en.wikipedia.org/wiki/Common_Log_Format
	// example 127.0.0.1 user-identifier frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326

	user := "-"
	if ctx.Req.URL.User != nil {
		user = ctx.Req.URL.User.Username()
	}
	common_log_format_parts := []string{
		ctx.Req.RemoteAddr,
		"-",
		user,
		time.Now().Format("[02/Jan/2006:15:04:05 -0700]"),
		`"` + ctx.Req.Method,
		ctx.Req.URL.RequestURI(),
		ctx.Req.Proto + `"`,
		strconv.FormatInt(int64(ctx.StatusCode), 10),
		strconv.FormatInt(int64(ctx.ContentLength), 10),
		"\n",
	}
	fmt.Print(strings.Join(common_log_format_parts, " "))

}

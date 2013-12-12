package grunway

import (
	"fmt"
	"github.com/amattn/deeperror"
	"strconv"
	"strings"
	"time"
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
	if ctx.R.URL.User != nil {
		user = ctx.R.URL.User.Username()
	}
	common_log_format_parts := []string{
		ctx.R.RemoteAddr,
		"-",
		user,
		time.Now().Format("[02/Jan/2006:15:04:05 -0700]"),
		`"` + ctx.R.Method,
		ctx.R.URL.RequestURI(),
		ctx.R.Proto + `"`,
		strconv.FormatInt(int64(ctx.StatusCode), 10),
		strconv.FormatInt(int64(ctx.ContentLength), 10),
		"\n",
	}
	fmt.Print(strings.Join(common_log_format_parts, " "))

}

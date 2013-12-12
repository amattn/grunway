package grunway

import (
	"github.com/amattn/deeperror"
)

type PrehandleProcessor interface {
	Process(routePtr *Route, ctx *Context) (isFinished bool, derr *deeperror.DeepError)
}

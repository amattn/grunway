package grunway

type RouteHandler func(*Context) RouteHandlerResult

type RouteHandlerResult struct {
	rerr *RouteError
	pmap PayloadsMap
	crr  CustomRouteResponse
}

type RouteError struct {
	code   int    // HTTP Status code
	errNo  int64  // Internal error number
	errStr string // Client Visible error message
}

type CustomRouteResponse func(*Context)

// #######
// #       #####  #####   ####  #####
// #       #    # #    # #    # #    #
// #####   #    # #    # #    # #    #
// #       #####  #####  #    # #####
// #       #   #  #   #  #    # #   #
// ####### #    # #    #  ####  #    #
//

func NewRouteError(code int, errNo int64, errStr string) *RouteError {
	rerr := new(RouteError)
	rerr.code = code
	rerr.errNo = errNo
	rerr.errStr = errStr
	return rerr
}

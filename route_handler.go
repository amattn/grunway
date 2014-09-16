package grunway

type RouteHandler func(*Context) RouteHandlerResult

type RouteHandlerResult struct {
	rerr *RouteError
	pmap PayloadsMap
	crr  CustomRouteResponse
}

type RouteError struct {
	statusCode int // HTTP Status code
	errorInfo  ErrorInfo
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

func NewRouteError(code int, errInfo ErrorInfo) *RouteError {
	rerr := new(RouteError)
	rerr.statusCode = code
	rerr.errorInfo = errInfo
	return rerr
}

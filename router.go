package grunway

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"sort"
	"strings"
)

const (
	NotFoundPrefix                            = "404 Not Found"
	NotFoundErrorNumber                       = 4040000404
	BadRequestPrefix                          = "400 Bad Request"
	BadRequestErrorNumber                     = 4000000000
	BadRequestSyntaxErrorPrefix               = BadRequestPrefix + ": Syntax Error"
	BadRequestSyntaxErrorErrorNumber          = 4000000001
	BadRequestMissingPrimaryKeyErrorNumber    = 4000000002
	BadRequestExtraneousPrimaryKeyErrorNumber = 4000000003
	// BadRequestMissingPrimaryKeyPrefix    = BadRequestPrefix + ": Missing Id"
	// BadRequestExtraneousPrimaryKeyPrefix = BadRequestPrefix + ": Extraneous Id"

	InternalServerErrorPrefix = "500 Internal Server Error"
)

type PayloadController interface {
}

type Router struct {
	BasePath string

	MiddlewareProcessors []MiddlewareProcessor
	PostProcessors       []PostProcessor

	Controllers map[string]PayloadController // key is entity name
	RouteMap    map[string]*Route            // key is entity name
}

func NewRouter() *Router {

	router := new(Router)

	router.Controllers = make(map[string]PayloadController)
	router.RouteMap = make(map[string]*Route)

	router.MiddlewareProcessors = []MiddlewareProcessor{}
	router.PostProcessors = []PostProcessor{
		new(CommonLogger),
	}
	return router
}

//  #####
// #     #  ####  #    # ###### #  ####
// #       #    # ##   # #      # #    #
// #       #    # # #  # #####  # #
// #       #    # #  # # #      # #  ###
// #     # #    # #   ## #      # #    #
//  #####   ####  #    # #      #  ####
//

// Configuration of Router

func (router *Router) RegisterEntity(name string, payloadController PayloadController) {
	payloadControllerType := reflect.TypeOf(payloadController)
	payloadControllerValue := reflect.ValueOf(payloadController)

	if isValid, reason := ValidateEntityName(name); isValid == false {
		log.Fatalln("Invalid Enitity name:'", name, "'", reason)
	}
	if payloadController == nil {
		log.Fatalln("untypedHandlerWrapper currently must not be nil")
	}

	router.Controllers[name] = payloadController

	authenticator, _ := payloadController.(AuthHandler)

	for i := 0; i < payloadControllerType.NumMethod(); i++ {

		potentialHandlerMethod := payloadControllerType.Method(i)
		potentialHandlerName := potentialHandlerMethod.Name
		if len(potentialHandlerName) > 0 && potentialHandlerName[0] == strings.ToUpper(potentialHandlerName)[0] {
			// skip unexported methods
			unknownhandler := payloadControllerValue.MethodByName(potentialHandlerName).Interface()
			router.AddEntityRoute(name, payloadControllerType.String(), potentialHandlerName, unknownhandler, authenticator)
		}
	}
}

func (router *Router) AddEntityRoute(entityName, controllerName, handlerName string, unknownhandler interface{}, authenticator AuthHandler) {

	// simple first:
	if strings.Contains(handlerName, MAGIC_HANDLER_KEYWORD) == false {
		// just skip it
		return
	}

	isValid, reason, handler := ValidateHandler(unknownhandler)
	if isValid == false {
		errNum := int64(3230075622)
		errMsg := fmt.Sprintln(errNum, "Handler Validation Failure:", "entityName:", entityName, "controllerName:", controllerName, "Invalid Handler:", handlerName, "reason:", reason)
		// derr := deeperror.New(errNum, errMsg, nil)
		log.Println(errMsg)
		// skip... invalid prefix
		return
	}

	routePtr := new(Route)
	routePtr.Path = entityName + "/"
	routePtr.EntityName = entityName
	routePtr.Handler = handler
	routePtr.HandlerName = handlerName
	routePtr.ControllerName = controllerName

	// Step 1 Check for Auth prrefix
	deauthedHandlerName := handlerName
	if strings.HasPrefix(handlerName, MAGIC_AUTH_REQUIRED_PREFIX) {
		deauthedHandlerName = handlerName[len(MAGIC_AUTH_REQUIRED_PREFIX):]
		routePtr.RequiresAuth = true
		if authenticator == nil {
			log.Fatalf("1323798307 Auth required handler defined (%s), but controller (%s) does not implement AuthHandler", handlerName, controllerName)
			return
		}
		routePtr.Authenticator = authenticator
	}

	// step 2 Find method
	var versionActionHandlerName string
	switch {
	case strings.HasPrefix(deauthedHandlerName, MAGIC_GET_HANDLER_PREFIX):
		routePtr.Method = "GET"
		versionActionHandlerName = deauthedHandlerName[len(MAGIC_GET_HANDLER_PREFIX):]
	case strings.HasPrefix(deauthedHandlerName, MAGIC_POST_HANDLER_PREFIX):
		routePtr.Method = "POST"
		versionActionHandlerName = deauthedHandlerName[len(MAGIC_POST_HANDLER_PREFIX):]
	case strings.HasPrefix(deauthedHandlerName, MAGIC_PUT_HANDLER_PREFIX):
		routePtr.Method = "PUT"
		versionActionHandlerName = deauthedHandlerName[len(MAGIC_PUT_HANDLER_PREFIX):]
	case strings.HasPrefix(deauthedHandlerName, MAGIC_DELETE_HANDLER_PREFIX):
		routePtr.Method = "DELETE"
		versionActionHandlerName = deauthedHandlerName[len(MAGIC_DELETE_HANDLER_PREFIX):]
	case strings.HasPrefix(deauthedHandlerName, MAGIC_PATCH_HANDLER_PREFIX):
		routePtr.Method = "PATCH"
		versionActionHandlerName = deauthedHandlerName[len(MAGIC_PATCH_HANDLER_PREFIX):]
	case strings.HasPrefix(deauthedHandlerName, MAGIC_HEAD_HANDLER_PREFIX):
		routePtr.Method = "HEAD"
		versionActionHandlerName = deauthedHandlerName[len(MAGIC_HEAD_HANDLER_PREFIX):]
	default:
		// skip... it's not a known prefix
		log.Println("1860816435 Skipping Route:", entityName, controllerName, handlerName)
		return
	}

	// do a bit of primite parsing:

	if isValid, reason := ValidateHandlerName(handler); isValid == false {
		log.Fatalln("1411397818 entity name:", routePtr.EntityName, "method:", routePtr.Method, "route:", routePtr.Path, "Invalid Handler:", handlerName, "reason:", reason)
	}

	// log.Println("versionActionHandlerName", versionActionHandlerName)

	versionStr, action := parseVersionFromPrefixlessHandlerName(versionActionHandlerName)
	if versionStr == "" {
		log.Println("1259486570 Skipping Route:", entityName, controllerName, handlerName)
		// skip... invalid prefix
		return
	}

	// log.Println("version, action", version, action)
	routePtr.Action = action
	routePtr.Path += action
	routePtr.VersionStr = versionStr

	setRoute(router.RouteMap, routePtr.Method, routePtr.VersionStr, routePtr.Action, routePtr)
}

// Convenience method
func (router *Router) AllRoutesCount() int {
	return len(router.RouteMap)
}

// Basically just used for logging and debugging.
// the first addon is a prefix, all remaining addons are treated as suffixes and appended to the end
func (router *Router) AllRoutesDescription(addons ...string) []string {
	// log.Println("104194464 All Routes:")

	var prefix string
	var suffix string
	if len(addons) >= 1 {
		prefix = addons[0]
	}
	if len(addons) >= 2 {
		suffix = strings.Join(addons[1:], " ")
	}

	count := len(router.RouteMap)

	routeKeys := make([]string, 0, count)
	for routeKey, _ := range router.RouteMap {
		routeKeys = append(routeKeys, routeKey)
	}
	sort.Strings(routeKeys)

	lines := make([]string, 0, count)

	for _, routeKey := range routeKeys {
		routePtr := router.RouteMap[routeKey]
		method, versionStr, entityName, action := routeComponents(routeKey)
		handlerType := reflect.TypeOf(routePtr.Handler)

		if action == "" {
			action = "<NONE>"
		}

		line := fmt.Sprintln(
			method,
			fmt.Sprintf("%vv%v/%v", router.BasePath, versionStr, routePtr.Path),
			"Entity:", entityName,
			"Action:", action,
			routePtr.ControllerName,
			routePtr.HandlerName,
			handlerType,
		)

		line = strings.Join([]string{prefix, line, suffix}, " ")
		line = strings.TrimSpace(line)

		lines = append(lines, line)
	}
	// log.Println("104194464 End Routes")

	// log.Println("104194464 RouteKeys")
	// for routeKey, _ := range router.RouteMap {
	// 	log.Println(routeKey)
	// }
	return lines
}

// Basically just used for logging and debugging.
// the first addon is a prefix, all remaining addons are treated as suffixes and appended to the end
func (router *Router) AllRoutesSummary(addons ...string) string {
	lines := router.AllRoutesDescription(addons...)
	lines = append(lines, "") // basically, append an newline at the end.
	return strings.Join(lines, "\n")
}

func (router *Router) LogAllRoutes(addons ...string) {
	lines := router.AllRoutesDescription(addons...)
	for _, line := range lines {
		log.Println(line)
	}
}

//  #####                              #     # ####### ####### ######
// #     # ###### #####  #    # ###### #     #    #       #    #     #
// #       #      #    # #    # #      #     #    #       #    #     #
//  #####  #####  #    # #    # #####  #######    #       #    ######
//       # #      #####  #    # #      #     #    #       #    #
// #     # #      #   #   #  #  #      #     #    #       #    #
//  #####  ###### #    #   ##   ###### #     #    #       #    #
//

// ServeHTTP does the basics:
// 1. Any pre-handler stuff
// 2. parse the route
// 3. lookup route
// 4. validate/auth route
// 5. Auth (if necessary)
// 6. Middleware
// 7. call handler method

// all writes to the responseWriter are done through writePayloadWrapper.

// any post handler stuff should be called in writePayloadWrapper

func (router *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 1. Any pre-handler stuff
	// TODO

	ctx := new(Context) // needs a leakybucket
	ctx.w = w
	ctx.Req = req
	ctx.router = router

	// 2. parse the route
	endpoint, clientDeepErr, serverDeepErr := parsePath(req.URL, router.BasePath)
	ctx.End = endpoint

	if clientDeepErr != nil {
		// log.Println("clientDeepErr", clientDeepErr)
		code := http.StatusBadRequest
		if clientDeepErr.StatusCode > 299 && clientDeepErr.StatusCode < 999 {
			code = clientDeepErr.StatusCode
		}
		ctx.SendSimpleErrorPayload(code, clientDeepErr.Num, fmt.Sprintf("%d %s (err code: %d)", code, BadRequestSyntaxErrorPrefix, clientDeepErr.Num))
		// log.Println("clientDeepErr.Num", clientDeepErr.Num)
		return
	}

	if serverDeepErr != nil {
		log.Println("serverDeepErr", serverDeepErr)
		code := http.StatusInternalServerError
		if serverDeepErr.StatusCode > 299 && serverDeepErr.StatusCode < 999 {
			code = serverDeepErr.StatusCode
		}
		ctx.SendSimpleErrorPayload(code, serverDeepErr.Num, fmt.Sprintf("%d %s (err code: %d)", code, InternalServerErrorPrefix, serverDeepErr.Num))
		return
	}

	router.handleContext(ctx, req)
}

func (router *Router) handleContext(ctx *Context, req *http.Request) {

	// 3. lookup the handler method
	routePtr, err := getRoute(router.RouteMap, req.Method, ctx.End.VersionStr, ctx.End.EntityName, ctx.End.Action)
	if err != nil || routePtr == nil {
		// log.Println("404 routekey", routeKey(req.Method, ctx.End.VersionStr, ctx.End.EntityName, ctx.End.Action))
		// log.Printf("404 for Method:%v, Endpoint %+v, routePtr:%+v, err:%v", req.Method, ctx.E, routePtr, err)
		// http.NotFound(w, req)
		ctx.SendSimpleErrorPayload(http.StatusNotFound, NotFoundErrorNumber, "404 Not Found")
		return
	}

	// log.Println("req.Method", req.Method)
	// log.Println("ctx.End.PrimaryKey", ctx.End.PrimaryKey)
	// log.Println("ctx.End.Extras", ctx.End.Extras)

	// 4. Some basic validation

	if req.Method == "POST" && ctx.End.PrimaryKey != 0 && len(ctx.End.Extras) == 1 {
		// log.Printf("400 for Method:%v, Endpoint %+v, routePtr:%+v, err:%v", req.Method, ctx.E, routePtr, err)
		// don't use http.Error!  use our sendErrorPayload instead
		// http.Error(w, BadRequestExtraneousPrimaryKeyPrefix, http.StatusBadRequest)
		ctx.SendSimpleErrorPayload(http.StatusBadRequest, BadRequestExtraneousPrimaryKeyErrorNumber, BadRequestSyntaxErrorPrefix)
		return
	}
	// Read and update require primary key
	if (req.Method == "GET" || req.Method == "PATCH" || req.Method == "PUT") && ctx.End.PrimaryKey == 0 && len(ctx.End.Extras) == 0 {
		// log.Printf("400 for Method:%v, Endpoint %+v, routePtr:%+v, err:%v", req.Method, ctx.E, routePtr, err)
		ctx.SendSimpleErrorPayload(http.StatusBadRequest, BadRequestMissingPrimaryKeyErrorNumber, BadRequestSyntaxErrorPrefix)
		return
	}

	// 5. Auth

	if routePtr.RequiresAuth {
		// log.Println("RequiresAuth = true")
		isAuthorized, failureToAuthErrorNum := routePtr.Authenticator.PerformAuth(routePtr, ctx)
		if isAuthorized == false {
			ctx.SendSimpleErrorPayload(http.StatusForbidden, int64(failureToAuthErrorNum), "Forbidden")
			return
		}
	}

	// 6. Middleware

	for _, middleware := range router.MiddlewareProcessors {
		middleware.Process(routePtr, ctx)
	}

	// 7. call handler method
	rhr := routePtr.Handler(ctx)
	if rhr.rerr != nil {
		rtErr := rhr.rerr
		ctx.SendErrorInfoPayload(rtErr.statusCode, rtErr.errorInfo)
	} else if rhr.pmap != nil {
		ctx.WrapAndSendPayloadsMap(rhr.pmap)
	} else if rhr.crr != nil {
		rhr.crr(ctx)
	} else {
		ctx.SendSimpleErrorPayload(http.StatusInternalServerError, 2302586595, "Invalid Handler response")
	}

	// 8. any post-handler stuff

	// TODO

}

// RouteMap helpers
const ROUTE_MAP_SEPARATOR = "-{&|!?}-"

func getRoute(routeMap map[string]*Route, method, versionString, entityName, action string) (*Route, error) {
	rk := routeKey(method, versionString, entityName, action)
	// log.Println("rk", rk)
	return routeMap[rk], nil
}
func setRoute(routeMap map[string]*Route, method, versionString, action string, route *Route) error {
	rk := routeKey(method, versionString, route.EntityName, action)
	// log.Println("rk", rk)
	routeMap[rk] = route
	return nil
}
func routeKey(method, versionString, entityName, action string) string {
	return routeKeyJoinString(method, versionString, entityName, action)
}
func routeComponents(routeKey string) (method, versionString, entityName, action string) {
	components := strings.Split(routeKey, ROUTE_MAP_SEPARATOR)
	entityName = components[0]
	method = components[1]
	versionString = components[2]
	action = components[3]
	return
}

func routeKeyJoinString(method, versionString, entityName, action string) string {
	// The order is fairly arbitrary, but makes for nicely ordered list of routes
	// when we sort be routeKey.  (only important for AllRoutesSummary)
	keyComponents := []string{
		strings.ToLower(entityName),
		method,
		versionString,
		strings.ToLower(action)}
	return strings.Join(keyComponents, ROUTE_MAP_SEPARATOR)
}

func routeKeyFormatString(method, versionString, entityName, action string) string {
	return fmt.Sprintf("%s%s%s%s%d%s%s",
		strings.ToLower(entityName),
		ROUTE_MAP_SEPARATOR,
		method,
		ROUTE_MAP_SEPARATOR,
		versionString,
		ROUTE_MAP_SEPARATOR,
		action,
	)
}

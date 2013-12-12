package grunway

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/amattn/deeperror"
)

const (
	NotFoundPrefix                      = "404 Not Found"
	NotFoundErrNo                       = 4040000404
	BadRequestPrefix                    = "400 Bad Request"
	BadRequestErrNo                     = 4000000000
	BadRequestSyntaxErrorPrefix         = BadRequestPrefix + ": Syntax Error"
	BadRequestSyntaxErrorErrNo          = 4000000001
	BadRequestMissingPrimaryKeyErrNo    = 4000000002
	BadRequestExtraneousPrimaryKeyErrNo = 4000000003
	// BadRequestMissingPrimaryKeyPrefix    = BadRequestPrefix + ": Missing Id"
	// BadRequestExtraneousPrimaryKeyPrefix = BadRequestPrefix + ": Extraneous Id"

	InternalServerErrorPrefix = "500 Internal Server Error"
)

type PayloadController interface {
}

type AuthenticatingPayloadController interface {
	PayloadController
	GetSecretKey(publicKey string) (secretKey string, errNum int)
}

type Router struct {
	BasePath string

	PrehandleProcessors []PrehandleProcessor

	Controllers map[string]PayloadController // key is entity name
	RouteMap    map[string]*Route            // key is entity name
}

func NewRouter() *Router {

	router := new(Router)

	router.Controllers = make(map[string]PayloadController)
	router.RouteMap = make(map[string]*Route)
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

	authenticator, _ := payloadController.(AuthenticatingPayloadController)

	for i := 0; i < payloadControllerType.NumMethod(); i++ {

		potentialHandlerMethod := payloadControllerType.Method(i)
		potentialHandlerName := potentialHandlerMethod.Name
		unknownhandler := payloadControllerValue.MethodByName(potentialHandlerName).Interface()
		router.AddEntityRoute(name, payloadControllerType.String(), potentialHandlerName, unknownhandler, authenticator)
	}
}

func (router *Router) AddEntityRoute(entityName, controllerName, handlerName string, unknownhandler interface{}, authenticator AuthenticatingPayloadController) {

	// simple first:
	if strings.Contains(handlerName, MAGIC_HANDLER_KEYWORD) == false {
		// just skip it
		return
	}

	isValid, reason, handler := ValidateHandler(unknownhandler)
	if isValid == false {
		errMsg := fmt.Sprintln("entityName:", entityName, "controllerName:", controllerName, "Invalid Handler:", handlerName, "reason:", reason)
		derr := deeperror.New(3230075622, errMsg, nil)
		log.Println("Handler Validation Failure:", derr)
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
			log.Fatalf("1323798307 Auth required handler defined (%s), but controller (%s) does not implement AuthenticatingPayloadController", handlerName, controllerName)
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
		return
	}

	// do a bit of primite parsing:

	if isValid, reason := ValidateHandlerName(handler); isValid == false {
		log.Fatalln("1411397818 entity name:", routePtr.EntityName, "method:", routePtr.Method, "route:", routePtr.Path, "Invalid Handler:", handlerName, "reason:", reason)
	}

	// log.Println("versionActionHandlerName", versionActionHandlerName)

	versionStr, action := parseVersionFromPrefixlessHandlerName(versionActionHandlerName)
	if versionStr == "" {
		// skip... invalid prefix
		return
	}

	// log.Println("version, action", version, action)
	routePtr.Action = action
	routePtr.Path += action
	routePtr.VersionStr = versionStr

	setRoute(router.RouteMap, routePtr.Method, routePtr.VersionStr, routePtr.Action, routePtr)
}

// Basically just used for logging and debugging.
func (router *Router) AllRoutesSummary() string {
	// log.Println("All Routes:")

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
		lines = append(lines, line)
	}
	// log.Println("End Routes.")

	// log.Println("RouteKeys")
	// for routeKey, _ := range router.RouteMap {
	// 	log.Println(routeKey)
	// }

	return strings.Join(lines, "")
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
// 8. any post-handler stuff (logging, etc.)
func (router *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 1. Any pre-handler stuff
	// TODO

	// 2. parse the route
	endpoint, clientDeepErr, serverDeepErr := parsePath(req.URL, router.BasePath)

	ctx := new(Context) // needs a leakybucket
	ctx.w = w
	ctx.R = req
	ctx.E = endpoint

	if clientDeepErr != nil {
		// log.Println("clientDeepErr", clientDeepErr)
		code := http.StatusBadRequest
		if clientDeepErr.StatusCode > 299 && clientDeepErr.StatusCode < 999 {
			code = clientDeepErr.StatusCode
		}
		ctx.SendErrorPayload(code, clientDeepErr.Num, fmt.Sprintf("%d %s (err code: %d)", code, BadRequestSyntaxErrorPrefix, clientDeepErr.Num))
		// log.Println("clientDeepErr.Num", clientDeepErr.Num)
		return
	}

	if serverDeepErr != nil {
		log.Println("serverDeepErr", serverDeepErr)
		code := http.StatusInternalServerError
		if serverDeepErr.StatusCode > 299 && serverDeepErr.StatusCode < 999 {
			code = serverDeepErr.StatusCode
		}
		ctx.SendErrorPayload(code, serverDeepErr.Num, fmt.Sprintf("%d %s (err code: %d)", code, InternalServerErrorPrefix, serverDeepErr.Num))
		return
	}

	// 3. lookup the handler method
	routePtr, err := getRoute(router.RouteMap, req.Method, endpoint.VersionStr, endpoint.EntityName, endpoint.Action)
	if err != nil || routePtr == nil {
		// log.Println("404 routekey", routeKey(req.Method, endpoint.VersionStr, endpoint.EntityName, endpoint.Action))
		// log.Printf("404 for Method:%v, Endpoint %+v, routePtr:%+v, err:%v", req.Method, endpoint, routePtr, err)
		// http.NotFound(w, req)
		ctx.SendErrorPayload(http.StatusNotFound, NotFoundErrNo, "404 Not Found")
		return
	}

	// log.Println("req.Method", req.Method)
	// log.Println("endpoint.PrimaryKey", endpoint.PrimaryKey)
	// log.Println("endpoint.Extras", endpoint.Extras)

	// 4. Some basic validation

	if req.Method == "POST" && endpoint.PrimaryKey != 0 && len(endpoint.Extras) == 1 {
		// log.Printf("400 for Method:%v, Endpoint %+v, routePtr:%+v, err:%v", req.Method, endpoint, routePtr, err)
		// don't use http.Error!  use our sendErrorPayload instead
		// http.Error(w, BadRequestExtraneousPrimaryKeyPrefix, http.StatusBadRequest)
		ctx.SendErrorPayload(http.StatusBadRequest, BadRequestExtraneousPrimaryKeyErrNo, BadRequestSyntaxErrorPrefix)
		return
	}
	// Read and update require primary key
	if (req.Method == "GET" || req.Method == "PATCH" || req.Method == "PUT") && endpoint.PrimaryKey == 0 && len(endpoint.Extras) == 0 {
		// log.Printf("400 for Method:%v, Endpoint %+v, routePtr:%+v, err:%v", req.Method, endpoint, routePtr, err)
		ctx.SendErrorPayload(http.StatusBadRequest, BadRequestMissingPrimaryKeyErrNo, BadRequestSyntaxErrorPrefix)
		return
	}

	// 5. Auth

	if routePtr.RequiresAuth {
		// log.Println("RequiresAuth = true")
		isAuthorized, failureToAuthErrorNum := performAuth(routePtr, ctx)
		if isAuthorized == false {
			ctx.SendErrorPayload(http.StatusForbidden, int64(failureToAuthErrorNum), "Forbidden")
			return
		}
	}

	// 6. Middleware

	for _, middleware := range router.PrehandleProcessors {
		middleware.Process(routePtr, ctx)
	}

	// 7. call handler method

	routePtr.Handler(ctx)

	// 8. any post-handler stuff
	// logging
	commonLogFormat(ctx)

	// TODO
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

package grunway

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"sort"
	"strings"

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

type Router struct {
	BasePath string

	Controllers map[string]interface{}  // key is entity name
	Payloads    map[string]reflect.Type // key is entity name

	RouteMap map[string]*Route
}

func NewRouter() *Router {

	router := new(Router)

	router.Controllers = make(map[string]interface{})
	router.Payloads = make(map[string]reflect.Type)
	router.RouteMap = make(map[string]*Route)
	return router
}

// Configuration of Router

func (router *Router) RegisterEntity(name string, payloadController, entityPayload interface{}) {
	payloadControllerType := reflect.TypeOf(payloadController)
	payloadControllerValue := reflect.ValueOf(payloadController)
	entityPayloadType := reflect.TypeOf(entityPayload)

	if isValid, reason := ValidateEntityName(name); isValid == false {
		log.Fatalln("Invalid Enitity name:'", name, "'", reason)
	}
	if payloadController == nil {
		log.Fatalln("untypedHandlerWrapper currently must not be nil")
	}
	if entityPayload == nil {
		log.Fatalln("untypedHandlerWrapper currently must not be nil")
	}

	log.Println("Registering payloadController:", payloadControllerType, "for payload:", entityPayloadType)
	router.Controllers[name] = payloadController
	router.Payloads[name] = entityPayloadType

	for i := 0; i < payloadControllerType.NumMethod(); i++ {

		potentialHandlerMethod := payloadControllerType.Method(i)
		potentialHandlerName := potentialHandlerMethod.Name
		unknownhandler := payloadControllerValue.MethodByName(potentialHandlerName).Interface()
		router.AddEntityRoute(name, payloadControllerType.String(), potentialHandlerName, unknownhandler)
	}
}

func (router *Router) AddEntityRoute(entityName, controllerName, handlerName string, unknownhandler interface{}) {

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

	var versionActionHandlerName string
	switch {
	case strings.HasPrefix(handlerName, MAGIC_GET_HANDLER_PREFIX):
		routePtr.Method = "GET"
		versionActionHandlerName = handlerName[len(MAGIC_GET_HANDLER_PREFIX):]
	case strings.HasPrefix(handlerName, MAGIC_POST_HANDLER_PREFIX):
		routePtr.Method = "POST"
		versionActionHandlerName = handlerName[len(MAGIC_POST_HANDLER_PREFIX):]
	case strings.HasPrefix(handlerName, MAGIC_PUT_HANDLER_PREFIX):
		routePtr.Method = "PUT"
		versionActionHandlerName = handlerName[len(MAGIC_PUT_HANDLER_PREFIX):]
	case strings.HasPrefix(handlerName, MAGIC_DELETE_HANDLER_PREFIX):
		routePtr.Method = "DELETE"
		versionActionHandlerName = handlerName[len(MAGIC_DELETE_HANDLER_PREFIX):]
	case strings.HasPrefix(handlerName, MAGIC_PATCH_HANDLER_PREFIX):
		routePtr.Method = "PATCH"
		versionActionHandlerName = handlerName[len(MAGIC_PATCH_HANDLER_PREFIX):]
	case strings.HasPrefix(handlerName, MAGIC_HEAD_HANDLER_PREFIX):
		routePtr.Method = "HEAD"
		versionActionHandlerName = handlerName[len(MAGIC_HEAD_HANDLER_PREFIX):]
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

// ServeHTTP does the basics:
// 1. Any pre-handler stuff
// 2. parse the route
// 3. lookup & call handler method
// 4. any post-handler stuff
func (router *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	// 1. Any pre-handler stuff
	// TODO

	// 2. parse the route
	endpoint, clientDeepErr, serverDeepErr := parsePath(req.URL, router.BasePath)

	ctxPtr := new(Context) // needs a leakybucket
	ctxPtr.W = w
	ctxPtr.R = req
	ctxPtr.E = endpoint

	if clientDeepErr != nil {
		http.Error(w, fmt.Sprintf("%s (err code: %d)", BadRequestSyntaxErrorPrefix, clientDeepErr.Num), http.StatusBadRequest)
		return
	}

	if serverDeepErr != nil {
		http.Error(w, fmt.Sprintf("500 Internal Server Error (err code: %d)", serverDeepErr.Num), http.StatusInternalServerError)
		return
	}

	// 3a. lookup the handler method
	routePtr, err := getRoute(router.RouteMap, req.Method, endpoint.VersionStr, endpoint.EntityName, endpoint.Action)
	if err != nil || routePtr == nil {
		// log.Println("404 routekey", routeKey(req.Method, endpoint.Version, endpoint.EntityName, endpoint.Action))
		log.Printf("404 for Method:%v, Endpoint %+v, routePtr:%+v, err:%v", req.Method, endpoint, routePtr, err)
		// http.NotFound(w, req)
		ctxPtr.SendErrorPayload(http.StatusNotFound, NotFoundErrNo, "404 Not Found")
		return
	}

	// log.Println("req.Method", req.Method)
	// log.Println("endpoint.PrimaryKey", endpoint.PrimaryKey)
	// log.Println("endpoint.Extras", endpoint.Extras)

	// 3b Some basic validation

	if req.Method == "POST" && endpoint.PrimaryKey != 0 && len(endpoint.Extras) == 1 {
		log.Printf("400 for Method:%v, Endpoint %+v, routePtr:%+v, err:%v", req.Method, endpoint, routePtr, err)
		// don't use http.Error!  use our sendErrorPayload instead
		// http.Error(w, BadRequestExtraneousPrimaryKeyPrefix, http.StatusBadRequest)
		ctxPtr.SendErrorPayload(http.StatusBadRequest, BadRequestExtraneousPrimaryKeyErrNo, BadRequestSyntaxErrorPrefix)
		return
	}
	// Read and update require primary key
	if (req.Method == "GET" || req.Method == "PATCH" || req.Method == "PUT") && endpoint.PrimaryKey == 0 && len(endpoint.Extras) == 0 {
		log.Printf("400 for Method:%v, Endpoint %+v, routePtr:%+v, err:%v", req.Method, endpoint, routePtr, err)
		ctxPtr.SendErrorPayload(http.StatusBadRequest, BadRequestMissingPrimaryKeyErrNo, BadRequestSyntaxErrorPrefix)
		return
	}

	typedHandler := routePtr.Handler

	// 3c. call handler method

	typedHandler(ctxPtr)

	// 4. any post-handler stuff

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

package grunway

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

const (
	MAGIC_AUTH_REQUIRED_PREFIX  = "Auth"
	MAGIC_HANDLER_KEYWORD       = "Handler"
	MAGIC_GET_HANDLER_PREFIX    = "GetHandler"    // CRUD: read
	MAGIC_POST_HANDLER_PREFIX   = "PostHandler"   // CRUD: create
	MAGIC_PUT_HANDLER_PREFIX    = "PutHandler"    // CRUD: update (the whole thing)
	MAGIC_PATCH_HANDLER_PREFIX  = "PatchHandler"  // CRUD: update (just a field or two)
	MAGIC_DELETE_HANDLER_PREFIX = "DeleteHandler" // CRUD: delete (duh)
	MAGIC_HEAD_HANDLER_PREFIX   = "HeadHandler"   // usually when you just want to check Etags or something.
)

const VERSION_BIT_DEPTH = 16

type VersionUint uint16

type Route struct {
	RequiresAuth   bool
	Method         string
	Path           string
	VersionStr     string
	EntityName     string
	Action         string
	Authenticator  AuthenticatingPayloadController
	Handler        RouteHandler
	HandlerName    string // not actually used except for logging and debugging
	ControllerName string // not actually used except for logging and debugging
}

type RouteHandler func(*Context) (*RouteError, PayloadsMap, CustomRouteResponse)

type RouteError struct {
	code   int    // HTTP Status code
	errNo  int64  // Internal error number
	errStr string // Client Visible error message
}

func NewRouteError(code int, errNo int64, errStr string) *RouteError {
	rerr := new(RouteError)
	rerr.code = code
	rerr.errNo = errNo
	rerr.errStr = errStr
	return rerr
}

type CustomRouteResponse func(*Context)

func parseVersionFromPrefixlessHandlerName(versionActionHandlerName string) (vStr string, action string) {
	re := regexp.MustCompile("^V([0-9]+)(.*)")
	matches := re.FindStringSubmatch(versionActionHandlerName)
	// log.Printf("%q\n", matches)

	if len(matches) < 3 {
		log.Fatalln("2457509067 Regular Expression failure.  Please file a bug")
	}

	vStr = matches[1]
	vStr = strings.TrimLeft(vStr, "0")
	_, err := strconv.ParseUint(vStr, 10, VERSION_BIT_DEPTH)
	if err != nil {
		return "", ""
	}

	action = matches[2]
	action = strings.ToLower(action)

	return
}

// Validation

func ValidateEntityName(name string) (isValid bool, reason string) {
	if len(name) < 1 {
		return false, "name must have at least one character"
	}
	// TODO: check for valid url chars: [a-Z 0-9 _ -]
	return true, ""
}
func ValidateHandlerName(handler interface{}) (isValid bool, reason string) {
	// TODO length? not much here really.
	return true, ""
}
func ValidateHandler(unknownHandler interface{}) (isValid bool, reason string, handler RouteHandler) {

	// We have to do some type gymnastics here.  first check the if the method matches the raw function type...
	validHandler, ok := unknownHandler.(func(*Context) (*RouteError, PayloadsMap, CustomRouteResponse))

	if ok == false {
		return false, "wrong function type, expected function type of RouteHandler", nil
	}

	// ...then convert the raw funtion type to the typed RouteHandler
	handler = validHandler
	return true, "", handler
}

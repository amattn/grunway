package grunway

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/amattn/deeperror"
)

type Context struct {
	Req *http.Request // original http request
	End Endpoint      // parsed endpoint information

	// only populated after a call to ctx.RequestBody()
	cachedRequestBody      []byte
	cachedRequestBodyError error

	// exposing the responseWriter tends to induce bugs.  We keep this internal for now.
	w http.ResponseWriter

	// router
	router *Router

	// only populated after auth
	PublicKey string // for Auth'd requests, will be set to public key if Auth was successful, "" otherwise

	// generic maps for middleware to stuff arbitrary data
	middleware map[string]interface{}
	postware   map[string]interface{}

	// only populated after a write

	written       bool // true after a write, false before.  Used to prevent "double writes".
	StatusCode    int  // The http status code written out. Only populated after a write.
	ContentLength int  // The number of bytes written out. Only populated after a write.
}

func (ctx *Context) AddHeader(key, value string) {
	ctx.w.Header().Add(key, value)
}
func (ctx *Context) DelHeader(key string) {
	ctx.w.Header().Del(key)
}
func (ctx *Context) GetHeader(key string) string {
	return ctx.w.Header().Get(key)
}
func (ctx *Context) SetHeader(key, value string) {
	ctx.w.Header().Set(key, value)
}

func (ctx *Context) RequestBody() ([]byte, error) {
	if ctx.cachedRequestBody != nil {
		return ctx.cachedRequestBody, nil
	}
	if ctx.cachedRequestBodyError != nil {
		return nil, ctx.cachedRequestBodyError
	}

	ctx.cachedRequestBody, ctx.cachedRequestBodyError = ioutil.ReadAll(ctx.Req.Body)
	return ctx.cachedRequestBody, ctx.cachedRequestBodyError
}

//  #####
// #     # #####   ##   #    # #####    ##   #####  #####
// #         #    #  #  ##   # #    #  #  #  #    # #    #
//  #####    #   #    # # #  # #    # #    # #    # #    #
//       #   #   ###### #  # # #    # ###### #####  #    #
// #     #   #   #    # #   ## #    # #    # #   #  #    #
//  #####    #   #    # #    # #####  #    # #    # #####
//
// ######
// #     # ######  ####  #    # #      #####  ####
// #     # #      #      #    # #        #   #
// ######  #####   ####  #    # #        #    ####
// #   #   #           # #    # #        #        #
// #    #  #      #    # #    # #        #   #    #
// #     # ######  ####   ####  ######   #    ####
//

func (ctx *Context) MakeRouteHandlerResultError(code int, errNo int64, errStr string) RouteHandlerResult {
	return RouteHandlerResult{NewRouteError(code, errNo, errStr), nil, nil}
}
func (ctx *Context) MakeRouteHandlerResultAlert(code int, errNo int64, alert string) RouteHandlerResult {
	return ctx.MakeRouteHandlerResultCustom(func(innerCtx *Context) {
		sendErrorPayload(innerCtx, code, errNo, "", alert)
	})
}
func (ctx *Context) MakeRouteHandlerResultPayloads(payloads ...Payload) RouteHandlerResult {
	return RouteHandlerResult{nil, MakePayloadMapFromPayloads(payloads...), nil}
}
func (ctx *Context) MakeRouteHandlerResultGenericJSON(v interface{}) RouteHandlerResult {
	jsonBytes, err := json.Marshal(v)
	code := http.StatusOK
	if err != nil {
		code = http.StatusInternalServerError
		return RouteHandlerResult{NewRouteError(code, 3913952842, "Internal Server Error"), nil, nil}
	} else {
		return RouteHandlerResult{nil, nil, func(innerCtx *Context) {
			if rw, isResponseWriter := innerCtx.w.(http.ResponseWriter); isResponseWriter {
				rw.WriteHeader(code)
				if len(jsonBytes) == 0 {
					log.Println("jsonBytes", jsonBytes, innerCtx.Req.URL)
				}
				bytesWritten, err := rw.Write(jsonBytes)
				if err != nil {
					log.Println("3952513088 WRITE ERROR", err)
				}
				innerCtx.ContentLength = bytesWritten
			}
		}}
	}
}
func (ctx *Context) MakeRouteHandlerResultCustom(crr CustomRouteResponse) RouteHandlerResult {
	return RouteHandlerResult{nil, nil, crr}
}
func (ctx *Context) MakeRouteHandlerResultOk() RouteHandlerResult {
	return ctx.MakeRouteHandlerResultCustom(func(innerCtx *Context) {
		sendOkPayload(innerCtx)
	})
}
func (ctx *Context) MakeRouteHandlerResultNotFound(errNo int64) RouteHandlerResult {
	return ctx.MakeRouteHandlerResultCustom(func(innerCtx *Context) {
		sendNotFoundPayload(innerCtx, errNo)
	})
}

//  #####
// #     # #    #  ####  #####  ####  #    #
// #       #    # #        #   #    # ##  ##
// #       #    #  ####    #   #    # # ## #
// #       #    #      #   #   #    # #    #
// #     # #    # #    #   #   #    # #    #
//  #####   ####   ####    #    ####  #    #
//
// ######
// #     # ######  ####  #    # #      #####  ####
// #     # #      #      #    # #        #   #
// ######  #####   ####  #    # #        #    ####
// #   #   #           # #    # #        #        #
// #    #  #      #    # #    # #        #   #    #
// #     # ######  ####   ####  ######   #    ####
//

func (ctx *Context) WrapAndSendPayload(payload Payload) {
	if payload == nil {
		ctx.SendErrorPayload(http.StatusInternalServerError, 388359273, "")
		return
	}
	wrapAndSendPayload(ctx, payload)
}

// for a slice of Entities
func (ctx *Context) WrapAndSendPayloadsList(payloadList []Payload) {
	wrapAndSendPayloadsList(ctx, payloadList)
}

// for a map of slices of Entities
func (ctx *Context) WrapAndSendPayloadsMap(payloads PayloadsMap) {
	wrapAndSendPayloadsMap(ctx, payloads)
}

// Error
func (ctx *Context) SendErrorPayload(code int, errNo int64, errStr string) {
	enduserErrMsg := errStr
	if len(errStr) == 0 {
		enduserErrMsg = http.StatusText(code)
	}
	sendErrorPayload(ctx, code, errNo, enduserErrMsg, "")
}

// Alert payloads are designed as a general notification service for clients (ie client must upgrade, server is in maint mode, etc.)
func (ctx *Context) SendAlertPayload(code int, errNo int64, errStr, alert string) {
	sendErrorPayload(ctx, code, errNo, errStr, alert)
}

func (ctx *Context) DecodeResponseBodyOrSendError(pc PayloadController, payloadReference interface{}) interface{} {
	requestBody := ctx.Req.Body
	if requestBody == nil {
		errStr := BadRequestPrefix + ": Expected non-empty body"
		ctx.SendErrorPayload(http.StatusBadRequest, 3003399819, errStr)
		return nil
	}
	defer requestBody.Close()

	decoder := json.NewDecoder(requestBody)
	err := decoder.Decode(payloadReference)

	if err != nil {
		errStr := BadRequestPrefix + ": Cannot parse body"
		derr := deeperror.New(3005488054, errStr, err)
		log.Println("derr", derr)
		ctx.SendErrorPayload(http.StatusBadRequest, 3005488054, errStr)
		return nil
	}

	return payloadReference
}

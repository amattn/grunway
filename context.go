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

func (ctx *Context) MakeRouteHandlerResultError(code int, errNo int64, errMsg string) RouteHandlerResult {
	errInfo := ErrorInfo{
		ErrorNumber:  errNo,
		ErrorMessage: errMsg,
	}
	return ctx.MakeRouteHandlerResultErrorInfo(code, errInfo)
}
func (ctx *Context) MakeRouteHandlerResultDebugError(code int, errNo int64, errMsg string, debugNo int64, debugMsg string) RouteHandlerResult {
	errInfo := ErrorInfo{
		ErrorNumber:  errNo,
		ErrorMessage: errMsg,
		DebugNumber:  debugNo,
		DebugMessage: debugMsg,
	}
	return ctx.MakeRouteHandlerResultErrorInfo(code, errInfo)
}
func (ctx *Context) MakeRouteHandlerResultErrorInfo(code int, errInfo ErrorInfo) RouteHandlerResult {
	rerr := NewRouteError(code, errInfo)
	return RouteHandlerResult{rerr, nil, nil}
}

func (ctx *Context) MakeRouteHandlerResultAlert(code int, errNo int64, alert string) RouteHandlerResult {
	return ctx.MakeRouteHandlerResultCustom(func(innerCtx *Context) {
		sendErrorPayload(innerCtx, code, ErrorInfo{ErrorNumber: errNo}, alert)
	})
}
func (ctx *Context) MakeRouteHandlerResultPayloads(payloads ...Payload) RouteHandlerResult {
	return RouteHandlerResult{nil, MakePayloadMapFromPayloads(payloads...), nil}
}
func (ctx *Context) MakeRouteHandlerResultGenericJSON(v interface{}) RouteHandlerResult {
	jsonBytes, err := json.Marshal(v)
	code := http.StatusOK
	if err != nil {
		rerr := NewRouteError(
			http.StatusInternalServerError,
			ErrorInfo{
				ErrorNumber:  3913952842,
				ErrorMessage: "Internal Server Error",
			})
		return RouteHandlerResult{rerr, nil, nil}
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
		ctx.SendSimpleErrorPayload(http.StatusInternalServerError, 388359273, "")
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
func (ctx *Context) SendSimpleErrorPayload(code int, errNo int64, errorMsg string) {
	ctx.SendErrorInfoPayload(code, ErrorInfo{
		ErrorNumber:  errNo,
		ErrorMessage: errorMsg,
	})
}
func (ctx *Context) SendErrorInfoPayload(code int, errInfo ErrorInfo) {
	enduserErrMsg := errInfo.ErrorMessage
	if len(enduserErrMsg) == 0 {
		enduserErrMsg = http.StatusText(code)
	}
	sendErrorPayload(ctx, code, errInfo, "")
}

// Alert payloads are designed as a general notification service for clients (ie client must upgrade, server is in maint mode, etc.)
func (ctx *Context) SendSimpleAlertPayload(code int, errNo int64, errMsg, alert string) {
	sendErrorPayload(ctx, code, ErrorInfo{ErrorNumber: errNo, ErrorMessage: errMsg}, alert)
}

func (ctx *Context) DecodeResponseBodyOrSendError(pc PayloadController, payloadReference interface{}) interface{} {
	requestBody := ctx.Req.Body
	if requestBody == nil {
		errMsg := BadRequestPrefix + ": Expected non-empty body"
		ctx.SendSimpleErrorPayload(http.StatusBadRequest, 3003399819, errMsg)
		return nil
	}
	defer requestBody.Close()

	decoder := json.NewDecoder(requestBody)
	err := decoder.Decode(payloadReference)

	if err != nil {
		errMsg := BadRequestPrefix + ": Cannot parse body"
		derr := deeperror.New(3005488054, errMsg, err)
		log.Println("derr", derr)
		ctx.SendSimpleErrorPayload(http.StatusBadRequest, derr.Num, errMsg)
		return nil
	}

	return payloadReference
}

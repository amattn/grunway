package grunway

import (
	"encoding/json"
	"github.com/amattn/deeperror"
	"log"
	"net/http"
)

type Context struct {
	R *http.Request
	w http.ResponseWriter
	E Endpoint

	// router
	router *Router

	// only populated after auth
	PublicKey string // for Auth'd requests, will be set to public key if Auth was successful, "" otherwise

	// generic maps for middleware to stuff arbitrary data
	middleware map[string]interface{}
	postware   map[string]interface{}

	// only populated after a write
	written       bool
	StatusCode    int
	ContentLength int
}

func (c *Context) AddHeader(key, value string) {
	c.w.Header().Add(key, value)
}
func (c *Context) DelHeader(key string) {
	c.w.Header().Del(key)
}
func (c *Context) GetHeader(key string) string {
	return c.w.Header().Get(key)
}
func (c *Context) SetHeader(key, value string) {
	c.w.Header().Set(key, value)
}

// conveninece wrapper, designed to generate return values inside a Route Handler
func (c *Context) ReturnRouteError(code int, errNo int64, errStr string) (*RouteError, PayloadsMap, CustomRouteResponse) {
	return NewRouteError(code, errNo, errStr), nil, nil
}

// conveninece wrapper, designed to generate return values inside a Route Handler
func (c *Context) ReturnPayload(payload Payload) (*RouteError, PayloadsMap, CustomRouteResponse) {
	return nil, MakePayloadMapFromPayload(payload), nil
}

// conveninece wrapper, designed to generate return values inside a Route Handler
func (c *Context) ReturnCustom(crh CustomRouteResponse) (*RouteError, PayloadsMap, CustomRouteResponse) {
	return nil, nil, crh
}

func (c *Context) ReturnOK() (*RouteError, PayloadsMap, CustomRouteResponse) {
	return c.ReturnCustom(func(ctx *Context) {
		sendOkPayload(ctx)
	})
}

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
func (c *Context) SendErrorPayload(code int, errNo int64, errStr string) {
	enduserErrMsg := errStr
	if len(errStr) == 0 {
		enduserErrMsg = http.StatusText(code)
	}
	sendErrorPayload(c, code, errNo, enduserErrMsg, "")
}

// Alert payloads are designed as a general notification service for clients (ie client must upgrade, server is in maint mode, etc.)
func (c *Context) SendAlertPayload(code int, errNo int64, errStr, alert string) {
	sendErrorPayload(c, code, errNo, errStr, alert)
}

func (c *Context) DecodeResponseBodyOrSendError(pc PayloadController, payloadReference interface{}) interface{} {
	requestBody := c.R.Body
	if requestBody == nil {
		errStr := BadRequestPrefix + ": Expected non-empty body"
		c.SendErrorPayload(http.StatusBadRequest, 3003399819, errStr)
		return nil
	}
	defer requestBody.Close()

	decoder := json.NewDecoder(requestBody)
	err := decoder.Decode(payloadReference)

	if err != nil {
		errStr := BadRequestPrefix + ": Cannot parse body"
		derr := deeperror.New(3005488054, errStr, err)
		log.Println("derr", derr)
		c.SendErrorPayload(http.StatusBadRequest, 3005488054, errStr)
		return nil
	}

	return payloadReference
}

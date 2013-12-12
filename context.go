package grunway

import (
	"encoding/json"
	"github.com/amattn/deeperror"
	"log"
	"net/http"
)

type Context struct {
	w       http.ResponseWriter
	written bool
	R       *http.Request
	E       Endpoint

	PublicKey string // for Auth'd requests, will be set to public key if Auth was successful, "" otherwise

	middleware map[string]interface{}
}

func (c *Context) Add(key, value string) {
	c.w.Header().Add(key, value)
}
func (c *Context) Del(key string) {
	c.w.Header().Del(key)
}
func (c *Context) Get(key string) string {
	return c.w.Header().Get(key)
}
func (c *Context) Set(key, value string) {
	c.w.Header().Set(key, value)
}

func (c *Context) WrapAndSendPayload(payload interface{}) {
	wrapAndSendPayload(c, payload)
}

// for a slice of Entities
func (c *Context) WrapAndSendPayloadList(payloadList []interface{}) {
	wrapAndSendPayloadList(c, payloadList)
}

// Error
func (c *Context) SendErrorPayload(code int, errNo int64, errStr string) {
	sendErrorPayload(c, code, errNo, errStr, "")
}

// Alert payloads are designed as a general notification service for clients (ie client must upgrade, server is in maint mode, etc.)
func (c *Context) SendAlertPayload(code int, errNo int64, errStr, alert string) {
	sendErrorPayload(c, code, errNo, errStr, alert)
}

// Ok payload is just a json dict w/ one kv: errNo == 0
func (c *Context) SendOkPayload() {
	sendOkPayload(c)
}

func (c *Context) DecodeResponseBodyOrSendError(pc PayloadController, payloadReference interface{}) interface{} {
	requestBody := c.R.Body
	if requestBody == nil {
		errStr := BadRequestPrefix + ": Expected non-empty body"
		c.SendErrorPayload(http.StatusBadRequest, 4003399819, errStr)
		return nil
	}
	defer requestBody.Close()

	decoder := json.NewDecoder(requestBody)
	err := decoder.Decode(payloadReference)

	if err != nil {
		errStr := BadRequestPrefix + ": Cannot parse body"
		derr := deeperror.New(4005488054, errStr, err)
		log.Println("derr", derr)
		c.SendErrorPayload(http.StatusBadRequest, 4005488054, errStr)
		return nil
	}

	return payloadReference
}

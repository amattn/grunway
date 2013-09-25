package grunway

import (
	"encoding/json"
	"github.com/amattn/deeperror"
	"log"
	"net/http"
)

type Context struct {
	W http.ResponseWriter
	R *http.Request
	E Endpoint
}

func (c *Context) WrapAndSendPayload(payload interface{}) {
	wrapAndSendPayload(c.W, payload)
}

// for a slice of Entities
func (c *Context) WrapAndSendPayloadList(payloadList []interface{}) {
	wrapAndSendPayloadList(c.W, payloadList)
}

// Error
func (c *Context) SendErrorPayload(code int, errNo int64, errStr string) {
	sendErrorPayload(c.W, code, errNo, errStr, "")
}

// Alert payloads are designed as a general notification service for clients (ie client must upgrade, server is in maint mode, etc.)
func (c *Context) SendAlertPayload(code int, errNo int64, errStr, alert string) {
	sendErrorPayload(c.W, code, errNo, errStr, alert)
}

// Ok payload is just a json dict w/ one kv: errNo == 0
func (c *Context) SendOkPayload() {
	sendOkPayload(c.W)
}

func (c *Context) DecodeResponseBodyOrSendError(pc PayloadController) interface{} {
	requestBody := c.R.Body
	if requestBody == nil {
		errStr := BadRequestPrefix + ": Expected non-empty body"
		c.SendErrorPayload(http.StatusBadRequest, 4003399819, errStr)
		return nil
	}
	defer requestBody.Close()

	decoder := json.NewDecoder(requestBody)
	payloadReference := pc.NewPaylodReference()
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

package grunway

import (
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

// Error or alerts
func (c *Context) SendErrorPayload(code int, errNo int64, errStr, alert string) {
	sendErrorPayload(c.W, code, errNo, errStr, alert)
}

// Ok payload is just a json dict w/ one kv: errNo == 0
func (c *Context) SendOkPayload() {
	sendOkPayload(c.W)
}

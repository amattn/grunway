package grunway

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/amattn/deeperror"
)

// This will typically be serialized into a JSON formatted string
type PayloadWrapper struct {
	PayloadType string        `json:",omitempty"` // optional, typically used for sanity checking
	PayloadList []interface{} `json:",omitempty"` // ALWAYS an array of objects. Typically, these are arrays of objects designed to be deserialized into entity structs (eg []BookPayload, []AuthorPayload)
	ErrNo       int64         `json:",omitempty"` // will be 0 on successful responses, non-zero otherwise
	ErrStr      string        `json:",omitempty"` // end-user appropriate error message
	Alert       string        `json:",omitempty"` // used when the client end user needs to be alerted of something: (eg, maintenance mode, downtime, sercurity, required update, etc.)
}

// If a payload implements this method, then the wrapper will autopopulate the PayloadType field
type TypedPayload interface {
	PayloadType() string
}

func NewPayloadWrapper() *PayloadWrapper {
	// needs a leaky bucket
	return new(PayloadWrapper)
}

// for a single Enitity
func wrapAndSendPayload(ctx *Context, payload interface{}) {
	wrapAndSendPayloadList(ctx, []interface{}{payload})
}

// for a slice of Entities
func wrapAndSendPayloadList(ctx *Context, payloadList []interface{}) {
	payloadWrapper := NewPayloadWrapper()

	var payloadType string

	for _, payload := range payloadList {
		typedPayload, isTP := payload.(TypedPayload)
		if isTP {
			if payloadType == "" {
				payloadType = typedPayload.PayloadType()
			}
			if payloadType != typedPayload.PayloadType() {
				payloadType = ""
				break
			}
		}
	}

	payloadWrapper.PayloadType = payloadType
	payloadWrapper.PayloadList = payloadList
	sendPayloadWrapper(ctx, http.StatusOK, payloadWrapper)
}

// Error or alerts
func sendErrorPayload(ctx *Context, code int, errNo int64, errStr, alert string) {
	payloadWrapper := NewPayloadWrapper()
	payloadWrapper.ErrNo = errNo
	payloadWrapper.ErrStr = errStr
	payloadWrapper.Alert = alert

	if errNo != 0 {
		ctx.Add("X-ErrorNum", fmt.Sprintf("%d", errNo))
	}
	if len(errStr) > 1 {
		ctx.Add("X-ErrorStr", fmt.Sprintf("%s", errStr))
	}
	if len(alert) > 1 {
		ctx.Add("X-Alert", fmt.Sprintf("%s", alert))
	}

	sendPayloadWrapper(ctx, code, payloadWrapper)
}

// Ok payloadWrapper is just a json dict w/ one kv: ErrNo == 0
func sendOkPayload(ctx *Context) {
	payloadWrapper := NewPayloadWrapper()
	sendPayloadWrapper(ctx, http.StatusOK, payloadWrapper)
}

func sendPayloadWrapper(ctx *Context, code int, payloadWrapper *PayloadWrapper) {
	if ctx.written == true {
		derr := deeperror.New(3314606687, "ERROR attempt to write multiple times to same writer", nil)
		log.Println(derr)
		return
	}

	ctx.written = true

	ctx.Add("Content-Type", "application/json")

	if code != http.StatusOK {
		if rw, isResponseWriter := ctx.w.(http.ResponseWriter); isResponseWriter {
			rw.WriteHeader(code)
		}
	}

	enc := json.NewEncoder(ctx.w)
	jsonErr := enc.Encode(payloadWrapper)
	if jsonErr != nil {

		derr := deeperror.NewHTTPError(3589720731, "Unexpeced error encoding json", jsonErr, http.StatusInternalServerError)
		responseWriter, ok := ctx.w.(http.ResponseWriter)
		if ok {
			errStr := fmt.Sprintln(derr.Num, derr.EndUserMsg)
			http.Error(responseWriter, errStr, derr.StatusCode)
		}
		log.Println(derr)
	}
}

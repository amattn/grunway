package grunway

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/amattn/deeperror"
)

// This will typically be serialized into a JSON formatted string
type PayloadWrapper struct {
	PayloadType string        `json:",omitempty"` // optional, typically used for sanity checking
	PayloadList []interface{} `json:",omitempty"` // ALWAYS an array of objects. Typically, these are arrays of objects designed to be deserialized into entity structs (eg []BookPayload, []AuthorPayload)
	ErrNo       int64         // will be 0 on successful responses, non-zero otherwise
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
func wrapAndSendPayload(w io.Writer, payload interface{}) {
	wrapAndSendPayloadList(w, []interface{}{payload})
}

// for a slice of Entities
func wrapAndSendPayloadList(w io.Writer, payloadList []interface{}) {
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
	sendPayloadWrapper(w, http.StatusOK, payloadWrapper)
}

// Error or alerts
func sendErrorPayload(w io.Writer, code int, errNo int64, errStr, alert string) {
	payloadWrapper := NewPayloadWrapper()
	payloadWrapper.ErrNo = errNo
	payloadWrapper.ErrStr = errStr
	payloadWrapper.Alert = alert
	sendPayloadWrapper(w, code, payloadWrapper)
}

// Ok payloadWrapper is just a json dict w/ one kv: ErrNo == 0
func sendOkPayload(w io.Writer) {
	payloadWrapper := NewPayloadWrapper()
	sendPayloadWrapper(w, http.StatusOK, payloadWrapper)
}

func sendPayloadWrapper(w io.Writer, code int, payloadWrapper *PayloadWrapper) {
	if code != http.StatusOK {
		if rw, isResponseWriter := w.(http.ResponseWriter); isResponseWriter {
			rw.WriteHeader(code)
		}
	}

	enc := json.NewEncoder(w)
	jsonErr := enc.Encode(payloadWrapper)
	if jsonErr != nil {

		derr := deeperror.NewHTTPError(1589720731, "Unexpeced error encoding json", jsonErr, http.StatusInternalServerError)
		responseWriter, ok := w.(http.ResponseWriter)
		if ok {
			errStr := fmt.Sprintln(derr.Num, derr.EndUserMsg)
			http.Error(responseWriter, errStr, derr.StatusCode)
		}
		log.Println(derr)
	}
}

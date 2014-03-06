package grunway

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"

	"github.com/amattn/deeperror"
)

// If a payload implements this method, then the wrapper will autopopulate the PayloadType field
type Payload interface {
	PayloadType() string
}

// ALWAYS a map of array of objects.
// key is type, value is list of payloads of that type
// Typically, these are arrays of objects designed to be deserialized into entity structs (eg []BookPayload, []AuthorPayload)
type PayloadsMap map[string][]Payload

// This will typically be serialized into a JSON formatted string
type PayloadWrapper struct {
	Payloads PayloadsMap `json:",omitempty"` // key is type, value is list of payloads of that type
	ErrNo    int64       `json:",omitempty"` // will be 0 on successful responses, non-zero otherwise
	ErrStr   string      `json:",omitempty"` // end-user appropriate error message
	Alert    string      `json:",omitempty"` // used when the client end user needs to be alerted of something: (eg, maintenance mode, downtime, sercurity, required update, etc.)
}

//  #####
// #     # #####  ######   ##   ##### ######
// #       #    # #       #  #    #   #
// #       #    # #####  #    #   #   #####
// #       #####  #      ######   #   #
// #     # #   #  #      #    #   #   #
//  #####  #    # ###### #    #   #   ######
//

func NewPayloadWrapper() *PayloadWrapper {
	// needs a leaky bucket
	return new(PayloadWrapper)
}

func MakePayloadMapFromPayload(payload Payload) PayloadsMap {
	return MakePayloadMapFromPayloadsList([]Payload{payload})
}

func MakePayloadMapFromPayloads(payloads ...Payload) PayloadsMap {
	return MakePayloadMapFromPayloadsList(payloads)
}
func MakePayloadMapFromPayloadsList(payloadsList []Payload) PayloadsMap {
	payloads := make(PayloadsMap)

	for _, payload := range payloadsList {
		ptype := payload.PayloadType()
		plist, exists := payloads[ptype]
		if exists == true {
			plist = append(plist, payload)
		} else {
			payloads[ptype] = []Payload{payload}
		}
	}
	return payloads
}

//  #####
// #     # ###### #    # #####
// #       #      ##   # #    #
//  #####  #####  # #  # #    #
//       # #      #  # # #    #
// #     # #      #   ## #    #
//  #####  ###### #    # #####
//

// for a single Enitity
func wrapAndSendPayload(ctx *Context, payload Payload) {
	wrapAndSendPayloadsList(ctx, []Payload{payload})
}

// for a slice of Entities
func wrapAndSendPayloadsList(ctx *Context, payloadsList []Payload) {
	wrapAndSendPayloadsMap(ctx, MakePayloadMapFromPayloadsList(payloadsList))
}

func wrapAndSendPayloadsMap(ctx *Context, pldsMap PayloadsMap) {
	payloadWrapper := NewPayloadWrapper()
	payloadWrapper.Payloads = pldsMap
	writePayloadWrapper(ctx, http.StatusOK, payloadWrapper)
}

// Error or alerts
func sendErrorPayload(ctx *Context, code int, errNo int64, errStr, alert string) {
	payloadWrapper := NewPayloadWrapper()
	payloadWrapper.ErrNo = errNo
	payloadWrapper.ErrStr = errStr
	payloadWrapper.Alert = alert

	if errNo != 0 {
		ctx.AddHeader("X-ErrorNum", fmt.Sprintf("%d", errNo))
	}
	if len(errStr) > 1 {
		ctx.AddHeader("X-ErrorStr", fmt.Sprintf("%s", errStr))
	}
	if len(alert) > 1 {
		ctx.AddHeader("X-Alert", fmt.Sprintf("%s", alert))
	}

	writePayloadWrapper(ctx, code, payloadWrapper)
}

// Ok payloadWrapper is just a json dict w/ one kv: ErrNo == 0
func sendOkPayload(ctx *Context) {
	payloadWrapper := NewPayloadWrapper()
	writePayloadWrapper(ctx, http.StatusOK, payloadWrapper)
}

// All output goes through here.
func writePayloadWrapper(ctx *Context, code int, payloadWrapper *PayloadWrapper) {
	if ctx.written == true {
		derr := deeperror.New(3314606687, "ERROR attempt to write multiple times to same writer", nil)
		log.Println(derr)
		return
	}

	ctx.written = true
	ctx.StatusCode = code
	ctx.SetHeader(httpHeaderContentType, httpHeaderContentTypeJSON)

	// This is the old way.  it doesn't give us status info.
	// enc := json.NewEncoder(ctx.w)
	// jsonErr := enc.Encode(payloadWrapper)

	// two ways to get status info, a counting buffer or just encode to bytes first.
	// counting buffer is more efficient, but I don't need the docs to implement the eoncode to bytes.
	// consider a counting buffer if memory usage or garbage collection pauses start hurting.

	jsonBytes, jsonErr := json.Marshal(payloadWrapper)

	if jsonErr != nil {
		derr := deeperror.NewHTTPError(3589720731, "Fatal Internal Output Error", jsonErr, http.StatusInternalServerError)
		responseWriter, ok := ctx.w.(http.ResponseWriter)
		if ok {
			ctx.AddHeader("X-ErrorNum", fmt.Sprintf("%d", derr.Num))
			ctx.AddHeader("X-ErrorStr", fmt.Sprintf("%s", derr.EndUserMsg))
			responseWriter.WriteHeader(derr.StatusCode)
			fmt.Fprintf(responseWriter, "{\"ErrNo\":%d,\"ErrStr\":%s}", derr.Num, derr.EndUserMsg)
		}
		log.Println(derr)
	} else {
		// At this point, everything is a-ok...  just write out.
		if rw, isResponseWriter := ctx.w.(http.ResponseWriter); isResponseWriter {
			rw.WriteHeader(code)
			if len(jsonBytes) == 0 {
				log.Println("jsonBytes", jsonBytes, ctx.R.URL)
			}
			bytesWritten, err := rw.Write(jsonBytes)
			if err != nil {
				log.Println("3952513088 WRITE ERROR", err)
			}
			ctx.ContentLength = bytesWritten
		}
	}

	for _, postproc := range ctx.router.PostProcessors {
		postproc.Process(ctx)
	}
}

// ######
// #     #   ##   #####   ####  ######
// #     #  #  #  #    # #      #
// ######  #    # #    #  ####  #####
// #       ###### #####       # #
// #       #    # #   #  #    # #
// #       #    # #    #  ####  ######
//

type unmarshallingPayloadWrapper struct {
	Payloads map[string][]json.RawMessage `json:",omitempty"` // key is type, value is list of payloads of that type
	ErrNo    int64                        `json:",omitempty"` // will be 0 on successful responses, non-zero otherwise
	ErrStr   string                       `json:",omitempty"` // end-user appropriate error message
	Alert    string                       `json:",omitempty"` // used when the client end user needs to be alerted of something: (eg, maintenance mode, downtime, sercurity, required update, etc.)
}

func UnmarshalPayloadWrapper(jsonBytes []byte, supportedPayloads ...Payload) (*PayloadWrapper, error) {

	if len(supportedPayloads) == 0 {
		return nil, deeperror.New(2911466683, "must supply supportedPayloads", nil)
	}
	var upw unmarshallingPayloadWrapper
	var pw PayloadWrapper
	err := json.Unmarshal(jsonBytes, &upw)
	if err != nil {
		return nil, deeperror.New(3679523795, "Parse Error, Unexpectd format", err)
	}

	// do the easy stuff first
	pw.ErrNo = upw.ErrNo
	pw.ErrStr = upw.ErrStr
	pw.Alert = upw.Alert
	pw.Payloads = make(PayloadsMap)

	payloadTypeReflecMap := make(map[string]reflect.Type)
	for _, payload := range supportedPayloads {
		payloadTypeReflecMap[payload.PayloadType()] = reflect.TypeOf(payload)
	}

	for pTypeString, rawJsonList := range upw.Payloads {
		pTypeReflect := payloadTypeReflecMap[pTypeString]
		payloadList := make([]Payload, 0, len(rawJsonList))
		for _, rawJson := range rawJsonList {
			pReflectValue := reflect.New(pTypeReflect)
			structPointer := pReflectValue.Interface()
			pld, ok := structPointer.(Payload)
			if ok == false {
				return nil, deeperror.New(3679523796, "Parse Error, Unexpectd payload", err)
			}

			err := json.Unmarshal(rawJson, pld)
			if err != nil {
				return nil, deeperror.New(3679523797, "Parse Error, Unexpectd payload", err)
			}

			payloadList = append(payloadList, pld)
		}

		pw.Payloads[pTypeString] = payloadList
	}

	return &pw, nil
}

package grunway

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/amattn/deeperror"
)

// import (
// 	"net/http"
// )

// This will typically be serialized into a JSON formatted string
type Payload struct {
	Entities []interface{} `json:",omitempty"` // ALWAYS an array of something. Typically arrays of entity structs (eg []BookPayload, []AuthorPayload)
	ErrNum   int64         // will be 0 on successful responses.
	ErrStr   string        `json:",omitempty"` // end-user appropriate error message
	Alert    string        `json:",omitempty"` // used when the client end user needs to be alerted of something: (eg, maintenance mode, downtime, sercurity, required update, etc.)
}

func NewPayload() *Payload {
	// needs a leaky bucket
	return new(Payload)
}

// for a single Enitity
func writeEntityPayload(w io.Writer, entity interface{}) {
	writeEntitiesPayload(w, []interface{}{entity})
}

// for a slice of Entities
func writeEntitiesPayload(w io.Writer, entities []interface{}) {
	payload := NewPayload()
	payload.Entities = entities
	writePayload(w, payload)
}

// Error or alerts
func writeErrorPayload(w io.Writer, errNum int64, errStr, alert string) {
	payload := NewPayload()
	payload.ErrNum = errNum
	payload.ErrStr = errStr
	payload.Alert = alert
	writePayload(w, payload)
}

// Ok payload is just a json dict w/ one kv: ErrNum == 0
func writeOkPayload(w io.Writer) {
	payload := NewPayload()
	writePayload(w, payload)
}

func writePayload(w io.Writer, payload *Payload) {
	enc := json.NewEncoder(w)
	jsonErr := enc.Encode(payload)
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

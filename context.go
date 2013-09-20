package grunway

import (
	"net/http"
)

type Context struct {
	W http.ResponseWriter
	R *http.Request
	E Endpoint
}

func (c *Context) RespondWithEntityPayload(entity interface{}) {

}

func (c *Context) WriteEntityPayload(entity interface{}) {
	writeEntityPayload(c.W, entity)
}

// for a slice of Entities
func (c *Context) WriteEntitiesPayload(entities []interface{}) {
	writeEntitiesPayload(c.W, entities)
}

// Error or alerts
func (c *Context) WriteErrorPayload(errNum int64, errStr, alert string) {
	writeErrorPayload(c.W, errNum, errStr, alert)
}

// Ok payload is just a json dict w/ one kv: ErrNum == 0
func (c *Context) WriteOkPayload() {
	writeOkPayload(c.W)
}

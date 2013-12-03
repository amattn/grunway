package grunway

import (
	"encoding/json"
	"net/http"
)

type CreatePayload interface {
	PayloadStore(controller PayloadController, c *Context) (didSucceed bool, errNo int64, responsePayload interface{})
}
type CreatePayloadValidator interface {
	PayloadIsValid(controller PayloadController, c *Context) (isValid bool, errNo int64)
}

func StandardCreateHandler(controller PayloadController, c *Context, createPayload CreatePayload) {
	// Get the request
	requestBody := c.R.Body
	if requestBody == nil {
		http.Error(c.W, "400 Bad Request: Expected non-empty body", http.StatusBadRequest)
		return
	}
	defer requestBody.Close()

	// parse the json
	decoder := json.NewDecoder(requestBody)
	err := decoder.Decode(createPayload)

	if err != nil {
		errStr := BadRequestPrefix + ": Cannot parse body"
		c.SendErrorPayload(http.StatusBadRequest, 199946555, errStr)
		return
	}

	// validate json
	validator, isValidator := createPayload.(CreatePayloadValidator)
	if isValidator {
		isValid, errNo := validator.PayloadIsValid(controller, c)
		if isValid == false {
			c.SendErrorPayload(http.StatusBadRequest, errNo, BadRequestPrefix)
			return
		}
	}

	// add entity to the model
	didSucceed, errNo, responsePayload := createPayload.PayloadStore(controller, c)
	if didSucceed == false {
		c.SendErrorPayload(http.StatusInternalServerError, errNo, InternalServerErrorPrefix)
		return
	}

	// response
	if responsePayload != nil {
		c.WrapAndSendPayload(responsePayload)
	} else {
		c.SendOkPayload()
	}
}

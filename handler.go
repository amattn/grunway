package grunway

import (
	"encoding/json"
	"net/http"

	"github.com/amattn/deeperror"
)

type CreatePerformer interface {
	PerformCreate(c *Context, createRequestPayload interface{}) (didSucceed bool, derr *deeperror.DeepError, responsePayload interface{})
}
type CreateValidator interface {
	CreatePayloadIsValid(c *Context, createRequestPayload interface{}) (isValid bool, errNo int64)
}

func StandardCreateHandler(controller CreatePerformer, c *Context, createRequestPayload interface{}) {
	// Get the request
	requestBody := c.R.Body
	if requestBody == nil {
		c.SendErrorPayload(http.StatusBadRequest, 3370318075, BadRequestPrefix+"Expected non-empty body")
		return
	}
	defer requestBody.Close()

	if c.E.PrimaryKey != 0 {
		c.SendErrorPayload(http.StatusBadRequest, BadRequestExtraneousPrimaryKeyErrNo, BadRequestSyntaxErrorPrefix+" Cannot set primary key")
		return
	}

	// parse the json
	decoder := json.NewDecoder(requestBody)
	err := decoder.Decode(createRequestPayload)

	if err != nil {
		errStr := BadRequestPrefix + ": Cannot parse body"
		c.SendErrorPayload(http.StatusBadRequest, 3540227685, errStr)
		return
	}

	// validate json
	validator, isValidator := controller.(CreateValidator)
	if isValidator {
		isValid, errNo := validator.CreatePayloadIsValid(c, createRequestPayload)
		if isValid == false {
			c.SendErrorPayload(http.StatusBadRequest, errNo, BadRequestPrefix)
			return
		}
	}

	// add entity to the model
	didSucceed, derr, responsePayload := controller.PerformCreate(c, createRequestPayload)
	if didSucceed == false {
		if derr != nil {
			if derr.StatusCode > 299 {
				c.SendErrorPayload(derr.StatusCode, derr.Num, derr.EndUserMsg)
			} else {
				c.SendErrorPayload(http.StatusInternalServerError, derr.Num, InternalServerErrorPrefix)
			}
		} else {
			c.SendErrorPayload(http.StatusInternalServerError, 3249936084, InternalServerErrorPrefix)
		}
		return
	}

	// response
	if responsePayload != nil {
		c.WrapAndSendPayload(responsePayload)
	} else {
		c.SendOkPayload()
	}
}

// #     #
// #     # #####  #####    ##   ##### ######
// #     # #    # #    #  #  #    #   #
// #     # #    # #    # #    #   #   #####
// #     # #####  #    # ######   #   #
// #     # #      #    # #    #   #   #
//  #####  #      #####  #    #   #   ######
//

type UpdatePerformer interface {
	PerformUpdate(c *Context, updateRequestPayload interface{}) (didSucceed bool, errNo int64, responsePayload interface{})
}
type UpdateValidator interface {
	UpdatePayloadIsValid(c *Context, updateRequestPayload interface{}) (isValid bool, errNo int64)
}

func StandardUpdateHandler(controller UpdatePerformer, c *Context, updateRequestPayload interface{}) {
	// Get the request
	requestBody := c.R.Body
	if requestBody == nil {
		c.SendErrorPayload(http.StatusBadRequest, 3851489100, BadRequestPrefix+"Expected non-empty body")
		return
	}
	defer requestBody.Close()

	// parse the json
	decoder := json.NewDecoder(requestBody)
	err := decoder.Decode(updateRequestPayload)

	if err != nil {
		errStr := BadRequestPrefix + ": Cannot parse body"
		c.SendErrorPayload(http.StatusBadRequest, 1858602328, errStr)
		return
	}

	// validate json
	validator, isValidator := controller.(UpdateValidator)
	if isValidator {
		isValid, errNo := validator.UpdatePayloadIsValid(c, updateRequestPayload)
		if isValid == false {
			c.SendErrorPayload(http.StatusBadRequest, errNo, BadRequestPrefix)
			return
		}
	}

	// add entity to the model
	didSucceed, errNo, responsePayload := controller.PerformUpdate(c, updateRequestPayload)
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

// ######
// #     # ###### #      ###### ##### ######
// #     # #      #      #        #   #
// #     # #####  #      #####    #   #####
// #     # #      #      #        #   #
// #     # #      #      #        #   #
// ######  ###### ###### ######   #   ######
//

type DeletePerformer interface {
	PerformDelete(c *Context) (didSucceed bool, errNo int64)
}
type DeleteValidator interface {
	DeleteRequestIsValid(c *Context) (isValid bool, errNo int64)
}

func StandardDeleteHandler(c *Context, controller DeletePerformer) {
	if c.E.PrimaryKey <= 0 {
		c.SendErrorPayload(http.StatusBadRequest, BadRequestMissingPrimaryKeyErrNo, BadRequestPrefix)
		return
	}

	// validate json
	validator, isValidator := controller.(DeleteValidator)
	if isValidator {
		isValid, errNo := validator.DeleteRequestIsValid(c)
		if isValid == false {
			c.SendErrorPayload(http.StatusBadRequest, errNo, BadRequestPrefix)
			return
		}
	}

	// delete entity
	didSucceed, errNo := controller.PerformDelete(c)
	if didSucceed == false {
		c.SendErrorPayload(http.StatusInternalServerError, errNo, InternalServerErrorPrefix)
		return
	}

	c.SendOkPayload()
}

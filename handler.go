package grunway

import (
	"encoding/json"
	"net/http"

	"github.com/amattn/deeperror"
)

type CreatePerformer interface {
	CreatePayloadIsValid(c *Context, createRequestPayload Payload) *deeperror.DeepError
	PerformCreate(c *Context, createRequestPayload Payload) (responsePayload Payload, derr *deeperror.DeepError)
}

func StandardCreateHandler(controller CreatePerformer, c *Context, createRequestPayload Payload) (*RouteError, PayloadsMap, CustomRouteResponse) {
	// Get the request
	requestBody := c.R.Body
	if requestBody == nil {
		return c.ReturnRouteError(http.StatusBadRequest, 3370318075, BadRequestPrefix+"Expected non-empty body")
	}
	defer requestBody.Close()

	if c.E.PrimaryKey != 0 {
		return c.ReturnRouteError(http.StatusBadRequest, BadRequestExtraneousPrimaryKeyErrNo, BadRequestSyntaxErrorPrefix+" Cannot set primary key")
	}

	// parse the json
	decoder := json.NewDecoder(requestBody)
	err := decoder.Decode(createRequestPayload)

	if err != nil {
		errStr := BadRequestPrefix + ": Cannot parse body"
		return c.ReturnRouteError(http.StatusBadRequest, 3540227685, errStr)
	}

	// validate json
	derr := controller.CreatePayloadIsValid(c, createRequestPayload)
	if derr != nil {
		// log.Println("CreatePayloadIsValid returned derr", derr)
		if derr.StatusCodeIsDefaultValue() {
			return c.ReturnRouteError(http.StatusBadRequest, derr.Num, derr.EndUserMsg)
		} else {
			return c.ReturnRouteError(derr.StatusCode, derr.Num, derr.EndUserMsg)
		}
	}

	// add entity to the model
	responsePayload, derr := controller.PerformCreate(c, createRequestPayload)
	if derr != nil {
		if derr.StatusCodeIsDefaultValue() {
			return c.ReturnRouteError(http.StatusInternalServerError, derr.Num, derr.EndUserMsg)
		} else {
			return c.ReturnRouteError(derr.StatusCode, derr.Num, derr.EndUserMsg)
		}
	}

	// response
	if responsePayload != nil {
		return c.ReturnPayload(responsePayload)
	} else {
		return c.ReturnOK()
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
	UpdatePayloadIsValid(c *Context, updateRequestPayload Payload) *deeperror.DeepError
	PerformUpdate(c *Context, updateRequestPayload Payload) (responsePayload Payload, derr *deeperror.DeepError)
}

func StandardUpdateHandler(controller UpdatePerformer, ctx *Context, updateRequestPayload Payload) {
	// Get the request
	requestBody := ctx.R.Body
	if requestBody == nil {
		ctx.SendErrorPayload(http.StatusBadRequest, 3851489100, BadRequestPrefix+"Expected non-empty body")
		return
	}
	defer requestBody.Close()

	// parse the json
	decoder := json.NewDecoder(requestBody)
	err := decoder.Decode(updateRequestPayload)

	if err != nil {
		errStr := BadRequestPrefix + ": Cannot parse body"
		ctx.SendErrorPayload(http.StatusBadRequest, 1858602328, errStr)
		return
	}

	// validate json
	derr := controller.UpdatePayloadIsValid(ctx, updateRequestPayload)
	if derr != nil {
		if derr.StatusCodeIsDefaultValue() {
			ctx.SendErrorPayload(http.StatusBadRequest, derr.Num, derr.EndUserMsg)
		} else {
			ctx.SendErrorPayload(derr.StatusCode, derr.Num, derr.EndUserMsg)
		}
		return
	}

	// add entity to the model
	responsePayload, derr := controller.PerformUpdate(ctx, updateRequestPayload)
	if derr != nil {
		if derr.StatusCodeIsDefaultValue() {
			ctx.SendErrorPayload(http.StatusInternalServerError, derr.Num, derr.EndUserMsg)
		} else {
			ctx.SendErrorPayload(derr.StatusCode, derr.Num, derr.EndUserMsg)
		}
		return
	}

	// response
	if responsePayload != nil {
		ctx.WrapAndSendPayload(responsePayload)
	} else {
		sendOkPayload(ctx)
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
	PerformDelete(ctx *Context) (didSucceed bool, errNo int64)
}
type DeleteValidator interface {
	DeleteRequestIsValid(ctx *Context) (isValid bool, errNo int64)
}

func StandardDeleteHandler(ctx *Context, controller DeletePerformer) {
	if ctx.E.PrimaryKey <= 0 {
		ctx.SendErrorPayload(http.StatusBadRequest, BadRequestMissingPrimaryKeyErrNo, BadRequestPrefix)
		return
	}

	// validate json
	validator, isValidator := controller.(DeleteValidator)
	if isValidator {
		isValid, errNo := validator.DeleteRequestIsValid(ctx)
		if isValid == false {
			ctx.SendErrorPayload(http.StatusBadRequest, errNo, BadRequestPrefix)
			return
		}
	}

	// delete entity
	didSucceed, errNo := controller.PerformDelete(ctx)
	if didSucceed == false {
		ctx.SendErrorPayload(http.StatusInternalServerError, errNo, InternalServerErrorPrefix)
		return
	}

	sendOkPayload(ctx)
}

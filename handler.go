package grunway

import (
	"encoding/json"
	"net/http"

	"github.com/amattn/deeperror"
)

type CreatePerformer interface {
	CreatePayloadIsValid(ctx *Context, createRequestPayload Payload) *deeperror.DeepError
	PerformCreate(ctx *Context, createRequestPayload Payload) (responsePayload Payload, derr *deeperror.DeepError)
}

func StandardCreateHandler(controller CreatePerformer, ctx *Context, createRequestPayload Payload) RouteHandlerResult {
	// Get the request
	requestBody := ctx.Req.Body
	if requestBody == nil {
		return ctx.MakeRouteHandlerResultError(http.StatusBadRequest, 3370318075, BadRequestPrefix+"Expected non-empty body")
	}
	defer requestBody.Close()

	if ctx.End.PrimaryKey != 0 {
		return ctx.MakeRouteHandlerResultError(http.StatusBadRequest, BadRequestExtraneousPrimaryKeyErrorNumber, BadRequestSyntaxErrorPrefix+" Cannot set primary key")
	}

	// parse the json
	decoder := json.NewDecoder(requestBody)
	err := decoder.Decode(createRequestPayload)

	if err != nil {
		errStr := BadRequestPrefix + ": Cannot parse body"
		return ctx.MakeRouteHandlerResultError(http.StatusBadRequest, 3540227685, errStr)
	}

	// validate json
	derr := controller.CreatePayloadIsValid(ctx, createRequestPayload)
	if derr != nil {
		// log.Println("CreatePayloadIsValid returned derr", derr)
		if derr.StatusCodeIsDefaultValue() {
			return ctx.MakeRouteHandlerResultError(http.StatusBadRequest, derr.Num, derr.EndUserMsg)
		} else {
			return ctx.MakeRouteHandlerResultError(derr.StatusCode, derr.Num, derr.EndUserMsg)
		}
	}

	// add entity to the model
	responsePayload, derr := controller.PerformCreate(ctx, createRequestPayload)
	if derr != nil {
		if derr.StatusCodeIsDefaultValue() {
			return ctx.MakeRouteHandlerResultError(http.StatusInternalServerError, derr.Num, derr.EndUserMsg)
		} else {
			return ctx.MakeRouteHandlerResultError(derr.StatusCode, derr.Num, derr.EndUserMsg)
		}
	}

	// response
	if responsePayload != nil {
		return ctx.MakeRouteHandlerResultPayloads(responsePayload)
	} else {
		return ctx.MakeRouteHandlerResultOk()
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
	UpdatePayloadIsValid(ctx *Context, updateRequestPayload Payload) *deeperror.DeepError
	PerformUpdate(ctx *Context, updateRequestPayload Payload) (responsePayload Payload, derr *deeperror.DeepError)
}

func StandardUpdateHandler(controller UpdatePerformer, ctx *Context, updateRequestPayload Payload) RouteHandlerResult {
	// Get the request
	requestBody := ctx.Req.Body
	if requestBody == nil {
		return ctx.MakeRouteHandlerResultError(http.StatusBadRequest, 3851489100, BadRequestPrefix+"Expected non-empty body")
	}
	defer requestBody.Close()

	// parse the json
	decoder := json.NewDecoder(requestBody)
	err := decoder.Decode(updateRequestPayload)

	if err != nil {
		errStr := BadRequestPrefix + ": Cannot parse body"
		return ctx.MakeRouteHandlerResultError(http.StatusBadRequest, 1858602328, errStr)
	}

	// validate json
	derr := controller.UpdatePayloadIsValid(ctx, updateRequestPayload)
	if derr != nil {
		if derr.StatusCodeIsDefaultValue() {
			return ctx.MakeRouteHandlerResultError(http.StatusBadRequest, derr.Num, derr.EndUserMsg)
		} else {
			return ctx.MakeRouteHandlerResultError(derr.StatusCode, derr.Num, derr.EndUserMsg)
		}
	}

	// add entity to the model
	responsePayload, derr := controller.PerformUpdate(ctx, updateRequestPayload)
	if derr != nil {
		if derr.StatusCodeIsDefaultValue() {
			return ctx.MakeRouteHandlerResultError(http.StatusInternalServerError, derr.Num, derr.EndUserMsg)
		} else {
			return ctx.MakeRouteHandlerResultError(derr.StatusCode, derr.Num, derr.EndUserMsg)
		}
	}

	// response
	if responsePayload != nil {
		return ctx.MakeRouteHandlerResultPayloads(responsePayload)
	} else {
		return ctx.MakeRouteHandlerResultOk()
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
	if ctx.End.PrimaryKey <= 0 {
		ctx.SendSimpleErrorPayload(http.StatusBadRequest, BadRequestMissingPrimaryKeyErrorNumber, BadRequestPrefix)
		return
	}

	// validate json
	validator, isValidator := controller.(DeleteValidator)
	if isValidator {
		isValid, errNo := validator.DeleteRequestIsValid(ctx)
		if isValid == false {
			ctx.SendSimpleErrorPayload(http.StatusBadRequest, errNo, BadRequestPrefix)
			return
		}
	}

	// delete entity
	didSucceed, errNo := controller.PerformDelete(ctx)
	if didSucceed == false {
		ctx.SendSimpleErrorPayload(http.StatusInternalServerError, errNo, InternalServerErrorPrefix)
		return
	}

	sendOkPayload(ctx)
}

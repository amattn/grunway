package grunway

import (
	"fmt"
	"log"

	"github.com/amattn/deeperror"
)

type AuthController struct {
	as AccountStore
}

type AccountResponsePayload struct {
	Id    int64
	Name  string
	Email string
}

func NewAccountResponsePayload() *AccountResponsePayload {
	return new(AccountResponsePayload) // leaky bucket?
}

//  #####
// #     # #####  ######   ##   ##### ######
// #       #    # #       #  #    #   #
// #       #    # #####  #    #   #   #####
// #       #####  #      ######   #   #
// #     # #   #  #      #    #   #   #
//  #####  #    # ###### #    #   #   ######
//

type AccountCreateRequestPayload struct {
	Name     string
	Email    string
	Password string
}

func NewAccountCreateRequestPayload() *AccountCreateRequestPayload {
	return new(AccountCreateRequestPayload) // leaky bucket?
}

func (authController *AuthController) PostHandlerV1(c *Context) {
	StandardCreateHandler(authController, c, NewAccountCreateRequestPayload())
}

func (authController *AuthController) CreatePayloadIsValid(c *Context, createRequestPayload interface{}) (isValid bool, errNo int64) {
	requestPayloadPtr, isExpectedType := createRequestPayload.(*AccountCreateRequestPayload)
	if isExpectedType == false {
		return false, 59244845
	}

	emailIsValid := SimpleEmailValidation(requestPayloadPtr.Email)
	if emailIsValid == false {
		return false, 512187273
	}

	passwordIsValue := SimplePasswordValidation(requestPayloadPtr.Password)
	if passwordIsValue == false {
		return false, 512187274
	}

	return true, 0
}

func (authController *AuthController) PerformCreate(c *Context, createRequestPayload interface{}) (bool, int64, interface{}) {
	requestPayloadPtr, isExpectedType := createRequestPayload.(*AccountCreateRequestPayload)
	if isExpectedType == false {
		return false, 3340732022, nil
	}

	entityPtr, err := authController.as.CreateAccount(requestPayloadPtr.Email, requestPayloadPtr.Password)
	if err != nil {
		derr := deeperror.New(600544904, "Could Not Create Account", err)
		derr.DebugMsg = fmt.Sprintf("authController.cs.CreateApp failure creating from requestPayloadPtr %+v", requestPayloadPtr)
		log.Println(derr)
		return false, derr.Num, nil
	}

	responsePayload := NewAccountResponsePayload()
	responsePayload.Email = entityPtr.Email
	responsePayload.Name = entityPtr.Name
	responsePayload.Id = entityPtr.Pkey

	return true, 0, responsePayload
}

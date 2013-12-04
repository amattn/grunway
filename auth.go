package grunway

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/amattn/deeperror"
)

type AccountController struct {
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

func (authController *AccountController) PostHandlerV1(c *Context) {
	StandardCreateHandler(authController, c, NewAccountCreateRequestPayload())
}

func (authController *AccountController) CreatePayloadIsValid(c *Context, createRequestPayload interface{}) (isValid bool, errNo int64) {
	requestPayloadPtr, isExpectedType := createRequestPayload.(*AccountCreateRequestPayload)
	if isExpectedType == false {
		return false, 59244845
	}

	if len(requestPayloadPtr.Name) > 256 {
		return false, 512187272
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

func (authController *AccountController) PerformCreate(c *Context, createRequestPayload interface{}) (bool, int64, interface{}) {
	requestPayloadPtr, isExpectedType := createRequestPayload.(*AccountCreateRequestPayload)
	if isExpectedType == false {
		return false, 3340732022, nil
	}

	entityPtr, err := authController.as.CreateAccount(requestPayloadPtr.Name, requestPayloadPtr.Email, requestPayloadPtr.Password)
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

// #
// #        ####   ####  # #    #
// #       #    # #    # # ##   #
// #       #    # #      # # #  #
// #       #    # #  ### # #  # #
// #       #    # #    # # #   ##
// #######  ####   ####  # #    #
//

type AuthController struct {
	as AccountStore
}

type AccountLoginRequest struct {
	Email    string
	Password string
}
type AccountLoginResponse struct {
	SecretKey string
}

func (authController *AuthController) PostHandlerV1(c *Context) {
	// Get the request
	requestBody := c.R.Body
	if requestBody == nil {
		http.Error(c.W, "400 Bad Request: Expected non-empty body", http.StatusBadRequest)
		return
	}
	defer requestBody.Close()

	// parse the json
	decoder := json.NewDecoder(requestBody)
	requestPayloadPtr := new(AccountLoginRequest)
	err := decoder.Decode(requestPayloadPtr)
	if err != nil {
		errStr := BadRequestPrefix + ": Cannot parse body"
		c.SendErrorPayload(http.StatusBadRequest, 5616956025, errStr)
		return
	}

	// validate json
	if len(requestPayloadPtr.Email) > MAX_EMAIL_LENGTH {
		c.SendErrorPayload(http.StatusBadRequest, 5616956026, BadRequestPrefix)
		return
	}
	if len(requestPayloadPtr.Password) > MAX_PASSWORD_LENGTH {
		c.SendErrorPayload(http.StatusBadRequest, 5616956027, BadRequestPrefix)
		return
	}

	// do the actual login
	acctPtr, err := authController.as.Login(requestPayloadPtr.Email, requestPayloadPtr.Password)
	if err != nil {
		c.SendErrorPayload(http.StatusForbidden, 5296511999, "Invalid email or password")
		return
	}
	// if acctPtr == nil {
	// 	c.SendErrorPayload(http.StatusInternalServerError, 535272094, "")
	// 	return
	// }

	// fetch SecretKey
	responsePayloadPtr := new(AccountLoginResponse)
	responsePayloadPtr.SecretKey = acctPtr.SecretKey

	// response
	c.WrapAndSendPayload(responsePayloadPtr)
}

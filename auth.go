package grunway

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/amattn/deeperror"
)

const (
	X_AUTH_PREFIX = "X-Auth-"
	X_AUTH_PUB    = X_AUTH_PREFIX + "Pub"
	X_AUTH_DATE   = X_AUTH_PREFIX + "Date"
	X_AUTH_SCHEME = X_AUTH_PREFIX + "Scheme"
	X_AUTH_SIG    = X_AUTH_PREFIX + "Sig"

	SCHEME_VERSION_1 = "S1-HMACSHA512"

	AUTH_DATE_TIME_FORMAT = "20060102T150405Z"
)

type AccountController struct {
	AS AccountStore
}
type AuthController struct {
	AS AccountStore
}

type AccountResponsePayload struct {
	Id        int64
	Name      string
	Email     string
	PublicKey string
}

func (payload *AccountResponsePayload) PayloadType() string {
	return "Account"
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

func (authController *AccountController) PostHandlerV1Create(c *Context) {
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

func (authController *AccountController) PerformCreate(c *Context, createRequestPayload interface{}) (bool, *deeperror.DeepError, interface{}) {
	requestPayloadPtr, isExpectedType := createRequestPayload.(*AccountCreateRequestPayload)
	if isExpectedType == false {
		return false, deeperror.NewHTTPError(3340732022, InternalServerErrorPrefix, nil, http.StatusInternalServerError), nil
	}

	if authController.AS == nil {
		return false, deeperror.NewHTTPError(3713900413, InternalServerErrorPrefix, nil, http.StatusInternalServerError), nil
	}

	acct, err := authController.AS.CreateAccount(requestPayloadPtr.Name, requestPayloadPtr.Email, requestPayloadPtr.Password)
	if err != nil {
		// attempt to figure out why we had a failure:

		isAvail, derr := authController.AS.EmailAddressAvailable(requestPayloadPtr.Email)
		if derr == nil {
			if isAvail == false {
				innerDerr := deeperror.NewHTTPError(300544903, "Could Not Create Account, Email address unavailable", derr, http.StatusConflict)
				return false, innerDerr, nil
			}
		}

		log.Println("isAvail", isAvail)
		log.Println("derr", derr)
		derr = deeperror.New(300544904, "Could Not Create Account", err)
		derr.DebugMsg = fmt.Sprintf("authController.cs.CreateApp failure creating from requestPayloadPtr %+v", requestPayloadPtr)

		return false, derr, nil
	}

	responsePld := NewAccountResponsePayload()
	responsePld.Email = acct.Email
	responsePld.Name = acct.Name
	responsePld.Id = acct.Pkey
	responsePld.PublicKey = acct.PublicKey

	return true, nil, responsePld
}

// #
// #        ####   ####  # #    #
// #       #    # #    # # ##   #
// #       #    # #      # # #  #
// #       #    # #  ### # #  # #
// #       #    # #    # # #   ##
// #######  ####   ####  # #    #
//

type AccountLoginRequest struct {
	Email    string
	Password string
}
type AccountLoginResponse struct {
	PublicKey string
}

func (authController *AuthController) PostHandlerV1Login(c *Context) {
	// Check our assumptions
	if authController.AS == nil {
		c.SendErrorPayload(http.StatusInternalServerError, 3256219140, "If you see this, we broke something fierce and we are very sorry.  Please report.")
		return
	}

	// Get the request
	requestBody := c.R.Body
	if requestBody == nil {
		c.SendErrorPayload(http.StatusBadRequest, 3775590199, "400 Bad Request: Expected non-empty body")
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
	acct, err := authController.AS.Login(requestPayloadPtr.Email, requestPayloadPtr.Password)
	if err != nil {
		c.SendErrorPayload(http.StatusForbidden, 5296511999, "Invalid email or password")
		return
	}
	// if acct == nil {
	// 	c.SendErrorPayload(http.StatusInternalServerError, 535272094, "")
	// 	return
	// }

	// fetch SecretKey
	responsePld := new(AccountLoginResponse)
	responsePld.PublicKey = acct.PublicKey

	// response
	c.WrapAndSendPayload(responsePld)
}

func (authController *AuthController) PostHandlerV1Logout(c *Context) {
	// Check our assumptions
	if authController.AS == nil {
		c.SendErrorPayload(http.StatusInternalServerError, 3256219141, "If you see this, we broke something fierce and we are very sorry.  Please report.")
	}

	c.SendErrorPayload(http.StatusInternalServerError, 2745166051, "TODO")
}

//    #
//   # #   #    # ##### #    #
//  #   #  #    #   #   #    #
// #     # #    #   #   ######
// ####### #    #   #   #    #
// #     # #    #   #   #    #
// #     #  ####    #   #    #
//

func performAuth(routePtr *Route, ctx *Context) (authenticationWasSucessful bool, failureToAuthErrorNum int) {

	// Verify date is not too old or too far in the future
	reqDates := ctx.R.Header[X_AUTH_DATE]
	// There should be only 1!
	if len(reqDates) != 1 {
		return false, 3756220698
	}
	// reqDate := reqDates[0]

	// TODO: for now, disable clock checking...
	// We cannot yet guaruntee the client has an accurate clock.
	// There are workarounds, like having the client ping the server for time and storing a diff, but save that for later

	// reqTime, err := time.Parse(AUTH_DATE_TIME_FORMAT, clientReqDate)
	// if err != nil {
	// 	return false, 3103849110
	// }
	// timediff := time.Now().Sub(reqTime)
	// // arbitrary threshold... try to take into account longish response times for mobile devices...
	// if math.Abs(timediff.Seconds()) > 600 {
	// 	log.Println("req too old or too far in the future", timediff.Seconds())
	// 	return 3422808969
	// }

	// get valid key
	publicKeys := ctx.R.Header[X_AUTH_PUB]
	// There should be only 1!
	if len(publicKeys) != 1 {
		return false, 3444855534
	}
	publicKey := publicKeys[0]

	secretKey, errInt := routePtr.Authenticator.GetSecretKey(publicKey)
	if errInt != 0 {
		return false, errInt
	}

	validationErrNum := validateSignature(secretKey, ctx.R.Method, ctx.R.URL, ctx.R.Header)
	if validationErrNum == 0 {
		ctx.PublicKey = publicKey
		return true, 0
	} else {
		ctx.PublicKey = "" // be a little paranoid here.
		return false, validationErrNum
	}
}
func stripAllWhitespace(s string) string {
	isWhitespace := func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}
	return strings.Map(isWhitespace, s)
}
func normalizeURI(url *url.URL) string {
	// NormalizedURI:all lowercase, strip all anchors (#loc), strip all parameters, strip any trailing /
	baseURI := url.RequestURI()
	// log.Println("baseURI", baseURI)
	strippedURI := strings.Split(baseURI, "?")[0]
	normalizedURI := strings.ToLower(strippedURI)
	normalizedURI = strings.TrimRight(normalizedURI, "/")
	// log.Println("normalizedURI", normalizedURI)
	return normalizedURI
}
func normalizeQuery(url *url.URL) string {
	return ""
}

func validateSignature(secretKey, method string, requestURL *url.URL, header http.Header) (validationErrNum int) {
	// log.Println("validateSignature")
	authHeaderKeys := []string{
		X_AUTH_DATE,
		X_AUTH_PUB,
		X_AUTH_SCHEME,
		X_AUTH_SIG,
	}

	for _, headerKey := range authHeaderKeys {
		headerVals := header[headerKey]
		if len(headerVals) != 1 {
			// log.Println(headerKey, "has len(vals) != 1, expected 1")
			return 32601110
		}
	}
	clientReqDate := (header[X_AUTH_DATE][0])
	_, err := time.Parse(AUTH_DATE_TIME_FORMAT, clientReqDate)
	if err != nil {
		// log.Println(cannot parse time)
		// log.Println("err", err)
		return 32601111
	}

	// TODO validate that this is a parseable date?
	clientReqPub := (header[X_AUTH_PUB][0])
	clientReqScheme := (header[X_AUTH_SCHEME][0])
	if clientReqScheme != SCHEME_VERSION_1 {
		// log.Println("Expected Scheme:S1-HMACSHA512 ")
		return 32601113
	}
	clientReqEncodedSig := (header[X_AUTH_SIG][0])
	clientReqSig, err := base64.URLEncoding.DecodeString(clientReqEncodedSig)
	if err != nil {
		// log.Println("base64.StdEncoding.DecodeString returned err:", err)
		return 32601114
	}

	// Generate signing key
	secretKeyHMAC := hmac.New(sha512.New, []byte(secretKey))
	secretKeyHMAC.Write([]byte(clientReqDate))
	signingKey := secretKeyHMAC.Sum(nil)

	// see auth.markdown for definitive definition.  as of this comment:
	// vStringToSign =
	//     HTTPRequestMethod + '\n' +
	//     NormalizedURI + '\n' +
	//     x-auth-date: YYYYMMDD'T'HHMMSS'Z' + '\n' +
	//     x-auth-pub: YYYYYYYY + '\n' +
	//     x-auth-scheme: S1-HMACSHA512

	grunwayReqComponents := []string{
		method,
		normalizeURI(requestURL),
		clientReqDate,
		clientReqPub,
		clientReqScheme,
	}

	stringToSign := strings.Join(grunwayReqComponents, "\n")

	signingKeyHMAC := hmac.New(sha512.New, signingKey)
	signingKeyHMAC.Write([]byte(stringToSign))
	expectedMAC := signingKeyHMAC.Sum(nil)
	// log.Println("signingKey", base64.StdEncoding.EncodeToString(signingKey))
	// log.Println("stringToSign", stringToSign)
	// log.Println("stringToSign len", len([]byte(stringToSign)))
	// log.Printf("stringToSign data % x", ([]byte(stringToSign)))
	// log.Println("clientReqSig", clientReqSig)
	// log.Println("clientReqSig", base64.URLEncoding.EncodeToString(clientReqSig))
	// log.Println("expectedMAC", expectedMAC)
	// log.Println("expectedMAC", base64.URLEncoding.EncodeToString(expectedMAC))
	if hmac.Equal(clientReqSig, expectedMAC) {
		return 0
	} else {
		return 32601119
	}
}

// ######
// #     # ######  ####  #    # ######  ####  #####
// #     # #      #    # #    # #      #        #
// ######  #####  #    # #    # #####   ####    #
// #   #   #      #  # # #    # #           #   #
// #    #  #      #   #  #    # #      #    #   #
// #     # ######  ### #  ####  ######  ####    #
//

// typically used for testing
func SignRequest(req *http.Request, publicKey, secretKey string) {
	now := time.Now()
	now = now.In(time.UTC)
	dateStr := now.Format(AUTH_DATE_TIME_FORMAT)
	req.Header.Add(X_AUTH_DATE, dateStr)
	req.Header.Add(X_AUTH_PUB, publicKey)
	req.Header.Add(X_AUTH_SCHEME, SCHEME_VERSION_1)

	// Generate signing key
	secretKeyHMAC := hmac.New(sha512.New, []byte(secretKey))
	secretKeyHMAC.Write([]byte(dateStr))
	vSigningKey := secretKeyHMAC.Sum(nil)

	// generate string to sign
	grunwayReqComponents := []string{
		req.Method,
		normalizeURI(req.URL),
		dateStr,
		publicKey,
		SCHEME_VERSION_1,
	}

	vStringToSign := strings.Join(grunwayReqComponents, "\n")

	signingKeyHMAC := hmac.New(sha512.New, vSigningKey)
	signingKeyHMAC.Write([]byte(vStringToSign))
	sig := signingKeyHMAC.Sum(nil)
	base64sig := base64.URLEncoding.EncodeToString(sig)
	req.Header.Add(X_AUTH_SIG, base64sig)
}

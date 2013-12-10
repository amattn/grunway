package grunway

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

// #     #
// #     # ##### # #      # ##### # ######  ####
// #     #   #   # #      #   #   # #      #
// #     #   #   # #      #   #   # #####   ####
// #     #   #   # #      #   #   # #           #
// #     #   #   # #      #   #   # #      #    #
//  #####    #   # ###### #   #   # ######  ####
//

func TestValidateSignature(t *testing.T) {
	requestURL, _ := url.Parse("http://example.com/search?q=what&z=z")

	encodedSig := "TEbv2Je9RfmYSxRjUJEEgoTsWwUPmw301N06ZqO7WS_8tL2IfKASeOEOWKidppUeEsv6VhA-RuaoTFapJu4HTw=="
	header := http.Header{
		"X-Auth-Date":   []string{"abc"},
		"X-Auth-Pub":    []string{"abc"},
		"X-Auth-Sig":    []string{encodedSig},
		"X-Auth-Scheme": []string{SCHEME_VERSION_1},
	}
	isValid := validateSignature("secretKey", "GET", requestURL, header)
	if isValid == false {
		t.Error("Expected valid signature")
	}
}

// ######
// #     #  ####  #    # ##### ######  ####
// #     # #    # #    #   #   #      #
// ######  #    # #    #   #   #####   ####
// #   #   #    # #    #   #   #           #
// #    #  #    # #    #   #   #      #    #
// #     #  ####   ####    #   ######  ####
//

func newTestRouter(as AccountStore) *Router {
	router := NewRouter()
	router.BasePath = "/api/"
	router.RegisterEntity("account", &AccountController{as})
	router.RegisterEntity("auth", &AuthController{as})
	router.RegisterEntity("qa", &QAController{})
	return router
}

type QAController struct {
}
type QARequest struct {
}
type QAResponse struct {
	Id       int64
	Question string
	Answer   string
}

func (controller *QAController) GetSecretKey(publicKey string) (string, int) {
	// soooo not secure.
	return "cba", 0
}
func (controller *QAController) AuthGetHandlerV1All(c *Context) {
	resp := new(QAResponse)
	resp.Id = c.E.PrimaryKey
	resp.Question = "What is your quest?"
	resp.Answer = "5 is right out"
	c.WrapAndSendPayload(resp)
}

func setupRouter(t *testing.T) *Router {
	host := "127.0.0.1"
	port := uint16(27500)

	AS := new(PostgresAccountStore)
	attribs := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d", "authtest", "authtest", "authtestdb", host, port)
	err := AS.Startup(attribs)
	if err != nil {
		log.Fatalln("9838924797 FATAL: Cannot startup PostgresAccountStore", host, port, err)
	}

	routerPtr := newTestRouter(AS)

	// routerPtr.LogRoutes()
	t.Log("All Routes:\n", routerPtr.AllRoutesSummary())

	return routerPtr
}

type EndpointTestPath struct {
	Method      string // GET, POST, PUT, DELETE, PATCH
	Headers     map[string]string
	Path        string
	RequestBody string

	ExpectedStatus       int
	ExpectedErrNo        int64
	ExpectedPayloadCount int
	ValidationFunc       func(t *testing.T, i int, etp EndpointTestPath, resp *http.Response, req *http.Request, pw *PayloadWrapper, payloadListBytes []byte)
}

func TestAPIRoutes(t *testing.T) {
	noHeaders := map[string]string{}
	router := setupRouter(t)
	ts := httptest.NewServer(router)
	defer ts.Close()

	var email string
	rand.Seed(time.Now().UnixNano())
	email = fmt.Sprintf("%d@test.test", rand.Int63())

	createPayload := `{"Name":"testApp","Email":"` + email + `","Password":"12345!@#$%"}`
	createBadEmailPayload := `{"Name":"testApp","Email":"intentionallyBadEamil","Password":"12345!@#$%"}`
	createBadPassword1Payload := `{"Name":"testApp","Email":"` + email + `"}`
	createBadPassword2Payload := `{"Name":"testApp","Email":"` + email + `", "Password":"a"}`
	createBadPassword3Payload := `{"Name":"testApp","Email":"` + email + `", "Password":"aaaaaaaa"}`
	// updatePayload := `{"PKey":1,"APIKey":"0","Name":"newName","Description":"newDesc"}`
	bogusPayload := `{"bogus":"bogus"}`

	loginPayload := `{"Email":"` + email + `","Password":"12345!@#$%"}`
	loginBadEmailPayload := `{"Email":"a","Password":"12345!@#$%"}`
	loginBadPassPayload := `{"Email":"` + email + `","Password":"1"}`

	var createdPK int64
	// validate creation
	checkCreateResponseBody := func(t *testing.T, i int, etp EndpointTestPath, resp *http.Response, req *http.Request, pw *PayloadWrapper, payloadListBytes []byte) {
		payloads := make([]*AccountResponsePayload, 0, len(pw.PayloadList))
		err := json.Unmarshal(payloadListBytes, &payloads)
		if err != nil {
			t.Errorf("9409264666 Cannot Unmarshal payloadListBytes:\n%s\n%v", string(payloadListBytes), err)
			return
		}

		if len(payloads) != 1 {
			t.Fatalf("Could not create account")
		}

		for i, payloadPtr := range payloads {
			createdPK = payloadPtr.Id
			if payloadPtr.Id <= 0 {
				t.Errorf("9180113630 %d Expected pkey > 0, got %d %+v", i, payloadPtr.Id, payloadPtr)
			}

			if payloadPtr.Email != email {
				t.Errorf("9953190685 expected %s, got %s", email, payloadPtr.Email)
			}
			if payloadPtr.Name != "testApp" {
				t.Errorf("9953190686 expected testApp, got %s", payloadPtr.Name)
			}
		}
	}

	// first just make one...  just to get a pk
	createETP := EndpointTestPath{"POST", noHeaders, "/api/v1/account/create", createPayload, http.StatusOK, 0, 1, checkCreateResponseBody}
	runETP(t, -1, createETP, ts)

	// now that we have a pkey, iterate through the remaining etps
	log.Println("createdPK", createdPK)
	pk := strconv.Itoa(int(createdPK))
	log.Println("pk", pk)

	// Auth stuff
	qaAllHeader := map[string]string{
		X_AUTH_SCHEME: SCHEME_VERSION_1,
		X_AUTH_DATE:   "abc",
		X_AUTH_PUB:    "abc",
		X_AUTH_SIG:    "yo482lqxi_r5XBI9WLtFdVi16SdzNBfQthNkUQjqr8G5yNNGBxY-yDIZqHEGbjh5sxcPjaB2-tbIBNWbWMvf1g==",
	}

	etps := []EndpointTestPath{

		// CREATE
		EndpointTestPath{"POST", noHeaders, "/api/v1/account/", "", http.StatusNotFound, NotFoundErrNo, 0, nil},
		EndpointTestPath{"POST", noHeaders, "/api/v1/account/create", createBadEmailPayload, http.StatusBadRequest, 512187273, 0, nil},
		EndpointTestPath{"POST", noHeaders, "/api/v1/account/create", createBadPassword1Payload, http.StatusBadRequest, 512187274, 0, nil},
		EndpointTestPath{"POST", noHeaders, "/api/v1/account/create", createBadPassword2Payload, http.StatusBadRequest, 512187274, 0, nil},
		EndpointTestPath{"POST", noHeaders, "/api/v1/account/create", createBadPassword3Payload, http.StatusBadRequest, 512187274, 0, nil},
		EndpointTestPath{"POST", noHeaders, "/api/v1/account/create", bogusPayload, http.StatusBadRequest, 512187273, 0, nil},
		EndpointTestPath{"POST", noHeaders, "/api/v1/account/1/create", createPayload, http.StatusBadRequest, BadRequestExtraneousPrimaryKeyErrNo, 0, nil},
		EndpointTestPath{"POST", noHeaders, "/api/v2234/account/create", "", http.StatusNotFound, NotFoundErrNo, 0, nil},
		EndpointTestPath{"POST", noHeaders, "/api/v1/bogus/", "", http.StatusNotFound, NotFoundErrNo, 0, nil},
		EndpointTestPath{"POST", noHeaders, "/api/v1/bogus/create", "", http.StatusNotFound, NotFoundErrNo, 0, nil},

		// LOGIN
		EndpointTestPath{"POST", noHeaders, "/api/v1/auth/login", loginPayload, http.StatusOK, 0, 1, nil},
		EndpointTestPath{"POST", noHeaders, "/api/v1/auth/login", loginBadEmailPayload, http.StatusForbidden, 5296511999, 0, nil},
		EndpointTestPath{"POST", noHeaders, "/api/v1/auth/login", loginBadPassPayload, http.StatusForbidden, 5296511999, 0, nil},

		// READ
		EndpointTestPath{"GET", noHeaders, "/api/v1/qa/all", "", http.StatusForbidden, 1444855534, 0, nil},
		EndpointTestPath{"GET", qaAllHeader, "/api/v1/qa/all", "", http.StatusOK, 0, 1, nil},
		// EndpointTestPath{"GET", "/api/v2341/account/" + pk, "", http.StatusNotFound, NotFoundErrNo, 0, nil},
		// EndpointTestPath{"GET", "/api/v1/account/", "", http.StatusBadRequest, BadRequestMissingPrimaryKeyErrNo, 0, nil},
		// EndpointTestPath{"GET", "/api/v1/bogus/" + pk, "", http.StatusNotFound, NotFoundErrNo, 0, nil},

		// // UPDATE
		// EndpointTestPath{"PUT", "/api/v1/account/" + pk, updatePayload, http.StatusOK, 0, 1, checkUpdateResponseBody},
	}

	for i, etp := range etps {
		runETP(t, i, etp, ts)
	}
}

func runETP(t *testing.T, i int, etp EndpointTestPath, ts *httptest.Server) {
	client := new(http.Client)

	var reqData io.Reader
	if len(etp.RequestBody) > 0 {
		reqData = strings.NewReader(etp.RequestBody)
	}

	req, err := http.NewRequest(etp.Method, ts.URL+etp.Path, reqData)
	for k, v := range etp.Headers {
		req.Header[k] = []string{v}
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != etp.ExpectedStatus {
		t.Error("9201570762", i, etp.Method, etp.Path, "expected ", etp.ExpectedStatus, ", got", resp.StatusCode)
	}

	pw, payloadListBytes := defaultResponseBodyValidator(t, i, etp, resp, req)
	if etp.ValidationFunc != nil {
		etp.ValidationFunc(t, i, etp, resp, req, pw, payloadListBytes)
	}
}

func defaultResponseBodyValidator(t *testing.T, i int, etp EndpointTestPath, resp *http.Response, req *http.Request) (pw *PayloadWrapper, payloadListBytes []byte) {
	if resp.Body == nil {
		t.Fatalf("9925996202 %d expected non-nil response body: %+v", i, resp)
	}
	responseBytes, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	pw = new(PayloadWrapper)
	err = json.Unmarshal(responseBytes, pw)
	if err != nil {
		if len(responseBytes) == 0 {
			t.Errorf("9925996203 %d Cannot decode: response body is empty", i)
		} else {
			t.Errorf("9925996204 %d Cannot decode response body %s, %v", i, string(responseBytes), err)
		}
	}

	if etp.ExpectedErrNo != pw.ErrNo {
		t.Errorf("9925996205 %d expected errNo = %d, got %d", i, etp.ExpectedErrNo, pw.ErrNo)
	}

	if etp.ExpectedPayloadCount != len(pw.PayloadList) {
		t.Errorf("9925996206 %d Expected Payload Count = %d, got %d", i, etp.ExpectedPayloadCount, len(pw.PayloadList))
	}

	payloadListBytes, err = json.Marshal(pw.PayloadList)
	if err != nil {
		t.Errorf("9925996207 Cannot Marshal PayloadList %+v", pw.PayloadList)
	}

	return
}

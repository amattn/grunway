package grunway

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func newTestRouter(as AccountStore) *Router {
	router := NewRouter()
	router.BasePath = "/api/"
	router.RegisterEntity("account", &AccountController{as})
	router.RegisterEntity("login", &AuthController{as})
	return router
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
	Path        string
	RequestBody string

	ExpectedStatus       int
	ExpectedErrNo        int64
	ExpectedPayloadCount int
	ValidationFunc       func(t *testing.T, i int, etp EndpointTestPath, resp *http.Response, req *http.Request, pw *PayloadWrapper, payloadListBytes []byte)
}

const (
	NotFoundString = NotFoundPrefix
)

func TestAPIRoutes(t *testing.T) {
	router := setupRouter(t)
	ts := httptest.NewServer(router)
	defer ts.Close()

	createPayload := `{"Name":"testApp","Email":"a@b.c","Password":"12345!@#$%"}`
	createBadEmailPayload := `{"Name":"testApp","Email":"intentionallyBadPass"}`
	createBadPassword1Payload := `{"Name":"testApp","Email":"a@b.c"}`
	createBadPassword2Payload := `{"Name":"testApp","Email":"a@b.c", "Password":"a"}`
	createBadPassword3Payload := `{"Name":"testApp","Email":"a@b.c", "Password":"aaaaaaaa"}`
	// updatePayload := `{"PKey":1,"APIKey":"0","Name":"newName","Description":"newDesc"}`
	bogusPayload := `{"bogus":"bogus"}`

	loginPayload := `{"Email":"a@b.c","Password":"12345!@#$%"}`
	loginBadEmailPayload := `{"Email":"a","Password":"12345!@#$%"}`
	loginBadPassPayload := `{"Email":"a@b.c","Password":"1"}`

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

			if payloadPtr.Email != "a@b.c" {
				t.Errorf("9953190685 expected a@b.c, got %s", payloadPtr.Email)
			}
			if payloadPtr.Name != "testApp" {
				t.Errorf("9953190686 expected testApp, got %s", payloadPtr.Name)
			}
		}
	}

	// first just make one...  just to get a pk
	createETP := EndpointTestPath{"POST", "/api/v1/account/", createPayload, http.StatusOK, 0, 1, checkCreateResponseBody}
	runETP(t, -1, createETP, ts)

	// now that we have a pkey, iterate through the remaining etps
	log.Println("createdPK", createdPK)
	pk := strconv.Itoa(int(createdPK))
	log.Println("pk", pk)

	etps := []EndpointTestPath{

		// CREATE
		EndpointTestPath{"POST", "/api/v1/account/", createBadEmailPayload, http.StatusBadRequest, 512187273, 0, nil},
		EndpointTestPath{"POST", "/api/v1/account/", createBadPassword1Payload, http.StatusBadRequest, 512187274, 0, nil},
		EndpointTestPath{"POST", "/api/v1/account/", createBadPassword2Payload, http.StatusBadRequest, 512187274, 0, nil},
		EndpointTestPath{"POST", "/api/v1/account/", createBadPassword3Payload, http.StatusBadRequest, 512187274, 0, nil},
		EndpointTestPath{"POST", "/api/v1/account/", bogusPayload, http.StatusBadRequest, 512187273, 0, nil},
		EndpointTestPath{"POST", "/api/v1/account/1", bogusPayload, http.StatusBadRequest, BadRequestExtraneousPrimaryKeyErrNo, 0, nil},
		EndpointTestPath{"POST", "/api/v2234/account/", "", http.StatusNotFound, NotFoundErrNo, 0, nil},
		EndpointTestPath{"POST", "/api/v1/bogus/", "", http.StatusNotFound, NotFoundErrNo, 0, nil},

		// LOGIN
		EndpointTestPath{"POST", "/api/v1/login", loginPayload, http.StatusOK, 0, 1, nil},
		EndpointTestPath{"POST", "/api/v1/login", loginBadEmailPayload, http.StatusForbidden, 5296511999, 0, nil},
		EndpointTestPath{"POST", "/api/v1/login", loginBadPassPayload, http.StatusForbidden, 5296511999, 0, nil},

		// READ
		// EndpointTestPath{"GET", "/api/v1/account/" + pk, "", http.StatusOK, 0, 1, nil},
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
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != etp.ExpectedStatus {
		t.Error("9584499607", i, etp.Method, etp.Path, "expected ", etp.ExpectedStatus, ", got", resp.StatusCode)
	}

	pw, payloadListBytes := defaultResponseBodyValidator(t, i, etp, resp, req)
	if etp.ValidationFunc != nil {
		etp.ValidationFunc(t, i, etp, resp, req, pw, payloadListBytes)
	}
}

func defaultResponseBodyValidator(t *testing.T, i int, etp EndpointTestPath, resp *http.Response, req *http.Request) (pw *PayloadWrapper, payloadListBytes []byte) {
	if resp.Body == nil {
		t.Fatalf("9849952252 %d expected non-nil response body: %+v", i, resp)
	}
	responseBytes, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	pw = new(PayloadWrapper)
	err = json.Unmarshal(responseBytes, pw)
	if err != nil {
		if len(responseBytes) == 0 {
			t.Errorf("9849952253 %d Cannot decode: response body is empty", i)
		} else {
			t.Errorf("9849952254 %d Cannot decode response body %s, %v", i, string(responseBytes), err)
		}
	}

	if etp.ExpectedErrNo != pw.ErrNo {
		t.Errorf("9849952255 %d expected errNo = %d, got %d", i, etp.ExpectedErrNo, pw.ErrNo)
	}

	if etp.ExpectedPayloadCount != len(pw.PayloadList) {
		t.Errorf("9849952256 %d Expected Payload Count = %d, got %d", i, etp.ExpectedPayloadCount, len(pw.PayloadList))
	}

	payloadListBytes, err = json.Marshal(pw.PayloadList)
	if err != nil {
		t.Errorf("9409264665 Cannot Marshal PayloadList %+v", pw.PayloadList)
	}

	return
}

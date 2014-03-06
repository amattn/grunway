package grunway

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNothing(t *testing.T) {
	// just a placeholder
}

type AuthorPayload struct {
	PKey int64
	Name string
}

func (payload AuthorPayload) PayloadType() string {
	return "author"
}

type AuthorController struct {
}

func (self *AuthorController) GetHandlerV1(ctx *Context) (*RouteError, PayloadsMap, CustomRouteResponse) {
	return ctx.ReturnOK()
}

type BookPayload struct {
	PKey     int64
	Name     string
	AuthorId uint64
}

func (payload BookPayload) PayloadType() string {
	return "book"
}

type BookController struct {
}

func (ctrlr *BookController) GetSecretKey(publicKey string) (string, int) {
	return "abc", 0
}
func (ctrlr *BookController) PerformAuth(routePtr *Route, ctx *Context) (authenticationWasSucessful bool, failureToAuthErrorNum int) {
	// hard code test to fail
	return false, 3220239796
}

func (self *BookController) GetHandlerV1(ctx *Context) (*RouteError, PayloadsMap, CustomRouteResponse) {

	var book BookPayload
	book.PKey = 1
	book.Name = "The Greatest Works of All Time"
	book.AuthorId = 1

	return nil, MakePayloadMapFromPayload(book), nil
}

func (self *BookController) GetHandlerV2(ctx *Context) (*RouteError, PayloadsMap, CustomRouteResponse) {
	return ctx.ReturnOK()
}
func (self *BookController) GetHandlerV003(ctx *Context) (*RouteError, PayloadsMap, CustomRouteResponse) {
	return ctx.ReturnOK()
}
func (self *BookController) GetHandlerV018(ctx *Context) (*RouteError, PayloadsMap, CustomRouteResponse) {
	return ctx.ReturnOK()
}
func (self *BookController) PostHandlerV1(ctx *Context) (*RouteError, PayloadsMap, CustomRouteResponse) {
	return ctx.ReturnOK()
}
func (self *BookController) PutHandlerV1(ctx *Context) (*RouteError, PayloadsMap, CustomRouteResponse) {
	return ctx.ReturnOK()
}
func (self *BookController) DeleteHandlerV1(ctx *Context) (*RouteError, PayloadsMap, CustomRouteResponse) {
	return ctx.ReturnOK()
}
func (self *BookController) GetHandlerV1All(ctx *Context) (*RouteError, PayloadsMap, CustomRouteResponse) {
	return ctx.ReturnOK()
}
func (self *BookController) GetHandlerV1Popular(ctx *Context) (*RouteError, PayloadsMap, CustomRouteResponse) {
	return ctx.ReturnOK()
}
func (self *BookController) AuthGetHandlerV1Login(ctx *Context) (*RouteError, PayloadsMap, CustomRouteResponse) {
	return ctx.ReturnOK()
}

func makeLibrary(t *testing.T) *Router {
	routerPtr := NewRouter()
	routerPtr.BasePath = "/api/"
	routerPtr.RegisterEntity("author", &AuthorController{})
	routerPtr.RegisterEntity("book", &BookController{})

	// routerPtr.LogRoutes()
	t.Log("All Routes:\n", routerPtr.AllRoutesSummary())

	return routerPtr
}

func TestRouterSetup(t *testing.T) {
	makeLibrary(t)
}

func TestRouter(t *testing.T) {
	router := makeLibrary(t)
	ts := httptest.NewServer(router)
	defer ts.Close()

	getURLAndStatusCodes := map[string]int{
		// CRUD
		"/api/":                http.StatusNotFound,
		"/api/v1/":             http.StatusNotFound,
		"/api/v1/book/1":       http.StatusOK,
		"/api/v2/book/1":       http.StatusOK,
		"/api/v3/book/1":       http.StatusOK,
		"/api/v4/book/1":       http.StatusNotFound,
		"/api/v18/book/1":      http.StatusOK,
		"/api/v1/book/":        http.StatusBadRequest,
		"/api/v1/book":         http.StatusBadRequest,
		"/api/v1/author/1":     http.StatusOK,
		"/api/v1/bogus/1":      http.StatusNotFound,
		"/api/v1/book/1/login": http.StatusForbidden,

		// Custom
		"/api/v1/book/Popular": http.StatusOK,
		"/api/v1/book/all":     http.StatusOK,
		"/api/v1/book/All":     http.StatusOK,
		"/api/v1/book/Bogus":   http.StatusNotFound,
	}

	commonTest := func(t *testing.T, response *http.Response, urlsuffix string) {
		ct := response.Header.Get(httpHeaderContentType)
		if ct != httpHeaderContentTypeJSON {
			// t.Logf("response %+v", response)
			// t.Logf("response.Header %+v", response.Header)
			t.Error(urlsuffix, "expected \"Content-Type\":", httpHeaderContentTypeJSON, ", got \"Content-Type\":", ct, "contentLength:", response.ContentLength)
		}
	}

	for urlsuffix, expectedStatusCode := range getURLAndStatusCodes {
		response, err := http.Get(ts.URL + urlsuffix)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		// t.Log("starting getURLAndStatusCodes", urlsuffix, expectedStatusCode, response.ContentLength)

		if response.StatusCode != expectedStatusCode {

			t.Error("GET", urlsuffix, "expected ", expectedStatusCode, ", got", response.StatusCode)
			body, _ := ioutil.ReadAll(response.Body)
			t.Error("response body:", string(body))
		}

		commonTest(t, response, urlsuffix)
		// t.Log("ending getURLAndStatusCodes", ts.URL+urlsuffix)
	}

	postURLAndStatusCodes := map[string]int{
		"/api/":         http.StatusNotFound,
		"/api/v1/":      http.StatusNotFound,
		"/api/v1/book/": http.StatusOK,

		"/api/v2/book/":   http.StatusNotFound,
		"/api/v3/book/":   http.StatusNotFound,
		"/api/v1/author/": http.StatusNotFound,
		"/api/v1/bogus/":  http.StatusNotFound,

		"/api/v1/book/1": http.StatusBadRequest, // Create (POST) should never have a pk
	}

	for urlsuffix, expectedStatusCode := range postURLAndStatusCodes {
		// t.Log("starting postURLAndStatusCodes, posting plain text")
		buf := bytes.NewBufferString("Hello World!")
		response, err := http.Post(ts.URL+urlsuffix, "text/plain", buf)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if response.StatusCode != expectedStatusCode {
			t.Error("POST", urlsuffix, "expected ", expectedStatusCode, ", got", response.StatusCode)
			body, _ := ioutil.ReadAll(response.Body)
			t.Error("response body:", string(body))
		}

		commonTest(t, response, urlsuffix)
		// t.Log("ending postURLAndStatusCodes, posting plain text")
	}

	putURLAndStatusCodes := map[string]int{
		"/api/":          http.StatusNotFound,
		"/api/v1/":       http.StatusNotFound,
		"/api/v1/book/1": http.StatusOK,

		"/api/v2/book/1":   http.StatusNotFound,
		"/api/v3/book/1":   http.StatusNotFound,
		"/api/v1/author/1": http.StatusNotFound,
		"/api/v1/bogus/1":  http.StatusNotFound,

		"/api/v1/book/": http.StatusBadRequest, // Update (PUT) should always have a pk
	}

	for urlsuffix, expectedStatusCode := range putURLAndStatusCodes {
		// t.Log("starting putURLAndStatusCodes, posting plain text")
		buf := bytes.NewBufferString("Hello World!")

		client := new(http.Client)
		req, err := http.NewRequest("PUT", ts.URL+urlsuffix, buf)
		response, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()

		if response.StatusCode != expectedStatusCode {
			t.Error("PUT", urlsuffix, "expected ", expectedStatusCode, ", got", response.StatusCode)
			body, _ := ioutil.ReadAll(response.Body)
			t.Error("response body:", string(body))
		}
		commonTest(t, response, urlsuffix)
		// t.Log("ending putURLAndStatusCodes, posting plain text")
	}
}

func TestPayload(t *testing.T) {
	router := makeLibrary(t)
	ts := httptest.NewServer(router)
	defer ts.Close()

	response, err := http.Get(ts.URL + "/api/v1/book/1")

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Error("9244741117", err)
	}
	defer response.Body.Close()
	bodyString := string(bodyBytes)
	// t.Log(bodyString)

	// decode the json,
	pw, err := UnmarshalPayloadWrapper(bodyBytes, BookPayload{}, AuthorPayload{})
	t.Logf("pw %+v", pw)

	if err != nil {
		t.Fatal("9063845927 JSON unmarshal failure", err, "\n", bodyString)
	}

	//check errNo == 0
	if pw.ErrNo != 0 {
		t.Errorf("918188683 expected non-zero ErrStr, got %d\npayload:%+v", pw.ErrNo, pw)
	}
	//check pk matches expected,

	if len(pw.Payloads) != 1 {
		t.Errorf("918188684 expected 1 payloadType, got %v\npayload:%+v", pw.ErrStr, pw)

	}

	t.Log("pw", pw)
	t.Log("pw.Payloads", pw.Payloads)

	for payloadType, payloadList := range pw.Payloads {

		t.Log("payloadType", payloadType)

		for i, untypedEntity := range payloadList {

			t.Log("untypedEntity", untypedEntity)
			jsonBytes, err := json.Marshal(untypedEntity)
			if err != nil {
				t.Fatalf("918188685 failure to marshal entity %+v", untypedEntity)
			}

			var bookPayload BookPayload
			err = json.Unmarshal(jsonBytes, &bookPayload)
			if err != nil {
				t.Fatalf("918188686 failure to Unmarshal entity %+v", string(jsonBytes))
			}

			if bookPayload.PKey != 1 {
				t.Errorf("918188687 i:%d Expected pk == 1, got %d\nentity:%+v\nbodyString:%v", i, bookPayload.PKey, pw, bodyString)
			}
		}
	}
}

// Benchmark our routeKey Algorithms.  this is called every request.

//As of 2013-09-19, Go 1.1, rMBP
//BenchmarkRouteKeyJoinString	 5000000	       483 ns/op
func BenchmarkRouteKeyJoinString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		routeKeyJoinString("GET", "1", "book", "all")
	}
}

//As of 2013-09-19, Go 1.1, rMBP
//BenchmarkRouteKeyFormatString	 1000000	      1249 ns/op
func BenchmarkRouteKeyFormatString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		routeKeyFormatString("GET", "1", "book", "all")
	}
}

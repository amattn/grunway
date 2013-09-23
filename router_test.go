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

type AuthorController struct {
}
type BookController struct {
}

type AuthorPayload struct {
	Entity
	Name string
}
type BookPayload struct {
	Entity
	Name     string
	AuthorId uint64
}

func (self *AuthorController) GetHandlerV1(ctx *Context) {

}

func (self *BookController) GetHandlerV1(ctx *Context) {

	var book BookPayload
	book.SetPrimaryKey(1)
	book.Name = "The Greatest Works of All Time"
	book.AuthorId = 1

	ctx.WrapAndSendPayload(book)
}
func (self *BookController) GetHandlerV2(ctx *Context) {

}
func (self *BookController) GetHandlerV003(ctx *Context) {

}
func (self *BookController) GetHandlerV018(ctx *Context) {

}
func (self *BookController) PostHandlerV1(ctx *Context) {

}
func (self *BookController) PutHandlerV1(ctx *Context) {

}
func (self *BookController) DeleteHandlerV1(ctx *Context) {

}
func (self *BookController) GetHandlerV1All(ctx *Context) {

}
func (self *BookController) GetHandlerV1Popular(ctx *Context) {

}

func makeLibrary(t *testing.T) *Router {
	routerPtr := NewRouter()
	routerPtr.BasePath = "/api/"
	routerPtr.RegisterEntity("author", &AuthorController{}, &AuthorPayload{})
	routerPtr.RegisterEntity("book", &BookController{}, &BookPayload{})

	// routerPtr.LogRoutes()
	t.Log("All Routes:\n", routerPtr.AllRoutesSummary())

	return routerPtr
}

func TestRouterSetup(t *testing.T) {
	makeLibrary(t)
}

func TestAPIRoutes(t *testing.T) {
	router := makeLibrary(t)
	ts := httptest.NewServer(router)
	defer ts.Close()

	getURLAndStatusCodes := map[string]int{
		// CRUD
		"/api/v1/book/1":   http.StatusOK,
		"/api/v2/book/1":   http.StatusOK,
		"/api/v3/book/1":   http.StatusOK,
		"/api/v4/book/1":   http.StatusNotFound,
		"/api/v18/book/1":  http.StatusOK,
		"/api/v1/book/":    http.StatusBadRequest,
		"/api/v1/book":     http.StatusBadRequest,
		"/api/v1/author/1": http.StatusOK,
		"/api/v1/bogus/1":  http.StatusNotFound,

		// Custom
		"/api/v1/book/Popular": http.StatusOK,
		"/api/v1/book/All":     http.StatusOK,
		"/api/v1/book/Bogus":   http.StatusNotFound,
	}

	for urlsuffix, expectedStatusCode := range getURLAndStatusCodes {
		response, err := http.Get(ts.URL + urlsuffix)
		if err != nil {
			t.Fatal(err)
		}

		if response.StatusCode != expectedStatusCode {
			t.Error("GET", urlsuffix, "expected ", expectedStatusCode, ", got", response.StatusCode)
		}
	}

	postURLAndStatusCodes := map[string]int{
		"/api/v1/book/": http.StatusOK,

		"/api/v2/book/":   http.StatusNotFound,
		"/api/v3/book/":   http.StatusNotFound,
		"/api/v1/author/": http.StatusNotFound,
		"/api/v1/bogus/":  http.StatusNotFound,

		"/api/v1/book/1": http.StatusBadRequest, // Create (POST) should never have a pk
	}

	for urlsuffix, expectedStatusCode := range postURLAndStatusCodes {
		buf := bytes.NewBufferString("Hello World!")
		response, err := http.Post(ts.URL+urlsuffix, "text/plain", buf)
		if err != nil {
			t.Fatal(err)
		}

		if response.StatusCode != expectedStatusCode {
			t.Error("POST", urlsuffix, "expected ", expectedStatusCode, ", got", response.StatusCode)
		}
	}

	putURLAndStatusCodes := map[string]int{
		"/api/v1/book/1": http.StatusOK,

		"/api/v2/book/1":   http.StatusNotFound,
		"/api/v3/book/1":   http.StatusNotFound,
		"/api/v1/author/1": http.StatusNotFound,
		"/api/v1/bogus/1":  http.StatusNotFound,

		"/api/v1/book/": http.StatusBadRequest, // Update (PUT) should always have a pk
	}

	for urlsuffix, expectedStatusCode := range putURLAndStatusCodes {
		buf := bytes.NewBufferString("Hello World!")

		client := new(http.Client)
		req, err := http.NewRequest("PUT", ts.URL+urlsuffix, buf)
		response, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if response.StatusCode != expectedStatusCode {
			t.Error("PUT", urlsuffix, "expected ", expectedStatusCode, ", got", response.StatusCode)
		}
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
	t.Log(bodyString)

	// decode the json,
	var pw PayloadWrapper
	json.Unmarshal(bodyBytes, &pw)

	//check errNo == 0
	if pw.ErrNo != 0 {
		t.Errorf("918188683 expected non-zero ErrStr, got %d\npayload:%+v", pw.ErrNo, pw)
	}
	//check pk matches expected,

	if len(pw.PayloadList) != 1 {
		t.Errorf("918188684 expected 1 entity, got %v\npayload:%+v", pw.ErrStr, pw)
	}

	for i, untypedEntity := range pw.PayloadList {

		jsonBytes, err := json.Marshal(untypedEntity)
		if err != nil {
			t.Fatalf("918188685 failure to marshal entity %+v", untypedEntity)
		}

		var bookPayload BookPayload
		err = json.Unmarshal(jsonBytes, &bookPayload)
		if err != nil {
			t.Fatalf("918188686 failure to Unmarshal entity %+v", string(jsonBytes))
		}

		if bookPayload.GetPrimaryKey() != 1 {
			t.Errorf("918188687 i:%d Expected pk == 1, got %d\nentity:%+v", i, bookPayload.GetPrimaryKey(), bookPayload)
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

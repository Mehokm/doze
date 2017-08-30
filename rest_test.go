package rest

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestController struct{}

func (t TestController) SimpleGet(c Context) ResponseSender {
	return NewOKJSONResponse(TestStruct{"Simple Get"})
}

func (t TestController) SimplePost(c Context) ResponseSender {
	var ts TestStruct

	c.BindJSONEntity(&ts)

	return NewOKJSONResponse(ts)
}

func (t TestController) SimplePut(c Context) ResponseSender {
	var ts TestStruct

	c.BindJSONEntity(&ts)
	ts.Message = ts.Message + " Updated"
	return NewOKJSONResponse(ts)
}

type TestStruct struct {
	Message string
}

const RestRoot = "/rest/api"

var (
	mux    *http.ServeMux
	server *httptest.Server
	r      router
)

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	r = DefaultRouter().RouteMap(
		NewRoute().Name("simpleGet").For(RestRoot+"/simpleget").With(http.MethodGet, TestController{}.SimpleGet),
		NewRoute().Name("simplePost").For(RestRoot+"/simplepost").With(http.MethodPost, TestController{}.SimplePost),
		NewRoute().Name("simplePut").For(RestRoot+"/simpleput").With(http.MethodPut, TestController{}.SimplePut),
	)
}

func teardown() {
	server.Close()
}

func TestRestMethodNotAllowed(t *testing.T) {
	setup()
	defer teardown()

	mux.Handle(RestRoot+"/", NewHandler(r))

	resp, _ := http.Get(server.URL + RestRoot + "/simplepost")

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode, "they should be equal")
}

func TestRestMethodNotFound(t *testing.T) {
	setup()
	defer teardown()

	mux.Handle(RestRoot+"/", NewHandler(r))

	resp, _ := http.Get(server.URL + RestRoot + "/notfound")

	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "they should be equal")
}

func TestRestSimpleGet(t *testing.T) {
	setup()
	defer teardown()

	mux.Handle(RestRoot+"/", NewHandler(r))

	resp, _ := http.Get(server.URL + RestRoot + "/simpleget")
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, `{"Message":"Simple Get"}`, string(body), "they should be equal")
}

func TestRestSimplePost(t *testing.T) {
	setup()
	defer teardown()

	mux.Handle(RestRoot+"/", NewHandler(r))

	resp, _ := http.Post(server.URL+RestRoot+"/simplepost", "application/json", strings.NewReader(`{"Message":"Simple Post"}`))
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, `{"Message":"Simple Post"}`, string(body), "they should be equal")
}

func TestRestSimplePut(t *testing.T) {
	setup()
	defer teardown()

	mux.Handle(RestRoot+"/", NewHandler(r))

	req, _ := http.NewRequest(http.MethodPut, server.URL+RestRoot+"/simpleput", strings.NewReader(`{"Message":"Simple Put"}`))
	resp, _ := http.DefaultClient.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, `{"Message":"Simple Put Updated"}`, string(body), "they should be equal")
}

func TestRestHandlerInterceptorTrue(t *testing.T) {
	setup()
	defer teardown()

	h := NewHandler(r)

	h.Intercept(func(c Context) bool {
		c.ResponseWriter.Header().Add("test", "true")

		return true
	})

	mux.Handle(RestRoot+"/", h)

	resp, _ := http.Get(server.URL + RestRoot + "/simpleget")

	assert.Equal(t, "true", resp.Header.Get("test"), "they should be equal")
}

func TestRestHandlerInterceptorFalse(t *testing.T) {
	setup()
	defer teardown()

	h := NewHandler(r)

	h.Intercept(func(c Context) bool {
		c.ResponseWriter.Header().Add("test", "false")

		return false
	})

	mux.Handle(RestRoot+"/", h)

	resp, _ := http.Get(server.URL + RestRoot + "/simpleget")
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, "false", resp.Header.Get("test"), "they should be equal")
	assert.Equal(t, "", string(body), "they should be equal")
}

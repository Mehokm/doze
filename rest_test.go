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

	c.BindJsonEntity(&ts)

	return NewOKJSONResponse(ts)
}

func (t TestController) SimplePut(c Context) ResponseSender {
	var ts TestStruct

	c.BindJsonEntity(&ts)
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
		NewRoute().Named("simpleGet").For(RestRoot+"/simpleget").With(MethodGet, TestController{}.SimpleGet),
		NewRoute().Named("simplePost").For(RestRoot+"/simplepost").With(MethodPost, TestController{}.SimplePost),
		NewRoute().Named("simplePut").For(RestRoot+"/simpleput").With(MethodPut, TestController{}.SimplePut),
	)

	mux.Handle(RestRoot+"/", NewHandler(r))
}

func teardown() {
	server.Close()
}

func TestRestMethodNotAllowed(t *testing.T) {
	setup()
	defer teardown()

	resp, _ := http.Get(server.URL + RestRoot + "/simplepost")

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode, "they should be equal")
}

func TestRestMethodNotFound(t *testing.T) {
	setup()
	defer teardown()

	resp, _ := http.Get(server.URL + RestRoot + "/notfound")

	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "they should be equal")
}

func TestRestSimpleGet(t *testing.T) {
	setup()
	defer teardown()

	resp, _ := http.Get(server.URL + RestRoot + "/simpleget")
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, string(body), `{"Message":"Simple Get"}`, "they should be equal")
}

func TestRestSimplePost(t *testing.T) {
	setup()
	defer teardown()

	resp, _ := http.Post(server.URL+RestRoot+"/simplepost", "application/json", strings.NewReader(`{"Message":"Simple Post"}`))
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, string(body), `{"Message":"Simple Post"}`, "they should be equal")
}

func TestRestSimplePut(t *testing.T) {
	setup()
	defer teardown()

	req, _ := http.NewRequest(MethodPut, server.URL+RestRoot+"/simpleput", strings.NewReader(`{"Message":"Simple Put"}`))
	resp, _ := http.DefaultClient.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, string(body), `{"Message":"Simple Put Updated"}`, "they should be equal")
}

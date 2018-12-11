package doze

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestController struct{}

func (t TestController) SimpleGet(c *Context) ResponseSender {
	return NewOKJSONResponse(TestStruct{"Simple Get"})
}

func (t TestController) SimplePost(c *Context) ResponseSender {
	var ts TestStruct

	c.BindJSONEntity(&ts)

	return NewOKJSONResponse(ts)
}

func (t TestController) SimplePut(c *Context) ResponseSender {
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
	r      RestRouter
)

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	r = Router("test")

	r.Add(NewRoute().Named("simpleGet").For(RestRoot+"/simpleget").With(http.MethodGet, TestController{}.SimpleGet))
	r.Add(NewRoute().Named("simplePost").For(RestRoot+"/simplepost").With(http.MethodPost, TestController{}.SimplePost))
	r.Add(NewRoute().Named("simplePut").For(RestRoot+"/simpleput").With(http.MethodPut, TestController{}.SimplePut))

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

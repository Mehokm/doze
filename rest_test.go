package rest

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestController struct{}

func (t TestController) TestAction(c Context) ResponseSender {
	return NewOKJSONResponse("Action Jackson")
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
		NewRoute().Named("notAllowed").For(RestRoot+"/notallowed").With(MethodPost, TestController{}.TestAction),
		NewRoute().Named("simpleGet").For(RestRoot+"/simpleget").With(MethodGet, TestController{}.TestAction),
	)

	mux.Handle(RestRoot+"/", NewHandler(r))
}

func teardown() {
	server.Close()
}

func Get(uri string) (string, error) {
	res, err := http.Get(server.URL + uri)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", err
	}
	return string(body), err
}

func TestRestMethodNotAllowed(t *testing.T) {
	setup()
	defer teardown()

	resp, _ := http.Get(server.URL + RestRoot + "/notallowed")

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode, "they should be equal")
}

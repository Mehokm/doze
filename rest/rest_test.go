package rest

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

type TestController struct {
	TestAction ControllerAction
}

const RestRoot = "/rest/api/"

func TestRestMethodNotAllowed(t *testing.T) {
	testController := TestController{}
	testController.TestAction = func(hb HttpBundle) Response {
		var body interface{}
		body = "Action Jackson"
		return NewOKJsonResponse(body)
	}

	router := DefaultRouter().RouteMap(
		NewRoute().For(RestRoot+"/test").With("GET", "TestAction").Using(testController),
	)

	rh := NewRestHandler(router)

	ts := httptest.NewServer(rh)
	defer ts.Close()

	fmt.Println(ts.URL + RestRoot + "test")
	res, err := http.Get(ts.URL + RestRoot + "test")
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(body))
}

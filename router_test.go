package rest

import (
	"reflect"
	"testing"
)

func TestRouter_GetRoute(t *testing.T) {
	router := DefaultRouter().RouteMap(
		NewRoute().Named("TestRoute").For("/people/{id:i}/details/{name:a}").With("GET", TestController{}.TestAction),
	)

	testRoute := router.GetRoute("TestRoute")

	if testRoute.Path != "/people/{id:i}/details/{name:a}" {
		t.Error("Expected Path to match '/people/{id:i}/details/{name:a}'")
	}

	if !reflect.DeepEqual(map[string]string{"GET": "TestAction"}, testRoute.action) {
		t.Error("Expected Action to match")
	}
}

package rest

import "testing"

func TestRouter_GetRoute(t *testing.T) {
	action := TestController{}.SimpleGet
	router := DefaultRouter().RouteMap(
		NewRoute().Named("TestRoute").For("/people/{id:i}/details/{name:a}").With("GET", action),
	)

	testRoute := router.GetRoute("TestRoute")

	if testRoute.Path != "/people/{id:i}/details/{name:a}" {
		t.Error("Expected Path to match '/people/{id:i}/details/{name:a}'")
	}

	// a := testRoute.action["GET"]
	// fmt.Println(&action == &a)
	// if !reflect.DeepEqual(map[string]ControllerAction{"GET": action}, testRoute.action) {
	// 	t.Error("Expected Action to match")
	// }
}

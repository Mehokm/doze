package rest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouterGetRouteWithName(t *testing.T) {
	router := DefaultRouter().RouteMap(
		NewRoute().Named("TestRoute").For("/people/{id:i}/details/{name:a}").With("GET", TestController{}.SimpleGet),
	)

	testRoute := router.GetRoute("TestRoute")

	assert.Equal(t, "/people/{id:i}/details/{name:a}", testRoute.path, "paths should match")
}

func TestRouterGetRouteWithoutName(t *testing.T) {
	router := DefaultRouter().RouteMap(
		NewRoute().For("/people/{id:i}/details/{name:a}").With("GET", TestController{}.SimpleGet),
	)

	testRoute := router.GetRoute("/people/{id:i}/details/{name:a}")

	assert.Equal(t, "/people/{id:i}/details/{name:a}", testRoute.path, "paths should match")
}

func TestRouterRouteMatch(t *testing.T) {
	router := DefaultRouter().RouteMap(
		NewRoute().For("/people/{id:i}/details/{name:a}").With("GET", TestController{}.SimpleGet),
	)

	route1 := router.Match("/people/10/details/job")

	assert.NotNil(t, route1, "route1 should not be nil")

	route2 := router.Match("/people/job/details/10")

	assert.Nil(t, route2, "route2 should be nil")
}

func TestRouteParams(t *testing.T) {
	router := DefaultRouter().RouteMap(
		NewRoute().For("/people/{id:i}/details/{name:a}").With("GET", TestController{}.SimpleGet),
	)

	route := router.Match("/people/10/details/job")

	assert.Len(t, route.Params(), 2, "length should be 2")
	assert.Equal(t, 10, route.Params()["id"], "they should match")
	assert.Equal(t, "job", route.Params()["name"], "they should match")
}

package rest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamedRouterMap(t *testing.T) {
	v1 := DefaultRouter().RouteMap(
		NewRoute().Named("TestRoute").For("/v1/people/{id:i}/details/{name:a}").With("GET", TestController{}.SimpleGet),
	)

	v2 := NewRouter("v2").RouteMap(
		NewRoute().Named("TestRoute").For("/v2/people/{id:i}/details/{name:a}").With("GET", TestController{}.SimpleGet),
	)

	assert.NotEqual(t, v1.Get("TestRoute"), v2.Get("TestRoute"), "should not be equal")
	assert.Equal(t, "/v1/people/{id:i}/details/{name:a}", v1.Get("TestRoute").Path, "should be equal")
	assert.Equal(t, "/v2/people/{id:i}/details/{name:a}", v2.Get("TestRoute").Path, "should be equal")
}

func TestRouterGetRouteWithName(t *testing.T) {
	router := DefaultRouter().RouteMap(
		NewRoute().Named("TestRoute").For("/people/{id:i}/details/{name:a}").With("GET", TestController{}.SimpleGet),
	)

	testRoute := router.Get("TestRoute")

	assert.Equal(t, "/people/{id:i}/details/{name:a}", testRoute.Path, "Paths should match")
}

func TestRouterGetRouteWithoutName(t *testing.T) {
	router := DefaultRouter().RouteMap(
		NewRoute().For("/people/{id:i}/details/{name:a}").With("GET", TestController{}.SimpleGet),
	)

	testRoute := router.Get("/people/{id:i}/details/{name:a}")

	assert.Equal(t, "/people/{id:i}/details/{name:a}", testRoute.Path, "Paths should match")
}

func TestRouterRouteMatch(t *testing.T) {
	router := DefaultRouter().RouteMap(
		NewRoute().For("/people/{id:i}/details/{name:a}").With("GET", TestController{}.SimpleGet),
		NewRoute().For("/people/{id}").With("GET", TestController{}.SimpleGet),
	)

	route1 := router.Match("/people/10/details/job")

	assert.NotNil(t, route1, "route1 should not be nil")

	route2 := router.Match("/people/job/details/10")

	assert.Nil(t, route2, "route2 should be nil")

	route3 := router.Match("/people/10")

	assert.NotNil(t, route3, "route3 should not be nil")

	route4 := router.Match("/people/job")

	assert.NotNil(t, route4, "route4 should not be nil")

	route5 := router.Match("/people/10/details/10")

	assert.Nil(t, route5, "route5 should be nil")
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

func TestRouteBuildShouldError(t *testing.T) {
	router := DefaultRouter().RouteMap(
		NewRoute().Named("test").For("/people/{id:i}/details/{name}").With("GET", TestController{}.SimpleGet),
	)

	m1 := map[string]interface{}{
		"id":   65,
		"not":  "valid",
		"name": "Joe",
	}

	s1, err1 := router.Get("test").Build(m1)

	assert.EqualError(t, err1, "wrong number of parameters: 3 given, 2 required", "they should match")
	assert.Equal(t, "", s1, "they should match")

	m2 := map[string]interface{}{
		"id":  65,
		"not": "valid",
	}

	s2, err2 := router.Get("test").Build(m2)

	assert.EqualError(t, err2, "parameter not valid: not", "they should match")
	assert.Equal(t, "", s2, "they should match")
}

func TestRouteBuild(t *testing.T) {
	router := DefaultRouter().RouteMap(
		NewRoute().Named("test").For("/people/{id:i}/details/{name}").With("GET", TestController{}.SimpleGet),
	)

	m := map[string]interface{}{
		"id":   65,
		"name": "Joe",
	}

	s, err := router.Get("test").Build(m)

	assert.Nil(t, err, "error should be nil")
	assert.Equal(t, "/people/65/details/Joe", s, "they should match")
}

func TestRouterPrefix(t *testing.T) {
	router := DefaultRouter().Prefix("/api/v3").RouteMap(
		NewRoute().Named("TestRoute").For("/people/{id:i}/details/{name:a}").With("GET", TestController{}.SimpleGet),
	)

	testRoute := router.Get("TestRoute")

	assert.Equal(t, "/api/v3/people/{id:i}/details/{name:a}", testRoute.Path, "Paths should match")
}

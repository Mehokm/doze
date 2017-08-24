package rest

import (
	"regexp"
	"strings"
	"sync"
)

type router struct {
	prefix     string
	Routes     map[string]*Route
	routingMap map[*Route]*regexp.Regexp
}

type routeBuilder struct {
	path      string
	actions   map[string]Action
	routeName string
}

var (
	routers map[string]router
	lock    sync.RWMutex
)

func init() {
	routers = make(map[string]router)
	routers["default"] = router{"", make(map[string]*Route), make(map[*Route]*regexp.Regexp)}
}

// DefaultRouter returns the router with name "default"
func DefaultRouter() router {
	lock.Lock()
	defer lock.Unlock()

	return routers["default"]
}

// Router returns a router specified by a name
func Router(name string) router {
	lock.Lock()
	defer lock.Unlock()

	// create new router if it doesn't exist
	if _, ok := routers[name]; !ok {
		routers[name] = router{"", make(map[string]*Route), make(map[*Route]*regexp.Regexp)}
	}

	return routers[name]
}

func (ro router) Prefix(prefix string) router {
	ro.prefix = prefix

	return ro
}

// NewRoute returns a wrapper to make a builder for Route
func NewRoute() *routeBuilder {
	return &routeBuilder{actions: make(map[string]Action)}
}

func (rb *routeBuilder) Name(name string) *routeBuilder {
	rb.routeName = name

	return rb
}

func (rb *routeBuilder) For(path string) *routeBuilder {
	rb.path = path

	return rb
}

func (rb *routeBuilder) With(method string, action Action) *routeBuilder {
	rb.actions[method] = action

	return rb
}

func (rb *routeBuilder) And(method string, action Action) *routeBuilder {
	return rb.With(method, action)
}

func (ro router) RouteMap(rbs ...*routeBuilder) router {
	for _, routeBuilder := range rbs {
		var paramNames []string

		parts := strings.Split(ro.prefix+routeBuilder.path, "/")
		for i := 0; i < len(parts); i++ {
			if len(parts[i]) > 0 && string(parts[i][0]) == "{" && string(parts[i][len(parts[i])-1]) == "}" {
				paramName := parts[i][1 : len(parts[i])-1]

				if index := strings.Index(parts[i], ":"); index >= 0 {
					paramName = parts[i][1:index]
				}

				paramNames = append(paramNames, paramName)
			}
		}

		route := &Route{
			Path:       ro.prefix + routeBuilder.path,
			Actions:    routeBuilder.actions,
			ParamNames: paramNames,
		}

		if routeBuilder.routeName == "" {
			ro.Routes[routeBuilder.path] = route
		}
		ro.Routes[routeBuilder.routeName] = route
	}

	return ro
}

func (ro router) Get(name string) *Route {
	return ro.Routes[name]
}

func (ro router) Match(test string) *Route {
	for _, route := range ro.Routes {
		u1 := NewRouteUri(route.Path)
		u2 := NewTestUri(test)

		um := UriMatcher{u1, u2}

		if um.match() {
			var values []interface{}

			for _, v := range um.test.params {
				values = append(values, v.value)
			}

			ro.SetParamValues(route, values)

			return route
		}
	}

	return nil
}

func (ro router) sortRoutes() {

}

func (ro router) SetParamNames(r *Route, pn []string) {
	r.ParamNames = pn
}

func (ro router) SetParamValues(r *Route, pv []interface{}) {
	r.ParamValues = pv
}

func (ro router) SetActions(r *Route, a map[string]Action) {
	r.Actions = a
}
